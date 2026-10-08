"""WHY: Recreate Hugo content from archived Jekyll sources so migration is repeatable without modifying those sources."""

from __future__ import annotations

import json
import re
import shutil
from dataclasses import dataclass
from datetime import date, datetime, time
from pathlib import Path
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parents[2]
DEST = ROOT / "hugo-site"
TZ = ZoneInfo("Europe/Amsterdam")
LEGACY_POST_URLS = json.loads(
    (Path(__file__).parent / "legacy-post-urls.json").read_text(encoding="utf-8")
)["posts"]
# Captured from the source Gists because the old Liquid embeds cannot render in Hugo.
GISTS = {
    "1415728": {
        "id_rsa.pub": "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAvwLdglcCqf/f+j+B8dZTVfRUPyhKCxDMEI8bnIbPRFsdCsEM1RlHV2D4bXuLctKgIey95oFvnZyrF9v4PoNSEgfyOiYjNbxmxL2g8LnskbRRM+vv4yH1ed/yENszE9563PSRcvp6NtvD3qTC5hNaYNMP2OO/SzFvS3gFeibolhJZy8AW/6Uv5tLsv4i1GakdKtepGEMTejlT6GrICAowEAjhFeaDDwuY9FCa1O1kednQZ/H+UTw+e9H5kTETc6ewaozT17lCDbcLqL1KdJrR0nMV7LO2OwG39zYDfmUVRd4HPTRX+C0nOETYHBS6ht4G8TQX2iQOSBrRC8KA1U1vlw== alessandrovermeulen@Orthesaurus.local"
    },
    "876971": {
        "gistfile1.css": '/*** Generic Text Markup ***/\n/*** Version: 0.2\n     Date: 18-3-2011\n     Author: Alessandro Vermeulen <alessandro@alessandrovermeulen.me>\n     Licence:\n       Copyright (c) 2011 Alessandro Vermeulen <alessandro@alessandrovermeulen.me>\n\n       Permission is hereby granted, free of charge, to any person obtaining a copy\n       of this software and associated documentation files (the "Software"), to deal\n       in the Software without restriction, including without limitation the rights\n       to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n       copies of the Software, and to permit persons to whom the Software is\n       furnished to do so, subject to the following conditions:\n\n       The above copyright notice and this permission notice shall be included in\n       all copies or substantial portions of the Software.\n\n       THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n       IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n       FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n       AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n       LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n       OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n       THE SOFTWARE.\n***/\nh1, h2, h3, h4, h5, h6 {\n  font-weight: bold;\n}\n\np {\n  text-align: justify;\n}\n\na {\n  text-decoration:none;\n}\n\np a {\n  font-weight: bold;\n}\n\na:hover {\n  text-decoration: underline;\n}\n\nh1 {\n  font-size:14pt;\n}\n\nh2 {\n  font-size: 13pt;\n}\n\nh3 {\n  font-size: 12pt;\n}\n\nh4, h5, h6 {\n  font-size: 11pt;\n}\n\n/* Give head space to text items */\n* + p, * + ul, * + ol, * + table {  \n  margin-top: 10px;\n}\n\n/* Give headers space from their preceeding text */\n* + h2, * + h3, * + h4, * + h5, * + h6 {\n  margin-top: 10px;\n}\n\n/* Align elements tightly to their headers, overriding previous statement\n   of give elements head space. */\nh1 + *, h2 + *, h3 + *, h4 + *, h5 + *, h6 + * {\n  margin-top:0;\n}\n\nul {\n  list-style-type: disc;\n  list-style-position: inside;\n}\n\nol {\n  list-style-type: decimal;\n  list-style-position: inside;\n}\n\ntable {\n  margin: 0 auto;\n  width: 90%;\n}\n/*** End Generic Text Markup ***/',
        "gistfile2.html": '<h1>A level one (1) heading</h1>\n<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed imperdiet, lacus id congue molestie, ipsum nibh aliquet erat, a eleifend justo quam a nisi. Nunc molestie rutrum vulputate. Sed vehicula rhoncus velit, nec lacinia nibh tristique eu. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Sed malesuada est et arcu lobortis viverra. Sed facilisis nulla ut diam adipiscing a dictum nunc feugiat. Proin vestibulum, sem sed tincidunt mattis, nulla massa aliquam enim, ac laoreet ipsum massa quis dui. Cras id metus nisi. Praesent congue blandit sapien et lacinia. Suspendisse varius risus ac magna pellentesque et euismod turpis laoreet. Morbi et tortor lorem. Etiam ac turpis nibh, ac iaculis diam. </p>\n\n<h2>A level two (2) heading</h2>\n<p>Een stukje tekst. Dat lang genoeg is om meerdere regels te beslaan. Als\nhet het goed is staat de titel hierboven een eindje van de vorige paragraaf\naf.</p>\n\n<p>Hierna volgt er een lijstje met wat elementen. Eens kijken wat dit\n  allemaal gaat doen.</p>\n  \n<ul>\n  <li>Item1</li>\n  <li>Item2</li>\n</ul>\n\n<p>En wat doet deze paragraaf? Die hopelijk langer is dan een enkele\n  regel zodat de pagina niet zo leeg lijkt.</p>\n  \n<h2>Kopje met meteen een lijst</h2>\n<ul>\n  <li>Item1</li>\n  <li>Item2</li>\n</ul>\n\n<p>En dan hebben we hier een leuke paragraaf die eigenlijk helemaal\n   nergens over gaat. Maar slechts hier staat om te laten zien hoe een\n   <a href="#">link</a> eruit ziet binnen een tekst.</p>\n      \n<h2>Lists</h2>\n<h3>Two lists after eachother</h3>\n<ul>\n  <li>List 1</li>\n</ul>\n<ul>\n  <li>List 2</li>\n</ul>\n\n<h3>Two ol lists after each other</h3>\n<ol>\n  <li>List 1</li>\n</ol>\n<ol>\n  <li>List 2</li>\n</ol>\n\n<h3>First ol, then ul</h3>\n\n<ol>\n  <li>Foo</li>\n</ol>\n<ul>\n  <li>Bar</li>\n</ul>\n\n<h3>First ul, then ol</h3>\n\n<ul>\n  <li>Bar</li>\n</ul>\n<ol>\n  <li>Foo</li>\n</ol>\n\n<h2>Tables</h2>\n\n<table>\n  <tr>\n    <td>A</td>\n    <td>B</td>\n  </tr>\n  <tr>\n    <td>C</td>\n    <td>D</td>\n  </tr>\n</table>\n\n<p>This is the paragraph after a table...</p>',
    },
    "1059096": {
        "gistfile1.bash": '# /usr/bin/env /bin/sh\nstart.applescript "`pwd`" $@',
        "start.applescript": '#! /usr/bin/env /usr/bin/osascript\non run argv\n  \n  tell application "Terminal"\n    set AppleScript\'s text item delimiters to " "\n\n    set cwd to first item of argv\n    set l to rest of argv\n\n    do script "cd " & cwd & " && " & (l as text)\n    activate\n  end tell\n\nend run\n',
    },
    "1902090": {
        "install.rb": '#! /usr/bin/ruby\nrequire \'readline\'\n\ndef command?(command)\n  system("which #{ command} > /dev/null 2>&1")\nend\n\ndef library?(lib) \n  system("ghc-pkg list | grep #{lib} > /dev/null 2>&1")\nend\n\ndef request(question, onYes)\n  stty_save = `stty -g`.chomp\n  trap(\'INT\') { system(\'stty\', stty_save); exit 1 }\n\n  line = Readline.readline(question, true)\n\n  if line.downcase == "y" || line == ""\n    return onYes.call\n  end\n\n  return false\n  \nend\n\ndef colorize(text, color_code)\n  "#{color_code}#{text}\\033[0m"\nend\n\ndef red(text); colorize(text, "\\033[31m"); end\ndef green(text); colorize(text, "\\033[32m"); end\ndef yellow(text); colorize(text, "\\033[33m"); end\ndef blue(text); colorize(text, "\\033[34m"); end\n\nclass Message\n  def self.ok(message)\n    puts("[  #{green("CHECK")}  ] #{message}")\n  end \n\n  def self.error(message)\n    puts("[  #{red("FAILED")} ] #{message}")\n  end\n  \n  def self.info(message)\n    puts("[  #{yellow("INFO")}   ] #{message}")\n  end\n  \n  def self.question(message)\n    puts("[  #{blue("REQ")}    ] #{message}")\n  end\nend\n\n# Message.ok ""\n# Message.error ""\n# Message.info ""\n# Message.question ""\n# exit\n\ndef check?(condition, message) \n  if condition\n    Message.ok(message)\n  else\n    Message.error(message)\n  end\n  condition\nend\n\nif ENV[\'UHC\'] == ""\n  Message.error "Please provide the $UHC variable before running this script!"\n  exit 1\nend\n\n\nrequiredCommands = ["git", "ghc", "cabal", "uuagc"]\nrequiredLibs     = ["uulib", "fgl", "network", "binary"]\n\ncabalInstall = "cabal install --user "\n\nrequiredCommands.each { | command | \n  if !check?((command? command), command)\n    exit 1\n  end\n}\n\nrequiredLibs.each {|lib|\n  command = cabalInstall + lib\n  if !check?((library? lib), lib)\n    if !request(blue("Install #{lib}? [Y/n]: "), lambda { system command })\n      exit 1\n    end\n    !check?((library? lib), lib)\n  end\n}\n\n\n# Check Cabal packages\n\n# Check github projects\nprojects = {"jcu"        => "https://github.com/UU-ComputerScience/JCU",\n            "uhcjs"      => "https://github.com/UU-ComputerScience/uhc-js",\n            "NanoProlog" => "https://github.com/UU-ComputerScience/NanoProlog"}\n\n# Check whether dir exists and if not pull it\nprojects.each { |key, path|\n  Message.info("Checking " + key)\n  if File.directory? key\n    command = "cd " + key + " && " + "git pull"\n  \n  else\n    command = "git clone " + path + " " + key\n  end\n  \n  if !system command\n    exit 1\n  end\n}        \n\nif !File.directory? "uu-tc-2009.2.2"\n  puts "Downloading uu-tc" \n  `curl -L -O http://www.cs.uu.nl/docs/vakken/b3tc/dist/uu-tc-2009.2.2.tar.gz`\n  puts "Extracting it"\n  `tar xvzf uu-tc-2009.2.2.tar.gz`\nelse\n  Message.ok "uu-tc library already downloaded."\nend\n           \n# Check snap version\n# \n\n# CD into JCU and try install\nMessage.info "Installing JCU"\nsystem "cd jcu && cabal install --user"\n\nMessage.info "Please make sure that you edit the connection string! And that PostgreSQL is installed!"\n\nif ENV[\'EDITOR\'] != \'\' \n  Message.info "- Opening the config file"\n  system "$EDITOR jcu/config/connection_string.conf"\nend\n\ncwd = Dir.pwd\n\n# CD into the Clientside app and build\nMessage.info "Please provide the $UHC variable yourself: export UHC=..."\nMessage.info "You can build UHC-JS with the following command:"\nMessage.info "    make uhc && make 101/ehclib EHC_VARIANT_TARGET=js"\nMessage.info "Building the client-side "\nsystem "export UHC_JSCRIPT=#{cwd}/uhcjs/uhc-js/src;export UHC_NANOPROLOG=#{cwd}/NanoProlog/src;export UHC_UU_TC=#{cwd}/uu-tc-2009.2.2/src;cd jcu/resources/static/hjs && make"\n\nMessage.info "If all went okay you can now cd into the #{yellow "`jcu`"} directory, execute the command jcu and browse to http://localhost:8000."'
    },
}


