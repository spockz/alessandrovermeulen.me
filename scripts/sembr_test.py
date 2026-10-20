# WHY: Test Semantic Line Breaks (SemBr) Markdown formatting edge cases.
import unittest
from scripts.sembr import format_markdown, split_into_sentences


class SembrFormattingTest(unittest.TestCase):
    def test_split_into_sentences_simple(self) -> None:
        text = "This is sentence one. This is sentence two! Is this sentence three?"
        expected = [
            "This is sentence one.",
            "This is sentence two!",
            "Is this sentence three?",
        ]
        self.assertEqual(split_into_sentences(text), expected)

    def test_split_preserves_abbreviations_and_numbers(self) -> None:
        text = "In GPT-5.6 Terra was fine, e.g. for review. GPT-6.1 Sol costs 0.25 of 5.6."
        expected = [
            "In GPT-5.6 Terra was fine, e.g. for review.",
            "GPT-6.1 Sol costs 0.25 of 5.6.",
        ]
        self.assertEqual(split_into_sentences(text), expected)

    def test_format_markdown_preserves_frontmatter(self) -> None:
        content = "---\n{\n  \"title\": \"Sample. Title.\"\n}\n---\nFirst sentence. Second sentence.\n"
        expected = "---\n{\n  \"title\": \"Sample. Title.\"\n}\n---\nFirst sentence.\nSecond sentence.\n"
        self.assertEqual(format_markdown(content), expected)

    def test_format_markdown_preserves_fenced_code_blocks(self) -> None:
        content = "```python\n# Some comment. Not a sentence.\nx = 1.0\n```\nFirst line. Second line.\n"
        expected = "```python\n# Some comment. Not a sentence.\nx = 1.0\n```\nFirst line.\nSecond line.\n"
        self.assertEqual(format_markdown(content), expected)

    def test_format_markdown_preserves_list_indentation(self) -> None:
        content = "  - Item one. Item one continuation.\n  - Item two.\n"
        expected = "  - Item one.\n    Item one continuation.\n  - Item two.\n"
        self.assertEqual(format_markdown(content), expected)

    def test_format_markdown_preserves_shortcodes(self) -> None:
        content = "{{< observation title=\"Test\" >}}\nSentence one. Sentence two.\n{{< /observation >}}\n"
        expected = "{{< observation title=\"Test\" >}}\nSentence one.\nSentence two.\n{{< /observation >}}\n"
        self.assertEqual(format_markdown(content), expected)


if __name__ == "__main__":
    unittest.main()
