#!/usr/bin/env python3
"""Trigger eval for the codex runner.

evals/cmd/evalrun drives the claude CLI only. Codex has no skill tool: skills
reach the model as a catalog of descriptions in the developer prompt, and a
skill "fires" when the model reads its SKILL.md through the shell. That makes
codex the sharper instrument for a description-compression question, because
the description is the only thing the model sees before it decides.

Two arms, same fixtures, same model: `before` = descriptions at HEAD,
`after` = the shortened descriptions in the working tree. Jobs are interleaved
arm-by-arm so a mid-run rate limit degrades both arms at the same point instead
of one of them, and a session that errored is recorded as an error rather than
as an empty skill set -- the failure mode that invalidated the earlier
claude-side comparison.
"""

import json
import os
import re
import subprocess
import sys
import tempfile
import threading
import time
from concurrent.futures import ThreadPoolExecutor

S = "/private/tmp/claude-501/-Users-eugeneshershen-golang-skills/390a5ebc-e53d-455a-b46e-7ec062f925f3/scratchpad"
REPO = "/Users/eugeneshershen/golang-skills"
MODEL = "gpt-5.6-luna"
EFFORT = "medium"
TIMEOUT = 240
PARALLEL = int(os.environ.get("J", "5"))

SKILL_PATH = re.compile(r"(?:^|[^A-Za-z0-9_-])(go-[a-z0-9-]+)/SKILL\.md")

out_lock = threading.Lock()
out_file = open(f"{S}/codextrig-results.jsonl", "a", buffering=1)


def parse(transcript):
    """Return (skills read, usage, completed) from a `codex exec --json` stream.

    Only the command of a shell call is searched, never its output: reading one
    SKILL.md echoes the whole file back, and skills cross-reference each other
    by name.
    """
    fired, usage, completed = set(), {}, False
    for line in transcript.splitlines():
        line = line.strip()
        if not line.startswith("{"):
            continue
        try:
            ev = json.loads(line)
        except json.JSONDecodeError:
            continue
        item = ev.get("item") or {}
        if item.get("type") == "command_execution":
            fired.update(SKILL_PATH.findall(item.get("command", "")))
        if ev.get("type") == "turn.completed":
            completed = True
            usage = ev.get("usage", {})
    return sorted(fired), usage, completed


def run_case(arm, ev):
    home = f"{S}/home-{arm}"
    work = tempfile.mkdtemp(prefix="codextrig-", dir=f"{S}/work")
    env = dict(os.environ)
    env.update(
        HOME=home,
        CODEX_HOME=f"{home}/.codex",
        PWD=work,
        OLDPWD=work,
        XDG_CONFIG_HOME=f"{home}/.config",
        XDG_DATA_HOME=f"{home}/.local/share",
        XDG_STATE_HOME=f"{home}/.local/state",
        XDG_CACHE_HOME=f"{home}/.cache",
    )
    cmd = [
        "codex", "exec", "--json", "--cd", work, "--skip-git-repo-check",
        "--ephemeral", "--color", "never", "-s", "read-only",
        "-c", 'approval_policy="never"', "-m", MODEL,
        "-c", f'model_reasoning_effort="{EFFORT}"', ev["query"],
    ]
    started = time.time()
    err = ""
    try:
        p = subprocess.run(cmd, cwd=work, env=env, capture_output=True,
                           text=True, timeout=TIMEOUT)
        transcript, stderr, code = p.stdout, p.stderr.strip(), p.returncode
    except subprocess.TimeoutExpired as e:
        transcript = (e.stdout or b"").decode() if isinstance(e.stdout, bytes) else (e.stdout or "")
        stderr, code = f"timed out after {TIMEOUT}s", -1
    got, usage, completed = parse(transcript)
    if code != 0:
        err = f"exit {code}: {stderr[:300]}"
    elif not completed:
        err = "session did not complete a turn: " + stderr[:300]
    want = ev["should_trigger"]
    ok = not err and (set(want) <= set(got) if want else not got)
    rec = {
        "arm": arm, "query": ev["query"], "set": ev["set"], "want": want,
        "got": got, "pass": ok, "error": err, "usage": usage,
        "seconds": round(time.time() - started, 1),
    }
    with out_lock:
        out_file.write(json.dumps(rec) + "\n")
        print(f"[{'PASS' if ok else 'FAIL'}] {arm:6s} want={want} got={got} "
              f"{ev['query'][:60]!r}{'  ' + err if err else ''}", flush=True)
    subprocess.run(["rm", "-rf", work], check=False)
    return rec


def main():
    evals = json.load(open(f"{REPO}/evals/evals.json"))["trigger_evals"]
    if len(sys.argv) > 1:
        evals = evals[: int(sys.argv[1])]
    os.makedirs(f"{S}/work", exist_ok=True)
    jobs = []
    for ev in evals:
        jobs.append(("before", ev))
        jobs.append(("after", ev))
    print(f"{len(evals)} cases x 2 arms = {len(jobs)} sessions, j={PARALLEL}\n", flush=True)
    with ThreadPoolExecutor(max_workers=PARALLEL) as pool:
        list(pool.map(lambda j: run_case(*j), jobs))
    print("\ndone", flush=True)


if __name__ == "__main__":
    main()