def parse_front_matter(text: str) -> tuple[dict[str, object], str]:
    if not text.startswith("---"):
        raise ValueError("missing YAML front matter")
    _, yaml, body = text.split("---", 2)
    result: dict[str, object] = {}
    list_key: str | None = None
    for line_number, line in enumerate(yaml.splitlines(), start=2):
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        item = re.match(r"^\s*-\s+(.*?)\s*$", line)
        if item and list_key:
            if result[list_key] is None:
                result[list_key] = []
            if not isinstance(result[list_key], list):
                raise ValueError(
                    f"list value expected at front matter line {line_number}"
                )
            result[list_key].append(parse_scalar(item.group(1)))
            continue
        if line.lstrip().startswith("-"):
            raise ValueError(
                f"unsupported list entry at front matter line {line_number}: {line}"
            )
        match = re.match(r"^([\w-]+):\s*(.*?)\s*$", line)
        if not match:
            raise ValueError(f"unsupported front matter line {line_number}: {line}")
        key, value = match.groups()
        list_key = None
        if value == "":
            result[key] = None
            list_key = key
        elif value == "[]":
            result[key] = []
        elif value.startswith("[") and value.endswith("]"):
            try:
                parsed = json.loads(value)
            except json.JSONDecodeError as error:
                raise ValueError(
                    f"unsupported inline list at front matter line {line_number}: {value}"
                ) from error
            if not isinstance(parsed, list):
                raise ValueError(f"expected list at front matter line {line_number}")
            result[key] = parsed
        else:
            result[key] = parse_scalar(value)
    return result, body.lstrip("\r\n")


