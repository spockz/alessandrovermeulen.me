# Generate local theme icons from the shared badge specification so updates stay deterministic and dependency-free.
import argparse
import re
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

import tomllib

GENERATED_COMMENT = (
    "<!-- Generated from data/project-badges.toml with make badges to preserve "
    "the shared project emblem style. -->\n"
)
SHAPE_ATTRIBUTES = {
    "path": {"d"},
    "circle": {"cx", "cy", "r"},
    "rect": {"x", "y", "width", "height", "rx"},
    "line": {"x1", "y1", "x2", "y2"},
    "polyline": {"points"},
    "polygon": {"points"},
}


def render_badge(style: dict, name: str, badge: dict) -> str:
    if not re.fullmatch(r"[a-z][a-z0-9]*(?:-[a-z0-9]+)*", name):
        raise ValueError(f"Invalid badge identifier: {name}")
    size = style["canvas"]
    inset = style["frame_inset"]
    if size <= 2 * inset or inset < 0 or style["stroke_width"] <= 0:
        raise ValueError("Badge canvas, frame inset, and stroke width are invalid")
    svg = ET.Element(
        "svg",
        {
            "xmlns": "http://www.w3.org/2000/svg",
            "viewBox": f"0 0 {size} {size}",
            "class": "project-badge",
            "data-project-badge": name,
            "aria-hidden": "true",
            "focusable": "false",
        },
    )
    ET.SubElement(
        svg,
        "rect",
        {
            "x": str(inset),
            "y": str(inset),
            "width": str(size - 2 * inset),
            "height": str(size - 2 * inset),
            "rx": str(style["frame_radius"]),
            "fill": "currentColor",
            "fill-opacity": str(style["surface_opacity"]),
            "stroke": "currentColor",
            "stroke-opacity": str(style["frame_opacity"]),
            "stroke-width": str(style["frame_stroke_width"]),
        },
    )
    symbol = ET.SubElement(
        svg,
        "g",
        {
            "fill": "none",
            "stroke": "currentColor",
            "stroke-width": str(style["stroke_width"]),
            "stroke-linecap": "round",
            "stroke-linejoin": "round",
        },
    )
    if not badge["shapes"]:
        raise ValueError(f"Badge {name} has no symbol")
    for shape in badge["shapes"]:
        tag = shape["tag"]
        if tag not in SHAPE_ATTRIBUTES:
            raise ValueError(f"Unsupported shape {tag} in badge {name}")
        attributes = {key: str(value) for key, value in shape.items() if key != "tag"}
        unknown = attributes.keys() - (SHAPE_ATTRIBUTES[tag] | {"fill"})
        if unknown or attributes.get("fill", "none") not in {"none", "currentColor"}:
            raise ValueError(f"Unsupported attributes in badge {name}: {attributes}")
        ET.SubElement(symbol, tag, attributes)
    ET.indent(svg, space="  ")
    return GENERATED_COMMENT + ET.tostring(svg, encoding="unicode") + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Regenerate consistent SVG project emblems."
    )
    parser.add_argument(
        "--check", action="store_true", help="Fail if generated badges need updating."
    )
    args = parser.parse_args()
    site = Path(__file__).resolve().parents[1]
    catalog = tomllib.loads((site / "data/project-badges.toml").read_text())
    output = site / "assets/icons"
    expected = {
        output / f"project-{name}.svg": render_badge(catalog["style"], name, badge)
        for name, badge in catalog["badges"].items()
    }
    changed = {
        path: source
        for path, source in expected.items()
        if not path.exists() or path.read_text() != source
    }
    obsolete = set(output.glob("project-*.svg")) - expected.keys()
    if args.check:
        if changed or obsolete:
            print("Project badges are out of date; run make badges:", file=sys.stderr)
            for path in sorted(changed.keys() | obsolete):
                print(f"  {path.relative_to(site)}", file=sys.stderr)
            return 1
        print(f"All {len(expected)} project badges match the shared style.")
        return 0
    output.mkdir(parents=True, exist_ok=True)
    for path in obsolete:
        path.unlink()
    for path, source in changed.items():
        path.write_text(source, encoding="utf-8")
    print(f"Generated {len(expected)} project badges; updated {len(changed)} files.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, KeyError, TypeError, ValueError) as error:
        raise SystemExit(f"Project badge generation failed: {error}") from error
