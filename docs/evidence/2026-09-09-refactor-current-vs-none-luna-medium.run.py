from pathlib import Path
import json, subprocess, time
base = Path(__file__).parent
meta = json.loads((base / "manifest.json").read_text())
start = time.time()
with (base / "run.log").open("w") as log:
    p = subprocess.Popen(meta["args"], cwd=Path(meta["repo"]) / "evals", stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    for line in p.stdout:
        log.write(line); log.flush(); print(line, end="", flush=True)
    code = p.wait()
(base / "execution.json").write_text(json.dumps({"started_unix": start, "finished_unix": time.time(), "wall_seconds": time.time()-start, "exit_code": code}, indent=2)+"\n")
raise SystemExit(code)
