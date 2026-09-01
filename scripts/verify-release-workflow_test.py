from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
GUARD = ROOT / "scripts" / "verify-release-workflow.py"
WORKFLOW = ROOT / ".github" / "workflows" / "release-images.yml"


class ReleaseWorkflowPolicyTest(unittest.TestCase):
    def run_guard(self, content: str) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory() as directory:
            candidate = Path(directory) / "release.yml"
            candidate.write_text(content, encoding="utf-8")
            return subprocess.run(
                [sys.executable, str(GUARD), str(candidate)],
                check=False,
                capture_output=True,
                text=True,
            )

    def test_reviewed_workflow_passes(self) -> None:
        result = self.run_guard(WORKFLOW.read_text(encoding="utf-8"))
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_rejects_privileged_pull_request_trigger(self) -> None:
        content = WORKFLOW.read_text(encoding="utf-8").replace(
            "workflow_dispatch:", "pull_request_target:", 1
        )
        self.assertNotEqual(self.run_guard(content).returncode, 0)

    def test_rejects_mutable_or_duplicate_action_reference(self) -> None:
        content = WORKFLOW.read_text(encoding="utf-8") + "\n      - uses: actions/checkout@v7\n"
        self.assertNotEqual(self.run_guard(content).returncode, 0)

    def test_rejects_repository_secret_expression(self) -> None:
        content = WORKFLOW.read_text(encoding="utf-8") + "\n# ${{ secrets.REGISTRY_TOKEN }}\n"
        self.assertNotEqual(self.run_guard(content).returncode, 0)


if __name__ == "__main__":
    unittest.main()
