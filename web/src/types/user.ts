export interface User {
  id: number
  username: string
  email: string
  avatar?: string
  nickname?: string
  phone?: string
  role: string
  status: number
  created_at: string
}

export interface UserConfig {
  text_provider: string
  text_model: string
  text_api_key_set: boolean
  image_provider: string
  image_model: string
  image_api_key_set: boolean
  video_provider: string
  video_model: string
  video_api_key_set: boolean
  default_style: string
  default_resolution: string
  default_fps: number
}
