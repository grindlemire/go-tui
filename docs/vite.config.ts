import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    chunkSizeWarningLimit: 900,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('/shiki/') || id.includes('/@shikijs/')) return 'shiki';
            if (id.includes('/react-markdown/') || id.includes('/remark-') || id.includes('/mdast-') || id.includes('/micromark') || id.includes('/unified/') || id.includes('/hast-')) return 'markdown';
            if (id.includes('/react-dom/')) return 'react-vendor';
            if (id.includes('/react-router')) return 'react-vendor';
            if (id.includes('/minisearch/')) return 'search';
          }
        },
      },
    },
  },
})
