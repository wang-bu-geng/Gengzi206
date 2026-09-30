import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

// 从 localStorage 获取保存的语言，默认为中文
const getStoredLanguage = (): string => {
  const stored = localStorage.getItem('language')
  if (stored) return stored
  
  // 自动检测浏览器语言
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) return 'zh-CN'
  return 'en-US'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getStoredLanguage(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  }
})

export default i18n

// 导出语言切换函数
export const setLanguage = (lang: string) => {
  i18n.global.locale.value = lang as any
  localStorage.setItem('language', lang)
}

export const getCurrentLanguage = () => {
  return i18n.global.locale.value
}
