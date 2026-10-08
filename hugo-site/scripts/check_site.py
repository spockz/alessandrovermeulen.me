# Verify published routes, feeds, local references, and loading budgets against the migration contract.
import argparse
import csv
import gzip
import hashlib
import io
import json
import re
import subprocess
import sys
import xml.etree.ElementTree as ET
from dataclasses import dataclass
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import unquote, urljoin, urlsplit


@dataclass(frozen=True)
class Reference:
    url: str
    asset: bool


@dataclass(frozen=True)
class FontScope:
    tag: str
    in_article: bool
    in_typography: bool
    in_pre: bool
    family: str
    weight: str
    style: str


ARTICLE_FONT_FILES = {
    "article-sans-v3052-latin.woff2": "sans-regular",
    "article-sans-v3052-bold-latin.woff2": "sans-bold",
    "article-sans-v3052-italic-latin.woff2": "sans-italic",
    "article-sans-v3052-bold-italic-latin.woff2": "sans-bold-italic",
    "article-mono-v2042-latin.woff2": "mono",
}


class Page(HTMLParser):
    def __init__(self, source: str):
        super().__init__(convert_charrefs=True)
        self.references: list[Reference] = []
        self.ids: set[str] = set()
        self.canonical: str | None = None
        self.redirect: str | None = None
        self.feed_links: list[dict[str, str | None]] = []
        self.used_fonts: set[str] = set()
        self.has_article_text = False
        self.blog_typography = False
        self.article_scopes: list[dict[str, str | None]] = []
        self.summary_scopes: list[dict[str, str | None]] = []
        self.scopes: list[FontScope] = []
        self.feed(source)

    @classmethod
    def from_file(cls, filename: Path):
        return cls(filename.read_text(encoding="utf-8"))

    def handle_starttag(self, tag: str, attributes: list[tuple[str, str | None]]):
        attrs = dict(attributes)
        parent = (
            self.scopes[-1]
            if self.scopes
            else FontScope("", False, False, False, "other", "400", "normal")
        )
        in_article = parent.in_article or (
            "article-content" in attrs.get("class", "").split()
        )
        typography = parent.in_typography or (
            "typography-prose" in attrs.get("class", "").split()
        )
        in_pre = parent.in_pre or tag == "pre"
        family, weight, style = parent.family, parent.weight, parent.style
        if {"article-content", "typography-summary"} & set(
            attrs.get("class", "").split()
        ):
            family, weight, style = "sans", "400", "normal"
        if tag in {"h1", "h2", "h3", "h4", "h5", "h6", "strong", "b", "th", "dt"}:
            weight = "700"
        if tag in {"em", "i", "cite", "blockquote"}:
            style = "italic"
        if tag == "blockquote" and typography:
            weight = "400"
        if tag == "figcaption":
            weight, style = "400", "normal"
        if tag in {"code", "kbd", "samp"}:
            if typography and not in_pre:
                family, weight, style = "mono", "400", "normal"
            else:
                family = "other"
        if tag == "pre":
            family = "other"
        if tag in {"script", "style", "textarea"}:
            family = "other"

        for declaration in attrs.get("style", "").split(";"):
            key, separator, value = declaration.partition(":")
            if not separator:
                continue
            key, value = key.strip().lower(), value.strip().lower()
            if key == "font-weight":
                if value not in {"inherit", "unset", "initial"}:
                    weight = normalize_weight(value)
            elif key == "font-style":
                if value not in {"inherit", "unset", "initial"}:
                    style = "italic" if value == "italic" else "normal"
            elif key == "font-family":
                selected = value.split(",", 1)[0].strip(" \t\"'").lower()
                if selected not in {"inherit", "unset", "initial"}:
                    family = font_family_for(selected)

        plan_weight = attrs.get("data-plan-weight")
        plan_style = attrs.get("data-plan-style")
        if plan_weight is not None:
            weight = normalize_weight(plan_weight)
        if plan_style is not None:
            style = "italic" if plan_style.lower() == "italic" else "normal"
        encoded_format = attrs.get("data-plan-inline")
        if encoded_format is not None:
            try:
                inline_format = json.loads(encoded_format)
            except json.JSONDecodeError as error:
                raise ValueError(f"Invalid data-plan-inline JSON: {error}") from error
            if not isinstance(inline_format, dict):
                raise ValueError("data-plan-inline must contain a JSON object")
            if {"family", "weight", "style"} <= inline_format.keys():
                inline_family = inline_format["family"]
                inline_weight = inline_format["weight"]
                inline_style = inline_format["style"]
                if not all(
                    isinstance(value, str)
                    for value in (inline_family, inline_weight, inline_style)
                ):
                    raise ValueError(
                        "data-plan-inline font family, weight, and style must be strings"
                    )
                inline_family = inline_family.lower()
                if inline_family not in {"article sans", "article mono"}:
                    raise ValueError(
                        f"Unsupported data-plan-inline font family: {inline_family}"
                    )
                if inline_weight not in {"400", "700"} or inline_style not in {
                    "normal",
                    "italic",
                }:
                    raise ValueError("Unsupported data-plan-inline weight or style")
                family = font_family_for(inline_family)
                weight = normalize_weight(inline_weight)
                style = "italic" if inline_style == "italic" else "normal"
            elif set(inline_format) == {"width"} and isinstance(
                inline_format["width"], (int, float)
            ):
                pass
            else:
                raise ValueError("Unexpected data-plan-inline metadata object")

        if tag not in {
            "area",
            "base",
            "br",
            "col",
            "embed",
            "hr",
            "img",
            "input",
            "link",
            "meta",
            "param",
            "source",
            "track",
            "wbr",
        }:
            self.scopes.append(
                FontScope(tag, in_article, typography, in_pre, family, weight, style)
            )
        if "article-content" in attrs.get("class", "").split():
            self.article_scopes.append(attrs)
        if "article-link__summary" in attrs.get("class", "").split():
            self.summary_scopes.append(attrs)
        if (
            tag == "meta"
            and attrs.get("name") == "article-typography"
            and attrs.get("content") == "planned"
        ):
            self.blog_typography = True
        if attrs.get("id"):
            self.ids.add(attrs["id"])
        if tag == "meta" and attrs.get("http-equiv", "").lower() == "refresh":
            self.redirect = attrs.get("content", "").split("url=", 1)[-1]
        if tag == "link" and attrs.get("rel") == "canonical":
            self.canonical = attrs.get("href")
        if (
            tag == "link"
            and "alternate" in attrs.get("rel", "").split()
            and attrs.get("type") in {"application/atom+xml", "application/rss+xml"}
            and any(scope.tag == "head" for scope in self.scopes)
        ):
            self.feed_links.append(attrs)
        for attribute in ("href", "src", "poster"):
            value = attrs.get(attribute)
            if value:
                asset = attribute != "href" or (
                    tag == "link"
                    and attrs.get("rel") in {"stylesheet", "icon", "preload"}
                )
                self.references.append(Reference(url=value, asset=asset))
        if attrs.get("srcset"):
            for candidate in attrs["srcset"].split(","):
                self.references.append(
                    Reference(url=candidate.strip().split()[0], asset=True)
                )

    def handle_endtag(self, tag: str):
        for index in range(len(self.scopes) - 1, -1, -1):
            if self.scopes[index].tag == tag:
                del self.scopes[index:]
                break

    def handle_data(self, data: str):
        if not data or data.isspace() or not self.scopes:
            return
        scope = self.scopes[-1]
        if not (scope.in_article or scope.in_typography) or scope.in_pre:
            return
        self.has_article_text = True
        if scope.family == "mono":
            self.used_fonts.add("mono")
            return
        if scope.family != "sans":
            return
        weight = "bold" if normalize_weight(scope.weight) == "700" else "regular"
        style = "italic" if scope.style.lower() == "italic" else "normal"
        suffix = (
            f"{weight}-{style}"
            if weight == "bold" and style == "italic"
            else (
                "bold"
                if weight == "bold"
                else "italic"
                if style == "italic"
                else "regular"
            )
        )
        self.used_fonts.add(f"sans-{suffix}")


