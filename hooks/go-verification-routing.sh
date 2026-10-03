#!/usr/bin/env bash
# PreToolUse Bash: complete the gate-owner load and receipt attempt before
# repeating runtime checks for the last Go edit. Never executes the command.
set -u
command -v python3 >/dev/null 2>&1 || exit 0
export GOLANG_SKILLS_VERIFICATION_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
python3 -c '
import hashlib,json,os,re,shlex,subprocess,sys
from pathlib import Path
try:
    event=json.load(sys.stdin)
except (ValueError,TypeError):
    sys.exit(0)
stopping=event.get("hook_event_name")=="Stop"
if (not stopping and (event.get("hook_event_name")!="PreToolUse" or event.get("tool_name")!="Bash")) or os.environ.get("GOLANG_SKILLS_VERIFICATION_GATE")=="off":
    sys.exit(0)
session=event.get("session_id") or "default"
command=(event.get("tool_input") or {}).get("command", "")
if not isinstance(session,str) or not re.fullmatch(r"[A-Za-z0-9_-]+",session) or not isinstance(command,str):
    sys.exit(0)
root=Path(os.environ["GOLANG_SKILLS_VERIFICATION_ROOT"])
base=Path(os.environ.get("CLAUDE_PLUGIN_DATA") or os.environ.get("TMPDIR","/tmp")+"/golang-skills-hooks")
loaded_file=base/"routing"/session/"loaded"
loaded=set(loaded_file.read_text().splitlines()) if loaded_file.is_file() else set()
if not loaded.intersection({"go-code","go-code-refactor"}):
    sys.exit(0)
if not (root/"skills/go-linting/SKILL.md").is_file():
    sys.exit(0)
state=base/"verification"/hashlib.sha256(session.encode()).hexdigest()
if not (state/"latest").is_file():
    sys.exit(0)
if stopping:
    text=event.get("last_assistant_message") or ""
    claims={m.group(1).lower() for m in re.finditer(r"\b(gofmt|vet|fix|lint|test)\s*(?:[:=]\s*)?pass\s*\([^)]*(?:\bhook\b|\breceipt\b)",text,re.I)}
    if not claims: sys.exit(0)
    try:
        latest=json.loads((state/"latest").read_text())
        attempted=json.loads((state/"attempted").read_text()) if (state/"attempted").is_file() else {}
        credits=set(attempted.get("credits",[])) if attempted.get("run_dir")==latest["run_dir"] else set()
        if claims & credits:
            fresh=subprocess.run([sys.executable,str(root/"hooks/go-check-receipt.py"),"current-credits",latest["run_dir"],*sorted(claims & credits)],capture_output=True,text=True,timeout=8)
            credits=set(json.loads(fresh.stdout)) if fresh.returncode==0 else set()
    except (OSError,ValueError,subprocess.TimeoutExpired): credits=set()
    unsupported=claims-credits
    if unsupported:
        reason="golang-skills verification gate: unsupported hook-pass report for "+", ".join(sorted(unsupported))+". Current verified credits: "+(", ".join(sorted(credits)) or "none")+". Correct the report to observed/unavailable or verify the current receipt. A direct rerun is pass (direct), not pass (hook)."
        if event.get("stop_hook_active"):
            print(json.dumps({"continue":False,"stopReason":reason}))
        else: print(json.dumps(dict(decision="block",reason=reason)))
    sys.exit(0)

def without_output_redirects(text):
    # Preserve quoted words, including bash -c bodies; forms parses those
    # recursively. Only unquoted output redirections are removed.
    word = r"(?:[^\s;&|()<>\"\x27\\]|\\.|\"(?:\\.|[^\"\\])*\"|\x27[^\x27]*\x27)+"
    pattern = re.compile(
        r"(?P<quoted>\"(?:\\.|[^\"\\])*\"|\x27[^\x27]*\x27|\\.)"
        r"|(?P<redirect>(?:(?<![\w])(?:[0-9]+|&))?>>?(?:&[0-9-]+|[ \t]*"+word+r"))")
    return pattern.sub(lambda m: m.group("quoted") or " ", text)

