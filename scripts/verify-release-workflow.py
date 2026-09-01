#!/usr/bin/env python
from __future__ import annotations

import re
import sys
from pathlib import Path

SHA = re.compile(r"^[0-9a-f]{40}$")
USES = re.compile(r"^\s*-?\s*uses:\s*([^\s#]+)", re.MULTILINE)
FORBIDDEN = (
    "pull_request_target",
    "secrets.",
    "secrets[",
    "--password",
    ".env",
)


def verify(path: Path) -> list[str]:
    if not path.is_file():
        return [f"missing release workflow: {path}"]
    text = path.read_text(encoding="utf-8")
    errors: list[str] = []
    if "workflow_dispatch:" not in text:
        errors.append("release workflow must be manually dispatchable")
    if not re.search(r"(?m)^permissions:\s*$", text):
        errors.append("release workflow must declare permissions")
    if not re.search(r"(?m)^\s{2}contents:\s*read\s*$", text):
        errors.append("release workflow must keep contents read-only")
    if not re.search(r"(?m)^\s{2}packages:\s*write\s*$", text):
        errors.append("release workflow needs packages write permission")
    lowered = text.lower()
    for forbidden in FORBIDDEN:
        if forbidden in lowered:
            errors.append(f"forbidden workflow token: {forbidden}")
    references = USES.findall(text)
    if not references:
        errors.append("release workflow must use reviewed actions")
    for reference in references:
        if reference.startswith("./"):
            continue
        if "@" not in reference or not SHA.fullmatch(reference.rsplit("@", 1)[1]):
            errors.append(f"action is not pinned to a full SHA: {reference}")
    required = (
        "ghcr.io/",
        "juntly-api",
        "juntly-frontend",
        "provenance: false",
        "sbom: false",
    )
    for value in required:
        if value not in text:
            errors.append(f"missing immutable-image control: {value}")
    return errors


def main() -> int:
    target = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(".github/workflows/release-images.yml")
    errors = verify(target)
    if errors:
        print("\n".join(errors), file=sys.stderr)
        return 1
    print("release workflow policy passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
