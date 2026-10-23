// Validate build-time line plans against the reader's actual font and geometry, keeping responsive prose as the default.
(() => {
  const tolerance = 0.4;
  const fontProperties = [
    "fontFamily",
    "fontSize",
    "fontWeight",
    "fontStyle",
    "fontStretch",
    "fontKerning",
    "fontVariantLigatures",
    "fontFeatureSettings",
    "fontVariationSettings",
    "letterSpacing",
    "textTransform",
  ];
  const edgeProperties = [
    "paddingLeft",
    "paddingRight",
    "borderLeftWidth",
    "borderRightWidth",
    "marginLeft",
    "marginRight",
  ];
  const measurementProperties = [
    ...fontProperties,
    ...edgeProperties,
    "borderLeftStyle",
    "borderRightStyle",
    "whiteSpace",
    "verticalAlign",
  ];

  function contentWidth(borderBoxWidth, style) {
    const edges = [
      "paddingLeft",
      "paddingRight",
      "borderLeftWidth",
      "borderRightWidth",
    ].map((property) => Number.parseFloat(style[property]));
    if (
      !Number.isFinite(borderBoxWidth) ||
      edges.some((value) => !Number.isFinite(value) || value < 0)
    )
      return Number.NaN;
    return borderBoxWidth - edges.reduce((total, value) => total + value, 0);
  }

  function fontFamily(style) {
    return style.fontFamily
      .split(",")[0]
      .replaceAll('"', "")
      .trim()
      .toLowerCase();
  }

  function pixels(value) {
    return value === "normal" ? 0 : Number.parseFloat(value);
  }

  function normalTextStyle(style, expected = {}) {
    return (
      style.fontWeight === (expected.weight ?? "400") &&
      style.fontStyle === (expected.style ?? "normal") &&
      ["normal", "100%"].includes(style.fontStretch) &&
      style.fontKerning === "normal" &&
      style.fontVariantLigatures === "normal" &&
      style.fontFeatureSettings === "normal" &&
      style.fontVariationSettings === "normal" &&
      Math.abs(pixels(style.letterSpacing) - (expected.letterSpacing ?? 0)) <
        0.01 &&
      style.textTransform === (expected.transform ?? "none") &&
      ["normal", "0px"].includes(style.wordSpacing)
    );
  }

  function inlineStyleMatches(style, expected) {
    return (
      ["article sans", "article mono"].includes(expected.family) &&
      ["400", "700"].includes(expected.weight) &&
      ["normal", "italic"].includes(expected.style) &&
      (expected.family !== "article mono" ||
        (expected.weight === "400" && expected.style === "normal")) &&
      ["none", "uppercase", "lowercase", "capitalize"].includes(
        expected.transform,
      ) &&
      Number.isFinite(expected.size) &&
      expected.size > 0 &&
      Number.isFinite(expected.letterSpacing) &&
      fontFamily(style) === expected.family &&
      Math.abs(Number.parseFloat(style.fontSize) - expected.size) < 0.01 &&
      normalTextStyle(style, expected) &&
      edgeProperties.every(
        (property) =>
          Number.isFinite(expected[property]) &&
          ((!property.startsWith("padding") &&
            !property.startsWith("border")) ||
            expected[property] >= 0) &&
          Math.abs(Number.parseFloat(style[property]) - expected[property]) <
            0.01,
      )
    );
  }

  function fontFaceMatches(face, expected) {
    const weight = { normal: "400", bold: "700" }[face.weight] ?? face.weight;
    return (
      face.family.replaceAll('"', "").trim().toLowerCase() ===
        expected.family &&
      weight === expected.weight &&
      face.style === expected.style &&
      face.status === "loaded"
    );
  }

  function fontsReady(faces, expected) {
    return expected.every((font) =>
      faces.some((face) => fontFaceMatches(face, font)),
    );
  }

  function imageStyleMatches(width, style, expected) {
    return (
      Number.isFinite(expected.width) &&
      expected.width > 0 &&
      Math.abs(width - expected.width) < 0.01 &&
      edgeProperties.every(
        (property) => Number.parseFloat(style[property]) === 0,
      )
    );
  }

  function assessPlan(geometry, lines) {
    const fontSize = Number.parseFloat(geometry.style.fontSize);
    if (
      ![fontSize, geometry.size, geometry.width, geometry.plannedWidth].every(
        (value) => Number.isFinite(value) && value > 0,
      )
    ) {
      return { valid: false, reason: "invalid geometry" };
    }
    if (!geometry.fontReady)
      return { valid: false, reason: "font unavailable" };
    const style = geometry.style;
    if (fontFamily(style) !== "article sans") {
      return { valid: false, reason: "font changed" };
    }
    if (Math.abs(fontSize - geometry.size) > 0.01) {
      return { valid: false, reason: "text size changed" };
    }
    if (
      !["400", "700"].includes(geometry.weight ?? "400") ||
      !["normal", "italic"].includes(geometry.fontStyle ?? "normal") ||
      !normalTextStyle(style, {
        weight: geometry.weight,
        style: geometry.fontStyle,
      }) ||
      style.textIndent !== "0px" ||
      geometry.inlineMismatch
    ) {
      return { valid: false, reason: "text styling changed" };
    }
    if (
      !Number.isFinite(geometry.width) ||
      Math.abs(geometry.width - geometry.plannedWidth) >
        (geometry.widthTolerance ?? tolerance)
    ) {
      return { valid: false, reason: "column width changed" };
    }
    if (lines === undefined) return { valid: true, adjustments: [] };
    if (!lines.length) return { valid: false, reason: "line plan missing" };
    const adjustments = [];
    for (const line of lines) {
      if (
        !Number.isFinite(line.actual) ||
        !Number.isFinite(line.expected) ||
        Math.abs(line.actual - line.expected) > tolerance
      ) {
        return { valid: false, reason: "font metrics changed" };
      }
      const delta = geometry.width - line.actual;
      if (
        !Number.isInteger(line.spaces) ||
        line.spaces < 0 ||
        (delta < -tolerance && line.spaces === 0)
      ) {
        return { valid: false, reason: "line cannot fit" };
      }
      const adjustment =
        line.spaces && (!line.final || delta < 0) ? delta / line.spaces : 0;
      if (adjustment < -geometry.size * 0.0825 - 0.05) {
        return { valid: false, reason: "line cannot fit" };
      }
      adjustments.push(adjustment);
    }
    return { valid: true, adjustments };
  }

  function sourceText(text) {
    return text.replaceAll("\u00ad", "");
  }

  function selectWidthPlan(bundle, width) {
    if (!Number.isFinite(width) || width <= 0) return null;
    const plannedWidth = Math.floor(width);
    if (!Object.hasOwn(bundle.widths, plannedWidth)) return null;
    const pathID = bundle.widths[plannedWidth];
    const path = bundle.paths[pathID];
    if (!Number.isInteger(pathID) || !Array.isArray(path) || !path.length) {
      throw new Error("Invalid paragraph plan path");
    }
    return { width: plannedWidth, pathID, path };
  }

  function lineSegments(bundle, selection, text, atoms = []) {
    const insideAtom = (offset) =>
      atoms.some((atom) => atom.start <= offset && offset < atom.end);
    let previousEnd = 0;
    const segments = selection.path.map((id, index) => {
      const line = bundle.lines[id];
      if (
        !Number.isInteger(id) ||
        !line ||
        line.start !== previousEnd ||
        !Number.isInteger(line.end) ||
        line.end <= line.start ||
        line.end > text.length ||
        !Number.isFinite(line.natural) ||
        line.natural < 0 ||
        !Number.isInteger(line.spaces) ||
        line.spaces < 0
      ) {
        throw new Error("Invalid paragraph line range");
      }
      if (atoms.some((atom) => atom.start < line.end && line.end < atom.end))
        throw new Error("Planned line splits an atomic inline box");
      let visibleStart = line.start;
      let visibleEnd = line.end;
      while (
        text[visibleStart] === " " &&
        visibleStart < visibleEnd &&
        !insideAtom(visibleStart)
      )
        visibleStart++;
      while (
        visibleEnd > visibleStart &&
        (text[visibleEnd - 1] === "\n" ||
          (text[visibleEnd - 1] === " " && !insideAtom(visibleEnd - 1)))
      )
        visibleEnd--;
      if (line.hyphen) {
        if (text[visibleEnd - 1] !== "\u00ad")
          throw new Error("Missing discretionary hyphen at chosen break");
        visibleEnd--;
      }
      if (visibleEnd <= visibleStart) throw new Error("Empty paragraph line");
      previousEnd = line.end;
      return {
        ...line,
        visibleStart,
        visibleEnd,
        final: line.ragged === true || index === selection.path.length - 1,
        isLast: index === selection.path.length - 1,
      };
    });
    if (previousEnd !== text.length)
      throw new Error("Incomplete paragraph plan");
    return segments;
  }

  if (typeof module !== "undefined")
    module.exports = {
      assessPlan,
      sourceText,
      selectWidthPlan,
      lineSegments,
      contentWidth,
      inlineStyleMatches,
      fontFaceMatches,
      fontsReady,
      imageStyleMatches,
    };
  if (typeof document === "undefined") return;

  const pendingScopes = new Set();
  let framePending = false;
  let layoutRevision = 0;

  function flushPlans() {
    framePending = false;
    const revision = layoutRevision;
    const scopes = [...pendingScopes];
    pendingScopes.clear();
    const plans = scopes.map((validate) => validate()).filter(Boolean);
    const probe = document.createElement("div");
    probe.setAttribute("aria-hidden", "true");
    Object.assign(probe.style, {
      position: "absolute",
      visibility: "hidden",
      left: "-100000px",
      top: "0",
      width: "max-content",
    });
    // Share layout work across scopes instead of alternating reads and writes per paragraph.
    for (const plan of plans) plan.run("rewrite");
    for (const plan of plans) plan.run("prepare", { probe, revision });
    if (probe.childElementCount) {
      document.body.append(probe);
      try {
        for (const plan of plans) plan.run("measure");
      } finally {
        probe.remove();
      }
    }
    for (const plan of plans) plan.run("commit");
  }

  function readStyle(node) {
    const style = getComputedStyle(node);
    return Object.fromEntries(
      [...measurementProperties, "wordSpacing", "textIndent"].map(
        (property) => [property, style[property]],
      ),
    );
  }

  function removePlanMarkup(paragraph) {
    for (const generated of paragraph.querySelectorAll(
      ".planned-break, .planned-hyphen",
    ))
      generated.remove();
    for (const wrapper of paragraph.querySelectorAll(
      ".planned-line, .planned-gap",
    ))
      wrapper.replaceWith(...wrapper.childNodes);
    paragraph.normalize();
  }

  function sourceNodes(paragraph) {
    const walker = document.createTreeWalker(
      paragraph,
      NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT,
      {
        acceptNode: (node) =>
          node.nodeType === Node.TEXT_NODE ||
          node.nodeName === "IMG" ||
          (node.nodeName === "BR" && !node.classList.contains("planned-break"))
            ? NodeFilter.FILTER_ACCEPT
            : NodeFilter.FILTER_SKIP,
      },
    );
    const nodes = [];
    let offset = 0;
    while (walker.nextNode()) {
      const node = walker.currentNode;
      const text =
        node.nodeName === "IMG"
          ? "\ufffc"
          : node.nodeName === "BR"
            ? "\n"
            : node.textContent;
      nodes.push({ node, text, start: offset, end: offset + text.length });
      offset += text.length;
    }
    return nodes;
  }

  function outermostAtom(piece, paragraph) {
    let atom = piece;
    for (
      let parent = piece.parentElement;
      parent !== paragraph;
      parent = parent.parentElement
    ) {
      if (!parent) throw new Error("Planned piece escaped its block");
      if (parent.hasAttribute("data-plan-atom")) atom = parent;
    }
    return atom;
  }

  function atomicRanges(nodes, paragraph) {
    const atoms = new Map();
    for (const source of nodes) {
      if (source.node.nodeName === "BR") continue;
      for (
        let ancestor =
          source.node.nodeType === Node.TEXT_NODE
            ? source.node.parentElement
            : source.node;
        ancestor !== paragraph;
        ancestor = ancestor.parentElement
      ) {
        if (!ancestor) throw new Error("Planned source escaped its block");
        if (ancestor.hasAttribute("data-plan-atom")) {
          const range = atoms.get(ancestor);
          if (range) range.end = source.end;
          else atoms.set(ancestor, { start: source.start, end: source.end });
        }
      }
    }
    return [...atoms.values()];
  }

  function rewriteParagraph(paragraph, bundle, selection) {
    removePlanMarkup(paragraph);
    const nodes = sourceNodes(paragraph);
    const segments = lineSegments(
      bundle,
      selection,
      nodes.map((source) => source.text).join(""),
      atomicRanges(nodes, paragraph),
    );
    const pieces = segments.map(() => []);
    for (const source of nodes) {
      const fragment = document.createDocumentFragment();
      let nodeWrapper;
      function append(start, end, className, index) {
        const from = Math.max(start, source.start);
        const to = Math.min(end, source.end);
        if (from >= to) return;
        const span = document.createElement("span");
        span.className = className;
        if (source.node.nodeName === "IMG" || source.node.nodeName === "BR") {
          if (source.node.nodeName === "BR" && className !== "planned-gap")
            throw new Error("Original break must end a planned line");
          nodeWrapper = span;
        } else
          span.textContent = source.text.slice(
            from - source.start,
            to - source.start,
          );
        if (className === "planned-line") {
          span.dataset.line = index;
          pieces[index].push(span);
        }
        fragment.append(span);
      }
      for (const [index, line] of segments.entries()) {
        if (line.start >= source.end) break;
        if (line.end <= source.start) continue;
        append(line.start, line.visibleStart, "planned-gap", index);
        append(line.visibleStart, line.visibleEnd, "planned-line", index);
        append(line.visibleEnd, line.end, "planned-gap", index);
      }
      source.node.replaceWith(fragment);
      if (nodeWrapper) nodeWrapper.append(source.node);
    }
    for (const [index, line] of segments.entries()) {
      const first = pieces[index][0];
      const last = pieces[index].at(-1);
      if (!first || !last) throw new Error("Planned line has no text pieces");
      first.dataset.natural = line.natural;
      first.dataset.spaces = line.spaces;
      if (line.final) first.dataset.final = "true";
      if (line.hyphen) {
        const dash = document.createElement("span");
        dash.className = "planned-hyphen";
        dash.setAttribute("aria-hidden", "true");
        last.append(dash);
      }
      if (!line.isLast) {
        const br = document.createElement("span");
        br.className = "planned-break";
        br.setAttribute("aria-hidden", "true");
        outermostAtom(last, paragraph).after(br);
      }
    }
  }

  function prepareLines(paragraph, style, styles) {
    const probe = document.createElement("div");
    Object.assign(probe.style, {
      whiteSpace: "nowrap",
      width: "max-content",
      padding: "0",
      border: "0",
      wordSpacing: "0px",
      textIndent: "0px",
      hyphens: "none",
    });
    for (const property of fontProperties)
      probe.style[property] = style[property];
    function sourceStyle(node) {
      if (!styles.has(node)) styles.set(node, readStyle(node));
      return styles.get(node);
    }
    const lines = [];
    const piecesByLine = new Map();
    for (const piece of paragraph.querySelectorAll("[data-line]")) {
      const key = piece.dataset.line;
      if (!piecesByLine.has(key)) piecesByLine.set(key, []);
      piecesByLine.get(key).push(piece);
    }
    for (const first of paragraph.querySelectorAll("[data-natural]")) {
      const pieces = piecesByLine.get(first.dataset.line) ?? [];
      const row = document.createElement("div");
      row.style.width = "max-content";
      const ancestors = new Map();
      for (const piece of pieces) {
        const ancestry = [];
        for (
          let ancestor = piece.parentElement;
          ancestor !== paragraph;
          ancestor = ancestor.parentElement
        ) {
          if (!ancestor) throw new Error("Planned piece escaped its block");
          ancestry.push(ancestor);
        }
        let parent = row;
        for (const ancestor of ancestry.reverse()) {
          let box = ancestors.get(ancestor);
          if (!box) {
            box = ancestor.cloneNode(false);
            box.removeAttribute("id");
            const ancestorStyle = sourceStyle(ancestor);
            for (const property of measurementProperties)
              box.style[property] = ancestorStyle[property];
            box.style.whiteSpace = ancestor.closest("[data-plan-atom]")
              ? "pre"
              : "nowrap";
            box.style.wordSpacing = "0px";
            ancestors.set(ancestor, box);
            parent.append(box);
          }
          parent = box;
        }
        const copy = piece.cloneNode(true);
        const pieceStyle = sourceStyle(piece);
        for (const property of measurementProperties)
          copy.style[property] = pieceStyle[property];
        copy.style.whiteSpace = piece.closest("[data-plan-atom]")
          ? "pre"
          : "nowrap";
        copy.style.wordSpacing = "0px";
        for (const dash of copy.querySelectorAll(".planned-hyphen"))
          dash.style.display = "inline";
        const sourceImages = piece.querySelectorAll("img");
        for (const [index, image] of [
          ...copy.querySelectorAll("img"),
        ].entries()) {
          image.removeAttribute("id");
          const imageStyle = getComputedStyle(sourceImages[index]);
          for (const property of [
            ...measurementProperties,
            "width",
            "height",
            "minWidth",
            "maxWidth",
            "boxSizing",
          ])
            image.style[property] = imageStyle[property];
          image.style.display = "inline-block";
          image.style.maxWidth = "none";
        }
        parent.append(copy);
      }
      probe.append(row);
      lines.push({
        row,
        pieces,
        expected: Number(first.dataset.natural),
        spaces: Number(first.dataset.spaces),
        final: first.dataset.final === "true",
      });
    }
    return { probe, lines };
  }

  function inlineContracts(paragraph) {
    return [...paragraph.querySelectorAll("[data-plan-inline]")].map(
      (node) => ({
        node,
        expected: JSON.parse(node.dataset.planInline),
      }),
    );
  }

  function requiredFonts(paragraph, contracts) {
    return [
      {
        family: "article sans",
        weight: paragraph.dataset.planWeight ?? "400",
        style: paragraph.dataset.planStyle ?? "normal",
      },
      ...contracts
        .filter(({ node }) => node.nodeName !== "IMG")
        .map(({ expected }) => expected),
    ];
  }

  function setupScope(scope) {
    const comparison = scope.closest(".typography-comparison");
    const status = comparison?.querySelector(".typography-status");
    const paragraphs = [...scope.querySelectorAll("[data-plan-width]")].filter(
      (paragraph) => paragraph.closest("[data-plan-lines]") === scope,
    );
    const entries = paragraphs.map((paragraph) => ({
      paragraph,
      bundle: null,
      pathID: null,
    }));
    let failed = false;
    let selectionDeferred = false;

    function fail(error) {
      failed = true;
      layoutRevision++;
      scope.classList.remove("planned-layout");
      if (status)
        status.textContent = "Responsive wrapping: layout check failed.";
      console.warn(
        "Article layout validation failed; using responsive wrapping.",
        error,
      );
    }

    function validate() {
      if (failed) return null;
      try {
        const plan = preparePlan();
        return {
          run(stage, probe) {
            if (failed) return;
            try {
              plan[stage](probe);
            } catch (error) {
              fail(error);
            }
          },
        };
      } catch (error) {
        fail(error);
        return null;
      }
    }

    function readGeometry(entry) {
      const { paragraph } = entry;
      const style = readStyle(paragraph);
      const contracts = inlineContracts(paragraph);
      const width = contentWidth(
        paragraph.getBoundingClientRect().width,
        style,
      );
      return {
        fontReady:
          Boolean(document.fonts) &&
          fontsReady([...document.fonts], requiredFonts(paragraph, contracts)),
        style,
        size: Number(paragraph.dataset.planSize),
        weight: paragraph.dataset.planWeight ?? "400",
        fontStyle: paragraph.dataset.planStyle ?? "normal",
        width,
        plannedWidth: Math.floor(width),
        widthTolerance: 1,
        inlineMismatch: contracts.some(({ node, expected }) => {
          const inlineStyle = readStyle(node);
          return !(node.nodeName === "IMG"
            ? imageStyleMatches(
                node.getBoundingClientRect().width,
                inlineStyle,
                expected,
              )
            : inlineStyleMatches(inlineStyle, expected));
        }),
      };
    }

    function preparePlan() {
      const selected =
        !comparison ||
        comparison.querySelector(".typography-mode-planned").checked;
      let result = { valid: false, reason: "native mode selected" };
      const candidates = [];
      const measurements = [];
      const prepared = [];
      function reject(entry, error) {
        result = { valid: false, reason: "line plan invalid" };
        if (entry.error !== error.message) {
          entry.error = error.message;
          console.warn(
            "Article block layout validation failed; using responsive wrapping.",
            error,
          );
        }
      }
      if (selected && paragraphs.length) {
        for (const entry of entries) {
          try {
            const { paragraph } = entry;
            const geometry = readGeometry(entry);
            result = assessPlan(geometry);
            if (!result.valid) continue;
            if (entry.bundle === null) {
              entry.bundle = JSON.parse(paragraph.dataset.linePlans);
              entry.pathID = entry.bundle.widths[paragraph.dataset.planWidth];
            }
            const selection = selectWidthPlan(entry.bundle, geometry.width);
            if (!selection) {
              result = { valid: false, reason: "column width not planned" };
              continue;
            }
            candidates.push({ entry, selection, geometry, ready: false });
          } catch (error) {
            reject(entry, error);
          }
        }
      }
      return {
        rewrite() {
          for (const candidate of candidates) {
            const { entry, selection } = candidate;
            const { paragraph } = entry;
            try {
              if (entry.pathID !== selection.pathID) {
                const textSelection = window.getSelection();
                if (
                  textSelection &&
                  !textSelection.isCollapsed &&
                  [...Array(textSelection.rangeCount).keys()].some((index) =>
                    textSelection.getRangeAt(index).intersectsNode(paragraph),
                  )
                ) {
                  selectionDeferred = true;
                  result = { valid: false, reason: "text selection active" };
                  continue;
                }
                layoutRevision++;
                paragraph.classList.remove("planned-paragraph");
                rewriteParagraph(paragraph, entry.bundle, selection);
                entry.pathID = selection.pathID;
              }
              if (paragraph.dataset.planWidth !== String(selection.width)) {
                layoutRevision++;
                paragraph.dataset.planWidth = selection.width;
              }
              candidate.ready = true;
            } catch (error) {
              reject(entry, error);
            }
          }
        },
        prepare({ probe, revision }) {
          const styles = new Map();
          for (const {
            entry,
            selection,
            ready,
            geometry: originalGeometry,
          } of candidates) {
            if (!ready) continue;
            try {
              const geometry =
                layoutRevision !== revision
                  ? readGeometry(entry)
                  : originalGeometry;
              result = assessPlan(geometry);
              if (!result.valid) continue;
              const current = selectWidthPlan(entry.bundle, geometry.width);
              if (
                !current ||
                current.pathID !== selection.pathID ||
                current.width !== selection.width
              ) {
                result = { valid: false, reason: "column width changed" };
                schedule();
                continue;
              }
              const measurement = prepareLines(
                entry.paragraph,
                geometry.style,
                styles,
              );
              probe.append(measurement.probe);
              measurements.push({ entry, geometry, lines: measurement.lines });
            } catch (error) {
              reject(entry, error);
            }
          }
        },
        measure() {
          for (const measurement of measurements) {
            try {
              for (const line of measurement.lines)
                line.actual = line.row.getBoundingClientRect().width;
              result = assessPlan(measurement.geometry, measurement.lines);
              if (!result.valid) continue;
              prepared.push({
                ...measurement,
                adjustments: result.adjustments,
              });
              measurement.entry.error = null;
            } catch (error) {
              reject(measurement.entry, error);
            }
          }
        },
        commit() {
          for (const { lines, adjustments } of prepared) {
            for (const [index, line] of lines.entries()) {
              const spacing = `${adjustments[index]}px`;
              for (const piece of line.pieces)
                if (piece.style.getPropertyValue("--plan-spacing") !== spacing)
                  piece.style.setProperty("--plan-spacing", spacing);
            }
          }
          const active = new Set(prepared.map(({ entry }) => entry.paragraph));
          for (const paragraph of paragraphs)
            paragraph.classList.toggle(
              "planned-paragraph",
              active.has(paragraph),
            );
          scope.classList.toggle("planned-layout", active.size > 0);
          if (status) {
            const message = active.size
              ? `Planned lines active in ${prepared.length} prose blocks at ${[...new Set(prepared.map(({ geometry }) => geometry.plannedWidth))].join(", ")} px.${prepared.length < paragraphs.length ? ` ${paragraphs.length - prepared.length} ${paragraphs.length - prepared.length === 1 ? "uses" : "use"} native wrapping.` : ""}`
              : `Responsive wrapping: ${result.reason}.`;
            if (status.textContent !== message) status.textContent = message;
          }
        },
      };
    }

    function schedule() {
      pendingScopes.add(validate);
      if (!framePending) {
        framePending = true;
        requestAnimationFrame(flushPlans);
      }
    }

    comparison?.addEventListener("change", (event) => {
      if (event.target.matches(".typography-size")) {
        scope.style.fontSize = event.target.value
          ? `${event.target.value}px`
          : "";
      }
      schedule();
    });
    if (typeof ResizeObserver !== "undefined") {
      let previousGeometry;
      new ResizeObserver(() => {
        const style = getComputedStyle(paragraphs[0] || scope);
        const geometry = `${scope.getBoundingClientRect().width}:${fontProperties.map((property) => style[property]).join(":")}`;
        if (geometry !== previousGeometry) {
          previousGeometry = geometry;
          schedule();
        }
      }).observe(scope);
    }
    if (typeof MutationObserver !== "undefined") {
      const changes = new MutationObserver(schedule);
      changes.observe(scope, {
        attributes: true,
        subtree: true,
        attributeFilter: ["class", "style"],
      });
      changes.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ["class", "style"],
      });
      changes.observe(document.body, {
        attributes: true,
        attributeFilter: ["class", "style"],
      });
      changes.observe(document.head, {
        attributes: true,
        childList: true,
        characterData: true,
        subtree: true,
      });
    }
    window.addEventListener("resize", schedule, { passive: true });
    window.addEventListener("pageshow", schedule);
    window.addEventListener("focus", schedule);
    document.addEventListener("selectionchange", () => {
      if (selectionDeferred && window.getSelection()?.isCollapsed) {
        selectionDeferred = false;
        schedule();
      }
    });
    document.addEventListener("visibilitychange", schedule);
    if (!document.fonts) {
      schedule();
      return;
    }
    document.fonts.addEventListener("loading", () => {
      schedule();
    });
    document.fonts.addEventListener("loadingdone", schedule);
    document.fonts.addEventListener("loadingerror", schedule);
    const requests = new Map();
    for (const { paragraph } of entries) {
      try {
        for (const font of requiredFonts(paragraph, inlineContracts(paragraph)))
          requests.set(`${font.family}:${font.weight}:${font.style}`, font);
      } catch (error) {
        console.warn(
          "Article block font contract is invalid; using responsive wrapping.",
          error,
        );
      }
    }
    for (const font of requests.values()) {
      document.fonts
        .load(`${font.style} ${font.weight} 16px "${font.family}"`)
        .then((faces) => {
          if (!fontsReady(faces, [font]))
            console.warn(
              "Article font face unavailable; using responsive wrapping.",
              font,
            );
          schedule();
        })
        .catch((error) => {
          console.warn(
            "Article font unavailable; using responsive wrapping.",
            error,
          );
          schedule();
        });
    }
    schedule();
  }

  function restoreClipboard(event) {
    const selection = window.getSelection();
    if (!event.clipboardData || !selection?.rangeCount) return;
    const range = selection.getRangeAt(0);
    const fragment = range.cloneContents();
    const ancestor =
      range.commonAncestorContainer.nodeType === Node.ELEMENT_NODE
        ? range.commonAncestorContainer
        : range.commonAncestorContainer.parentElement;
    if (
      !ancestor?.closest("[data-hyphenate-lang]") &&
      !fragment.querySelector(
        ".planned-line, .planned-gap, .planned-break, .planned-hyphen, .planned-prose-block, [data-plan-inline]",
      )
    )
      return;
    for (const generated of fragment.querySelectorAll(
      ".planned-break, .planned-hyphen",
    ))
      generated.remove();
    for (const wrapper of fragment.querySelectorAll(
      ".planned-line, .planned-gap, .planned-prose-block",
    ))
      wrapper.replaceWith(...wrapper.childNodes);
    const walker = document.createTreeWalker(fragment, NodeFilter.SHOW_TEXT);
    while (walker.nextNode())
      walker.currentNode.textContent = sourceText(
        walker.currentNode.textContent,
      );
    for (const node of fragment.querySelectorAll(
      "[data-plan-width], [data-plan-inline], [data-plan-atom]",
    )) {
      node.classList.remove("planned-paragraph");
      for (const attribute of [
        "data-plan-width",
        "data-plan-size",
        "data-plan-weight",
        "data-plan-style",
        "data-line-plans",
        "data-plan-inline",
        "data-plan-atom",
      ])
        node.removeAttribute(attribute);
    }
    for (const scope of fragment.querySelectorAll("[data-plan-lines]")) {
      scope.classList.remove("planned-layout", "typography-prose");
      for (const attribute of [
        "data-plan-lines",
        "data-plan-generated",
        "data-hyphenate-lang",
      ])
        scope.removeAttribute(attribute);
    }
    function plain(node) {
      if (node.nodeType === Node.TEXT_NODE) return node.textContent;
      if (node.nodeName === "BR") return "\n";
      const text = [...node.childNodes].map(plain).join("");
      if (/^(TH|TD)$/.test(node.nodeName)) return `${text}\t`;
      return /^(P|LI|H[1-6]|PRE|BLOCKQUOTE|FIGURE|FIGCAPTION|DT|DD|TR)$/.test(
        node.nodeName,
      )
        ? `\n${text}\n`
        : text;
    }
    const text = plain(fragment)
      .replace(/\n{3,}/g, "\n\n")
      .replace(/^\n+|\n+$/g, "");
    const container = document.createElement("div");
    container.append(fragment);
    event.clipboardData.setData("text/plain", text);
    event.clipboardData.setData("text/html", container.innerHTML);
    event.preventDefault();
  }

  for (const scope of document.querySelectorAll("[data-plan-lines]"))
    setupScope(scope);
  document.addEventListener("copy", restoreClipboard);
})();
