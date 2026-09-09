from pathlib import Path
import json, subprocess, sys, time
base = Path(__file__).parent
runner, model, label = sys.argv[1:]
args = ["go", "run", "./cmd/abrun", "-runner", runner, "-model", model,
        "-effort", "medium", "-tasks", "gw0,gw1,gw2,gw3,gw4", "-arms", "baseline",
        "-n", "2", "-j", "2", "-seed", "1", "-timeout", "10m", "-keep",
        "-prompt", (base / "prompt.txt").read_text().strip(), "-out", str(base / (label + ".json"))]
start = time.time()
with (base / (label + ".log")).open("w") as log:
    proc = subprocess.Popen(args, cwd=base / "repo/evals", stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, text=True)
    for line in proc.stdout:
        log.write(line)
        log.flush()
        print(line, end="", flush=True)
    code = proc.wait()
(base / (label + ".execution.json")).write_text(json.dumps({
    "args": args, "started_unix": start, "finished_unix": time.time(),
    "wall_seconds": time.time() - start, "exit_code": code}, indent=2) + "\n")
raise SystemExit(code)