def parse_scalar(value: str) -> object:
    if value in ("", "null", "Null", "NULL", "~"):
        return None
    if value.lower() in ("true", "false"):
        return value.lower() == "true"
    if value.startswith('"'):
        try:
            return json.loads(value)
        except json.JSONDecodeError as error:
            raise ValueError(f"unsupported double-quoted scalar: {value}") from error
    if value.startswith("'") and value.endswith("'"):
        return value[1:-1].replace("''", "'")
    if value.startswith(("{", "[", "&", "*", "|", ">")):
        raise ValueError(f"unsupported YAML scalar: {value}")
    return value


def post_slug(source: Path) -> str:
    return re.sub(r"^\d{4}-\d{2}-\d{2}-", "", source.stem).lower()


def source_date(source: Path, metadata: dict[str, object]) -> datetime:
    raw = metadata.get("date")
    if raw:
        parsed = datetime.strptime(str(raw), "%Y-%m-%d %H:%M").replace(tzinfo=TZ)
    else:
        match = re.match(r"^(\d{4}-\d{2}-\d{2})-", source.name)
        if not match:
            raise ValueError(f"cannot infer date from {source}")
        parsed = datetime.combine(date.fromisoformat(match.group(1)), time.min)
    return parsed.replace(tzinfo=TZ)


