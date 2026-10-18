#!/usr/bin/env python3
"""Collect dependency license/notice files from `go list -m -json all`."""
import json
import pathlib
import re
import sys

out = pathlib.Path(sys.argv[1])
out.mkdir(parents=True, exist_ok=True)
data = sys.stdin.read()
decoder = json.JSONDecoder()
while data.strip():
    module, offset = decoder.raw_decode(data.lstrip())
    data = data.lstrip()[offset:]
    if module.get("Main") or not module.get("Dir"):
        continue
    directory = pathlib.Path(module["Dir"])
    target = out / re.sub(r"[^a-zA-Z0-9._-]", "_", module["Path"])
    files = [p for p in directory.iterdir() if p.is_file() and
             p.name.upper().startswith(("LICENSE", "COPYING", "NOTICE"))]
    if not files:
        files = [p for p in directory.iterdir() if p.is_file() and
                 p.name.upper().startswith("README") and
                 "license" in p.read_text(errors="replace").lower()]
    if not files:
        raise SystemExit(f"No license found for {module['Path']}; review before releasing")
    target.mkdir(exist_ok=True)
    (target / "SOURCE.txt").write_text(f"{module['Path']} {module.get('Version', '')}\n")
    for source in files:
        (target / source.name).write_bytes(source.read_bytes())