def font_family_for(value: str) -> str:
    return {
        "article sans": "sans",
        "article mono": "mono",
    }.get(value, "other")


def normalize_weight(value: str) -> str:
    value = value.strip().lower()
    if value in {"bold", "bolder"}:
        return "700"
    if value in {"normal", "lighter"}:
        return "400"
    try:
        return "700" if float(value) >= 600 else "400"
    except ValueError:
        return "400"


def output_file(public: Path, route: str) -> Path:
    result = public / unquote(urlsplit(route).path).lstrip("/")
    return result / "index.html" if result.is_dir() or route.endswith("/") else result


def require(condition: bool, message: str):
    if not condition:
        raise AssertionError(message)


def check_routes(site: Path, public: Path, manifest: dict):
    legacy_urls = json.loads((site / "scripts" / "legacy-post-urls.json").read_text())[
        "posts"
    ]
    imported_urls = {
        Path(entry["source"]).name: entry["canonical_url"]
        for entry in manifest["routes"]
        if entry["source"].startswith("_posts/") and not entry["draft"]
    }
    require(
        imported_urls == legacy_urls,
        "Published archive URLs must match the verified live site",
    )
    rows = csv.DictReader(
        io.StringIO(
            subprocess.check_output(["hugo", "list", "all"], cwd=site, text=True)
        )
    )
    listed = {row["path"]: row for row in rows if row["kind"] == "page"}
    published_rows = list(
        csv.DictReader(
            io.StringIO(
                subprocess.check_output(
                    ["hugo", "list", "published"], cwd=site, text=True
                )
            )
        )
    )
    published = {
        row["path"]: urlsplit(row["permalink"]).path
        for row in published_rows
        if row["kind"] == "page"
    }
    require(
        len(manifest["routes"]) == 37, "Migration must account for 36 posts and About"
    )
    for entry in manifest["routes"]:
        canonical = entry["canonical_url"]
        row = listed.get(entry["destination"])
        require(
            row is not None and urlsplit(row["permalink"]).path == canonical,
            f"Changed route: {entry}",
        )
        require(
            (row["draft"] == "true") == entry["draft"], f"Changed draft status: {entry}"
        )
        filename = output_file(public, canonical)
        aliases = entry.get("aliases", [])
        if entry["draft"]:
            require(
                not filename.is_file(), f"Unpublished source draft leaked: {canonical}"
            )
            for alias in aliases:
                require(
                    not output_file(public, alias).is_file(),
                    f"Unpublished source draft alias leaked: {alias}",
                )
            continue
        require(filename.is_file(), f"Missing page: {canonical}")
        disk_page = Page.from_file(filename)
        require(
            disk_page.redirect is None
            and disk_page.canonical == "https://alessandrovermeulen.me" + canonical,
            f"Invalid canonical: {canonical}",
        )
        for alias in aliases:
            alias_file = output_file(public, alias)
            require(alias_file.is_file(), f"Missing compatibility alias: {alias}")
            alias_page = Page.from_file(alias_file)
            target = "https://alessandrovermeulen.me" + canonical
            require(
                alias_page.redirect == target and alias_page.canonical == target,
                f"Alias must point directly to its canonical article: {alias}",
            )
    published_routes = set(published.values())
    for route in published_routes:
        require(
            output_file(public, route).is_file(), f"Published page missing: {route}"
        )
    for path, route in published.items():
        if not path.startswith("posts/"):
            continue
        page = Page.from_file(output_file(public, route))
        require(
            page.blog_typography and len(page.article_scopes) == 1,
            f"Blog typography missing: {route}",
        )
        scope = page.article_scopes[0]
        require(
            scope.get("data-plan-lines") == "responsive"
            and scope.get("data-plan-generated") == "true"
            and "typography-prose" in scope.get("class", "").split(),
            f"Blog default layout missing: {route}",
        )
    namespace = {"sitemap": "http://www.sitemaps.org/schemas/sitemap/0.9"}
    sitemap = ET.parse(public / "sitemap.xml")
    locations = {
        node.text for node in sitemap.findall("sitemap:url/sitemap:loc", namespace)
    }
    require(
        {"https://alessandrovermeulen.me" + route for route in published_routes}
        <= locations,
        "Sitemap must list every published canonical URL",
    )
    for entry in manifest["routes"]:
        for alias in entry.get("aliases", []):
            require(
                "https://alessandrovermeulen.me" + alias not in locations,
                f"Sitemap must exclude aliases: {alias}",
            )
        if entry["draft"]:
            require(
                "https://alessandrovermeulen.me" + entry["canonical_url"]
                not in locations,
                "Sitemap must exclude drafts",
            )
    require(
        "Sitemap: https://alessandrovermeulen.me/sitemap.xml"
        in (public / "robots.txt").read_text(),
        "robots.txt must advertise the sitemap",
    )
    return {
        urlsplit(row["permalink"]).path
        for row in published_rows
        if row["kind"] == "page" and row["section"] == "posts"
    }


