# Keep local and hosted Hugo release checks compatible without weakening the version pin.
import os
import subprocess
import tempfile
import unittest
from pathlib import Path


class HugoVersionCheckTest(unittest.TestCase):
    def check_version(self, output, exit_code=0):
        site = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as directory:
            hugo = Path(directory) / "hugo"
            hugo.write_text(
                '#!/bin/sh\nprintf "%s\\n" "$TEST_HUGO_OUTPUT"\n'
                'exit "$TEST_HUGO_EXIT_CODE"\n',
                encoding="utf-8",
            )
            hugo.chmod(0o755)
            return subprocess.run(
                ["make", "--no-print-directory", "check-hugo"],
                cwd=site,
                env={
                    **os.environ,
                    "PATH": directory + os.pathsep + os.environ["PATH"],
                    "TEST_HUGO_OUTPUT": output,
                    "TEST_HUGO_EXIT_CODE": str(exit_code),
                },
                capture_output=True,
                text=True,
                check=False,
            )

    def test_accepts_pinned_release_with_build_metadata(self):
        versions = [
            "v0.166.0",
            "v0.166.0+extended+withdeploy",
            "v0.166.0-78400b4de8adc99273d3e674ed6c4d14c3bdf669+extended",
            "v0.166.0-78400b4de8adc99273d3e674ed6c4d14c3bdf669",
        ]
        for version in versions:
            with self.subTest(version=version):
                result = self.check_version(f"hugo {version} linux/amd64")
                self.assertEqual(result.returncode, 0, result.stderr)

    def test_rejects_other_releases_and_prereleases(self):
        versions = [
            "v0.165.0+extended",
            "v0.166.1-78400b4de8adc99273d3e674ed6c4d14c3bdf669+extended",
            "v0.167.0",
            "v0.166.00",
            "v0.166.0-DEV+extended",
            "v0.166.0-rc.1+extended",
            "v0.166.0-not-a-hash",
        ]
        for version in versions:
            with self.subTest(version=version):
                result = self.check_version(f"hugo {version} linux/amd64")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(f"Expected Hugo v0.166.0, got {version}", result.stderr)

    def test_rejects_missing_version(self):
        result = self.check_version("")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Expected Hugo v0.166.0, got (missing version)", result.stderr)

    def test_rejects_failed_hugo_command_even_with_valid_output(self):
        result = self.check_version("hugo v0.166.0+extended", exit_code=1)
        self.assertNotEqual(result.returncode, 0)


if __name__ == "__main__":
    unittest.main()
