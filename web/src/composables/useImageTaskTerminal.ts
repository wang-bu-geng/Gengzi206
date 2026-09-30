import { useGenerationTerminal } from '@/composables/useGenerationTerminal'

/**
 * 图片类生成任务 → 生成终端桥接（角色立绘 / 三视图 / 场景图 / 批量）
 *
 * 数据全部来自真实轮询：
 * - image_generation_status: pending → processing → completed/failed
 * - 成功判据由调用方提供（图片身份变化 / 参考图集增长 / 场景图出现）
 * 不写入任何无真实来源的遥测（GPU/FPS/码率等）。
 */

export interface TerminalImageTarget {
  id: number | string
  /** 日志中的技术标签，如 CHAR_10「方源」 */
  tag: string
  /** 成功判据：相对提交时刻，该实体是否已产出新图 */
  hasResult: (entity: any) => boolean
}

export interface TrackImageTasksOptions {
  /** 红色大字，如 角色「方源」 / 三视图「方源」 / 批量立绘 ×3 */
  subject: string
  /** 技术标签，如 角色立绘 CHARACTER_PORTRAIT */
  chip: string
  targets: TerminalImageTarget[]
  /** 每轮轮询调用（通常是 loadDramaData / loadScenes） */
  refresh: () => unknown | Promise<unknown>
  /** 取该实体最新数据，返回 null 表示实体已不存在 */
  findEntity: (id: number | string) => any
  intervalMs?: number
  /** 最大轮询次数（含首次） */
  maxAttempts?: number
  /** 全部结束（成功/失败/超时）回调 */
  onSettled?: (summary: { success: number; failed: number; timeout: number }) => void
}

type Phase = 'queued' | 'processing' | 'success' | 'failed'

interface Runtime {
  target: TerminalImageTarget
  phase: Phase
  /** 超时后是否仍未见结果 */
  timedOut: boolean
}

const activeTimers = new Set<number>()

/** 组件卸载时调用，防止泄漏 */
export function stopAllImageTaskTimers() {
  activeTimers.forEach((id) => window.clearInterval(id))
  activeTimers.clear()
}

