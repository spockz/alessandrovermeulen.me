// Prove that matching declared fonts alone cannot enable stale geometry or shaping plans.
const assert = require("node:assert/strict");
const { test } = require("node:test");
const {
  assessPlan,
  sourceText,
  selectWidthPlan,
  lineSegments,
  contentWidth,
  inlineStyleMatches,
  fontFaceMatches,
  fontsReady,
  imageStyleMatches,
} = require("../assets/js/planned-layout.js");

const inline = (overrides = {}) => ({
  family: "article sans",
  size: 16,
  weight: "400",
  style: "normal",
  transform: "none",
  letterSpacing: 0,
  paddingLeft: 0,
  paddingRight: 0,
  borderLeftWidth: 0,
  borderRightWidth: 0,
  marginLeft: 0,
  marginRight: 0,
  ...overrides,
});

const style = {
  fontFamily: '"Article Sans", sans-serif',
  fontSize: "16px",
  fontWeight: "400",
  fontStyle: "normal",
  fontStretch: "100%",
  fontKerning: "normal",
  fontVariantLigatures: "normal",
  fontFeatureSettings: "normal",
  fontVariationSettings: "normal",
  letterSpacing: "normal",
  textTransform: "none",
  wordSpacing: "0px",
  textIndent: "0px",
  paddingLeft: "0px",
  paddingRight: "0px",
  borderLeftWidth: "0px",
  borderRightWidth: "0px",
  marginLeft: "0px",
  marginRight: "0px",
};
const geometry = {
  fontReady: true,
  style,
  size: 16,
  width: 576,
  plannedWidth: 576,
};
const lines = [
  { actual: 571, expected: 571.1, spaces: 10, final: false },
  { actual: 580, expected: 580, spaces: 8, final: false },
  { actual: 250, expected: 250, spaces: 4, final: true },
];

test("matching real metrics permit stretch, shrink, and a ragged final line", () => {
  const result = assessPlan(geometry, lines);
  assert.equal(result.valid, true);
  assert.deepEqual(result.adjustments, [0.5, -0.5, 0]);
  assert.equal(
    assessPlan(
      {
        ...geometry,
        style: { ...style, fontFamily: '"article sans", sans-serif' },
      },
      lines,
    ).valid,
    true,
  );
});

test("matching family and size still reject an unloaded or substituted font", () => {
  assert.equal(
    assessPlan({ ...geometry, fontReady: false }, lines).valid,
    false,
  );
  assert.equal(
    assessPlan(geometry, [{ ...lines[0], actual: 568 }]).reason,
    "font metrics changed",
  );
  assert.equal(
    assessPlan(
      { ...geometry, style: { ...style, fontFamily: "Arial, sans-serif" } },
      lines,
    ).reason,
    "font changed",
  );
});

test("reader size, width, and inline style changes turn plans off", () => {
  assert.equal(
    assessPlan({ ...geometry, width: 352 }, lines).reason,
    "column width changed",
  );
  assert.equal(
    assessPlan({ ...geometry, style: { ...style, fontSize: "20px" } }, lines)
      .reason,
    "text size changed",
  );
  for (const property of [
    "fontWeight",
    "fontFeatureSettings",
    "letterSpacing",
    "wordSpacing",
  ]) {
    assert.equal(
      assessPlan(
        { ...geometry, style: { ...style, [property]: "changed" } },
        lines,
      ).valid,
      false,
      property,
    );
  }
  assert.equal(
    assessPlan({ ...geometry, inlineMismatch: true }, lines).valid,
    false,
  );
});

test("list indentation and block padding select the remaining content width", () => {
  assert.equal(contentWidth(550, { ...style, paddingLeft: "6px" }), 544);
  assert.equal(
    contentWidth(576, {
      ...style,
      paddingLeft: "12px",
      paddingRight: "8px",
      borderLeftWidth: "2px",
      borderRightWidth: "1px",
    }),
    553,
  );
  assert.equal(
    assessPlan(
      {
        ...geometry,
        width: 553,
        plannedWidth: 553,
        style: { ...style, paddingLeft: "12px", paddingRight: "8px" },
      },
      [{ actual: 550, expected: 550, spaces: 6 }],
    ).valid,
    true,
  );
  assert.ok(
    Number.isNaN(contentWidth(550, { ...style, paddingLeft: "unknown" })),
  );
});

