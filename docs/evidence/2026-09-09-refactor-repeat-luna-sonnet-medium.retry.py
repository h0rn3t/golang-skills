from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
import json, subprocess, time
base = Path(__file__).parent
slots = json.loads((base / "retry-slots.json").read_text())

def run(slot):
    label = "sonnet-retry-" + slot["task"] + "-" + str(slot["rep"])
    args = ["go", "run", "./cmd/abrun", "-runner", "claude", "-model", "claude-sonnet-5",
            "-effort", "medium", "-tasks", slot["task"], "-arms", "baseline",
            "-n", "1", "-j", "1", "-seed", "1", "-timeout", "10m", "-keep",
            "-prompt", (base / "retry-prompt.txt").read_text().strip(), "-out", str(base / (label + ".json"))]
    start = time.time()
    result = subprocess.run(args, cwd=base / "repo/evals", capture_output=True, text=True)
    (base / (label + ".log")).write_text(result.stdout + result.stderr)
    (base / (label + ".execution.json")).write_text(json.dumps({"slot": slot, "args": args,
        "wall_seconds": time.time() - start, "exit_code": result.returncode}, indent=2) + "\n")
    print(label + "\n" + result.stdout + result.stderr, flush=True)
    return result.returncode

with ThreadPoolExecutor(max_workers=2) as pool:
    codes = list(pool.map(run, slots))
raise SystemExit(int(any(codes)))
