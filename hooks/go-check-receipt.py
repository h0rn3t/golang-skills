#!/usr/bin/env python3
"""Internal edit-hook check runner; immutable receipts, never a gate policy."""

import hashlib
import json
import os
from pathlib import Path
import re
import signal
import stat
import shutil
import shlex
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone


def digest(data):
    return hashlib.sha256(data).hexdigest()


def capture(argv, cwd, logical=False):
    try:
        environment = dict(os.environ)
        if logical:
            environment["PWD"] = os.path.abspath(cwd)
        p = subprocess.Popen(argv, cwd=cwd, stdout=subprocess.PIPE,
                             stderr=subprocess.STDOUT, env=environment, start_new_session=True)
        try:
            output, _ = p.communicate(timeout=min(8, remaining()))
        except subprocess.TimeoutExpired:
            os.killpg(p.pid, signal.SIGKILL)
            p.communicate()
            return 124, "metadata probe deadline exceeded"
        return p.returncode, output.decode("utf-8", "replace").strip()
    except OSError as e:
        return 127, str(e)


def remaining():
    return max(0.001, float(os.environ.get("GOLANG_SKILLS_HOOK_DEADLINE", time.time() + 135)) - time.time())


def snapshot(file, check, config, filter_kind, receipt_dir):
    """Conservative: any unreadable/oversize input prevents reuse."""
    cwd = str(Path(file).resolve().parent)
    problems, entries = [], []
    tool = "gofmt" if check == "gofmt" else "golangci-lint" if check == "lint" else "go"
    executable = shutil.which(tool)
    version_args = [tool, "version"] if tool != "gofmt" else ["go", "version"]
    rc, version = capture(version_args, cwd)
    if rc or not executable:
        problems.append("tool version unavailable")
    rc, raw = capture(["go", "env", "-json"], cwd)
    env = {}
    try:
        if rc:
            raise ValueError("go env failed")
        env = json.loads(raw)
    except (ValueError, TypeError):
        problems.append("Go build environment unavailable")
    # Hash potentially credential-bearing settings; never print/store them.
    stable_env = dict(env)
    stable_env["GOGCCFLAGS"] = re.sub(r"-(ffile|fdebug)-prefix-map=.*?go-build\d+=/tmp/go-build",
                                    r"-\1-prefix-map=<go-build>=/tmp/go-build", env.get("GOGCCFLAGS", ""))
    entries.append(("go-env", digest(json.dumps(stable_env, sort_keys=True).encode())))
    build_env = {k: v for k, v in os.environ.items()
                 if k not in ("PWD", "OLDPWD", "SHLVL", "_", "XPC_SERVICE_NAME", "XPC_FLAGS",
                              "AI_AGENT", "CLAUDE_CODE_EXECPATH", "CLAUDE_PROJECT_DIR", "GIT_EDITOR")
                 and not k.startswith(("GOLANG_SKILLS_", "CLAUDE_PLUGIN_"))}
    entries.append(("build-env", digest(json.dumps(build_env, sort_keys=True).encode())))
    roots, pending = set(), []
    gomod = env.get("GOMOD")
    if gomod and gomod != os.devnull and Path(gomod).is_file():
        pending.append(str(Path(gomod).resolve().parent))
    elif check == "gofmt":
        roots.add(cwd)
    else:
        problems.append("no module")
    work = env.get("GOWORK")
    if work and work != "off":
        rc, data = capture(["go", "work", "edit", "-json", work], cwd)
        try:
            if rc:
                raise ValueError()
            obj = json.loads(data)
            pending.extend(str((Path(work).parent / x["DiskPath"]).resolve())
                           for x in obj.get("Use") or [])
            pending.extend(str((Path(work).parent / x["New"]["Path"]).resolve())
                           for x in obj.get("Replace") or [] if not x["New"].get("Version"))
        except (ValueError, KeyError, TypeError):
            problems.append("workspace inputs unavailable")
    # Resolve only local replacements; never download dependencies for a receipt.
    while pending:
        root = pending.pop()
        if root in roots:
            continue
        roots.add(root)
        rc, data = capture(["go", "mod", "edit", "-json", str(Path(root) / "go.mod")], cwd)
        try:
            if rc:
                raise ValueError()
            obj = json.loads(data)
            pending.extend(str((Path(root) / x["New"]["Path"]).resolve())
                           for x in obj.get("Replace") or [] if not x["New"].get("Version"))
        except (ValueError, KeyError, TypeError):
            problems.append("local module inputs unavailable")
    extras = [executable, env.get("GOENV"), config]
    if work and work != "off":
        extras.extend([work, work + ".sum"])
    # A compiler change with the same version is not the same toolchain.
    goroot = env.get("GOROOT", "")
    host = env.get("GOHOSTOS", "") + "_" + env.get("GOHOSTARCH", "")
    extras.extend([shutil.which("go"), str(Path(goroot) / "pkg" / "tool" / host / "compile")])
    size, count = 0, 0

    def add(path, optional=False):
        nonlocal size, count
        p = Path(path)
        try:
            if not p.exists() and optional:
                entries.append((str(p), "absent"))
                return
            fd = os.open(p, os.O_RDONLY | os.O_NONBLOCK)
            with os.fdopen(fd, "rb") as stream:
                info = os.fstat(stream.fileno())
                if not stat.S_ISREG(info.st_mode):
                    raise ValueError("nonregular input: " + str(p))
                size += info.st_size
                if size > 128 * 1024 * 1024:
                    raise ValueError("input budget exceeded")
                content = stream.read()
            count += 1
            if size > 128 * 1024 * 1024 or count > 20000:
                raise ValueError("input budget exceeded")
            # Metadata detects rewrites that restore the original bytes.
            link_info = p.lstat()
            entries.append((str(p), digest(content), info.st_mtime_ns, info.st_ctime_ns,
                            info.st_ino, link_info.st_ctime_ns, os.readlink(p) if p.is_symlink() else ""))
        except (OSError, ValueError) as e:
            problems.append(str(e))

    try:
        for root in sorted(roots):
            if not Path(root).is_dir():
                problems.append("input root missing: " + root)
            for cur, dirs, files in os.walk(root, onerror=lambda e: problems.append(str(e))):
                if remaining() < 1:
                    raise ValueError("hook input snapshot deadline exceeded")
                dirs[:] = sorted(x for x in dirs if x != ".git"
                                 and str(Path(cur, x).resolve()) != str(Path(receipt_dir).resolve()))
                for name in dirs:
                    if Path(cur, name).is_symlink():
                        problems.append("directory symlink inputs unsupported: " + str(Path(cur, name)))
                dirs[:] = [x for x in dirs if not Path(cur, x).is_symlink()]
                for name in sorted(files):
                    if name != ".git":
                        add(str(Path(cur, name)))
                if size > 128 * 1024 * 1024 or count > 20000:
                    raise ValueError("input budget exceeded")
        for path in extras:
            if path and path != "off":
                add(str(Path(path).resolve()), optional=path == env.get("GOENV") or path.endswith(".sum"))
    except ValueError as e:
        problems.append(str(e))
    if filter_kind == "new-since-HEAD":
        rc, head = capture(["git", "rev-parse", "HEAD"], cwd)
        if rc:
            problems.append("lint revision unavailable")
        entries.append(("HEAD", head))
    toolchain = {key: env.get(key, "") for key in
                 ("GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "GOTOOLCHAIN", "GOROOT")}
    return digest(json.dumps(sorted(entries), sort_keys=True).encode()), problems, version, toolchain


def status_for(check, code, output):
    if code in (124, 126, 127, 137) or code < 0:
        return "unavailable"
    if check == "lint" and code:
        return "fail" if code == 1 and re.search(r"^/.*:\d+:\d+: ", output, re.M) else "unavailable"
    if check == "fix":
        if re.search(r"^--- ", output, re.M):
            return "fail"
        return "pass" if code == 0 and not output.strip() else "unavailable"
    if check == "gofmt":
        return "fail" if code or output.strip() else "pass"
    return "pass" if code == 0 else "fail"


def now():
    return datetime.now(timezone.utc).isoformat()


def atomic_json(path, data):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix=".pending-", dir=path.parent)
    with os.fdopen(fd, "w") as stream:
        json.dump(data, stream)
    os.replace(temporary, path)


def begin(directory, file, session, state_base):
    state = Path(state_base).resolve() / "verification" / digest(session.encode())
    info = dict(run_dir=str(Path(directory).resolve()), package=str(Path(file).resolve().parent), state=str(state))
    atomic_json(Path(directory) / ".session", info)
    atomic_json(state / "latest", info)


def mark_attempt(directory, credits, replace=True):
    meta = Path(directory) / ".session"
    if not meta.is_file():
        return
    info = json.loads(meta.read_text())
    state = Path(info["state"])
    latest = json.loads((state / "latest").read_text())
    if latest["run_dir"] == str(Path(directory).resolve()):
        if not replace and (state / "attempted").is_file():
            previous = json.loads((state / "attempted").read_text())
            if previous.get("run_dir") == latest["run_dir"]:
                credits = sorted(set(credits) | set(previous.get("credits", [])))
        atomic_json(state / "attempted", dict(run_dir=latest["run_dir"], credits=credits))


def run(args):
    directory, check, file, config, filter_kind, planned_status, reason, *command = args
    before, errors_before, version, toolchain = snapshot(file, check, config, filter_kind, directory)
    started, code, output = now(), None, ""
    if planned_status == "run":
        try:
            if remaining() < 5:
                status, reason = "unavailable", "hook deadline exceeded before check"
            else:
                p = subprocess.Popen(command, cwd=str(Path(file).resolve().parent),
                                     stdout=subprocess.PIPE, stderr=subprocess.STDOUT, start_new_session=True)
                try:
                    raw, _ = p.communicate(timeout=max(0.001, remaining() - 4))
                    code = p.returncode
                except subprocess.TimeoutExpired:
                    os.killpg(p.pid, signal.SIGKILL)
                    raw, _ = p.communicate()
                    code = 124
                    raw += b"\nedit-hook global deadline exceeded\n"
                output = raw.decode("utf-8", "replace")
                status = status_for(check, code, output)
        except OSError as e:
            code, output = 127, str(e) + "\n"
            status = "unavailable"
    else:
        status = planned_status
    ended = now()
    after, errors_after, _, _ = snapshot(file, check, config, filter_kind, directory)
    record = dict(version=1, check=check, status=status, exit_code=code, command=command,
                  cwd=str(Path(file).resolve().parent), file=str(Path(file).absolute()),
                  scope=dict(kind="file" if check == "gofmt" else "package",
                             target=str(Path(file).resolve()) if check == "gofmt" else str(Path(file).resolve().parent)),
                  config=config, filter=filter_kind, tool_version=version,
                  toolchain=toolchain,
                  started=started, ended=ended, inputs_before=before, inputs_after=after,
                  reusable=status == "pass" and before == after and not errors_before and not errors_after,
                  reason=reason, snapshot_errors=errors_before + errors_after, diagnostic=output)
    fd, temporary = tempfile.mkstemp(prefix=".pending-", dir=directory)
    with os.fdopen(fd, "w") as stream:
        json.dump(record, stream, ensure_ascii=True)
    os.replace(temporary, str(Path(directory) / (check + ".json")))
    sys.stdout.write(output)
    return code if code is not None and code >= 0 else 0 if code is None else 128 - code


def verify(path, expected=None):
    record = json.loads(Path(path).read_text())
    current, errors, _, _ = snapshot(record["file"], record["check"], record["config"],
                                  record["filter"], str(Path(path).parent))
    valid = (record.get("version") == 1 and record.get("status") == "pass"
             and record.get("reusable") is True and not errors
             and record["inputs_before"] == record["inputs_after"] == current)
    matching = (expected is not None and len(expected) >= 2
                and str(Path(expected[0]).resolve()) == record["cwd"]
                and expected[1:] == record["command"])
    credit = valid and matching
    if expected is not None:
        mark_attempt(Path(path).parent, [record["check"]] if credit else [], replace=not valid)
    reason = errors or ("" if valid else "stale or non-passing receipt")
    if valid and expected is not None and not matching:
        reason = "cwd or command differs from the recorded check"
    print(json.dumps(dict(valid=valid, check=record["check"], command=record["command"],
                          cwd=record["cwd"], scope=record["scope"], config=record["config"],
                          filter=record["filter"], hook_credit=credit, reason=reason)))
    return 0 if (credit if expected is not None else valid) else 1


def verify_gate(directory, target):
    target = str(Path(target).resolve())
    checks, cache = [], {}
    for check in ("gofmt", "vet", "fix", "test", "lint"):
        path = Path(directory) / (check + ".json")
        try:
            r = json.loads(path.read_text())
            kind = "vet" if check in ("vet", "fix", "test") else check
            key = (r["file"], kind, r["config"], r["filter"])
            if key not in cache:
                cache[key] = snapshot(r["file"], kind, r["config"], r["filter"], directory)
            current, errors, _, _ = cache[key]
            valid = (r.get("version") == 1 and r["status"] == "pass" and r["reusable"] is True
                     and not errors and r["inputs_before"] == r["inputs_after"] == current)
            # The batch selects package checks. File gofmt and short tests
            # still need the selected broader formatting/race gate.
            eligible = r["cwd"] == target and r["scope"] == dict(kind="package", target=target)
            command = r["command"]
            if check == "vet": eligible = eligible and command == ["go", "vet", "."]
            elif check == "fix": eligible = eligible and command == ["go", "fix", "-diff", "."]
            elif check == "lint":
                expected_lint = ["timeout", "60", "golangci-lint", "run", "--allow-parallel-runners", "--path-mode=abs",
                                 "--output.text.print-issued-lines=false", "--show-stats=false", ".", "--config", r["config"]]
                eligible = eligible and r["filter"] == "all" and bool(r["config"]) and command == expected_lint
            else: eligible = False
            checks.append(dict(check=check, status=r["status"], valid=valid, hook_credit=valid and eligible,
                               scope=r["scope"], command=command, config=r["config"], filter=r["filter"],
                               reason=errors or ("" if valid and eligible else "run the selected check directly: stale, unavailable or broader scope/flags")))
        except (OSError, ValueError, KeyError, IndexError) as error:
            checks.append(dict(check=check, status="unavailable", valid=False, hook_credit=False, reason=str(error)))
    credited = [row["check"] for row in checks if row["hook_credit"]]
    mark_attempt(directory, credited)
    print(json.dumps(dict(scope=dict(kind="package", target=target), checks=checks,
                          required_direct=[row["check"] for row in checks if not row["hook_credit"]],
                          note="No build, race or applicable govulncheck evidence in these receipts; complete the selected gate.")))
    # A failed attempt releases direct checks but never grants credit.
    return 0 if credited else 1


def current_credits(directory, names):
    verified, cache = [], {}
    for name in names:
        if name not in ("gofmt", "vet", "fix", "test", "lint"):
            continue
        try:
            r = json.loads((Path(directory) / (name + ".json")).read_text())
            kind = "vet" if name in ("vet", "fix", "test") else name
            key = (r["file"], kind, r["config"], r["filter"])
            if key not in cache:
                cache[key] = snapshot(r["file"], kind, r["config"], r["filter"], directory)
            current, errors, _, _ = cache[key]
            if r["status"] == "pass" and r["reusable"] and not errors and r["inputs_before"] == r["inputs_after"] == current:
                verified.append(name)
        except (OSError, ValueError, KeyError):
            continue
    print(json.dumps(verified))


def report(directory, context_only=False):
    rows = []
    commands = []
    skill = "go-linting"
    manifest = Path(__file__).resolve().parent.parent / ".claude-plugin" / "plugin.json"
    if manifest.is_file():
        namespace = json.loads(manifest.read_text()).get("name")
        if isinstance(namespace, str) and re.fullmatch(r"[A-Za-z0-9_.-]+", namespace):
            skill = namespace + ":" + skill
    for check in ("gofmt", "vet", "fix", "test", "lint"):
        path = Path(directory) / (check + ".json")
        r = json.loads(path.read_text())
        rows.append(f'{check}: {r["status"]}; exit={r["exit_code"]}; scope={r["scope"]["kind"]}; '
                    f'filter={r["filter"]}; state-stable={r["reusable"]}; reuse=unverified; reason={r["reason"]}')
        if r["reusable"]:
            argv = ["bash", str(Path(__file__).resolve().with_suffix(".sh")),
                    str(path.resolve()), r["cwd"], *r["command"]]
            commands.append(" ".join(shlex.quote(arg) for arg in argv))
    context = f"With a Bash tool, before the final report load the `{skill}` skill (Skill tool, name `{skill}`), then verify matching receipts below or run the selected checks directly.\n"
    context += "Edit-hook receipts: " + directory + "\nReuse is UNVERIFIED; recorded pass is not pass (hook).\n" + "\n".join(rows)
    package = json.loads((Path(directory) / "vet.json").read_text())["cwd"]
    batch = ["bash", str(Path(__file__).resolve().with_suffix(".sh")), "--gate", str(Path(directory).resolve()), package]
    context += "\nFirst verification action after the final edit (Bash tool):\n" + " ".join(shlex.quote(arg) for arg in batch)
    context += "\nUse only hook_credit=true checks from that JSON. Never write lint pass for reuse=unverified, even with a disclaimer. Run required_direct checks and the selected build/race gate."
    if commands:
        context += "\nFor each check matching your selected gate, run its Bash command below. Only hook_credit=true permits pass (hook); exit 0 from a state-only verifier does not.\n" + "\n".join(commands)
    context += "\nFile-only gofmt and package checks cannot satisfy a wider gate. Package short tests have no race check. If verification is unavailable, run the check directly or report it unavailable."
    if context_only:
        print(context)
    else:
        print(json.dumps(dict(hookSpecificOutput=dict(hookEventName="PostToolUse", additionalContext=context))))


if __name__ == "__main__":
    try:
        if sys.argv[1] == "run":
            sys.exit(run(sys.argv[2:]))
        if sys.argv[1] == "begin":
            begin(*sys.argv[2:])
            sys.exit(0)
        if sys.argv[1] == "verify-gate":
            sys.exit(verify_gate(*sys.argv[2:]))
        if sys.argv[1] == "current-credits":
            current_credits(sys.argv[2], sys.argv[3:])
            sys.exit(0)
        if sys.argv[1] == "verify":
            sys.exit(verify(sys.argv[2], sys.argv[3:] or None))
        if sys.argv[1] in ("report", "context"):
            report(sys.argv[2], context_only=sys.argv[1] == "context")
        elif sys.argv[1] == "module":
            code, module = capture(["go", "env", "GOMOD"], sys.argv[2])
            sys.exit(0 if code == 0 and module != os.devnull and Path(module).is_file() else 1)
        elif sys.argv[1] == "config":
            code, output = capture(["golangci-lint", "config", "path"], sys.argv[2], logical=True)
            print(output)
            sys.exit(code)
        else:
            raise ValueError("unknown receipt operation")
    except (OSError, ValueError, KeyError, IndexError) as error:
        print("edit-hook receipt unavailable: " + str(error), file=sys.stderr)
        sys.exit(2)