export function useImageTaskTerminal() {
  const t = useGenerationTerminal()

  /**
   * 提交失败（HTTP 4xx/5xx，如欠费 429）时由调用方调用：
   * 打开终端并立即给出错误结论。
   */
  const failImmediately = (cfg: { subject: string; chip: string; message: string }) => {
    t.startSession({
      kind: 'image',
      title: '正在生成',
      subject: cfg.subject,
      chip: cfg.chip,
      meta: '任务提交失败',
    })
    t.setProgress(null)
    t.log(`> TASK_SUBMIT_FAILED // ${cfg.message}`, 'error')
    t.endSession({ result: 'error', text: '任务提交失败，请检查模型配置或账户额度' })
  }

  const track = (opts: TrackImageTasksOptions) => {
    const intervalMs = opts.intervalMs ?? 5000
    const maxAttempts = opts.maxAttempts ?? 24
    const multi = opts.targets.length > 1

    t.startSession({
      kind: 'image',
      title: '正在生成',
      subject: opts.subject,
      chip: opts.chip,
      meta: multi ? `队列 ${opts.targets.length} 个任务 // 轮询真实状态` : '任务已提交 // 轮询真实状态',
    })
    t.setProgress(multi ? 0 : null)
    t.log(`> SESSION_OPEN // ${opts.chip}`)
    opts.targets.forEach((tg) => {
      t.log(`> ${tg.tag} 已进入生成队列 // QUEUED`)
    })

    const runtimes: Runtime[] = opts.targets.map((target) => ({
      target,
      phase: 'queued' as Phase,
      timedOut: false,
    }))

    let attempts = 0
    let stopped = false

    const settle = (result: 'success' | 'error', text: string) => {
      if (timer) {
        window.clearInterval(timer)
        activeTimers.delete(timer)
      }
      stopped = true
      const success = runtimes.filter((r) => r.phase === 'success').length
      const failed = runtimes.filter((r) => r.phase === 'failed').length
      const timeout = runtimes.filter((r) => r.timedOut).length
      t.endSession({ result, text, autoMinimizeMs: result === 'success' ? 5000 : undefined })
      opts.onSettled?.({ success, failed, timeout })
    }

    const tick = async () => {
      if (stopped) return
      attempts += 1
      try {
        await opts.refresh()
      } catch {
        // 单次刷新失败不致命，继续下一轮
        t.log('> STATE_REFRESH_FAILED // 本轮状态刷新失败，重试中', 'warning')
        return
      }

      for (const rt of runtimes) {
        if (rt.phase === 'success' || rt.phase === 'failed') continue
        const entity = opts.findEntity(rt.target.id)
        const status: string = entity?.image_generation_status || ''

        if (status === 'failed') {
          rt.phase = 'failed'
          const msg = entity?.image_generation_error || '未返回错误详情'
          t.log(`> ${rt.target.tag} 生成失败 // ${msg}`, 'error')
          continue
        }

        if (status === 'processing' && rt.phase === 'queued') {
          rt.phase = 'processing'
          t.log(`> ${rt.target.tag} 模型渲染中 // PROCESSING`)
          continue
        }

        // completed 状态不会被后端透出（字段仅暴露 pending/processing/failed）；
        // 以「字段为空 + 成功判据成立」为完成。
        if (!status && entity && rt.target.hasResult(entity)) {
          rt.phase = 'success'
          t.log(`> ${rt.target.tag} 成片已回传 // RENDER_SUCCESS`, 'success')
          continue
        }
      }

      if (multi) {
        const settledCount = runtimes.filter(
          (r) => r.phase === 'success' || r.phase === 'failed',
        ).length
        const pct = Math.round((settledCount / runtimes.length) * 100)
        t.setProgress(pct)
        const ok = runtimes.filter((r) => r.phase === 'success').length
        const bad = runtimes.filter((r) => r.phase === 'failed').length
        t.setMeta(`已完成 ${settledCount}/${runtimes.length} // 成功 ${ok} 失败 ${bad}`)
      }

      const unresolved = runtimes.filter((r) => r.phase !== 'success' && r.phase !== 'failed')
      if (unresolved.length === 0) {
        const ok = runtimes.filter((r) => r.phase === 'success').length
        const bad = runtimes.filter((r) => r.phase === 'failed').length
        if (bad === 0) {
          t.log(`> SESSION_COMPLETE // ${ok}/${runtimes.length} 成片回传`, 'success')
          settle('success', `生成完成 · ${ok} 个成片`)
        } else if (ok === 0) {
          t.log(`> SESSION_ERROR // ${bad}/${runtimes.length} 任务失败`, 'error')
          settle('error', `${bad} 个任务生成失败，可重试`)
        } else {
          t.log(`> SESSION_PARTIAL // 成功 ${ok} // 失败 ${bad}`, 'warning')
          settle('success', `部分完成 · 成功 ${ok} / 失败 ${bad}`)
        }
        return
      }

      if (attempts >= maxAttempts) {
        unresolved.forEach((rt) => {
          rt.timedOut = true
          t.log(`> ${rt.target.tag} 轮询超时 // 任务可能仍在后台进行`, 'warning')
        })
        const ok = runtimes.filter((r) => r.phase === 'success').length
        if (ok > 0) {
          settle('success', `部分成片已回传 · ${ok} 个完成，其余仍在后台`)
        } else {
          t.log('> SESSION_TIMEOUT // 等待超时，可稍后刷新页面查看', 'error')
          settle('error', '生成超时，任务可能仍在后台进行')
        }
      }
    }

    const timer = window.setInterval(tick, intervalMs)
    activeTimers.add(timer)
    // 立即执行首轮，缩短「提交→状态」感知间隔
    void tick()

    return () => {
      if (!stopped) {
        stopped = true
        window.clearInterval(timer)
        activeTimers.delete(timer)
      }
    }
  }

  return { track, failImmediately }
}
