# Protect the archived downloads after removing the old site's duplicate copies from the checkout.
import json
import shutil
import tempfile
import unittest
from pathlib import Path

from check_site import check_downloads


class DownloadIntegrityTest(unittest.TestCase):
    def setUp(self):
        self.site = Path(__file__).resolve().parents[1]
        self.manifest = json.loads((self.site / "migration.json").read_text())

    def test_retained_downloads_match_original_hashes(self):
        check_downloads(self.site / "static", self.manifest)

    def test_changed_published_download_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            public = Path(directory)
            shutil.copytree(self.site / "static/downloads", public / "downloads")
            with (public / "downloads/cv.pdf").open("ab") as output:
                output.write(b"changed")
            with self.assertRaisesRegex(
                AssertionError, "Download changed: downloads/cv.pdf"
            ):
                check_downloads(public, self.manifest)

    def test_missing_published_download_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            public = Path(directory)
            shutil.copytree(self.site / "static/downloads", public / "downloads")
            (public / "downloads/report.pdf").unlink()
            with self.assertRaises(FileNotFoundError):
                check_downloads(public, self.manifest)

    def test_incomplete_hash_contract_is_rejected(self):
        manifest = {"downloads": dict(self.manifest["downloads"])}
        del manifest["downloads"]["downloads/report.pdf"]
        with self.assertRaisesRegex(
            AssertionError, "all three original download hashes"
        ):
            check_downloads(self.site / "static", manifest)


if __name__ == "__main__":
    unittest.main()