def check_feed(public: Path, post_routes: set[str]):
    namespace = {"atom": "http://www.w3.org/2005/Atom"}
    feed = ET.parse(public / "atom.xml")
    require(
        feed.getroot().tag == "{http://www.w3.org/2005/Atom}feed",
        "Legacy subscription endpoint must contain Atom",
    )
    entries = feed.findall("atom:entry", namespace)
    require(
        len(entries) == len(post_routes),
        "Atom feed must include every published post and exclude drafts",
    )
    links = {
        urlsplit(entry.find("atom:link", namespace).attrib["href"]).path
        for entry in entries
    }
    require(links == post_routes, "Atom entries must match the published post routes")
    for entry in entries:
        require(
            entry.find("atom:id", namespace).text
            == entry.find("atom:link", namespace).attrib["href"],
            "Atom IDs must use the canonical article URLs",
        )
        require(
            entry.find("atom:content", namespace).text is not None,
            "Feed must retain article contents",
        )


def check_references(public: Path, pages: dict[Path, Page]):
    for filename, page in pages.items():
        relative = filename.relative_to(public).as_posix()
        route = "/" + relative.removesuffix("index.html")
        for reference in page.references:
            parsed = urlsplit(
                urljoin("https://alessandrovermeulen.me" + route, reference.url)
            )
            if parsed.scheme not in {"http", "https"}:
                continue
            if parsed.hostname != "alessandrovermeulen.me":
                require(
                    not reference.asset,
                    f"External loading dependency: {filename}: {reference.url}",
                )
                continue
            target = output_file(public, parsed.path)
            require(
                target.is_file(), f"Broken local reference: {relative}: {reference.url}"
            )
            if (
                parsed.fragment
                and target.suffix == ".html"
                and Page.from_file(target).redirect is None
            ):
                require(
                    unquote(parsed.fragment) in Page.from_file(target).ids,
                    f"Missing anchor: {relative}: {reference.url}",
                )


