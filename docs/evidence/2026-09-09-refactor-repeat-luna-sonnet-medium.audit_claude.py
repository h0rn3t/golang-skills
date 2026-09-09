from pathlib import Path
import hashlib, json

EXPECTED = "1ee8827c4ca7e69705cc164b134cb4adf05443b290a1cb76974645031be61414"

def audit(run):
    work = Path(run["workdir"]).resolve()
    trace = Path(run["trace_path"])
    valid_paths = {str(p.resolve()) for p in work.glob(".eval-plugin-*/skills/go-code-refactor/SKILL.md")
                   if hashlib.sha256(p.read_bytes()).hexdigest() == EXPECTED}
    calls, reads, model_ids = {}, [], set()
    init_plugin_ok = False
    for line_number, line in enumerate(trace.read_text().splitlines(), 1):
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("type") == "system" and event.get("subtype") == "init":
            model_ids.add(event.get("model", ""))
            for plugin in event.get("plugins", []):
                path = Path(plugin.get("path", ".")) / "skills/go-code-refactor/SKILL.md"
                if plugin.get("name") == "golang-skills" and str(path.resolve()) in valid_paths:
                    init_plugin_ok = True
        message = event.get("message", {})
        content = message.get("content", []) if isinstance(message, dict) else []
        if not isinstance(content, list):
            continue
        for block in content:
            if not isinstance(block, dict):
                continue
            if block.get("type") == "tool_use":
                calls[block["id"]] = block
            if block.get("type") != "tool_result" or block.get("is_error"):
                continue
            call = calls.get(block.get("tool_use_id"), {})
            args = call.get("input", {})
            text = str(block.get("content", ""))
            if call.get("name") == "Skill" and args.get("skill") == "golang-skills:go-code-refactor" and init_plugin_ok and "Launching skill" in text:
                reads.append({"line": line_number, "method": "Skill", "input": args})
            if call.get("name") == "Read" and "name: go-code-refactor" in text:
                path = Path(args.get("file_path", "."))
                if not path.is_absolute():
                    path = work / path
                if str(path.resolve()) in valid_paths:
                    reads.append({"line": line_number, "method": "Read", "input": args})
    return {"assigned_skill_loaded": bool(reads), "reads": reads,
            "models": sorted(model_ids), "expected_skill_sha256": EXPECTED,
            "trace_sha256": hashlib.sha256(trace.read_bytes()).hexdigest()}
