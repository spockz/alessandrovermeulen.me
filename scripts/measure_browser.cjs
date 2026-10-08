// Measure the production HTML in an isolated browser so optional theme features cannot hide loading costs.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || "playwright");

const baseURL = process.argv[2] || "http://127.0.0.1:1314/";
const output = path.resolve(process.argv[3] || ".cache/review");
const article =
  "/2013/07/13/the-difference-between-shallow-and-deep-embedding/";
const conditions = {
  viewport: { width: 390, height: 844 },
  latencyMs: 150,
  downloadBytesPerSecond: 200000,
  uploadBytesPerSecond: 75000,
  cpuSlowdown: 4,
  compressed: false,
};

async function main() {
  const browser = await chromium.launch({ headless: true });
  const results = [];
  await fs.mkdir(output, { recursive: true });
  try {
    for (const route of ["/", article]) {
      for (let run = 1; run <= 3; run++) {
        const context = await browser.newContext({
          viewport: conditions.viewport,
          isMobile: true,
          hasTouch: true,
          deviceScaleFactor: 1,
          colorScheme: "light",
        });
        try {
          const page = await context.newPage();
          const errors = [];
          page.on("pageerror", (error) => errors.push(error.message));
          page.on("response", (response) => {
            if (response.status() >= 400)
              errors.push(`${response.status()} ${response.url()}`);
          });
          page.on("requestfailed", (request) =>
            errors.push(`${request.url()}: ${request.failure().errorText}`),
          );
          await page.addInitScript(() => {
            window.loadingMetrics = { lcp: null, cls: 0 };
            new PerformanceObserver((list) => {
              window.loadingMetrics.lcp = list.getEntries().at(-1).startTime;
            }).observe({ type: "largest-contentful-paint", buffered: true });
            new PerformanceObserver((list) => {
              for (const entry of list.getEntries()) {
                if (!entry.hadRecentInput)
                  window.loadingMetrics.cls += entry.value;
              }
            }).observe({ type: "layout-shift", buffered: true });
          });
          const cdp = await context.newCDPSession(page);
          await cdp.send("Network.enable");
          await cdp.send("Network.setCacheDisabled", { cacheDisabled: true });
          await cdp.send("Network.emulateNetworkConditions", {
            offline: false,
            latency: conditions.latencyMs,
            downloadThroughput: conditions.downloadBytesPerSecond,
            uploadThroughput: conditions.uploadBytesPerSecond,
            connectionType: "cellular4g",
          });
          await cdp.send("Emulation.setCPUThrottlingRate", {
            rate: conditions.cpuSlowdown,
          });
          const response = await page.goto(new URL(route, baseURL).href, {
            waitUntil: "networkidle",
          });
          assert.equal(response.status(), 200);
          const metrics = await page.evaluate(() => ({
            ...window.loadingMetrics,
            fcp: performance.getEntriesByName("first-contentful-paint")[0]
              ?.startTime,
            requests: performance.getEntriesByType("resource").map((entry) => ({
              url: entry.name,
              transferBytes: entry.transferSize,
            })),
            viewportWidth: innerWidth,
            documentWidth: document.documentElement.scrollWidth,
          }));
          assert.ok(
            metrics.lcp !== null,
            "LCP observer must produce a measurement",
          );
          assert.ok(metrics.fcp > 0, "FCP must be measured");
          assert.ok(
            metrics.documentWidth <= metrics.viewportWidth,
            "Page must fit the mobile viewport",
          );
          assert.deepEqual(
            errors,
            [],
            "Browser errors and failed requests must be resolved",
          );
          assert.ok(
            metrics.requests.every(
              (request) =>
                new URL(request.url).origin === new URL(baseURL).origin,
            ),
            "Representative pages must load only local resources",
          );
          results.push({ route, run, ...metrics });
          if (run === 1) {
            const label = route === "/" ? "home-mobile" : "article-mobile";
            await page.screenshot({
              path: path.join(output, `${label}.png`),
              fullPage: true,
            });
          }
        } finally {
          await context.close();
        }
      }
    }
    const context = await browser.newContext({
      viewport: { width: 1440, height: 1000 },
      colorScheme: "light",
    });
    try {
      const page = await context.newPage();
      await page.goto(baseURL, { waitUntil: "networkidle" });
      await page.screenshot({
        path: path.join(output, "home-desktop.png"),
        fullPage: true,
      });
      await page
        .getByRole("button", { name: "Switch to dark appearance" })
        .click();
      assert.equal(
        await page
          .getByRole("button", { name: "Switch to light appearance" })
          .count(),
        1,
      );
      await page.screenshot({
        path: path.join(output, "home-dark.png"),
        fullPage: true,
      });
    } finally {
      await context.close();
    }
    const report = { conditions, results };
    await fs.writeFile(
      path.join(output, "browser.json"),
      `${JSON.stringify(report, null, 2)}\n`,
    );
    console.log(JSON.stringify(report, null, 2));
  } finally {
    await browser.close();
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
