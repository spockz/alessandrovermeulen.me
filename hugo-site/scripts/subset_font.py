# Create matching build and browser subsets; renaming avoids the upstream reserved name and keeps line metrics identical.
import argparse
from pathlib import Path

from fontTools import subset
from fontTools.ttLib import TTFont


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=Path)
    parser.add_argument("--font", choices=("sans", "mono"), default="sans")
    parser.add_argument("--style", choices=("regular", "bold", "italic", "bold-italic"), default="regular")
    args = parser.parse_args()
    if args.font == "mono" and args.style != "regular":
        parser.error("--style applies only to --font sans")
    site = Path(__file__).resolve().parents[1]
    font = TTFont(args.source, recalcTimestamp=False)
    options = subset.Options()
    options.layout_features = ["kern", "liga", "clig", "calt", "mark", "mkmk", "ccmp", "rlig", "locl"]
    options.hinting = False
    options.name_IDs = [0, 1, 2, 3, 4, 5, 6, 13, 14, 16, 17]
    options.name_legacy = True
    options.name_languages = ["*"]
    worker = subset.Subsetter(options=options)
    worker.populate(unicodes=set(range(0x20, 0x180)) | set(range(0x2000, 0x2070)))
    worker.subset(font)
    profiles = {
        "sans": {
            "family": "Article Sans",
            "styles": {
                "regular": ("Regular", "ArticleSans-Regular", "ArticleSans-Regular-3.052-Latin", "Regular", "regular"),
                "bold": ("Bold", "ArticleSans-Bold", "ArticleSans-Bold-3.052-Latin", "Bold", "bold"),
                "italic": ("Italic", "ArticleSans-Italic", "ArticleSans-Italic-3.052-Latin", "Italic", "italic"),
                "bold-italic": ("Bold Italic", "ArticleSans-BoldItalic", "ArticleSans-BoldItalic-3.052-Latin", "BoldItalic", "bold-italic"),
            },
        },
        "mono": {
            "family": "Article Mono",
            "styles": {
                "regular": ("Regular", "ArticleMono-Regular", "ArticleMono-Regular-2.042-Latin", "Regular", "regular"),
            },
        },
    }
    profile = profiles[args.font]
    family = profile["family"]
    style_name, postscript, unique, file_style, style_slug = profile["styles"][args.style]
    suffix = "" if args.font == "mono" else f"-{file_style}"
    woff_suffix = "" if args.font == "mono" or args.style == "regular" else f"-{style_slug}"
    names = {
        1: family,
        2: style_name,
        3: unique,
        4: f"{family} {style_name}",
        6: postscript,
        16: family,
        17: style_name,
    }
    for record in font["name"].names:
        if record.nameID in names:
            record.string = names[record.nameID].encode(record.getEncoding())
    ttf = site / "scripts/typography/font" / f"Article{'Sans' if args.font == 'sans' else 'Mono'}{suffix or '-Regular'}.ttf"
    woff = site / "static/fonts" / f"article-{'sans' if args.font == 'sans' else 'mono'}-v{'3052' if args.font == 'sans' else '2042'}{woff_suffix}-latin.woff2"
    ttf.parent.mkdir(parents=True, exist_ok=True)
    woff.parent.mkdir(parents=True, exist_ok=True)
    font.save(ttf)
    font.flavor = "woff2"
    font.save(woff)
    restored = TTFont(woff)
    for table in ("hmtx", "cmap", "GPOS", "GSUB"):
        if table not in font:
            continue
        if restored[table].compile(restored) != font[table].compile(font):
            raise ValueError(f"Font metrics/layout differ in {table}")
    print(f"Matching fonts: {ttf.stat().st_size:,} byte TTF, {woff.stat().st_size:,} byte WOFF2")


if __name__ == "__main__":
    main()
