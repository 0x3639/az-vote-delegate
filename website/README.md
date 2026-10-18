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
the standalone glyph uses the brand green fill. The upstream MIT license is
preserved in `vendor/zenon-design-system/LICENSE`. Space Grotesk and JetBrains
Mono are bundled through their Fontsource packages. Site code and documentation
use the repository's GPL-3.0 license.