@dataclass(frozen=True)
class PostRoute:
    canonical_url: str
    aliases: list[str]


def post_route(source: Path, metadata: dict[str, object]) -> PostRoute:
    slug = post_slug(source)
    dated = source_date(source, metadata).strftime("/%Y/%m/%d/") + slug + "/"
    canonical = LEGACY_POST_URLS.get(source.name)
    if canonical is None:
        if metadata.get("published") is not False:
            raise ValueError(f"missing verified live URL for {source}")
        canonical = dated
    return PostRoute(
        canonical_url=canonical,
        aliases=[url for url in (f"/posts/{slug}/", dated) if url != canonical],
    )


def rewrite_post_links(body: str, source: Path) -> str:
    known_urls = {
        post_slug(Path(filename)): url for filename, url in LEGACY_POST_URLS.items()
    }
    known_urls.update({url.rsplit("/", 2)[1]: url for url in LEGACY_POST_URLS.values()})

    def post_link(match: re.Match[str]) -> str:
        slug = match.group("slug").lower()
        if slug not in known_urls:
            raise ValueError(f"unknown linked post {slug} in {source}")
        return f"{match.group('prefix')}{known_urls[slug]}"

    return re.sub(
        r"(?P<prefix>\]\(|\bhref=[\"'])(?:https?://alessandrovermeulen\.me)?/?(?:(?:[^/?#\s\"'()]+/)*\d{4}/\d{2}/\d{2}/|posts/)(?P<slug>[^/?#\s\"'()]+)/?",
        post_link,
        body,
    )


def language(filename: str) -> str:
    return {
        "css": "css",
        "html": "html",
        "bash": "sh",
        "applescript": "applescript",
        "rb": "ruby",
        "pub": "text",
    }.get(filename.rsplit(".", 1)[-1], "text")


