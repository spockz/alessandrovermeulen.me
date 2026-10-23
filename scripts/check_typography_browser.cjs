// Check rendered glyphs and line fragments because final font measurements hide loading flashes and CSS rewrapping.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { webkit } = require(process.env.PLAYWRIGHT_MODULE || "playwright");

const baseURL = process.argv[2] || "http://localhost:1313/";
const output = path.resolve(".cache/review");
const routes = [
  "/2015/08/31/finaglecon/",
  "/2013/07/13/the-difference-between-shallow-and-deep-embedding/",
  "/2014/07/15/notes-on-the-advanced-akka-course/",
  "/posts/",
  "/posts/page/2/",
  "/",
  "/tags/scala/",
  "/posts/a-fresh-leaf/",
  "/about/",
];

async function inspect(page) {
  return page.evaluate(() => {
    const failures = [];
    // WebKit can apply fractional word spacing differently at a text-fragment boundary.
    const widthTolerance = 1;
    let column = null;
    const header = document.querySelector("#single_header");
    if (header) {
      const article = header.closest("article");
      const content = article.querySelector(".article-content");
      const box = content.getBoundingClientRect();
      const main = document
        .querySelector("#main-content")
        .getBoundingClientRect();
      const sidebar = article.querySelector(".toc-right");
      const centered = !sidebar || sidebar.getClientRects().length === 0;
      const offset = box.left + box.width / 2 - (main.left + main.width / 2);
      column = { width: box.width, centered, offset };
      if (centered) {
        if (Math.abs(offset) > widthTolerance)
          failures.push(`Article column is off-center by ${offset}px`);
        const expectedWidth = Math.min(
          main.width,
          36 * parseFloat(getComputedStyle(document.documentElement).fontSize),
        );
        if (Math.abs(box.width - expectedWidth) > widthTolerance)
          failures.push(`Article column width changed: ${box.width}px`);
        for (const element of [
          header,
          article.querySelector(":scope > footer"),
        ]) {
          if (!element) continue;
          const rect = element.getBoundingClientRect();
          if (
            Math.abs(rect.left - box.left) > widthTolerance ||
            Math.abs(rect.width - box.width) > widthTolerance
          )
            failures.push(`${element.tagName} does not align with its article`);
        }
      }
    }
    const main = document.querySelector("#main-content");
    const list = main.querySelector(
      ":scope > section:has(> .article-link--simple), :scope > section > section:has(> .article-link--simple)",
    );
    if (list) {
      const box = list.getBoundingClientRect();
      const mainBox = main.getBoundingClientRect();
      const offset =
        box.left + box.width / 2 - (mainBox.left + mainBox.width / 2);
      const expectedWidth = Math.min(
        mainBox.width,
        36 * parseFloat(getComputedStyle(document.documentElement).fontSize),
      );
      column = { width: box.width, centered: true, offset };
      if (Math.abs(offset) > widthTolerance)
        failures.push(`Post listing is off-center by ${offset}px`);
      if (Math.abs(box.width - expectedWidth) > widthTolerance)
        failures.push(`Post listing width changed: ${box.width}px`);
      const alignmentTargets =
        list.parentElement === main
          ? main.querySelectorAll(":scope > :is(header, section, ul)")
          : [list.parentElement];
      for (const element of alignmentTargets) {
        const rect = element.getBoundingClientRect();
        if (
          Math.abs(rect.left - box.left) > widthTolerance ||
          Math.abs(rect.width - box.width) > widthTolerance
        )
          failures.push(
            `${element.tagName} does not align with its post listing`,
          );
      }
      if (list.parentElement !== main) {
        const profile = main
          .querySelector(":scope > article")
          .getBoundingClientRect();
        if (
          Math.abs(profile.left - mainBox.left) > widthTolerance ||
          Math.abs(profile.width - mainBox.width) > widthTolerance
        )
          failures.push(
            "Centering recent articles must preserve the home profile layout",
          );
      }
      for (const summary of list.querySelectorAll(".typography-summary")) {
        const rect = summary.getBoundingClientRect();
        if (
          Math.abs(rect.left - box.left) > widthTolerance ||
          Math.abs(rect.width - expectedWidth) > widthTolerance
        )
          failures.push(
            "Post summary must retain the full centered text width",
          );
      }
    }
    const active = [...document.querySelectorAll(".planned-paragraph")];
    for (const [index, paragraph] of active.entries()) {
      const style = getComputedStyle(paragraph);
      if (
        style.whiteSpace !== "nowrap" ||
        style.textAlign !== "start" ||
        style.hyphens !== "none"
      )
        failures.push(
          `Block ${index}: native wrapping or justification overrides the plan`,
        );
      const box = paragraph.getBoundingClientRect();
      const left =
        box.left +
        parseFloat(style.paddingLeft) +
        parseFloat(style.borderLeftWidth);
      const right =
        box.right -
        parseFloat(style.paddingRight) -
        parseFloat(style.borderRightWidth);
      const own = (node) => node.closest("[data-plan-width]") === paragraph;
      for (const first of [
        ...paragraph.querySelectorAll("[data-natural]"),
      ].filter(own)) {
        const pieces = [
          ...paragraph.querySelectorAll(`[data-line="${first.dataset.line}"]`),
        ].filter(own);
        const rects = [];
        for (const piece of pieces) {
          let extent = piece;
          for (
            let parent = piece.parentElement;
            parent !== paragraph;
            parent = parent.parentElement
          ) {
            if (!parent) throw new Error("Line piece escaped its paragraph");
            if (parent.hasAttribute("data-plan-atom")) extent = parent;
          }
          const fragments = [...extent.getClientRects()];
          rects.push(...fragments);
        }
        if (!rects.length) {
          failures.push(
            `Block ${index}, line ${first.dataset.line}: no rendered text`,
          );
          continue;
        }
        if (
          Math.max(...rects.map((r) => r.top)) >=
          Math.min(...rects.map((r) => r.bottom))
        )
          failures.push(
            `Block ${index}, line ${first.dataset.line}: pieces occupy different visual lines`,
          );
        const actualLeft = Math.min(...rects.map((r) => r.left));
        const actualRight = Math.max(...rects.map((r) => r.right));
        if (
          actualLeft < left - widthTolerance ||
          actualRight > right + widthTolerance
        )
          failures.push(
            `Block ${index}, line ${first.dataset.line}: rendered text exceeds its column by ${actualRight - right}px`,
          );
        if (
          !first.hasAttribute("data-final") &&
          Math.abs(actualRight - actualLeft - (right - left)) > widthTolerance
        )
          failures.push(
            `Block ${index}, line ${first.dataset.line}: rendered width differs by ${actualRight - actualLeft - (right - left)}px`,
          );
      }
    }
    return {
      active: active.length,
      failures,
      column,
      overflow: document.documentElement.scrollWidth > innerWidth,
      links: [
        ...document.querySelectorAll(
          ".article-content a, .article-link--simple a",
        ),
      ].map((a) => ({
        id: a.id,
        href: a.getAttribute("href"),
      })),
      code: [...document.querySelectorAll("pre")].map((pre) => pre.textContent),
    };
  });
}

