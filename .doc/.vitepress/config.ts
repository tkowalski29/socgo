import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'SocGo Documentation',
  description: 'Platforma do zarządzania treściami w mediach społecznościowych',
  
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'API', link: '/api/' },
      { text: 'Architecture', link: '/architecture/' },
      { text: 'User Guide', link: '/guide/' },
      { text: 'Config', link: '/configuration/' }
    ],

    sidebar: {
      '/': [
        {
          text: 'Introduction',
          items: [
            { text: 'Overview', link: '/' },
            { text: 'Quick Start', link: '/getting-started' },
            { text: 'Installation', link: '/installation' }
          ]
        },
        {
          text: 'Configuration',
          items: [
            { text: 'Overview', link: '/configuration/' },
            { text: 'Config File', link: '/configuration/config-file' },
            { text: 'Environment Variables', link: '/configuration/environment' },
            { text: 'OAuth Setup', link: '/configuration/oauth' }
          ]
        },
        {
          text: 'Architecture',
          items: [
            { text: 'System Overview', link: '/architecture/' },
            { text: 'Database Design', link: '/architecture/database' },
            { text: 'Service Layer', link: '/architecture/services' },
            { text: 'Authentication', link: '/architecture/auth' }
          ]
        },
        {
          text: 'API Reference',
          items: [
            { text: 'Overview', link: '/api/' },
            { text: 'OAuth Endpoints', link: '/api/oauth' },
            { text: 'Posts Management', link: '/api/posts' },
            { text: 'Calendar & Scheduling', link: '/api/calendar' },
            { text: 'Notifications', link: '/api/notifications' },
            { text: 'Settings', link: '/api/settings' }
          ]
        },
        {
          text: 'Social Media Providers',
          items: [
            { text: 'Overview', link: '/providers/' },
            { text: 'Facebook Integration', link: '/providers/facebook' },
            { text: 'Instagram Integration', link: '/providers/instagram' },
            { text: 'TikTok Integration', link: '/providers/tiktok' }
          ]
        },
        {
          text: 'User Guide',
          items: [
            { text: 'Getting Started', link: '/guide/' },
            { text: 'Managing Posts', link: '/guide/posts' },
            { text: 'Scheduling Content', link: '/guide/scheduling' },
            { text: 'Provider Setup', link: '/guide/providers' },
            { text: 'Notifications', link: '/guide/notifications' }
          ]
        },
        {
          text: 'Development',
          items: [
            { text: 'Overview', link: '/development/' },
            { text: 'Contributing', link: '/development/contributing' },
            { text: 'Testing', link: '/development/testing' },
            { text: 'Deployment', link: '/development/deployment' },
            { text: 'Code Structure', link: '/development/code-structure' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/tkowalski/socgo' }
    ],

    search: {
      provider: 'local'
    }
  },

  markdown: {
    config: (md) => {
      md.use(require('markdown-it-mermaid').default)
    }
  }
})