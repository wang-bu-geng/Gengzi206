import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './assets/styles/element/index.scss'

import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import i18n from './locales'

// JetBrains Mono 本地化引入（技术标签/编号/日志），避免 Google Fonts 外链不稳定
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/500.css'
import '@fontsource/jetbrains-mono/700.css'
import './assets/styles/main.css'

// Apply theme before app mounts to prevent flash
// 赛博终端以深色（CRT）为默认主题，仅当用户显式选择浅色时才翻转
const savedTheme = localStorage.getItem('theme')
if (savedTheme !== 'light') {
  document.documentElement.classList.add('dark')
}

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(ElementPlus)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')