def convert_body(body: str, source: Path) -> str:
    def gist(match: re.Match[str]) -> str:
        gist_id = match.group(1)
        if gist_id not in GISTS:
            raise ValueError(f"unresolved gist {gist_id} in {source}")
        result = []
        for filename, content in GISTS[gist_id].items():
            result.append(
                f"`{filename}`\n\n```{language(filename)}\n{content.rstrip()}\n```"
            )
        return "\n\n".join(result)

    body = re.sub(r"\{%\s*gist\s+(\d+)\s*%\}", gist, body)
    body = re.sub(r"\{\{:\s*toc\s*\}\}", "", body, flags=re.IGNORECASE)
    body = re.sub(r"\{:\s*target=[\"']_blank[\"']\s*\}", "", body)
    body = rewrite_post_links(body, source)
    body = body.replace(
        "http://alessandrovermeulen.me/projects/lhs2texhl/",
        "https://github.com/spockz/lhs2texhl",
    )
    if source.stem == "2013-05-19-why-you-should-switch-to-declarative-programming":
        body = body.replace(
            "[Lloyd, J.W., Practical Advantages of Declarative Programming](ftp://clip.dia.fi.upm.es/pub/papers/PARFORCE/second_review/D.WP3.1.M2.3.ps.Z)",
            '<a href="ftp://clip.dia.fi.upm.es/pub/papers/PARFORCE/second_review/D.WP3.1.M2.3.ps.Z">Lloyd, J.W., Practical Advantages of Declarative Programming</a>',
        )
    caption = re.compile(
        r'\[caption[^\]]*caption="([^"]*)"[^\]]*\](.*?)\[/caption\]', re.DOTALL
    )

    def convert_caption(match: re.Match[str]) -> str:
        caption_text, inner = match.groups()
        if "wp-content/" in inner:
            return f"<p><small>Archived image unavailable: {caption_text}.</small></p>"
        return f"<figure>\n{inner.strip()}\n<figcaption>{caption_text}</figcaption>\n</figure>"

    body = caption.sub(convert_caption, body)
    wp_image = re.compile(
        r'<a\b[^>]*href=["\'][^"\']*wp-content/[^"\']*["\'][^>]*>.*?</a>', re.DOTALL
    )

    def unavailable_image(match: re.Match[str]) -> str:
        if "<img" not in match.group(0).lower():
            return match.group(0)
        alt = re.search(r'alt=["\']([^"\']+)["\']', match.group(0))
        description = alt.group(1) if alt else "archived screenshot"
        return f"<p><small>Archived image unavailable: {description}.</small></p>"

    body = wp_image.sub(unavailable_image, body)
    readme_link = re.compile(
        r'<a\b[^>]*href=["\']https?://alessandrovermeulen\.me/wp-content/2009/09/README\.html["\'][^>]*>(.*?)</a>',
        re.DOTALL,
    )
    body = readme_link.sub(
        r"\1 <small>(download unavailable in the archive)</small>",
        body,
    )
    local_images = {
        "http://farm8.staticflickr.com/7025/6767517435_b075d1c686_z.jpg": "/images/archive/6767517435_b075d1c686_z.jpg",
        "http://farm8.staticflickr.com/7155/6767517241_eacd8ec0ed_z.jpg": "/images/archive/6767517241_eacd8ec0ed_z.jpg",
        "http://farm8.staticflickr.com/7313/9243796534_60bb926e44_o.png": "/images/archive/9243796534_60bb926e44_o.png",
        "http://farm4.staticflickr.com/3804/9241070475_9d48236aa7_o.png": "/images/archive/9241070475_9d48236aa7_o.png",
    }
    for original, local in local_images.items():
        body = body.replace(original, local)
    dimensions = {
        "6767517435_b075d1c686_z.jpg": (639, 413),
        "6767517241_eacd8ec0ed_z.jpg": (639, 413),
        "9243796534_60bb926e44_o.png": (563, 155),
        "9241070475_9d48236aa7_o.png": (430, 410),
    }

    def add_dimensions(match: re.Match[str]) -> str:
        image = match.group(0)
        src = re.search(r'src="([^"]+)"', image)
        filename = Path(src.group(1)).name if src else ""
        if filename not in dimensions:
            return image
        width, height = dimensions[filename]
        for name, value in (("width", width), ("height", height)):
            attr = re.compile(rf'\b{name}="\d+"')
            if attr.search(image):
                image = attr.sub(f'{name}="{value}"', image)
            else:
                image = image[:-1].rstrip() + f' {name}="{value}">'
        return image

    body = re.sub(r"<img\b[^>]*>", add_dimensions, body, flags=re.IGNORECASE)
    body = re.sub(
        r"\[php\](.*?)\[/php\]",
        lambda match: f"```php\n{match.group(1).strip()}\n```",
        body,
        flags=re.DOTALL | re.IGNORECASE,
    )
    body = re.sub(
        r"<img\b(?![^>]*\bloading=)([^>]*)>",
        r'<img\1 loading="lazy">',
        body,
        flags=re.IGNORECASE,
    )
    return body.rstrip() + "\n"