test("inline code has a pinned smaller font and two measured padding edges", () => {
  const mono = {
    ...style,
    fontFamily: '"Article Mono", monospace',
    fontSize: "14px",
    paddingLeft: "5px",
    paddingRight: "5px",
    marginLeft: "0px",
    marginRight: "0px",
  };
  const expected = inline({
    family: "article mono",
    size: 14,
    paddingLeft: 5,
    paddingRight: 5,
  });
  assert.equal(inlineStyleMatches(mono, expected), true);
  for (const change of [
    { fontFamily: "monospace" },
    { fontSize: "16px" },
    { paddingLeft: "0px" },
    { paddingRight: "6px" },
    { fontWeight: "600" },
  ])
    assert.equal(inlineStyleMatches({ ...mono, ...change }, expected), false);
  const link = { ...style, marginLeft: "0px", marginRight: "0px" };
  assert.equal(inlineStyleMatches(link, inline()), true);
  assert.equal(
    inlineStyleMatches({ ...link, paddingLeft: "5px" }, inline()),
    false,
  );
});

test("mixed inline fonts match their declared metrics rather than their parent", () => {
  for (const [weight, fontStyle, size] of [
    ["700", "normal", 16],
    ["400", "italic", 16],
    ["700", "italic", 12.8],
    ["400", "normal", 12],
    ["700", "italic", 10.5],
  ]) {
    const actual = {
      ...style,
      fontWeight: weight,
      fontStyle,
      fontSize: `${size}px`,
    };
    const expected = inline({ weight, style: fontStyle, size });
    assert.equal(inlineStyleMatches(actual, expected), true);
    assert.equal(
      inlineStyleMatches({ ...actual, fontFamily: "sans-serif" }, expected),
      false,
    );
    assert.equal(
      inlineStyleMatches({ ...actual, fontSize: `${size + 1}px` }, expected),
      false,
    );
  }
  const expected = inline({
    transform: "uppercase",
    letterSpacing: 0.25,
    paddingLeft: 3,
    marginRight: -1,
  });
  const actual = {
    ...style,
    textTransform: "uppercase",
    letterSpacing: "0.25px",
    paddingLeft: "3px",
    marginRight: "-1px",
  };
  assert.equal(inlineStyleMatches(actual, expected), true);
  assert.equal(
    inlineStyleMatches({ ...actual, textTransform: "none" }, expected),
    false,
  );
  assert.equal(
    inlineStyleMatches({ ...actual, letterSpacing: "normal" }, expected),
    false,
  );
  assert.equal(
    inlineStyleMatches({ ...actual, marginRight: "0px" }, expected),
    false,
  );
});

test("italic and bold block roots permit only their declared face and size", () => {
  for (const weight of ["400", "700"]) {
    for (const fontStyle of ["normal", "italic"]) {
      const block = {
        ...geometry,
        weight,
        fontStyle,
        size: 14,
        style: { ...style, fontWeight: weight, fontStyle, fontSize: "14px" },
      };
      assert.equal(assessPlan(block, lines).valid, true);
      assert.equal(
        assessPlan(
          { ...block, style: { ...block.style, fontSize: "16px" } },
          lines,
        ).reason,
        "text size changed",
      );
      assert.equal(
        assessPlan(
          {
            ...block,
            style: {
              ...block.style,
              fontStyle: fontStyle === "normal" ? "italic" : "normal",
            },
          },
          lines,
        ).reason,
        "text styling changed",
      );
    }
  }
});

test("loaded regular faces cannot stand in for an italic or bold face", () => {
  const regular = {
    family: '"Article Sans"',
    weight: "400",
    style: "normal",
    status: "loaded",
  };
  const boldItalic = { ...regular, weight: "700", style: "italic" };
  const mono = { ...regular, family: "Article Mono" };
  const expected = [
    inline(),
    inline({ weight: "700", style: "italic" }),
    inline({ family: "article mono", size: 14 }),
  ];
  assert.equal(fontsReady([regular], [expected[0]]), true);
  assert.equal(fontsReady([regular, mono], expected), false);
  assert.equal(fontsReady([regular, boldItalic, mono], expected), true);
  assert.equal(
    fontFaceMatches({ ...boldItalic, status: "loading" }, expected[1]),
    false,
  );
  assert.equal(
    fontFaceMatches({ ...boldItalic, family: "Fallback Sans" }, expected[1]),
    false,
  );
  assert.equal(
    fontFaceMatches({ ...boldItalic, weight: "400 700" }, expected[1]),
    false,
  );
});

test("a rejected mixed-font block leaves independently matching blocks valid", () => {
  const blocks = [
    geometry,
    {
      ...geometry,
      weight: "400",
      fontStyle: "italic",
      style: { ...style, fontStyle: "italic" },
      fontReady: false,
    },
    { ...geometry, inlineMismatch: true },
    { ...geometry, size: 14, style: { ...style, fontSize: "14px" } },
  ];
  assert.deepEqual(
    blocks.map((block) => assessPlan(block, lines).valid),
    [true, false, false, true],
  );
  assert.deepEqual(
    blocks.map(
      (block) =>
        assessPlan(
          { ...block, style: { ...block.style, fontSize: "20px" } },
          lines,
        ).valid,
    ),
    [false, false, false, false],
  );
});