def page_weight(public: Path, route: str, pages: dict[Path, Page]) -> dict:
    page_file = output_file(public, route)
    page = Page.from_file(page_file)
    resources = {page_file}
    resources.update(
        output_file(public, reference.url)
        for reference in page.references
        if reference.asset
    )
    for filename in resources.copy():
        if filename.suffix != ".css":
            continue
        css = filename.read_text()
        require(
            "@import" not in css,
            "Imported stylesheets require recursive budget measurement",
        )
        urls = re.findall(r"url\(\s*(?:\"([^\"]*)\"|'([^']*)'|([^\s)]*))\s*\)", css)
        for groups in urls:
            url = next(value for value in groups if value)
            if url.startswith(("data:", "#")):
                continue
            base = (
                "https://alessandrovermeulen.me/"
                + filename.relative_to(public).as_posix()
            )
            parsed = urlsplit(urljoin(base, url))
            require(
                parsed.hostname == "alessandrovermeulen.me",
                f"External CSS dependency: {url}",
            )
            target = output_file(public, parsed.path)
            require(target.is_file(), f"Missing CSS dependency: {url}")
            font_face = ARTICLE_FONT_FILES.get(target.name)
            if font_face is not None and font_face not in page.used_fonts:
                continue
            if (
                target.suffix in {".woff2", ".woff", ".ttf", ".otf"}
                and font_face is None
                and not page.has_article_text
            ):
                continue
            resources.add(target)
    files = [
        {
            "file": filename.relative_to(public).as_posix(),
            "raw_bytes": filename.stat().st_size,
            "gzip_bytes": len(
                gzip.compress(filename.read_bytes(), compresslevel=9, mtime=0)
            ),
        }
        for filename in sorted(resources)
    ]
    compressed = sum(item["gzip_bytes"] for item in files)
    code = sum(
        item["gzip_bytes"]
        for item in files
        if Path(item["file"]).suffix in {".html", ".css", ".js"}
    )
    require(code < 100_000, f"HTML/CSS/JS budget exceeded: {route}: {code}")
    require(
        compressed < 200_000, f"Initial transfer budget exceeded: {route}: {compressed}"
    )
    return {"route": route, "gzip_bytes": compressed, "files": files}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("public", type=Path)
    args = parser.parse_args()
    site = Path(__file__).resolve().parents[1]
    public = args.public.resolve()
    manifest = json.loads((site / "migration.json").read_text())
    pages = {filename: Page.from_file(filename) for filename in public.rglob("*.html")}
    post_routes = check_routes(site, public, manifest)
    check_feed(public, post_routes)
    check_references(public, pages)
    for filename, page in pages.items():
        if page.redirect is None:
            require(
                len(page.feed_links) == 1
                and page.feed_links[0].get("type") == "application/atom+xml"
                and page.feed_links[0].get("href") == "/atom.xml",
                f"Page head must advertise the shared Atom feed exactly once: {filename}",
            )
        for summary in page.summary_scopes:
            require(
                summary.get("data-plan-lines") == "responsive"
                and summary.get("data-plan-generated") == "true"
                and "typography-summary" in summary.get("class", "").split(),
                f"Summary typography missing: {filename}",
            )
        if page.summary_scopes:
            require(
                any(
                    "planned-layout.min." in reference.url
                    for reference in page.references
                ),
                f"Summary layout guard missing: {filename}",
            )
    for name in ("cv.pdf", "report.pdf", "report-on-getting-rid-of-js.pdf"):
        original = site.parent / "downloads" / name
        migrated = public / "downloads" / name
        require(
            hashlib.sha256(original.read_bytes()).digest()
            == hashlib.sha256(migrated.read_bytes()).digest(),
            f"Download changed: {name}",
        )
    weights = [
        page_weight(public, route, pages)
        for route in (
            "/",
            "/projects/",
            "/posts/",
            "/2013/07/13/the-difference-between-shallow-and-deep-embedding/",
            "/2015/08/31/finaglecon/",
        )
    ]
    report = {
        "posts": len(post_routes),
        "html_pages": len(pages),
        "weights": weights,
    }
    output = site / ".cache" / "review" / "site.json"
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps(report, indent=2))


if __name__ == "__main__":
    try:
        main()
    except (
        AssertionError,
        OSError,
        RuntimeError,
        ValueError,
        ET.ParseError,
        subprocess.CalledProcessError,
    ) as error:
        sys.exit(str(error))
