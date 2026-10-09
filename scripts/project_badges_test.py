# Exercise the real regeneration command so missing, edited, and future project emblems cannot silently drift from the style.
import shutil
import subprocess
import sys
import tempfile
import unittest
import xml.etree.ElementTree as ET
from pathlib import Path


class ProjectBadgeRegenerationTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.site = Path(self.directory.name)
        source = Path(__file__).resolve().parents[1]
        (self.site / "scripts").mkdir()
        (self.site / "data").mkdir()
        shutil.copyfile(
            source / "scripts/generate_project_badges.py",
            self.site / "scripts/generate_project_badges.py",
        )
        self.catalog = self.site / "data/project-badges.toml"
        shutil.copyfile(source / "data/project-badges.toml", self.catalog)

    def run_generator(self, *arguments):
        return subprocess.run(
            [
                sys.executable,
                str(self.site / "scripts/generate_project_badges.py"),
                *arguments,
            ],
            capture_output=True,
            text=True,
            check=False,
        )

    def test_detects_missing_and_edited_badges_and_restores_them(self):
        missing = self.run_generator("--check")
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn("run make badges", missing.stderr)
        generated = self.run_generator()
        self.assertEqual(generated.returncode, 0, generated.stderr)
        self.assertEqual(self.run_generator("--check").returncode, 0)
        badge = next((self.site / "assets/icons").glob("project-*.svg"))
        original = badge.read_text()
        edited = ET.parse(badge)
        edited.find("{http://www.w3.org/2000/svg}g").set("stroke-width", "12")
        edited.write(badge, encoding="utf-8")
        stale = self.run_generator("--check")
        self.assertNotEqual(stale.returncode, 0)
        self.assertIn(badge.name, stale.stderr)
        restored = self.run_generator()
        self.assertEqual(restored.returncode, 0, restored.stderr)
        self.assertEqual(badge.read_text(), original)
        self.assertEqual(self.run_generator("--check").returncode, 0)

    def test_generates_new_projects_and_removes_retired_emblems(self):
        original = self.catalog.read_text()
        self.catalog.write_text(
            original + '\n[badges.future-project]\nconcept = "A new project"\n'
            'shapes = [{tag = "circle", cx = 48, cy = 48, r = 12}]\n',
            encoding="utf-8",
        )
        generated = self.run_generator()
        self.assertEqual(generated.returncode, 0, generated.stderr)
        future = self.site / "assets/icons/project-future-project.svg"
        self.assertTrue(future.is_file())
        self.assertEqual(self.run_generator("--check").returncode, 0)
        self.catalog.write_text(original, encoding="utf-8")
        stale = self.run_generator("--check")
        self.assertNotEqual(stale.returncode, 0)
        self.assertIn(future.name, stale.stderr)
        regenerated = self.run_generator()
        self.assertEqual(regenerated.returncode, 0, regenerated.stderr)
        self.assertFalse(future.exists())
        self.assertEqual(self.run_generator("--check").returncode, 0)


if __name__ == "__main__":
    unittest.main()