def migrate_posts() -> list[dict[str, object]]:
    sources = sorted((ROOT / "_posts").glob("*"))
    if len(sources) != 36:
        raise RuntimeError(f"expected 36 posts, found {len(sources)}")
    destination = DEST / "content" / "posts"
    destination.mkdir(parents=True, exist_ok=True)
    routes = []
    for source in sources:
        metadata, body = parse_front_matter(source.read_text(encoding="utf-8-sig"))
        title = metadata.get("title")
        if title is None:
            raise ValueError(f"missing title in {source}")
        if not isinstance(title, str):
            raise TypeError(f"title must be a string in {source}")
        timestamp = source_date(source, metadata)
        slug = post_slug(source)
        route = post_route(source, metadata)
        draft = metadata.get("published") is False
        front_matter: dict[str, object] = {
            "title": title,
            "date": timestamp.isoformat(),
            "slug": slug,
            "url": route.canonical_url,
            "aliases": route.aliases,
            "draft": draft,
        }
        if "tags" in metadata:
            front_matter["tags"] = metadata["tags"]
        if "categories" in metadata:
            source_categories = metadata["categories"]
            front_matter["categories"] = (
                [source_categories]
                if isinstance(source_categories, str)
                else source_categories
            )
        elif "category" in metadata:
            category = metadata["category"]
            front_matter["categories"] = (
                [category] if isinstance(category, str) else category
            )
        if source.name == "2014-12-04-orchestration-support-announced-on-dockercon.md":
            front_matter["summary"] = (
                "Docker's new tools announced at DockerCon address provisioning, clustering, and application orchestration."
            )
        output = "---\n# WHY: Preserve this archived article's metadata and exact historical URL.\n"
        output += json.dumps(front_matter, ensure_ascii=False, indent=2) + "\n---\n\n"
        output += convert_body(body, source)
        target = destination / f"{slug}.md"
        target.write_text(output, encoding="utf-8")
        routes.append(
            {
                "source": source.relative_to(ROOT).as_posix(),
                "destination": target.relative_to(DEST).as_posix(),
                "canonical_url": route.canonical_url,
                "aliases": route.aliases,
                "draft": draft,
            }
        )
    return routes


def migrate_about() -> None:
    _, body = parse_front_matter((ROOT / "about.md").read_text(encoding="utf-8-sig"))
    body = body.replace(
        "[photography](/photography)", "[photography](/category/photography/)"
    )
    body = (
        "<p><small>This biography is preserved from the 2016 site and may be out of date.</small></p>\n\n"
        + body
    )
    front_matter = {"title": "About", "url": "/about/", "draft": False}
    output = "---\n# WHY: Mark this historical biography as an archive while keeping its existing page route.\n"
    output += json.dumps(front_matter, ensure_ascii=False, indent=2) + "\n---\n\n"
    output += convert_body(body, ROOT / "about.md")
    (DEST / "content" / "about.md").write_text(output, encoding="utf-8")


def copy_downloads() -> int:
    sources = sorted((ROOT / "downloads").glob("*"))
    destination = DEST / "static" / "downloads"
    destination.mkdir(parents=True, exist_ok=True)
    for source in sources:
        if not source.is_file():
            raise ValueError(f"unexpected downloads entry: {source}")
        shutil.copy2(source, destination / source.name)
    return len(sources)


