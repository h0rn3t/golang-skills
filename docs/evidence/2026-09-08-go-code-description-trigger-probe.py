#!/usr/bin/env python3
"""Trigger probe for the go-code description edit, opencode runner.

Two arms differing in exactly one line: `before` = the go-code description at
HEAD, `after` = the broadened one in the working tree. Everything else in both
skill trees is the current working tree, so the contract-intake section is
present in both arms and cannot be the reason a skill fires.

A skill counts as fired the same way abrun counts it for opencode: a call to the
`skill` tool carrying its name. Scoring matches evals/evals.json semantics --
a positive case passes when every expected skill fired, a negative control
passes only when nothing fired at all.
"""

import json
import os
import subprocess
import tempfile
import threading
import time
from concurrent.futures import ThreadPoolExecutor

S = "/private/tmp/claude-501/-Users-eugeneshershen-golang-skills/99e0f2c3-5a17-408a-b5c5-65e0b4bbc384/scratchpad"
MODEL = "opencode-go/deepseek-v4-flash"
TIMEOUT = 300
PARALLEL = int(os.environ.get("J", "5"))

out_lock = threading.Lock()
out_file = open(f"{S}/trigprobe-results.jsonl", "a", buffering=1)


def parse(transcript):
    """Return (skills fired, completed) from an `opencode run --format json` stream."""
    fired, saw_text = set(), False
    for line in transcript.splitlines():
        line = line.strip()
        if not line.startswith("{"):
            continue
        try:
            ev = json.loads(line)
        except json.JSONDecodeError:
            continue
        part = ev.get("part") or {}
        if ev.get("type") == "tool_use" and part.get("tool") == "skill":
            name = ((part.get("state") or {}).get("input") or {}).get("name")
            if isinstance(name, str) and name.startswith("go-"):
                fired.add(name)
        if ev.get("type") == "text" and part.get("text"):
            saw_text = True
    return sorted(fired), saw_text


def run_case(arm, ev):
    home = f"{S}/trig-home-{arm}"
    work = tempfile.mkdtemp(prefix="trigprobe-", dir=f"{S}/trigwork")
    env = dict(os.environ)
    env.update(
        HOME=home,
        PWD=work,
        OLDPWD=work,
        XDG_CONFIG_HOME=f"{home}/.config",
        XDG_DATA_HOME=f"{home}/.local/share",
        XDG_STATE_HOME=f"{home}/.local/state",
        XDG_CACHE_HOME=f"{home}/.cache",
    )
    cmd = ["opencode", "run", "--format", "json", "--auto", "--model", MODEL, ev["query"]]
    started = time.time()
    err = ""
    try:
        p = subprocess.run(cmd, cwd=work, env=env, capture_output=True, text=True, timeout=TIMEOUT)
        transcript, stderr, code = p.stdout, p.stderr.strip(), p.returncode
    except subprocess.TimeoutExpired as e:
        transcript = (e.stdout or b"").decode() if isinstance(e.stdout, bytes) else (e.stdout or "")
        stderr, code = f"timed out after {TIMEOUT}s", -1
    got, completed = parse(transcript)
    if code != 0:
        err = f"exit {code}: {stderr[:200]}"
    elif not completed:
        err = "session produced no message: " + stderr[:200]
    want = ev["should_trigger"]
    ok = not err and (set(want) <= set(got) if want else not got)
    rec = {
        "arm": arm, "query": ev["query"], "set": ev["set"], "want": want,
        "got": got, "pass": ok, "error": err, "seconds": round(time.time() - started, 1),
    }
    with out_lock:
        out_file.write(json.dumps(rec) + "\n")
        print(f"[{'PASS' if ok else 'FAIL'}] {arm:6s} want={want} got={got} "
              f"{ev['query'][:55]!r}{'  ' + err if err else ''}", flush=True)
    subprocess.run(["rm", "-rf", work], check=False)
    return rec


def main():
    cases = json.load(open(f"{S}/trigger-subset.json"))
    os.makedirs(f"{S}/trigwork", exist_ok=True)
    jobs = []
    for ev in cases:
        jobs.append(("before", ev))
        jobs.append(("after", ev))
    print(f"{len(cases)} cases x 2 arms = {len(jobs)} sessions, j={PARALLEL}\n", flush=True)
    with ThreadPoolExecutor(max_workers=PARALLEL) as pool:
        recs = list(pool.map(lambda j: run_case(*j), jobs))
    print("\ndone", flush=True)
    for arm in ("before", "after"):
        rs = [r for r in recs if r["arm"] == arm]
        pos = [r for r in rs if r["want"]]
        neg = [r for r in rs if not r["want"]]
        gc = sum(1 for r in rs if "go-code" in r["got"])
        print(f"{arm:6s} positives {sum(1 for r in pos if r['pass'])}/{len(pos)}  "
              f"negatives {sum(1 for r in neg if r['pass'])}/{len(neg)}  "
              f"go-code fired {gc}/{len(rs)}  errors {sum(1 for r in rs if r['error'])}")


if __name__ == "__main__":
    main()