async function checkProjectBadges(page) {
  await page.emulateMedia({ media: "screen" });
  const response = await page.goto(new URL("/projects/", baseURL).href, {
    waitUntil: "networkidle",
  });
  assert.equal(response.status(), 200);
  await page.evaluate(() => document.fonts.ready);
  const report = [];
  for (const colorScheme of ["light", "dark"]) {
    await page.emulateMedia({ colorScheme });
    await page.waitForFunction(
      (scheme) =>
        document.documentElement.classList.contains("dark") ===
        (scheme === "dark"),
      colorScheme,
    );
    // A theme change can cancel a previous transition before its promise resolves.
    await page.waitForFunction(() =>
      document
        .getAnimations()
        .every(
          (animation) =>
            !animation.pending && animation.playState !== "running",
        ),
    );
    for (const width of [320, 834, 1280]) {
      await page.setViewportSize({ width, height: 874 });
      const result = await page.evaluate(() => {
        const badges = [...document.querySelectorAll(".project-badge")];
        return {
          colorScheme: document.documentElement.classList.contains("dark")
            ? "dark"
            : "light",
          cards: document.querySelectorAll(
            ".project-showcase .project-card-title > h3",
          ).length,
          overflow: document.documentElement.scrollWidth > innerWidth,
          badges: badges.map((svg) => {
            const box = svg.getBoundingClientRect();
            const symbol = svg.querySelector("g");
            const glyph = symbol.getBBox();
            const card = svg.closest("article");
            const metadata = card.querySelector(".project-meta");
            const footer = card.querySelector(".project-card-footer");
            const link = footer.querySelector("a").getBoundingClientRect();
            const languages = card.querySelector(
              ".project-card-title .project-languages",
            );
            const statusBox = metadata.getBoundingClientRect();
            const labels = [...languages.querySelectorAll(".project-label")];
            return {
              id: svg.dataset.projectBadge,
              width: box.width,
              height: box.height,
              color: getComputedStyle(svg).color,
              symbol: symbol.innerHTML,
              visible: glyph.width > 0 && glyph.height > 0,
              fits:
                glyph.x >= 12 &&
                glyph.y >= 12 &&
                glyph.x + glyph.width <= 84 &&
                glyph.y + glyph.height <= 84,
              decorative: svg.getAttribute("aria-hidden") === "true",
              metadata: metadata.dataset.projectMeta,
              languages: labels.map((label) => label.textContent.trim()),
              pills: labels.every((label) => {
                const style = getComputedStyle(label);
                return (
                  parseFloat(style.borderRadius) >=
                    label.getBoundingClientRect().height / 2 &&
                  style.borderTopStyle === "solid" &&
                  parseFloat(style.borderTopWidth) >= 1 &&
                  style.backgroundColor !== "rgba(0, 0, 0, 0)"
                );
              }),
              sameRow:
                Math.abs(
                  statusBox.top +
                    statusBox.height / 2 -
                    (link.top + link.height / 2),
                ) < 1,
              rightAligned:
                Math.abs(
                  statusBox.right - footer.getBoundingClientRect().right,
                ) < 1,
              statuses: metadata.querySelectorAll(
                ".project-ci-badge, .project-ci-unconfigured",
              ).length,
            };
          }),
        };
      });
      assert.equal(
        result.colorScheme,
        colorScheme,
        "Check the rendered theme, not a pending preference change",
      );
      assert.ok(result.cards > 0);
      assert.equal(
        result.overflow,
        false,
        `Project cards overflow at ${width}px`,
      );
      assert.equal(
        result.badges.length,
        result.cards,
        "Each card needs an emblem",
      );
      assert.equal(
        new Set(result.badges.map((badge) => badge.id)).size,
        result.cards,
      );
      assert.equal(
        new Set(result.badges.map((badge) => badge.symbol)).size,
        result.cards,
        "Project symbols must be distinct",
      );
      assert.equal(
        new Set(result.badges.map((badge) => badge.color)).size,
        1,
        "Emblems must share the theme palette",
      );
      for (const badge of result.badges) {
        assert.equal(badge.width, 72, badge.id);
        assert.equal(badge.height, 72, badge.id);
        assert.ok(badge.visible && badge.fits && badge.decorative, badge.id);
        assert.equal(
          badge.metadata,
          badge.id,
          "Metadata must match its project",
        );
        assert.ok(badge.languages.length > 0 && badge.languages.every(Boolean));
        assert.ok(badge.pills, `Languages must be pills: ${badge.id}`);
        assert.ok(
          badge.sameRow,
          `CI status must share the link row: ${badge.id}`,
        );
        assert.ok(
          badge.rightAligned,
          `CI status must align right: ${badge.id}`,
        );
        assert.equal(badge.statuses, 1, "Each card needs one honest CI status");
      }
      report.push({
        colorScheme,
        width,
        cards: result.cards,
        color: result.badges[0].color,
      });
      if (width === 1280)
        await page.screenshot({
          path: path.join(output, `project-badges-${colorScheme}.png`),
          fullPage: true,
          scale: "css",
        });
    }
  }
  await fs.writeFile(
    path.join(output, "webkit-project-badges.json"),
    JSON.stringify(report, null, 2) + "\n",
  );
}

