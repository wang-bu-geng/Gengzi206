import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag.startsWith('cropper-')
        }
      }
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3012,
    proxy: {
      '/api': {
        target: 'http://localhost:5678',
        changeOrigin: true
      },
      '/static': {
        target: 'http://localhost:5678',
        changeOrigin: true
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        // Three.js 独立分包：全局背景异步加载，不阻塞首屏
        manualChunks: {
          three: ['three']
        }
      }
    }
  }
})
