// Check real browser line fragments because font measurements alone cannot detect CSS rewrapping.
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
  "/",
  "/tags/scala/",
  "/about/",
];

async function inspect(page) {
  return page.evaluate(() => {
    const failures = [];
    // WebKit can apply fractional word spacing differently at a text-fragment boundary.
    const widthTolerance = 1;
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

async function main() {
  const browser = await webkit.launch({ headless: true });
  const report = [];
  await fs.mkdir(output, { recursive: true });
  try {
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
      for (const width of [320, 390, 402, 874, 1280]) {
        console.error(`Checking ${width}px`);
        await page.setViewportSize({ width, height: 874 });
        await page.evaluate(
          () =>
            new Promise((resolve) =>
              requestAnimationFrame(() => requestAnimationFrame(resolve)),
            ),
        );
        const result = await inspect(page);
        report.push({
          route,
          width,
          active: result.active,
          failures: result.failures,
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
