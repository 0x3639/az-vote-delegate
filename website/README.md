# Documentation site

Docusaurus documentation for az-vote-delegate. Site source and the Pages workflow
live on the `docusaurus` branch; application releases continue from `main`.

Public URL: https://www.0x3639.com/az-vote-delegate/

## Develop

Use Node.js 22 or newer, then run from this directory:

```sh
npm ci
npm start
```

Before publishing:

```sh
npm run typecheck
npm run build
npm run serve
```

The production build checks internal links and generates the local search index.
Search is available in the production preview. Dependencies are locked in
`package-lock.json`; fonts are bundled locally.

## Search, social previews, and model access

Every guide has its own description and canonical URL, a large Open Graph / X
card, and TechArticle structured data. WebSite metadata identifies the source
repository and license. Search results and 404 pages use `noindex`; the sitemap
contains the homepage and guides. No search-engine submission or account is
needed to build these files.

`docusaurus-plugin-llms` generates these on every production build:

- [`llms.txt`](https://www.0x3639.com/az-vote-delegate/llms.txt): a compact index with links to all guides.
- [`llms-full.txt`](https://www.0x3639.com/az-vote-delegate/llms-full.txt): the complete guide text in one fetch.
- Individual Markdown copies such as [`getting-started.md`](https://www.0x3639.com/az-vote-delegate/getting-started.md).

HTML pages advertise the index with `rel="describedby"`; guides advertise their
Markdown counterpart with `rel="alternate" type="text/markdown"` and a visible
footer link. The regular footer links to both text indexes. New `.md` guides in
`docs/` are included automatically, and `includeOrder` sets the preferred order.
These exports follow the [llms.txt proposal](https://llmstxt.org/); they make the
content easier to retrieve but do not guarantee a model or search engine will
use it. Agents can fetch them directly over HTTPS without executing JavaScript.

`npm run build` also runs `scripts/verify-discovery.mjs`. It checks that every
source guide survives intact in both exports, generated links resolve, canonical
URLs and schemas are correct, social assets exist, and utility pages stay out of
the index. Run `npm run verify:discovery` to recheck an existing build.

The effective robots file belongs at the **origin root**:
`https://www.0x3639.com/robots.txt`, controlled by the separate account Pages
repository. It returned 404 on 2026-10-18, so there were no robots-file crawl
restrictions. A file at `/az-vote-delegate/robots.txt` would not control crawling;
this project therefore does not publish a misleading substitute. If origin rules
are added later, permit `/az-vote-delegate/` for the intended crawlers and add:

```text
Sitemap: https://www.0x3639.com/az-vote-delegate/sitemap.xml
```

The generated social artwork is `static/img/social-card.png` (1730 × 909, PNG),
created with the built-in image generation tool. Its exact prompt is preserved
in `assets/social-card-prompt.txt`. The favicon uses the authentic AZ glyph on a
dark tile; SVG, multi-size ICO, 32px PNG, and 180px Apple touch versions are
committed. To regenerate the raster icons after editing `static/img/favicon.svg`,
install ImageMagick and run `sh scripts/generate-favicons.sh`.

Package overrides keep Docusaurus's build dependencies on patched releases of
tinypool, serialize-javascript, postcss-selector-parser, and uuid. Revisit these
overrides when upgrading Docusaurus. The deployed site is static HTML, CSS, and
JavaScript; no Node.js server is published.

As of 2026-10-18, `npm audit` still reports the upstream
[braces stack-exhaustion advisory](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm)
through build-time file matching. The latest braces release is affected and has
no patched release. This dependency is not shipped to site visitors. Build only
trusted source, and review the advisory when updating the toolchain.

## Publish

Push changes to `docusaurus`. `.github/workflows/docs.yml` builds the site and
deploys it through GitHub Pages Actions. Pages must use the Actions deployment
source and permit the `docusaurus` branch in the `github-pages` environment.
No `gh-pages` output branch is required. The project inherits the account's
existing Pages domain, `www.0x3639.com`; it has no separate domain configuration.

The project URL requires `baseUrl: '/az-vote-delegate/'`. Keep links inside the
site base-path aware. If moving to a custom domain, update `url` and `baseUrl`
together and configure Pages for that domain.

## Design attribution

The color, radius, and elevation tokens under `src/css/zenon/`, plus the
Accelerator-Z glyph at `static/img/az.svg`, come from
[digitalSloth/zenon-design-system](https://github.com/digitalSloth/zenon-design-system)
at revision `8d2f76d74a286bd722f1fb62f280d89740ca29a8`.

The token dark selector is adapted to Docusaurus's `data-theme` attribute, and
the standalone glyph uses the brand green fill. The favicon adds a dark tile to
that same glyph. The upstream MIT license is
preserved in `vendor/zenon-design-system/LICENSE`. Space Grotesk and JetBrains
Mono are bundled through their Fontsource packages. Site code and documentation
use the repository's GPL-3.0 license.