def write_route_manifest(routes: list[dict[str, object]]) -> None:
    routes.append(
        {
            "source": "about.md",
            "destination": "content/about.md",
            "canonical_url": "/about/",
            "draft": False,
        }
    )
    unresolved = [
        {
            "source": "_posts/2009-04-17-doctrine-meets-codeigniter.markdown",
            "path": "http://alessandrovermeulen.me/wp-content/2009/09/README.html",
            "status": "http-404",
            "resolution": "unavailable notice shown; broken hyperlink removed",
        },
        {
            "source": "_posts/2010-09-10-how-to-setup-afp-on-debian-lenny.markdown",
            "path": "http://alessandrovermeulen.me/wp-content/2010/09/Screen-shot-2010-09-10-at-12.24.28.png",
            "status": "http-404",
            "resolution": "unavailable notice replaces broken image",
        },
        {
            "source": "_posts/2010-09-10-how-to-setup-afp-on-debian-lenny.markdown",
            "path": "http://alessandrovermeulen.me/wp-content/2010/09/Screen-shot-2010-09-10-at-12.24.28-300x225.png",
            "status": "http-404",
            "resolution": "unavailable notice replaces broken image",
        },
        {
            "source": "_posts/2010-10-23-lhs2tex-hl-released.markdown",
            "path": "http://alessandrovermeulen.me/wp-content/2010/10/Screen-shot-2010-10-23-at-18.41.00.png",
            "status": "http-404",
            "resolution": "unavailable notice replaces broken image",
        },
        {
            "source": "_posts/2010-10-23-lhs2tex-hl-released.markdown",
            "path": "http://alessandrovermeulen.me/wp-content/2010/10/Screen-shot-2010-10-23-at-18.41.00-300x255.png",
            "status": "http-404",
            "resolution": "unavailable notice replaces broken image",
        },
    ]
    recovered = [
        {
            "source": "_posts/2012-01-26-getting-rid-of-javascript-with-haskell.md",
            "path": "http://farm8.staticflickr.com/7025/6767517435_b075d1c686_z.jpg",
            "destination": "static/images/archive/6767517435_b075d1c686_z.jpg",
            "recovery_url": "https://farm8.staticflickr.com/7025/6767517435_b075d1c686_z.jpg",
            "http_status": 200,
            "content_type": "image/jpeg",
            "dimensions": "639x413",
            "verified_image_content": True,
        },
        {
            "source": "_posts/2012-01-26-getting-rid-of-javascript-with-haskell.md",
            "path": "http://farm8.staticflickr.com/7155/6767517241_eacd8ec0ed_z.jpg",
            "destination": "static/images/archive/6767517241_eacd8ec0ed_z.jpg",
            "recovery_url": "https://farm8.staticflickr.com/7155/6767517241_eacd8ec0ed_z.jpg",
            "http_status": 200,
            "content_type": "image/jpeg",
            "dimensions": "639x413",
            "verified_image_content": True,
        },
        {
            "source": "_posts/2013-07-08-combining-graphviz-dot-and-tikz-with-dot2tex.md",
            "path": "http://farm8.staticflickr.com/7313/9243796534_60bb926e44_o.png",
            "destination": "static/images/archive/9243796534_60bb926e44_o.png",
            "recovery_url": "https://web.archive.org/web/20160528095748id_/http://farm8.staticflickr.com/7313/9243796534_60bb926e44_o.png",
            "recovery_http_status": 200,
            "recovery_content_type": "image/png",
            "original_host_http_status": 429,
            "dimensions": "563x155",
            "verified_image_content": True,
        },
        {
            "source": "_posts/2013-07-08-combining-graphviz-dot-and-tikz-with-dot2tex.md",
            "path": "http://farm4.staticflickr.com/3804/9241070475_9d48236aa7_o.png",
            "destination": "static/images/archive/9241070475_9d48236aa7_o.png",
            "recovery_url": "https://web.archive.org/web/20160528051606id_/http://farm4.staticflickr.com/3804/9241070475_9d48236aa7_o.png",
            "recovery_http_status": 200,
            "recovery_content_type": "image/png",
            "original_host_http_status": 429,
            "dimensions": "430x410",
            "verified_image_content": True,
        },
    ]
    manifest = {
        "routes": routes,
        "unresolved_references": unresolved,
        "recovered_references": recovered,
    }
    (DEST / "migration.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )


def main() -> None:
    routes = migrate_posts()
    migrate_about()
    downloads = copy_downloads()
    post_count = len(routes)
    write_route_manifest(routes)
    print(f"Migrated {post_count} posts, About, and {downloads} downloads.")


if __name__ == "__main__":
    main()