test("atomic code spacing excludes its fixed spaces and includes padding in measured width", () => {
  const result = assessPlan({ ...geometry, width: 100, plannedWidth: 100 }, [
    { actual: 94, expected: 94, spaces: 1, final: false },
  ]);
  assert.deepEqual(result, { valid: true, adjustments: [6] });
  assert.equal(
    assessPlan({ ...geometry, width: 100, plannedWidth: 100 }, [
      { actual: 84, expected: 94, spaces: 1, final: false },
    ]).reason,
    "font metrics changed",
  );
  assert.equal(
    inlineStyleMatches(
      {
        ...style,
        fontFamily: '"Article Mono"',
        fontSize: "14px",
        paddingLeft: "5px",
        paddingRight: "5px",
        wordSpacing: "6px",
      },
      inline({
        family: "article mono",
        size: 14,
        paddingLeft: 5,
        paddingRight: 5,
      }),
    ),
    false,
  );
});

test("missing and impossible line plans cannot activate fixed breaks", () => {
  assert.equal(assessPlan(geometry, []).valid, false);
  assert.equal(
    assessPlan({ ...geometry, width: Number.NaN }, lines).valid,
    false,
  );
  assert.equal(assessPlan({ ...geometry, size: 0 }, lines).valid, false);
  assert.equal(
    assessPlan(geometry, [
      { actual: 600, expected: 600, spaces: 0, final: false },
    ]).valid,
    false,
  );
  assert.equal(
    assessPlan(geometry, [
      { actual: 600, expected: 600, spaces: 2, final: false },
    ]).valid,
    false,
  );
});

test("clipboard wording drops discretionary characters without joining distinct words", () => {
  assert.equal(
    sourceText("hy\u00adphen\u00ada\u00adtion across two words"),
    "hyphenation across two words",
  );
  assert.equal(
    sourceText("first\u00ad line\nsecond line"),
    "first line\nsecond line",
  );
});

const text = "café soft\u00adhyphen text";
const bundle = {
  lines: [
    { start: 0, end: 10, natural: 80, spaces: 1, hyphen: true },
    { start: 10, end: 21, natural: 78, spaces: 1 },
    { start: 0, end: 21, natural: 160, spaces: 3 },
  ],
  paths: [[0, 1], [2]],
  widths: { 240: 0, 241: 0, 360: 1 },
};

test("fluid widths select the lower integer plan and share unchanged paths", () => {
  assert.deepEqual(selectWidthPlan(bundle, 240.75), {
    width: 240,
    pathID: 0,
    path: [0, 1],
  });
  assert.equal(selectWidthPlan(bundle, 241).pathID, 0);
  assert.equal(selectWidthPlan(bundle, 360).pathID, 1);
  for (const width of [239, 242, 577, 0, Number.NaN])
    assert.equal(selectWidthPlan(bundle, width), null);
  assert.throws(
    () => selectWidthPlan({ ...bundle, widths: { 240: 3 } }, 240),
    /Invalid paragraph plan path/,
  );
});

test("responsive geometry tolerance does not relax font measurement tolerance", () => {
  const responsive = { ...geometry, width: 576.75, widthTolerance: 1 };
  assert.equal(assessPlan(responsive, lines).valid, true);
  assert.equal(assessPlan({ ...responsive, width: 577.1 }, lines).valid, false);
  assert.equal(
    assessPlan(responsive, [{ ...lines[0], actual: 571.7 }]).reason,
    "font metrics changed",
  );
});

test("changing paths preserves complete text ranges and conditional hyphens", () => {
  const narrow = lineSegments(bundle, selectWidthPlan(bundle, 240), text);
  assert.deepEqual(
    narrow.map((line) => text.slice(line.visibleStart, line.visibleEnd)),
    ["café soft", "hyphen text"],
  );
  assert.deepEqual(
    narrow.map((line) => line.final),
    [false, true],
  );
  assert.equal(
    narrow.map((line) => text.slice(line.start, line.end)).join(""),
    text,
  );
  const wide = lineSegments(bundle, selectWidthPlan(bundle, 360), text);
  assert.equal(wide.length, 1);
  assert.equal(text.slice(wide[0].visibleStart, wide[0].visibleEnd), text);
});