def forms(text,depth=0):
    if depth>3: return []
    try:
        lexer=shlex.shlex(without_output_redirects(text),posix=True,punctuation_chars=";&|()\n")
        lexer.whitespace=" \t\r"; lexer.whitespace_split=True
        tokens=list(lexer)
    except ValueError:
        return []
    result=[]; segment=[]
    for token in tokens+[";"]:
        if token and all(x in ";&|()\n" for x in token):
            if segment:
                args=segment
                while args and re.match(r"^[A-Za-z_]\w*=",args[0]):
                    args=args[1:]
                if args and Path(args[0]).name=="env":
                    args=args[1:]
                    while args and (args[0].startswith("-") or re.match(r"^[A-Za-z_]\w*=",args[0])):
                        count=2 if args[0] in ("-C","--chdir","-u","--unset") else 1
                        args=args[count:]
                if args and args[0]=="command": args=args[1:]
                if args and Path(args[0]).name in ("timeout","gtimeout"):
                    args=args[1:]
                    while args and args[0].startswith("-"):
                        args=args[2:] if args[0] in ("-k","--kill-after","-s","--signal") else args[1:]
                    args=args[1:]
                if args:
                    name=Path(args[0]).name
                    if name=="bash" and len(args)>2 and args[1]=="-c": result+=forms(args[2],depth+1)
                    elif name=="bash" and len(args)>1 and Path(args[1]).name=="go-check-receipt.sh":
                        if Path(args[1]).resolve()==(root/"hooks/go-check-receipt.sh").resolve(): result.append(("verifier",args[2:]))
                    elif name=="go":
                        sub=args[1:]
                        if sub and sub[0]=="-C": sub=sub[2:]
                        elif sub and sub[0].startswith("-C="): sub=sub[1:]
                        if sub and sub[0] in ("build","vet","test","fix"): result.append(("check",[]))
                    elif name in ("gofmt","govulncheck") or (name=="golangci-lint" and len(args)>1 and args[1]=="run"):
                        result.append(("check",[]))
                    elif name=="bash" and len(args)>1 and Path(args[1]).name in ("verify-refactor.sh","pre-review.sh"):
                        result.append(("check",[]))
            segment=[]
        else: segment.append(token)
    return result

kinds=forms(command)
if not kinds: sys.exit(0)
try:
    latest=json.loads((state/"latest").read_text())
    attempted=json.loads((state/"attempted").read_text()) if (state/"attempted").is_file() else {}
except (OSError,ValueError):
    print("golang-skills verification gate: receipt state unavailable; run required checks directly and do not credit hooks.",file=sys.stderr)
    sys.exit(0)
namespace=""
manifest=root/".claude-plugin/plugin.json"
if manifest.is_file(): namespace=json.loads(manifest.read_text()).get("name") or ""
skill=(namespace+":" if namespace else "")+"go-linting"
recipe=" ".join(shlex.quote(x) for x in ["bash",str(root/"hooks/go-check-receipt.sh"),"--gate",latest["run_dir"],latest["package"]])
if "go-linting" not in loaded:
    print("golang-skills verification gate: before verifying Go work, load the `"+skill+"` skill (Skill tool, name `"+skill+"`), then run:\n"+recipe,file=sys.stderr)
    sys.exit(2)
valid_prefix=False
for kind,args in kinds:
    if kind=="check": break
    if kind=="verifier":
        if len(args)==3 and args[0]=="--gate":
            valid_prefix=(str(Path(args[1]).resolve())==latest["run_dir"] and str(Path(args[2]).resolve())==latest["package"])
        elif len(args)>=3:
            valid_prefix=Path(args[0]).is_file() and str(Path(args[0]).resolve().parent)==latest["run_dir"] and str(Path(args[1]).resolve())==latest["package"]
        if valid_prefix: break
if (valid_prefix and not any(kind=="check" for kind,_ in kinds)) or attempted.get("run_dir")==latest["run_dir"]:
    sys.exit(0)
print("golang-skills verification gate: before repeating checks for the last Go edit, run this verifier once (Bash tool):\n"+recipe+"\nA redirect or a pipe that only displays output (2>&1, | tail, | head) is still that command. Another verification command in the same call is not.\nOnly hook_credit=true checks may be reported as pass (hook). A failed verification allows direct checks; the next Go edit requires another attempt.",file=sys.stderr)
sys.exit(2)
'
