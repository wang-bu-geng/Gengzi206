import { inject, type Ref, type InjectionKey } from 'vue'
import type { Drama } from '@/types/drama'

export interface DramaContext {
  dramaId: string
  dramaData: Ref<Drama | undefined>
  loadDramaData: () => Promise<void>
  scenes: Ref<any[]>
  loadScenes: () => void
}

export const DRAMA_CONTEXT_KEY: InjectionKey<DramaContext> = Symbol('drama-context')

export function useDramaContext(): DramaContext {
  const ctx = inject(DRAMA_CONTEXT_KEY)
  if (!ctx) throw new Error('useDramaContext() 必须在 DramaManagement 壳组件内使用')
  return ctx
}
