<!-- Record the pinned font, its permitted modification, and how the two shipped encodings stay metrically identical. -->

# Article font

Article Sans is a renamed Latin subset family from Source Sans 3 **3.052**, using the upstream release tree at commit `87b37a2daaed80fcb8e8ccb0085c4d72ddade12e`. Each face has its own matching static font; browser-synthesized bold and italic are not required:

- Source: https://github.com/adobe-fonts/source-sans/blob/87b37a2daaed80fcb8e8ccb0085c4d72ddade12e/TTF/SourceSans3-Regular.ttf
- Source SHA-256: `4644c81b86ec9caaa76b634889968ed3c4f4f52f054855933acc7c2b21e53b0f`
- Subset TTF SHA-256: `8b12c26fb72ac940aef4afdb2c24be54b257787b5f1bf212e83d5f8832a083e2`
- Subset WOFF2 SHA-256: `e91bcf06a065cee91f8e8e9036512b1a641b8df0721747d3d1a8e319bfc1c200`

The SIL Open Font License is retained in `LICENSE.md` here and in `static/fonts/`. Every modified face is renamed Article Sans because the upstream license reserves the name Source. The subsets retain Latin/Latin Extended-A and general punctuation, kerning and normal shaping features; they remove hinting and unused optional alternates. Missing characters use browser fallback fonts and are excluded from planned paragraphs by the Go builder.

The additional faces use these official upstream TTFs and hashes:

| Article Sans face       | Source file              | Source SHA-256                                                     | Subset TTF SHA-256                                                 | Subset WOFF2 SHA-256                                               | TTF / WOFF2 bytes |
| ----------------------- | ------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ | ----------------- |
| Bold, weight 700        | `SourceSans3-Bold.ttf`   | `9214b9d95e4231c609802815c2646c98174e2102d0d37f88978a7f8e71006e6a` | `38fe9f0083ce2d5542b2eab5852c9945d85967cbc26803e9926183febef979ff` | `5ae0982f197215bb08eb2a180a8faafca25abd17be9c7737eb8efe87a3cb08e5` | 36,820 / 14,888   |
| Italic, weight 400      | `SourceSans3-It.ttf`     | `192afd78f0f54a3c69eaf02d43f4d9a821e9d6110e41d3d25d61a7385cd580e4` | `4798837181e4588daff60144bbf799381cae8d8bfc979c0e4a336d8583528d30` | `10e2685b8cbaebb5d6dc9ef620501f84fb540bd461765dce5cfb243ce010711f` | 34,844 / 14,572   |
| Bold Italic, weight 700 | `SourceSans3-BoldIt.ttf` | `7978291fc1bf314db887e0366853b33c5cf2e964c7b95cfb9ce403a6ec46a842` | `54aa24ec729884755dc2865c60a3a14287a69a7873a0380e38c3e3450129747e` | `ff6fdee3d99c21d3fc9fa20646439420479b888d70bc0a8f58ac51f0bc86da2b` | 34,704 / 14,544   |

The source URLs are:

- https://github.com/adobe-fonts/source-sans/blob/87b37a2daaed80fcb8e8ccb0085c4d72ddade12e/TTF/SourceSans3-Bold.ttf
- https://github.com/adobe-fonts/source-sans/blob/87b37a2daaed80fcb8e8ccb0085c4d72ddade12e/TTF/SourceSans3-It.ttf
- https://github.com/adobe-fonts/source-sans/blob/87b37a2daaed80fcb8e8ccb0085c4d72ddade12e/TTF/SourceSans3-BoldIt.ttf

`scripts/subset_font.py` creates both shipped files from the same in-memory font and verifies identical advance, character-map, kerning and substitution tables after WOFF2 decoding. The Go helper embeds the TTF; the browser downloads the matching WOFF2. CSS has no `local()` source. Font generation is an occasional maintenance operation, not a build dependency.

Reproduce with Python 3.12, FontTools 4.60.1 and Brotli 1.1.0. The default command generates the regular face; use `--style bold`, `--style italic`, or `--style bold-italic` with the matching upstream TTF for the other faces:

```sh
python3 scripts/subset_font.py /path/to/SourceSans3-Regular.ttf
python3 scripts/subset_font.py --style bold /path/to/SourceSans3-Bold.ttf
python3 scripts/subset_font.py --style italic /path/to/SourceSans3-It.ttf
python3 scripts/subset_font.py --style bold-italic /path/to/SourceSans3-BoldIt.ttf
```

The generated binaries and license are committed with the site. Replacing the font requires regenerating both encodings together, updating this provenance, and rebuilding every plan. Browser measurements independently check the font's actual chosen-line widths before activating plans.

## Article Mono

Article Mono is a Latin subset of Source Code Pro Regular **2.042**, from the official Adobe release commit `d3f1a5962cde503f9409c21e58527611d4a19ef1`:

- Source: https://github.com/adobe-fonts/source-code-pro/blob/d3f1a5962cde503f9409c21e58527611d4a19ef1/TTF/SourceCodePro-Regular.ttf
- Source SHA-256: `74bd80d3e42a08517cd7e1108ba3d86f2da29ac0f3065be95e0357956ab9db37`
- Subset TTF SHA-256: `5560b16b9b509df819ef864a8f28d2fa85760e7741c5f2892257e86476eb2ba4`
- Subset WOFF2 SHA-256: `0d32462b36962a176688ded28caeb1ea4128ead1c2f0bf2b72f204897ee1a44d`

The upstream SIL Open Font License and copyright notice are retained in `ArticleMono-LICENSE.md` in both font directories. The modified font is renamed Article Mono because Adobe reserves the name Source. It retains the same Latin/Latin Extended-A and general punctuation ranges, kerning and shaping features as Article Sans, with hinting removed. It is one regular face; bold and italics are not embedded.

Reproduce with Python 3.12.14, FontTools 4.60.1 and Brotli 1.1.0, using the same bundled runtime and cached font tools as for Article Sans:

```sh
PYTHONPATH=.cache/font-tools /Users/alessandro/.cache/codex-runtimes/codex-primary-runtime/dependencies/python/bin/python3 scripts/subset_font.py --font mono /path/to/SourceCodePro-Regular.ttf
```

The generator writes the embedded TTF and browser WOFF2 from the same subset and verifies matching `hmtx`, `cmap`, `GPOS` and `GSUB` tables when present. Inline-code plans and browser measurements use this same family and metrics.