async function hasPaintedText(paragraph, waitForFonts = true) {
  const previousFontWait = process.env.PW_TEST_SCREENSHOT_NO_FONTS_READY;
  try {
    // Playwright's font wait would hide the loading transition being checked.
    if (!waitForFonts) process.env.PW_TEST_SCREENSHOT_NO_FONTS_READY = "1";
    const visible = await paragraph.screenshot();
    const empty = await paragraph.screenshot({
      mask: [paragraph],
      maskColor: "#ffffff",
    });
    return !visible.equals(empty);
  } finally {
    if (previousFontWait === undefined)
      delete process.env.PW_TEST_SCREENSHOT_NO_FONTS_READY;
    else process.env.PW_TEST_SCREENSHOT_NO_FONTS_READY = previousFontWait;
  }
}

function articleFontLoaded() {
  return [...document.fonts].some(
    (face) =>
      face.family.toLowerCase().includes("article sans") &&
      face.weight === "400" &&
      face.style === "normal" &&
      face.status === "loaded",
  );
}

async function checkColdFontLoading(browser) {
  const report = [];
  for (const route of ["/about/", "/"]) {
    for (const width of [390, 1280]) {
      const context = await browser.newContext({
        viewport: { width, height: 874 },
        colorScheme: "light",
      });
      let releaseFonts;
      const fontGate = new Promise((resolve) => {
        releaseFonts = resolve;
      });
      try {
        const page = await context.newPage();
        let fontRequests = 0;
        await page.route("**/fonts/**", async (request) => {
          if (
            request.request().url().endsWith("/article-sans-v3052-latin.woff2")
          )
            fontRequests += 1;
          await fontGate;
          await request.continue();
        });
        const response = await page.goto(new URL(route, baseURL).href, {
          waitUntil: "domcontentloaded",
        });
        assert.equal(response.status(), 200, route);
        const paragraph = page
          .locator(".article-content p, .typography-summary p")
          .first();
        await page.waitForFunction(() => {
          const node = document.querySelector(
            ".article-content p, .typography-summary p",
          );
          return (
            node &&
            getComputedStyle(node)
              .fontFamily.toLowerCase()
              .includes("article sans") &&
            [...document.fonts].some((face) =>
              face.family.toLowerCase().includes("article sans"),
            )
          );
        });
        assert.equal(
          await page
            .locator(
              'head link[rel="preload"][as="font"][href="/fonts/article-sans-v3052-latin.woff2"][crossorigin]',
            )
            .count(),
          1,
          "The article font must be discovered in the page head",
        );
        const beforeSize = await paragraph.evaluate(
          (node) => getComputedStyle(node).fontSize,
        );
        assert.equal(beforeSize, "16px");
        assert.equal(
          await page.evaluate(articleFontLoaded),
          false,
          "Cold-load capture must happen before the article font is available",
        );
        assert.equal(
          await hasPaintedText(paragraph, false),
          false,
          `${route} at ${width}px must not paint larger fallback glyphs before the article font loads`,
        );
        releaseFonts();
        await page.evaluate(() => document.fonts.ready);
        assert.equal(
          await page.evaluate(articleFontLoaded),
          true,
          "Article text must use its downloaded font after loading",
        );
        await paragraph.waitFor({ state: "visible" });
        await page.waitForFunction(() =>
          document.querySelector(".planned-paragraph"),
        );
        assert.equal(
          await hasPaintedText(paragraph),
          true,
          "Loaded article text must be visible",
        );
        const afterSize = await paragraph.evaluate(
          (node) => getComputedStyle(node).fontSize,
        );
        assert.equal(afterSize, beforeSize);
        assert.equal(
          fontRequests,
          1,
          "Font preload must reuse the CSS request",
        );
        report.push({
          mode: "delayed font",
          route,
          width,
          beforeSize,
          afterSize,
          fontRequests,
        });
      } finally {
        releaseFonts();
        await context.unrouteAll({ behavior: "wait" });
        await context.close();
      }
    }
  }
  for (const mode of ["failed font", "disabled JavaScript"]) {
    const context = await browser.newContext({
      viewport: { width: 390, height: 874 },
      colorScheme: "light",
      javaScriptEnabled: mode !== "disabled JavaScript",
    });
    try {
      const page = await context.newPage();
      if (mode === "failed font")
        await page.route("**/fonts/**", (request) => request.abort());
      const response = await page.goto(new URL("/about/", baseURL).href, {
        waitUntil: "networkidle",
      });
      assert.equal(response.status(), 200);
      await page.evaluate(() => document.fonts.ready);
      const paragraph = page.locator(".article-content p").first();
      assert.equal(
        await hasPaintedText(paragraph),
        true,
        `${mode} must retain readable text`,
      );
      assert.equal(
        await page.locator(".planned-paragraph").count(),
        0,
        `${mode} must retain native wrapping`,
      );
      report.push({ mode, route: "/about/", width: 390 });
    } finally {
      await context.close();
    }
  }
  await fs.writeFile(
    path.join(output, "webkit-font-loading.json"),
    JSON.stringify(report, null, 2) + "\n",
  );
}