test("atomic ranges keep code spaces visible and reject breaks inside boxes", () => {
  const source = "a  code  z";
  const atomic = { start: 2, end: 8 };
  const plans = {
    lines: [
      { start: 0, end: 8, natural: 70, spaces: 1 },
      { start: 8, end: 10, natural: 10, spaces: 0 },
    ],
  };
  const selection = { path: [0, 1] };
  const segments = lineSegments(plans, selection, source, [atomic]);
  assert.deepEqual(
    segments.map((line) => source.slice(line.visibleStart, line.visibleEnd)),
    ["a  code ", "z"],
  );
  const broken = {
    lines: [
      { ...plans.lines[0], end: 7 },
      { ...plans.lines[1], start: 7 },
    ],
  };
  assert.throws(
    () => lineSegments(broken, selection, source, [atomic]),
    /splits an atomic inline box/,
  );
});

test("mandatory source breaks end ragged lines without changing complete source ranges", () => {
  const source = "first words \n next line words ending";
  const afterBreak = source.indexOf("\n") + 1;
  const plans = {
    lines: [
      { start: 0, end: afterBreak, natural: 60, spaces: 1, ragged: true },
      { start: afterBreak, end: source.length, natural: 90, spaces: 3 },
    ],
  };
  const segments = lineSegments(plans, { path: [0, 1] }, source);
  assert.deepEqual(
    segments.map((line) => source.slice(line.visibleStart, line.visibleEnd)),
    ["first words", "next line words ending"],
  );
  assert.deepEqual(
    segments.map((line) => line.final),
    [true, true],
  );
  assert.deepEqual(
    segments.map((line) => line.isLast),
    [false, true],
  );
  assert.equal(
    segments.map((line) => source.slice(line.start, line.end)).join(""),
    source,
  );
  assert.deepEqual(
    assessPlan(
      { ...geometry, width: 100, plannedWidth: 100 },
      segments.map((line) => ({
        ...line,
        actual: line.natural,
        expected: line.natural,
      })),
    ),
    { valid: true, adjustments: [0, 0] },
  );
  assert.deepEqual(
    assessPlan({ ...geometry, width: 100, plannedWidth: 100 }, [
      { actual: 101, expected: 101, spaces: 1, final: true },
    ]),
    { valid: true, adjustments: [-1] },
  );
});

test("a mandatory break followed by an ordinary line retains its distinct spacing policy", () => {
  const source = "first\n next middle ending";
  const plans = {
    lines: [
      { start: 0, end: 6, natural: 30, spaces: 0, ragged: true },
      { start: 6, end: 19, natural: 98, spaces: 1 },
      { start: 19, end: source.length, natural: 40, spaces: 0 },
    ],
  };
  const segments = lineSegments(plans, { path: [0, 1, 2] }, source);
  assert.deepEqual(
    segments.map((line) => line.final),
    [true, false, true],
  );
  assert.deepEqual(
    segments.map((line) => line.isLast),
    [false, false, true],
  );
  assert.deepEqual(
    assessPlan(
      { ...geometry, width: 100, plannedWidth: 100 },
      segments.map((line) => ({
        ...line,
        actual: line.natural,
        expected: line.natural,
      })),
    ).adjustments,
    [0, 2, 0],
  );
});

test("inline image tokens occupy one source position and require their measured width", () => {
  const source = "text \ufffc after";
  const plans = {
    lines: [
      { start: 0, end: 6, natural: 70, spaces: 1 },
      { start: 6, end: source.length, natural: 30, spaces: 0 },
    ],
  };
  const segments = lineSegments(plans, { path: [0, 1] }, source, [
    { start: 5, end: 6 },
  ]);
  assert.deepEqual(
    segments.map((line) => source.slice(line.visibleStart, line.visibleEnd)),
    ["text \ufffc", "after"],
  );
  assert.equal(imageStyleMatches(32, style, { width: 32 }), true);
  assert.equal(imageStyleMatches(31, style, { width: 32 }), false);
  assert.equal(
    imageStyleMatches(32, { ...style, marginRight: "1px" }, { width: 32 }),
    false,
  );
  assert.equal(imageStyleMatches(32, style, { width: Number.NaN }), false);
});

test("corrupt or incomplete source ranges fail explicitly", () => {
  const selection = selectWidthPlan(bundle, 240);
  assert.throws(
    () =>
      lineSegments(
        { ...bundle, paths: [[0]] },
        { ...selection, path: [0] },
        text,
      ),
    /Incomplete paragraph plan/,
  );
  assert.throws(
    () => lineSegments(bundle, selection, text.replace("\u00ad", "-")),
    /Missing discretionary hyphen/,
  );
  for (const changed of [
    { start: 9 },
    { end: 22 },
    { natural: null },
    { spaces: -1 },
  ]) {
    const broken = {
      ...bundle,
      lines: [bundle.lines[0], { ...bundle.lines[1], ...changed }],
    };
    assert.throws(
      () => lineSegments(broken, selection, text),
      /Invalid paragraph line range/,
    );
  }
});
