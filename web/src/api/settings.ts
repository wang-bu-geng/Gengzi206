import request from '../utils/request'

export const settingsAPI = {
  // 获取系统语言
  getLanguage() {
    return request.get<{ language: string }>('/settings/language')
  },

  // 更新系统语言
  updateLanguage(language: 'zh' | 'en') {
    return request.put<{ message: string; language: string }>('/settings/language', { language })
  }
}
