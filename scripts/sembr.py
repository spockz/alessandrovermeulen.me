# WHY: Enforce Semantic Line Breaks (one sentence per line) for Markdown prose.
# This simplifies git diffs, eliminates arbitrary column re-wrapping, and pairs with editor formatters.
import argparse
import re
import sys
from pathlib import Path

# Abbreviations, initials, and numeric patterns that must not trigger a sentence break.
_ABBREVIATIONS = (
    r"\b(?:e\.g|i\.e|vs|etc|al|approx|dr|mr|mrs|ms|prof|sr|jr|no|fig)\."
)
_INITIALS = r"\b[A-Z]\."
_NUMBERS = r"\b\d+\.\d+\b"
_VERSION = r"\bv?\d+(?:\.\d+)+(?:-[a-zA-Z0-9.]+)?\b"
_DOMAINS = r"\b[a-zA-Z0-9-]+\.(?:com|org|net|me|io|app|dev|edu|gov|html|md|go|py|cjs|js|ts|json|yaml|yml)\b"


def _mask_protected_tokens(text: str) -> tuple[str, list[str]]:
    """Mask non-terminating periods so they are preserved during sentence splitting."""
    tokens: list[str] = []

    def _replace(match: re.Match[str]) -> str:
        tokens.append(match.group(0))
        return f"\x00{len(tokens) - 1}\x00"

    pattern = re.compile(
        f"({_ABBREVIATIONS}|{_INITIALS}|{_VERSION}|{_NUMBERS}|{_DOMAINS})",
        re.IGNORECASE,
    )
    masked = pattern.sub(_replace, text)
    return masked, tokens


def _unmask_protected_tokens(text: str, tokens: list[str]) -> str:
    """Restore masked tokens to their original form."""
    def _restore(match: re.Match[str]) -> str:
        idx = int(match.group(1))
        return tokens[idx]

    return re.sub(r"\x00(\d+)\x00", _restore, text)


def split_into_sentences(text: str) -> list[str]:
    """Split a prose string into distinct sentences, respecting protected abbreviations."""
    if not text.strip():
        return []

    masked, tokens = _mask_protected_tokens(text)

    # Match sentence terminators (. ! ?) followed by whitespace or end of string,
    # optionally trailing closing quotes or parentheses.
    pattern = re.compile(r'([.?!]+[\'")\]]*)(?:\s+|$)')

    sentences: list[str] = []
    start = 0
    for match in pattern.finditer(masked):
        end = match.end()
        candidate = masked[start:end].strip()
        if candidate:
            sentences.append(_unmask_protected_tokens(candidate, tokens))
        start = end

    remainder = masked[start:].strip()
    if remainder:
        sentences.append(_unmask_protected_tokens(remainder, tokens))

    return sentences


def format_markdown(content: str) -> str:
    """Format Markdown content into semantic line breaks, preserving syntax structures."""
    lines = content.splitlines(keepends=True)
    out: list[str] = []

    in_frontmatter = False
    in_codeblock = False
    frontmatter_count = 0

    for line in lines:
        stripped = line.strip()

        # Handle YAML/JSON frontmatter boundaries
        if stripped == "---" and frontmatter_count < 2:
            frontmatter_count += 1
            in_frontmatter = (frontmatter_count == 1)
            out.append(line)
            continue

        if in_frontmatter:
            out.append(line)
            continue

        # Handle fenced code blocks
        if stripped.startswith("```") or stripped.startswith("~~~"):
            in_codeblock = not in_codeblock
            out.append(line)
            continue

        if in_codeblock:
            out.append(line)
            continue

        # Preserve structural elements unchanged
        if (
            not stripped
            or stripped.startswith("#")
            or stripped.startswith("|")
            or stripped.startswith("{{<")
            or stripped.startswith("{{%")
            or stripped.startswith("<!--")
            or (stripped.startswith("<") and not stripped.startswith("<http"))
        ):
            out.append(line)
            continue

        # Handle blockquotes
        quote_match = re.match(r"^(\s*>\s*)(.*)$", line.rstrip())
        if quote_match:
            prefix = quote_match.group(1)
            body = quote_match.group(2)
            sentences = split_into_sentences(body)
            if sentences:
                for s in sentences:
                    out.append(f"{prefix}{s}\n")
            else:
                out.append(line)
            continue

        # Handle list items (ordered and unordered)
        list_match = re.match(r"^(\s*(?:[-*+]|\d+\.)\s+)(.*)$", line.rstrip())
        if list_match:
            prefix = list_match.group(1)
            body = list_match.group(2)
            sub_indent = " " * len(prefix)
            sentences = split_into_sentences(body)
            if sentences:
                out.append(f"{prefix}{sentences[0]}\n")
                for s in sentences[1:]:
                    out.append(f"{sub_indent}{s}\n")
            else:
                out.append(line)
            continue

        # Standard prose paragraph line
        sentences = split_into_sentences(line.rstrip())
        if sentences:
            for s in sentences:
                out.append(f"{s}\n")
        else:
            out.append(line)

    return "".join(out)


def main() -> int:
    parser = argparse.ArgumentParser(description="Format Markdown with Semantic Line Breaks (SemBr).")
    parser.add_argument("files", nargs="*", help="Markdown files to process (reads stdin if omitted).")
    parser.add_argument("--check", action="store_true", help="Check formatting without modifying files; exit 1 on diff.")
    parser.add_argument("-w", "--write", action="store_true", help="Write formatted output directly to files in place.")
    args = parser.parse_args()

    # Process standard input if no files specified
    if not args.files or args.files == ["-"]:
        content = sys.stdin.read()
        formatted = format_markdown(content)
        if args.check:
            if content != formatted:
                sys.stderr.write("stdin: needs formatting\n")
                return 1
            return 0
        sys.stdout.write(formatted)
        return 0

    unformatted: list[Path] = []
    for file_path in args.files:
        path = Path(file_path)
        if not path.is_file():
            continue
        original = path.read_text(encoding="utf-8")
        formatted = format_markdown(original)
        if original != formatted:
            unformatted.append(path)
            if args.write:
                path.write_text(formatted, encoding="utf-8")

    if args.check and unformatted:
        for p in unformatted:
            sys.stderr.write(f"{p}: unformatted (requires semantic line breaks)\n")
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