async function main() {
  const browser = await webkit.launch({ headless: true });
  const report = [];
  await fs.mkdir(output, { recursive: true });
  try {
    await checkColdFontLoading(browser);
    const context = await browser.newContext({
      viewport: { width: 402, height: 874 },
      isMobile: true,
      hasTouch: true,
      deviceScaleFactor: 3,
      colorScheme: "dark",
    });
    const page = await context.newPage();
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    page.on("requestfailed", (request) =>
      errors.push(`${request.url()}: ${request.failure().errorText}`),
    );
    for (const route of routes) {
      console.error(`Checking WebKit typography: ${route}`);
      const response = await page.goto(new URL(route, baseURL).href, {
        waitUntil: "networkidle",
      });
      assert.equal(response.status(), 200, route);
      await page.evaluate(() => document.fonts.ready);
      await page.waitForFunction(() =>
        document.querySelector(".planned-layout"),
      );
      let identity;
      for (const width of [320, 390, 402, 834, 874, 1024, 1280]) {
        console.error(`Checking ${width}px`);
        await page.setViewportSize({ width, height: 874 });
        await page.evaluate(
          () =>
            new Promise((resolve) =>
              requestAnimationFrame(() => requestAnimationFrame(resolve)),
            ),
        );
        const result = await inspect(page);
        if (["/", "/posts/", "/posts/page/2/", "/tags/scala/"].includes(route))
          assert.ok(
            result.column?.centered,
            `${route} must have a checked centered article list`,
          );
        report.push({
          route,
          width,
          active: result.active,
          failures: result.failures,
          column: result.column,
        });
        assert.equal(
          result.overflow,
          false,
          `${route} overflows at ${width}px`,
        );
        assert.ok(
          result.active > 0,
          `${route} must retain planned typography at ${width}px`,
        );
        const current = { links: result.links, code: result.code };
        if (identity)
          assert.deepEqual(
            current,
            identity,
            "Resize must preserve links and code blocks",
          );
        identity = current;
        if (route === routes[0] && width === 402) {
          assert.ok(
            result.active >= 16,
            "Phone-width FinagleCon must retain working planned typography",
          );
          await page
            .locator(".typography-prose p.planned-paragraph")
            .first()
            .evaluate((p) =>
              p.scrollIntoView({ block: "start", behavior: "instant" }),
            );
          await page.screenshot({
            path: path.join(output, "webkit-iphone-fixed.png"),
            scale: "css",
          });
        }
        if (route === "/about/" && width === 402) {
          await page.screenshot({
            path: path.join(output, "webkit-about-planned.png"),
            scale: "css",
            fullPage: true,
          });
        }
        if (route === "/" && width === 1280) {
          await page.screenshot({
            path: path.join(output, "webkit-home-centered.png"),
            scale: "css",
            fullPage: true,
          });
        }
      }
    }
    await page.emulateMedia({ media: "print" });
    const printed = await page.evaluate(() => ({
      wrapping: [...document.querySelectorAll("[data-plan-width]")].map(
        (p) => getComputedStyle(p).whiteSpace,
      ),
      breaks: [
        ...document.querySelectorAll(".planned-break, .planned-hyphen"),
      ].map((p) => getComputedStyle(p).display),
    }));
    assert.ok(
      printed.wrapping.every((value) => value === "normal"),
      "Print must use native wrapping",
    );
    assert.ok(
      printed.breaks.every((value) => value === "none"),
      "Print must hide generated breaks and hyphens",
    );
    await checkProjectBadges(page);
    assert.deepEqual(
      errors,
      [],
      "Browser errors and failed requests must be resolved",
    );
    await fs.writeFile(
      path.join(output, "webkit-typography.json"),
      JSON.stringify(report, null, 2) + "\n",
    );
    console.log(JSON.stringify(report, null, 2));
    assert.deepEqual(
      report.filter((result) => result.failures.length),
      [],
      "Planned lines must render without rewrapping or width errors",
    );
  } finally {
    await browser.close();
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
