// Exercise native selection and clipboard behavior because generated layout must preserve ordinary reading interactions.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const { webkit, chromium } = require(
  process.env.PLAYWRIGHT_MODULE || "playwright",
);

const baseURL = process.argv[2] || "http://localhost:1313/";
const engine = process.env.TYPOGRAPHY_BROWSER || "webkit";
assert.ok(["webkit", "chromium"].includes(engine), "Unknown browser engine");

async function selectedText(page) {
  return page.evaluate(() =>
    window
      .getSelection()
      .toString()
      .replaceAll("\u00ad", "")
      .replace(/\s+/g, " ")
      .trim(),
  );
}

async function copiedText(page) {
  return page.evaluate(() => {
    const data = new DataTransfer();
    const event = new ClipboardEvent("copy", {
      clipboardData: data,
      cancelable: true,
    });
    document.dispatchEvent(event);
    return {
      text: data.getData("text/plain"),
      html: data.getData("text/html"),
      handled: event.defaultPrevented,
    };
  });
}

async function main() {
  const browser = await { webkit, chromium }[engine].launch({
    headless: true,
    ...(engine === "chromium" && process.env.CHROMIUM_EXECUTABLE
      ? { executablePath: process.env.CHROMIUM_EXECUTABLE }
      : {}),
  });
  try {
    const context = await browser.newContext({
      viewport: { width: 402, height: 874 },
      deviceScaleFactor: 3,
      isMobile: true,
      hasTouch: true,
      colorScheme: "dark",
    });
    const page = await context.newPage();
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    const response = await page.goto(
      new URL("/2015/08/31/finaglecon/", baseURL).href,
      { waitUntil: "networkidle" },
    );
    assert.equal(response.status(), 200);
    await page.evaluate(() => document.fonts.ready);
    await page.waitForFunction(() => document.querySelector(".planned-layout"));
    const paragraph = page
      .locator("p[data-plan-width]")
      .filter({ hasText: /^The use of/ });
    assert.equal(await paragraph.count(), 1);
    const expected = await paragraph.evaluate((p) =>
      p.textContent.replaceAll("\u00ad", "").replace(/\s+/g, " ").trim(),
    );
    await paragraph.evaluate((p) =>
      p.scrollIntoView({ block: "start", behavior: "instant" }),
    );
    const wordPoint = (p, word) => {
      const text = p.querySelector(".planned-line").firstChild;
      const start = text.data.indexOf(word);
      if (start < 0) throw new Error(`Expected ${word} on the first line`);
      const range = document.createRange();
      range.setStart(text, start);
      range.setEnd(text, start + 2);
      const box = range.getBoundingClientRect();
      return { x: (box.left + box.right) / 2, y: (box.top + box.bottom) / 2 };
    };
    const point = await paragraph.evaluate(wordPoint, "asyn");
    await page.mouse.click(point.x, point.y, { clickCount: 2 });
    const word = await selectedText(page);
    assert.equal(
      word,
      "asynchronous",
      "Word selection must cross a planned hyphenation break",
    );
    const wordCopy = await copiedText(page);
    assert.ok(wordCopy.handled);
    assert.equal(wordCopy.text, word);
    assert.doesNotMatch(wordCopy.html, /planned-|data-plan-|\u00ad/);
    const unbrokenPoint = await paragraph.evaluate(wordPoint, "Fu");
    await page.mouse.click(unbrokenPoint.x, unbrokenPoint.y, { clickCount: 2 });
    assert.equal(await selectedText(page), "Futures");
    const unbrokenCopy = await copiedText(page);
    assert.ok(
      unbrokenCopy.handled,
      "Copy cleanup must handle a single text node",
    );
    assert.equal(unbrokenCopy.text, "Futures");
    assert.doesNotMatch(unbrokenCopy.html, /\u00ad/);
    await page.mouse.click(point.x, point.y, { clickCount: 3 });
    assert.equal(
      await selectedText(page),
      expected,
      "Paragraph selection must cross every planned line",
    );
    const paragraphCopy = await copiedText(page);
    assert.equal(paragraphCopy.text, expected);
    assert.doesNotMatch(paragraphCopy.html, /planned-|data-plan-|\u00ad/);
    await page.screenshot({
      path: `.cache/review/${engine}-paragraph-selection.png`,
      scale: "css",
    });

    await page.setViewportSize({ width: 390, height: 874 });
    await page.evaluate(
      () =>
        new Promise((resolve) =>
          requestAnimationFrame(() => requestAnimationFrame(resolve)),
        ),
    );
    assert.equal(
      await selectedText(page),
      expected,
      "Resizing must retain the active selection",
    );
    assert.equal(
      await paragraph.evaluate((p) =>
        p.classList.contains("planned-paragraph"),
      ),
      false,
      "A selected paragraph must wrap natively while its new plan is deferred",
    );
    await page.evaluate(() => window.getSelection().removeAllRanges());
    await page.waitForFunction(() =>
      [...document.querySelectorAll("p.planned-paragraph")].some(
        (p) =>
          p.textContent.startsWith("The use of") &&
          p.dataset.planWidth === "342",
      ),
    );

    await page.setViewportSize({ width: 402, height: 874 });
    const writing = await page.goto(new URL("/posts/", baseURL).href, {
      waitUntil: "networkidle",
    });
    assert.equal(writing.status(), 200);
    await page.evaluate(() => document.fonts.ready);
    await page.waitForFunction(() =>
      document.querySelector(".typography-summary .planned-paragraph"),
    );
    const summaries = await page.evaluate(() =>
      [...document.querySelectorAll(".article-link__summary")].map((p) => ({
        generated: p.dataset.planGenerated,
        planned: !!p.querySelector("[data-line-plans]"),
        active: !!p.querySelector(".planned-paragraph"),
        width: p.getBoundingClientRect().width,
      })),
    );
    assert.ok(summaries.length > 0);
    assert.ok(
      summaries.every(
        (summary) => summary.generated === "true" && summary.planned,
      ),
      "Every Writing description must receive a plan",
    );
    assert.ok(summaries.some((summary) => summary.active));
    const summary = page
      .locator(".typography-summary .planned-paragraph")
      .first();
    await summary.evaluate((p) =>
      p.scrollIntoView({ block: "start", behavior: "instant" }),
    );
    await summary.locator(".planned-line").first().dblclick();
    assert.equal(
      await page.url(),
      new URL("/posts/", baseURL).href,
      "Selecting a description must not activate the overlaid article link",
    );
    assert.ok(
      (await selectedText(page)).length > 0,
      "Descriptions must remain selectable",
    );
    await page.evaluate(() => window.getSelection().removeAllRanges());
    await page.evaluate(() => window.scrollTo({ top: 0, behavior: "instant" }));
    await page.screenshot({
      path: `.cache/review/${engine}-writing-planned.png`,
      scale: "css",
    });
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth > innerWidth,
      ),
      false,
    );
    assert.deepEqual(errors, []);
    const report = {
      engine,
      word,
      paragraphLength: expected.length,
      summaries,
    };
    await fs.writeFile(
      `.cache/review/${engine}-selection.json`,
      JSON.stringify(report, null, 2) + "\n",
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
