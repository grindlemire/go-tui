// @ts-check

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'go-tui Docs',
  tagline: 'Documentation site placeholder',
  url: 'https://example.vercel.app',
  baseUrl: '/',

  organizationName: 'grindlemire',
  projectName: 'go-tui',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en']
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          path: 'docs',
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js')
        },
        blog: false,
        pages: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css')
        }
      })
    ]
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'go-tui',
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docsSidebar',
            position: 'left',
            label: 'Docs'
          },
          {
            href: 'https://github.com/grindlemire/go-tui',
            label: 'GitHub',
            position: 'right'
          }
        ]
      },
      footer: {
        style: 'dark',
        copyright: `Copyright ${new Date().getFullYear()} go-tui`
      },
      prism: {
        additionalLanguages: ['go']
      }
    })
};

module.exports = config;
