#!/usr/bin/env python3
import yaml, subprocess, pathlib, datetime, json

BASE = pathlib.Path(".")
RESULTS = BASE / "results"
LOGS = BASE / "log"
RESULTS.mkdir(exist_ok=True)
LOGS.mkdir(exist_ok=True)

with open("temp.yml", "r") as f:
    config = yaml.safe_load(f)

run_id = datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
summary = {
    "run_id": run_id,
    "started": datetime.datetime.now().isoformat(),
    "scans": []
}

for tech, data in config.get("tech", {}).items():
    input_file = pathlib.Path(data.get("input", f"subtech/{tech}.txt"))
    templates = data.get("templates", [])

    if not input_file.exists() or input_file.stat().st_size == 0:
        continue

    if not templates:
        continue

    out_file = RESULTS / f"{tech}-{run_id}.jsonl"
    log_file = LOGS / f"{tech}-{run_id}.log"

    cmd = [
        "nuclei",
        "-l", str(input_file),
        "-jsonl",
        "-severity", "low,medium,high,critical",
        "-rl", "50",
        "-c", "25",
        "-o", str(out_file)
    ]

    for template in templates:
        cmd += ["-t", template]

    with open(log_file, "w") as log:
        log.write("[CMD] " + " ".join(cmd) + "\n\n")
        p = subprocess.run(cmd, stdout=log, stderr=log)

    summary["scans"].append({
        "tech": tech,
        "input": str(input_file),
        "templates": templates,
        "result": str(out_file),
        "log": str(log_file),
        "exit_code": p.returncode
    })

summary["finished"] = datetime.datetime.now().isoformat()

with open(RESULTS / "latest-summary.json", "w") as f:
    json.dump(summary, f, indent=2)

print(json.dumps(summary, indent=2))