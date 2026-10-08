# Verify regenerated content against captured public URLs so reimporting cannot break existing article links.
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import migrate_content


class MigrationURLsTest(unittest.TestCase):
    def test_regenerated_archive_keeps_verified_urls_and_drafts(self):
        verified = json.loads(
            (Path(__file__).parent / "legacy-post-urls.json").read_text()
        )["posts"]
        with tempfile.TemporaryDirectory() as directory:
            destination = Path(directory)
            with patch.object(migrate_content, "DEST", destination):
                routes = migrate_content.migrate_posts()
            self.assertEqual(
                {
                    Path(route["source"]).name: route["canonical_url"]
                    for route in routes
                    if not route["draft"]
                },
                verified,
            )
            drafts = []
            for route in routes:
                text = (destination / route["destination"]).read_text()
                header, delimiter, _ = text.partition("\n---\n")
                self.assertTrue(delimiter)
                metadata = json.loads(
                    "\n".join(
                        line
                        for line in header.splitlines()[1:]
                        if not line.lstrip().startswith("#")
                    )
                )
                self.assertEqual(metadata["url"], route["canonical_url"])
                self.assertEqual(metadata["draft"], route["draft"])
                self.assertIn(f"/posts/{metadata['slug']}/", metadata["aliases"])
                if metadata["draft"]:
                    drafts.append(route["source"])
            self.assertEqual(
                drafts,
                [
                    "_posts/2012-01-07-creating-nice-looking-documents-with-lhs2TeX.markdown",
                    "_posts/2012-01-28-the-acta-debacle.md",
                ],
            )
            embedding = (
                destination
                / "content/posts/the-difference-between-shallow-and-deep-embedding.md"
            ).read_text()
            self.assertIn(
                "[folds](/2009/12/17/haskell-datatypes-and-folds-part-i/)", embedding
            )
            self.assertIn("([2](/2010/01/03/datatypes-and-folds-part-ii/))", embedding)

    def test_article_links_use_live_slugs_and_keep_queries_and_fragments(self):
        body = (
            "[folds](/2009/12/17/haskell-datatypes-and-folds/#definition) "
            "[preview](/posts/haskell-datatypes-and-folds/?source=archive#definition) "
            '<a href="http://alessandrovermeulen.me/2009/07/12/extending-the-v-in-mvc-revisited/">MVC</a> '
            "[relative](2013/05/19/why-you-should-switch-to-declarative-programming/) "
            "[external](https://example.org/2009/12/17/haskell-datatypes-and-folds/)"
        )
        expected = (
            "[folds](/2009/12/17/haskell-datatypes-and-folds-part-i/#definition) "
            "[preview](/2009/12/17/haskell-datatypes-and-folds-part-i/?source=archive#definition) "
            '<a href="/2009/07/12/extending-the-v-in-mvc---revisited/">MVC</a> '
            "[relative](/2013/05/19/why-you-should-switch-to-declarative-programming/) "
            "[external](https://example.org/2009/12/17/haskell-datatypes-and-folds/)"
        )
        self.assertEqual(
            migrate_content.rewrite_post_links(body, Path("example.md")), expected
        )


if __name__ == "__main__":
    unittest.main()
