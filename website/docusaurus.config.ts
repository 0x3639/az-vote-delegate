import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'az-vote-delegate',
  tagline: 'Copy pillar votes. Keep control of your own.',
  url: 'https://0x3639.github.io',
  baseUrl: '/az-vote-delegate/',
  trailingSlash: true,
  favicon: 'img/az.svg',
  organizationName: '0x3639',
  projectName: 'az-vote-delegate',
  onBrokenLinks: 'throw',
  onBrokenAnchors: 'throw',
  markdown: {hooks: {onBrokenMarkdownLinks: 'throw'}},
  i18n: {defaultLocale: 'en', locales: ['en']},
  presets: [['classic', {
    docs: {
      sidebarPath: './sidebars.ts',
      routeBasePath: '/',
      editUrl: 'https://github.com/0x3639/az-vote-delegate/edit/docusaurus/website/',
    },
    blog: false,
    theme: {customCss: './src/css/custom.css'},
    sitemap: {changefreq: 'weekly', priority: 0.6},
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
    metadata: [{name: 'description', content: 'Install az-vote-delegate on your Zenon pillar, choose a pillar to copy, and manage AZ proposal and phase-payout votes from a terminal menu.'}],
    colorMode: {defaultMode: 'dark', respectPrefersColorScheme: false},
    navbar: {
      title: 'az-vote-delegate',
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
        {title: 'Project', items: [{label: 'Command reference', to: '/commands'}, {label: 'GitHub', href: 'https://github.com/0x3639/az-vote-delegate'}, {label: 'Releases', href: 'https://github.com/0x3639/az-vote-delegate/releases'}]},
      ],
      copyright: 'az-vote-delegate · Community tooling for Zenon, Network of Momentum · GPL-3.0',
    },
    prism: {theme: prismThemes.github, darkTheme: prismThemes.dracula, additionalLanguages: ['bash', 'json', 'ini']},
  } satisfies Preset.ThemeConfig,
};
export default config;
