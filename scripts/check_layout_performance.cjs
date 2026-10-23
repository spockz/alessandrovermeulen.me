// Bound real Chromium layout work and exercise scope isolation so batching cannot trade correctness for speed.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { execFileSync } = require("node:child_process");
const { chromium, webkit } = require(
  process.env.PLAYWRIGHT_MODULE || "playwright",
);

const baseURL = process.argv[2] || "http://localhost:1313/";
const root = path.resolve(__dirname, "..");
const output = path.join(root, ".cache/review");

function omitAnalytics(page) {
  return page.route("https://cloud.umami.is/script.js", (request) =>
    request.fulfill({ contentType: "application/javascript", body: "" }),
  );
}

async function settle(page) {
  await page.evaluate(() => document.fonts.ready);
  await page.waitForFunction(() =>
    document.querySelector(".planned-paragraph"),
  );
  await page.evaluate(
    () =>
      new Promise((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(resolve)),
      ),
  );
  await page.waitForTimeout(100);
}

async function checkReflows(browser) {
  const report = [];
  for (const route of [
    "/",
    "/posts/",
    "/2015/08/31/finaglecon/",
    "/2013/07/13/the-difference-between-shallow-and-deep-embedding/",
  ]) {
    for (const width of [390, 1280]) {
      const context = await browser.newContext({
        viewport: { width, height: 874 },
      });
      try {
        const page = await context.newPage();
        await omitAnalytics(page);
        const cdp = await context.newCDPSession(page);
        await cdp.send("Performance.enable");
        await cdp.send("Emulation.setCPUThrottlingRate", { rate: 4 });
        async function metrics() {
          return Object.fromEntries(
            (await cdp.send("Performance.getMetrics")).metrics.map(
              ({ name, value }) => [name, value],
            ),
          );
        }
        const response = await page.goto(new URL(route, baseURL).href, {
          waitUntil: "networkidle",
        });
        assert.equal(response.status(), 200);
        await settle(page);
        const before = await metrics();
        await page.setViewportSize({ width: width - 12, height: 874 });
        await settle(page);
        const after = await metrics();
        const layouts = after.LayoutCount - before.LayoutCount;
        const styles = after.RecalcStyleCount - before.RecalcStyleCount;
        assert.ok(
          layouts <= 24,
          `${route} at ${width}px performed ${layouts} layouts after resize`,
        );
        assert.ok(
          styles <= 24,
          `${route} at ${width}px performed ${styles} style recalculations after resize`,
        );
        await page.waitForTimeout(150);
        const idle = await metrics();
        assert.equal(
          idle.LayoutCount,
          after.LayoutCount,
          "The planner must settle after resizing",
        );
        assert.equal(
          idle.RecalcStyleCount,
          after.RecalcStyleCount,
          "The planner must not keep invalidating its own styles",
        );
        assert.equal(
          await page.evaluate(
            () => document.documentElement.scrollWidth > innerWidth,
          ),
          false,
        );
        report.push({
          route,
          width,
          layouts,
          styles,
          layoutMs: (after.LayoutDuration - before.LayoutDuration) * 1000,
        });
      } finally {
        await context.close();
      }
    }
  }
  return report;
}

async function createFixture(browser) {
  const page = await browser.newPage();
  let assets;
  try {
    await omitAnalytics(page);
    const response = await page.goto(baseURL, { waitUntil: "networkidle" });
    assert.equal(response.status(), 200);
    assets = await page.evaluate(() =>
      [
        ...document.querySelectorAll(
          'head link[rel="stylesheet"], head script[src*="planned-layout.min."]',
        ),
      ]
        .map((node) => node.outerHTML)
        .join("\n"),
    );
  } finally {
    await page.close();
  }
  const directory = path.join(output, "batch-fixture");
  await fs.mkdir(directory, { recursive: true });
  const filename = path.join(directory, "index.html");
  await fs.writeFile(
    filename,
    `<!doctype html><html><head><meta charset="utf-8">${assets}<style>body{padding:24px}main{max-width:576px}#nested{padding-inline:16px}img{display:block;margin:0}</style></head><body><main class="article-content">
<section id="outer" class="typography-prose" lang="en-GB" data-hyphenate-lang="en-GB" data-plan-lines="responsive">
<p id="outer-text">Programming languages help us express complex systems with useful abstractions. Good explanations preserve the details that matter while making their relationships easier to understand.</p>
<section id="nested" class="typography-prose" lang="en-GB" data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p id="nested-text">Nested explanations use their own available column width. They must keep independent line measurements when another paragraph changes its layout at the same time.</p></section>
<p id="image-text">An inline image <img id="inline-image" src="/favicon.svg" width="40" height="16" alt=""> participates in the paragraph without changing the spacing inside its own box. The following words must still fit the selected lines.</p></section>
<section id="sibling" class="typography-prose" lang="en-GB" data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p id="sibling-text">Independent paragraphs on the same page share a measurement phase, while retaining their own font checks and line sequences. A change in one scope must never activate an invalid plan in another scope.</p></section>
</main></body></html>`,
  );
  execFileSync(
    path.join(root, ".cache/tools/typography"),
    ["process", directory],
    { cwd: root, stdio: "pipe" },
  );
  return fs.readFile(filename, "utf8");
}

