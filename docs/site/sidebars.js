/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docsSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Introduction',
      items: ['intro/what-is-the-framework', 'intro/why-it-exists']
    },
    {
      type: 'category',
      label: 'Getting Started',
      items: ['getting-started/install-and-run', 'getting-started/first-project']
    },
    {
      type: 'category',
      label: 'Core Concepts',
      items: ['core-concepts/mental-model']
    },
    'theme-showcase'
  ]
};

module.exports = sidebars;
