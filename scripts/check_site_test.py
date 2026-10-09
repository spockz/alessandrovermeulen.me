# Verify automatic post URLs and protect the original retained downloads.
import csv
import io
import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

from check_site import check_downloads


def list_fixture_pages(project: Path):
    site = Path(__file__).resolve().parents[1]
    output = subprocess.check_output(
        [
            "hugo",
            "list",
            "all",
            "--source",
            str(project),
            "--config",
            str(site / "config/_default/hugo.toml"),
        ],
        cwd=site,
        text=True,
    )
    return list(csv.DictReader(io.StringIO(output)))


class PostPermalinkTest(unittest.TestCase):
    def test_real_hugo_generates_post_urls_and_respects_overrides(self):
        cases = {
            "2026-10-08-a-fresh-leaf.md": {
                "front_matter": {"title": "A fresh leaf"},
                "route": "https://alessandrovermeulen.me/posts/a-fresh-leaf/",
            },
            "plain-filename.md": {
                "front_matter": {"title": "An ordinary article"},
                "route": "https://alessandrovermeulen.me/posts/an-ordinary-article/",
            },
            "2026-10-08-anniversary.md": {
                "front_matter": {"title": "Notes from 2025"},
                "route": "https://alessandrovermeulen.me/posts/notes-from-2025/",
            },
            "2026-10-08-different.md": {
                "front_matter": {
                    "title": "A revised headline",
                    "slug": "keep-this-name",
                },
                "route": "https://alessandrovermeulen.me/posts/keep-this-name/",
            },
            "2026-10-08-archived.md": {
                "front_matter": {
                    "title": "Pinned archive",
                    "slug": "ignored-name",
                    "url": "/2015/08/31/pinned-archive/",
                },
                "route": "https://alessandrovermeulen.me/2015/08/31/pinned-archive/",
            },
            "2026-10-08-bundle/index.md": {
                "front_matter": {"title": "A bundled article"},
                "route": "https://alessandrovermeulen.me/posts/a-bundled-article/",
            },
        }
        with tempfile.TemporaryDirectory() as directory:
            project = Path(directory)
            content = project / "content"
            for filename, case in cases.items():
                post = content / "posts" / filename
                post.parent.mkdir(parents=True, exist_ok=True)
                front_matter = {
                    "date": "2020-01-01T12:00:00+01:00",
                    "draft": False,
                    **case["front_matter"],
                }
                post.write_text(
                    "---\n" + json.dumps(front_matter) + "\n---\nExample prose.\n",
                    encoding="utf-8",
                )
            (content / "posts/_index.md").write_text(
                '---\n{"title": "Writing"}\n---\n', encoding="utf-8"
            )
            rows = list_fixture_pages(project)
        posts = {
            row["title"]: row["permalink"] for row in rows if row["kind"] == "page"
        }
        self.assertEqual(len(posts), len(cases))
        for filename, case in cases.items():
            with self.subTest(filename=filename):
                self.assertEqual(posts[case["front_matter"]["title"]], case["route"])
        section = next(row for row in rows if row["kind"] == "section")
        self.assertEqual(section["permalink"], "https://alessandrovermeulen.me/posts/")

    def test_hugo_new_generates_clean_title_and_url_from_dated_filename(self):
        site = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as directory:
            project = Path(directory)
            shutil.copytree(site / "archetypes", project / "archetypes")
            (project / "content").mkdir()
            subprocess.check_output(
                [
                    "hugo",
                    "new",
                    "content",
                    "posts/2026-10-08-a-fresh-leaf.md",
                    "--source",
                    str(project),
                    "--config",
                    str(site / "config/_default/hugo.toml"),
                ],
                cwd=site,
                text=True,
            )
            post = next(
                row for row in list_fixture_pages(project) if row["kind"] == "page"
            )
        self.assertEqual(post["title"], "A Fresh Leaf")
        self.assertEqual(post["slug"], "")
        self.assertEqual(post["draft"], "true")
        self.assertEqual(
            post["permalink"], "https://alessandrovermeulen.me/posts/a-fresh-leaf/"
        )


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
            with (public / "downloads/report.pdf").open("ab") as output:
                output.write(b"changed")
            with self.assertRaisesRegex(
                AssertionError, "Download changed: downloads/report.pdf"
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
        with self.assertRaisesRegex(AssertionError, "both retained download hashes"):
            check_downloads(self.site / "static", manifest)

    def test_retired_cv_is_not_published(self):
        with tempfile.TemporaryDirectory() as directory:
            public = Path(directory)
            shutil.copytree(self.site / "static/downloads", public / "downloads")
            (public / "downloads/cv.pdf").write_bytes(b"retired CV")
            with self.assertRaisesRegex(
                AssertionError, "Retired CV must not be published"
            ):
                check_downloads(public, self.manifest)


if __name__ == "__main__":
    unittest.main()
