/**
 * 自动保存 Composable
 * 使用 lodash debounce 对表单变更实现 300ms 延迟自动保存
 */

import { watch, type Ref } from 'vue'
import { ElMessage } from 'element-plus'
import debounce from 'lodash/debounce'

export interface UseAutoSaveOptions<T extends Record<string, any>> {
  /** 被监听的表单数据 */
  source: Ref<T>
  /** 保存函数，返回 true 表示成功 */
  saveFn: (data: T) => Promise<boolean>
  /** 是否启用自动保存 */
  enabled: Ref<boolean> | (() => boolean)
  /** debounce 延迟，默认 300ms */
  delay?: number
  /** 校验函数，返回不为 null 表示校验失败 */
  validate?: () => Promise<boolean>
  /** 是否为首轮（新建模式，第一次保存成功后切为编辑模式） */
  isNew?: Ref<boolean>
  /** 新建成功后回调（传入 API 返回的 ID） */
  onCreated?: (id: number | string) => void
}

export function useAutoSave<T extends Record<string, any>>(options: UseAutoSaveOptions<T>) {
  const delay = options.delay ?? 300

  const debouncedSave = debounce(async (data: T) => {
    const isEnabled = typeof options.enabled === 'function' ? options.enabled() : options.enabled.value
    if (!isEnabled) return

    // 校验
    if (options.validate) {
      const valid = await options.validate()
      if (!valid) return
    }

    try {
      await options.saveFn(data)
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  }, delay)

  // 深度监听表单变更
  const stopWatch = watch(
    options.source,
    (newVal) => {
      debouncedSave({ ...newVal })
    },
    { deep: true }
  )

  /** 立即保存（跳过 debounce），Ctrl+Enter 触发 */
  async function saveNow(): Promise<boolean> {
    debouncedSave.cancel()

    const isEnabled = typeof options.enabled === 'function' ? options.enabled() : options.enabled.value
    if (!isEnabled) return false

    if (options.validate) {
      const valid = await options.validate()
      if (!valid) return false
    }

    try {
      const result = await options.saveFn({ ...options.source.value })
      ElMessage.success('已保存')
      return true
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
      return false
    }
  }

  /** 销毁 */
  function destroy() {
    debouncedSave.cancel()
    stopWatch()
  }

  return { saveNow, destroy }
}
