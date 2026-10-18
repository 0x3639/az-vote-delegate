import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';
const sidebars: SidebarsConfig = {
  guide: [
    {type: 'link', label: 'Quick start', href: '/'},
    'getting-started',
    'delegation',
    'manual-voting',
    'updates',
    'service',
    'troubleshooting',
    'commands',
  ],
};
export default sidebars;
