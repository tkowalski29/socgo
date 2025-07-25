import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid({
  mermaid: {
    theme: 'default',
    securityLevel: 'loose',
    flowchart: {
      useMaxWidth: true,
      htmlLabels: true
    }
  },
  lang: 'pl-PL',
  title: 'SocGo',
  description: 'Dokumentacja systemu SocGo',
  
  // Base URL dla Netlify
  base: '/',
  
  // Clean URLs
  cleanUrls: true,
  
  themeConfig: {
    nav: [
      { text: 'Strona główna', link: '/' }
    ],
    
    sidebar: {},
    
    socialLinks: [
      { icon: 'github', link: 'https://github.com/tkowalski29/socgo' }
    ],
    
    editLink: {
      pattern: 'https://github.com/tkowalski29/socgo/edit/main/.doc/:path'
    },
    
    lastUpdated: true,
    
    search: {
      provider: 'local'
    }
  }
}) 