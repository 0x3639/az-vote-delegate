import assert from 'node:assert/strict';
import {readFile, readdir, stat} from 'node:fs/promises';
import path from 'node:path';

// Check the published artifacts, not only the configuration that generates them.
const root = path.resolve('build');
const site = 'https://www.0x3639.com/az-vote-delegate/';
const read = (name) => readFile(path.join(root, name), 'utf8');
const attrs = (tag) => Object.fromEntries([...tag.matchAll(/([\w:-]+)="([^"]*)"/g)].map((m) => [m[1], m[2]]));
const tags = (html, name) => [...html.matchAll(new RegExp(`<${name}\\b[^>]*>`, 'g'))].map((m) => attrs(m[0]));
const meta = (html, key) => tags(html, 'meta').filter((a) => a.name === key || a.property === key).map((a) => a.content);
const link = (html, rel) => tags(html, 'link').filter((a) => a.rel === rel);

async function checkTarget(rawUrl, relativeTo = site) {
  const url = new URL(rawUrl, relativeTo);
  if (!url.href.startsWith(site)) return;
  let relative = decodeURIComponent(url.pathname.slice(new URL(site).pathname.length));
  if (!relative || relative.endsWith('/')) relative += 'index.html';
  const target = path.resolve(root, relative);
  assert(target.startsWith(`${root}${path.sep}`), `Out-of-scope path: ${rawUrl}`);
  assert((await stat(target)).isFile(), `Missing target: ${rawUrl}`);
}

const index = await read('llms.txt');
const full = await read('llms-full.txt');
// The combined export nests each guide under an H2, shifting its headings.
const normalizeHeadings = (markdown) => markdown.replace(/^#{1,6} /gm, '# ');
assert(index.startsWith('# az-vote-delegate documentation\n'));
assert(full.length > index.length * 3, 'Full export is unexpectedly short');
const sitemap = await read('sitemap.xml');
assert(!sitemap.includes('/search/'), 'Search results must not be in the sitemap');
const routes = [...sitemap.matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1]);
assert(routes.includes(site), 'Homepage missing from sitemap');

const sourceDocs = (await readdir('docs')).filter((name) => name.endsWith('.md'));
for (const sourceName of sourceDocs) {
  const body = (await readFile(path.join('docs', sourceName), 'utf8')).replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, '').trim();
  const markdown = await read(sourceName);
  assert(markdown.includes(body), `Lost source content in ${sourceName}`);
  assert(normalizeHeadings(full).includes(normalizeHeadings(body)), `Lost source content in full bundle: ${sourceName}`);
  assert(index.includes(`${site}${sourceName}`), `Missing index entry: ${sourceName}`);
  assert(routes.includes(`${site}${sourceName.replace(/\.md$/, '')}/`), `Missing sitemap entry: ${sourceName}`);
  for (const match of markdown.matchAll(/\]\(([^\s)]+)(?:\s+"[^"]*")?\)/g)) {
    await checkTarget(match[1], `${site}${sourceName}`);
  }
}
for (const match of index.matchAll(/\]\(([^\s)]+)\)/g)) await checkTarget(match[1]);

for (const route of routes) {
  await checkTarget(route);
  const relative = route.slice(site.length);
  const html = await read(`${relative}index.html`);
  assert.equal(link(html, 'canonical')[0]?.href, route, `Canonical mismatch: ${route}`);
  assert.equal(meta(html, 'description').length, 1, `Description missing or duplicated: ${route}`);
  assert(meta(html, 'description')[0].length > 30);
  assert.equal(meta(html, 'og:url')[0], route);
  assert.equal(meta(html, 'og:image')[0], `${site}img/social-card.png`);
  assert.equal(meta(html, 'twitter:image')[0], `${site}img/social-card.png`);
  assert.equal(meta(html, 'twitter:card')[0], 'summary_large_image');
  assert(meta(html, 'og:image:alt')[0]);
  assert(!meta(html, 'robots').some((value) => value.includes('noindex')), `Guide blocked: ${route}`);
  assert.equal(link(html, 'describedby')[0]?.href, `${site}llms.txt`);
  for (const asset of [...link(html, 'icon'), ...link(html, 'apple-touch-icon')]) await checkTarget(asset.href);
  const schemas = [...html.matchAll(/<script\b[^>]*type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/g)].map((m) => JSON.parse(m[1]));
  assert(schemas.some((schema) => schema['@type'] === 'WebSite'), `Missing site schema: ${route}`);
  if (relative) {
    const markdown = link(html, 'alternate').find((a) => a.type === 'text/markdown');
    assert(markdown, `Missing Markdown discovery link: ${route}`);
    await checkTarget(markdown.href);
    assert(schemas.some((schema) => schema['@type'] === 'TechArticle' && schema.url === route));
  }
}
for (const page of ['search/index.html', '404.html']) {
  const html = await read(page);
  const robots = tags(html, 'meta').filter((a) => a.name === 'robots');
  assert(robots.length && robots.every((a) => a.content.includes('noindex')), `Utility page must be noindex: ${page}`);
}

const png = await readFile(path.join(root, 'img/social-card.png'));
assert.equal(png.subarray(1, 4).toString(), 'PNG');
assert.equal(png.readUInt32BE(16), 1730);
assert.equal(png.readUInt32BE(20), 909);
assert(png.length < 5_000_000, 'Social image exceeds 5 MB');
assert.equal(meta(await read('index.html'), 'og:image:width')[0], '1730');
assert.equal(meta(await read('index.html'), 'og:image:height')[0], '909');
console.log(`Verified ${routes.length} SEO pages, ${sourceDocs.length} complete Markdown exports, llms files, sitemap, schemas, and social assets.`);
