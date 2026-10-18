import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const siteUrl = 'https://www.0x3639.com/az-vote-delegate/';
const description = 'Install AZ Vote Delegate on your Zenon pillar, choose a pillar to copy, and manage AZ proposal and phase-payout votes from a terminal menu.';
const llmContext = `az-vote-delegate is a Linux terminal application and background service for Zenon Accelerator-Z voting, using a pillar producer key.
Dry run previews without submitting; Live submits votes. Manual overrides apply to individual proposal or phase request IDs and persist locally. Dynamic Plasma is not supported.
Start with Installation & setup for the install command, producer configuration, delegate selection, and voting mode.
Full documentation: [llms-full.txt](${siteUrl}llms-full.txt). Source and releases: [GitHub](https://github.com/0x3639/az-vote-delegate).`;

const config: Config = {
  title: 'AZ Vote Delegate',
  tagline: 'Copy pillar votes. Keep control of your own.',
  url: 'https://www.0x3639.com',
  baseUrl: '/az-vote-delegate/',
  trailingSlash: true,
  favicon: 'img/favicon.ico',
  organizationName: '0x3639',
  projectName: 'az-vote-delegate',
  onBrokenLinks: 'throw',
  onBrokenAnchors: 'throw',
  markdown: {hooks: {onBrokenMarkdownLinks: 'throw'}},
  i18n: {defaultLocale: 'en', locales: ['en']},
  headTags: [
    {tagName: 'link', attributes: {rel: 'icon', type: 'image/svg+xml', sizes: 'any', href: `${siteUrl}img/favicon.svg`}},
    {tagName: 'link', attributes: {rel: 'icon', type: 'image/png', sizes: '32x32', href: `${siteUrl}img/favicon-32.png`}},
    {tagName: 'link', attributes: {rel: 'apple-touch-icon', sizes: '180x180', href: `${siteUrl}img/apple-touch-icon.png`}},
    {tagName: 'link', attributes: {rel: 'describedby', type: 'text/plain', href: `${siteUrl}llms.txt`, title: 'Documentation index for language models'}},
    {tagName: 'link', attributes: {rel: 'sitemap', type: 'application/xml', href: `${siteUrl}sitemap.xml`}},
    {tagName: 'script', attributes: {type: 'application/ld+json'}, innerHTML: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'WebSite',
      '@id': `${siteUrl}#website`,
      name: 'AZ Vote Delegate documentation',
      url: siteUrl,
      description,
      inLanguage: 'en',
      about: {
        '@type': 'SoftwareSourceCode',
        name: 'AZ Vote Delegate',
        codeRepository: 'https://github.com/0x3639/az-vote-delegate',
        programmingLanguage: 'Go',
        license: 'https://www.gnu.org/licenses/gpl-3.0.html',
      },
    })},
  ],
  plugins: [['docusaurus-plugin-llms', {
    docsDir: [{path: 'docs', routeBasePath: '/', label: 'Guides'}],
    title: 'AZ Vote Delegate documentation',
    description,
    generateLLMsTxt: true,
    generateLLMsFullTxt: true,
    generateMarkdownFiles: true,
    includeOrder: ['getting-started.md', 'delegation.md', 'manual-voting.md', 'updates.md', 'service.md', 'troubleshooting.md', 'commands.md'],
    excludeImports: true,
    removeDuplicateHeadings: true,
    rootContent: llmContext,
    fullRootContent: llmContext,
  }]],
  presets: [['classic', {
    docs: {
      sidebarPath: './sidebars.ts',
      routeBasePath: '/',
      editUrl: 'https://github.com/0x3639/az-vote-delegate/edit/docusaurus/website/',
    },
    blog: false,
    theme: {customCss: './src/css/custom.css'},
    sitemap: {changefreq: 'weekly', priority: 0.6, ignorePatterns: ['**/search/**']},
  } satisfies Preset.Options]],
  themes: [['@easyops-cn/docusaurus-search-local', {
    hashed: 'filename',
    docsRouteBasePath: '/',
    indexBlog: false,
    indexPages: true,
    highlightSearchTermsOnTargetPage: true,
    searchResultLimits: 8,
    searchBarShortcutHint: true,
  }]],
  themeConfig: {
    image: 'img/social-card.jpg',
    metadata: [
      {name: 'theme-color', content: '#151515'},
      {name: 'twitter:card', content: 'summary_large_image'},
      {name: 'twitter:image:alt', content: 'az-vote-delegate — Copy pillar votes. Keep control of your own. Green connected pillars on a dark background.'},
      {property: 'og:site_name', content: 'AZ Vote Delegate documentation'},
      {property: 'og:type', content: 'website'},
      {property: 'og:locale', content: 'en_US'},
      {property: 'og:image:width', content: '1730'},
      {property: 'og:image:height', content: '909'},
      {property: 'og:image:type', content: 'image/jpeg'},
      {property: 'og:image:alt', content: 'az-vote-delegate — Copy pillar votes. Keep control of your own. Green connected pillars on a dark background.'},
    ],
    colorMode: {defaultMode: 'dark', respectPrefersColorScheme: false},
    navbar: {
      title: 'AZ Vote Delegate',
      logo: {alt: 'Accelerator-Z', src: 'img/az.svg'},
      items: [
        {to: '/', label: 'Quick start', position: 'left', activeBaseRegex: '^/az-vote-delegate/$'},
        {type: 'docSidebar', sidebarId: 'guide', label: 'Guides', position: 'left'},
        {to: '/troubleshooting', label: 'Troubleshooting', position: 'left'},
        {href: 'https://github.com/0x3639/az-vote-delegate/releases', label: 'Releases', position: 'right'},
        {href: 'https://github.com/0x3639/az-vote-delegate', label: 'GitHub', position: 'right'},
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {title: 'Get started', items: [{label: 'Quick start', to: '/'}, {label: 'Installation & setup', to: '/getting-started'}, {label: 'Choose a delegate', to: '/delegation'}]},
        {title: 'Manage your votes', items: [{label: 'Manual votes & phases', to: '/manual-voting'}, {label: 'Updates', to: '/updates'}, {label: 'Service & troubleshooting', to: '/service'}]},
        {title: 'Project', items: [{label: 'Command reference', to: '/commands'}, {label: 'For AI tools (llms.txt)', href: `${siteUrl}llms.txt`}, {label: 'All docs as text', href: `${siteUrl}llms-full.txt`}, {label: 'GitHub', href: 'https://github.com/0x3639/az-vote-delegate'}, {label: 'Releases', href: 'https://github.com/0x3639/az-vote-delegate/releases'}]},
      ],
      copyright: 'AZ Vote Delegate · Community tooling for Zenon, Network of Momentum · GPL-3.0',
    },
    prism: {theme: prismThemes.github, darkTheme: prismThemes.dracula, additionalLanguages: ['bash', 'json', 'ini']},
  } satisfies Preset.ThemeConfig,
};
export default config;
