#!/usr/bin/env python3
"""Audit one abrun report: did each session measure both LOC counts before its
first production edit, and did it actually read its own arm's go-code-refactor?

Assigned reads are derived from the trace, because the arm homes are gone by the
time this runs: a command naming a path under this session's own
abrun-codex-*/.codex/skills that exited 0 is a read that landed.
"""
import json, re, sys

OWN = re.compile(r"/abrun-codex-\d+/\.codex/skills/(go-[a-z0-9-]+)/SKILL\.md")

report = json.load(open(sys.argv[1]))
rows = []
for r in report["results"]:
    trace = r.get("trace_path")
    seq, assigned = [], set()
    if trace:
        for line in open(trace, errors="replace"):
            line = line.strip()
            if not line.startswith("{"):
                continue
            try:
                ev = json.loads(line)
            except Exception:
                continue
            if ev.get("type") != "item.completed":
                continue
            it = ev.get("item", {})
            t = it.get("type")
            if t == "command_execution":
                cmd = it.get("command", "")
                seq.append(("cmd", cmd))
                if str(it.get("exit_code")) == "0":
                    assigned.update(OWN.findall(cmd))
            elif t in ("patch_apply", "file_change"):
                seq.append(("edit", ""))
    first_edit = next((i for i, (t, _) in enumerate(seq) if t == "edit"), None)

    def step(needle):
        return next((i for i, (t, c) in enumerate(seq) if t == "cmd" and needle in c), None)

    lb, ld = step("loc-baseline"), step("loc-diff")
    before = lb is not None and (first_edit is None or lb < first_edit)
    rows.append({
        "arm": r["arm"], "task": r["task"],
        "loc_baseline_step": lb, "first_edit_step": first_edit, "loc_diff_step": ld,
        "measured_before_edit": before,
        "read_own_refactor": "go-code-refactor" in assigned,
        "assigned_reads": sorted(assigned),
        "dphys": r["delta"]["lines"], "dcode": r["delta"]["code"], "dfuncs": r["delta"]["funcs"],
        "dtypes": r["delta"]["types"],
        "build": r["build"], "golden": r["golden"], "edited": r["edited"],
        "line_gate": r["line_gate_pass"], "code_gate": r["code_gate_pass"],
        "empty_diff": r.get("empty_diff"),
        "foreign": [f["path"] for f in r.get("foreign_skills", [])],
        "unresolved": r.get("unresolved_skill_reads", []),
        "skills_flag": r.get("skills", []),
        "commands": r.get("commands"),
        "reported_counts": r.get("reported_counts"),
        "err": r.get("error", ""),
        "trace": trace, "workdir": r.get("workdir"),
    })

rows.sort(key=lambda x: (x["arm"], x["task"]))
hdr = (f'{"arm":10s} {"task":5s} {"locB":>4s} {"edit":>4s} {"locD":>4s} {"before":>6s} {"ownread":>7s} '
       f'{"dphys":>6s} {"dcode":>6s} {"dfun":>5s} {"gate":>5s} {"cgate":>6s} {"SKL":>3s} {"cmds":>4s}')
print(hdr)
for x in rows:
    print(f'{x["arm"]:10s} {x["task"]:5s} {str(x["loc_baseline_step"]):>4s} {str(x["first_edit_step"]):>4s} '
          f'{str(x["loc_diff_step"]):>4s} {str(x["measured_before_edit"]):>6s} {str(x["read_own_refactor"]):>7s} '
          f'{x["dphys"]:>6d} {x["dcode"]:>6d} {x["dfuncs"]:>5d} {str(x["line_gate"]):>5s} '
          f'{str(x["code_gate"]):>6s} {str(len(x["foreign"])):>3s} {str(x["commands"]):>4s}')

for arm in sorted({x["arm"] for x in rows}):
    a = [x for x in rows if x["arm"] == arm]
    clean = [x for x in a if not x["foreign"]]
    n, c = len(a), len(clean)
    print(f'\n{arm}: n={n}, own-skill reads {sum(x["read_own_refactor"] for x in a)}/{n}, '
          f'foreign {sum(1 for x in a if x["foreign"])}/{n}')
    print(f'  measured both counts before first edit: {sum(x["measured_before_edit"] for x in a)}/{n} '
          f'(clean runs {sum(x["measured_before_edit"] for x in clean)}/{c})')
    print(f'  build {sum(x["build"] for x in a)}/{n}, golden {sum(x["golden"] for x in a)}/{n}, '
          f'edited {sum(x["edited"] for x in a)}/{n}, empty diff {sum(1 for x in a if x["empty_diff"])}/{n}')
    if c:
        print(f'  clean means: Δphys {sum(x["dphys"] for x in clean)/c:+.1f}, '
              f'Δcode {sum(x["dcode"] for x in clean)/c:+.1f}, Δfuncs {sum(x["dfuncs"] for x in clean)/c:+.2f}; '
              f'physical gate {sum(x["line_gate"] for x in clean)}/{c}, code gate {sum(x["code_gate"] for x in clean)}/{c}')
    for x in a:
        if x["unresolved"]:
            print(f'  {x["task"]}: {len(x["unresolved"])} unresolved skill path(s), first {x["unresolved"][0]}')

json.dump(rows, open(sys.argv[2], "w"), indent=1, ensure_ascii=False)
print(f'\nwrote {sys.argv[2]}')
