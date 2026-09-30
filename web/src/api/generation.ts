import type {
    GenerateCharactersRequest,
    ParseScriptRequest,
    ParseScriptResult
} from '../types/generation'
import request from '../utils/request'

export const generationAPI = {
  generateCharacters(data: GenerateCharactersRequest) {
    return request.post<{ task_id: string; status: string; message: string }>('/generation/characters', data)
  },

  generateStoryboard(episodeId: string, model?: string) {
    return request.post<{ task_id: string; status: string; message: string }>(`/episodes/${episodeId}/storyboards`, { model })
  },

  parseScript(episodeId: string, data?: { model?: string }) {
    return request.post<{ task_id: string; status: string; message: string }>(`/episodes/${episodeId}/parse-script`, data || {})
  },

  // 按整段剧本内容解析拆分（注意：后端路由尚未实现，UploadScriptDialog 当前也未被任何页面挂载）
  parseScriptContent(data: ParseScriptRequest) {
    return request.post<ParseScriptResult>('/dramas/parse-script', data)
  },

  generateShots(episodeId: string, data?: { model?: string }) {
    return request.post<{ task_id: string; status: string; message: string }>(`/episodes/${episodeId}/generate-shots`, data || {})
  },

  getTaskStatus(taskId: string) {
    return request.get<{
      id: string
      type: string
      status: string
      progress: number
      message?: string
      error?: string
      result?: string
      created_at: string
      updated_at: string
      completed_at?: string
    }>(`/tasks/${taskId}`)
  }
  
}