async function checkScopes(browser, html) {
  const context = await browser.newContext({
    viewport: { width: 390, height: 874 },
  });
  try {
    const page = await context.newPage();
    await page.route("**/__typography_batch/", (request) =>
      request.fulfill({ contentType: "text/html", body: html }),
    );
    await page.goto(new URL("/__typography_batch/", baseURL).href, {
      waitUntil: "networkidle",
    });
    const ids = ["outer-text", "nested-text", "image-text", "sibling-text"];
    async function active(expected) {
      await page.waitForFunction(
        (wanted) =>
          wanted.every((id) =>
            document.getElementById(id).classList.contains("planned-paragraph"),
          ),
        expected,
      );
    }
    await active(ids);
    for (const width of [390, 402, 834]) {
      await page.setViewportSize({ width, height: 874 });
      await settle(page);
      await active(ids);
      const blocks = await page.evaluate(
        (ids) =>
          ids.map((id) => {
            const node = document.getElementById(id);
            const box = node.getBoundingClientRect();
            const lines = [...node.querySelectorAll("[data-natural]")].map(
              (first) => {
                const pieces = [
                  ...node.querySelectorAll(
                    `[data-line="${first.dataset.line}"]`,
                  ),
                ];
                const range = document.createRange();
                range.setStartBefore(pieces[0]);
                range.setEndAfter(pieces.at(-1));
                const rects = [...range.getClientRects()].filter(
                  (rect) => rect.width > 0 && rect.height > 0,
                );
                return {
                  width: range.getBoundingClientRect().width,
                  final: first.dataset.final === "true",
                  singleLine:
                    rects.length > 0 &&
                    Math.max(...rects.map((rect) => rect.top)) <
                      Math.min(...rects.map((rect) => rect.bottom)),
                };
              },
            );
            return {
              id,
              width: box.width,
              lines,
              scope: node.closest("[data-plan-lines]").id,
            };
          }),
        ids,
      );
      assert.deepEqual(
        blocks.map((block) => block.scope),
        ["outer", "nested", "outer", "sibling"],
      );
      for (const block of blocks) {
        assert.ok(block.lines.length > 0);
        for (const line of block.lines) {
          assert.ok(line.singleLine, `${block.id} wraps a planned line`);
          assert.ok(
            line.width <= block.width + 1,
            `${block.id} exceeds its own column`,
          );
          if (!line.final)
            assert.ok(
              Math.abs(line.width - block.width) <= 1,
              `${block.id} does not fill its own column`,
            );
        }
      }
      assert.equal(
        await page
          .locator("#inline-image")
          .evaluate((node) => node.getBoundingClientRect().width),
        40,
      );
    }
    await page.addStyleTag({
      content:
        "@media(max-height:760px){#image-text{letter-spacing:1px!important}}",
    });
    await settle(page);
    await active(ids);
    for (const height of [730, 874]) {
      await page.setViewportSize({ width: 834, height });
      if (height === 730) {
        await page.waitForFunction(() => {
          const paragraph = document.getElementById("image-text");
          return (
            getComputedStyle(paragraph).letterSpacing === "1px" &&
            !paragraph.classList.contains("planned-paragraph")
          );
        });
        await active(ids.filter((id) => id !== "image-text"));
      } else {
        await active(ids);
      }
    }
    await page.locator("#inline-image").evaluate((node) => {
      node.style.maxWidth = "20px";
    });
    await page.waitForFunction(
      () =>
        !document
          .getElementById("image-text")
          .classList.contains("planned-paragraph"),
    );
    await active(ids.filter((id) => id !== "image-text"));
    await page.locator("#inline-image").evaluate((node) => {
      node.style.maxWidth = "";
    });
    await active(ids);
    await page.waitForTimeout(150);
    assert.equal(
      await page.locator("body > div[aria-hidden]").count(),
      0,
      "Measurement probes must be removed",
    );
  } finally {
    await context.close();
  }
}

async function main() {
  await fs.mkdir(output, { recursive: true });
  const browser = await chromium.launch({ headless: true });
  let fixture;
  try {
    const report = await checkReflows(browser);
    fixture = await createFixture(browser);
    await checkScopes(browser, fixture);
    await fs.writeFile(
      path.join(output, "chromium-layout-performance.json"),
      JSON.stringify({ browser: browser.version(), report }, null, 2) + "\n",
    );
    console.log(JSON.stringify(report, null, 2));
  } finally {
    await browser.close();
  }
  const safari = await webkit.launch({ headless: true });
  try {
    await checkScopes(safari, fixture);
  } finally {
    await safari.close();
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
