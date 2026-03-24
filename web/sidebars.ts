import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'introduction',
    {
      type: 'category',
      label: 'Getting Started',
      items: ['getting-started/installation', 'getting-started/quickstart'],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/pushing',
        'guides/installing',
        'guides/teams',
        'guides/profiles',
        'guides/multi-tool',
      ],
    },
    {
      type: 'category',
      label: 'Server',
      items: ['server/deployment', 'server/configuration', 'server/api'],
    },
    {
      type: 'category',
      label: 'Reference',
      items: ['reference/cli', 'reference/architecture', 'reference/security'],
    },
  ],
};

export default sidebars;
