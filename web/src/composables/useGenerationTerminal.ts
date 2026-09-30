import { reactive, readonly } from "vue";

/**
 * 生成终端 —— 全局单例会话状态
 * 仅承载从真实任务轮询派生出的展示数据，不参与任何任务执行逻辑。
 * 硬约束：不写入无真实来源的遥测（FPS/码率/GPU 等）。
 */

export type TerminalKind =
  | "video-batch" // 批量分镜视频
  | "video-single" // 单个分镜视频
  | "storyboard" // AI 拆分分镜
  | "image"; // 角色 / 场景图片

export type TerminalLevel = "info" | "success" | "warning" | "error";

export interface TerminalLog {
  id: number;
  time: string;
  level: TerminalLevel;
  text: string;
}

export interface TerminalSession {
  /** 是否有会话（含已完成等待关闭） */
  active: boolean;
  /** 全屏面板是否被最小化 */
  minimized: boolean;
  /** 任务是否仍在运行 */
  running: boolean;
  kind: TerminalKind | "";
  /** 白色大字行，如「正在生成」 */
  title: string;
  /** 红色大字行，如「第01集」「分镜 #042」 */
  subject: string;
  /** 技术标签，如「批量渲染 BATCH_RENDER」 */
  chip: string;
  /** 真实进度 0~100；null 表示该任务类型不提供百分比 */
  progress: number | null;
  /** 右上系统状态 */
  systemStatus: string;
  /** 进度条下方的真实元信息行 */
  meta: string;
  logs: TerminalLog[];
  /** 会话结论：成功 / 失败 / 空 */
  result: "success" | "error" | "";
  resultText: string;
  startedAt: number;
  elapsed: number;
}

const MAX_LOGS = 220;

const state = reactive<TerminalSession>({
  active: false,
  minimized: false,
  running: false,
  kind: "",
  title: "",
  subject: "",
  chip: "",
  progress: null,
  systemStatus: "",
  meta: "",
  logs: [],
  result: "",
  resultText: "",
  startedAt: 0,
  elapsed: 0,
});

let logSeq = 0;
let elapsedTimer = 0;
let minimizeTimer = 0;

const pad = (n: number) => String(n).padStart(2, "0");

const nowTime = () => {
  const d = new Date();
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
};

const clearTimers = () => {
  if (elapsedTimer) window.clearInterval(elapsedTimer);
  if (minimizeTimer) window.clearTimeout(minimizeTimer);
  elapsedTimer = 0;
  minimizeTimer = 0;
};

const startElapsed = () => {
  if (elapsedTimer) window.clearInterval(elapsedTimer);
  state.elapsed = 0;
  elapsedTimer = window.setInterval(() => {
    state.elapsed = Math.floor((Date.now() - state.startedAt) / 1000);
  }, 1000);
};

/**
 * 开启（或复用同类）会话。同类会话进行中再次调用不会清空日志。
 */
const startSession = (cfg: {
  kind: TerminalKind;
  title: string;
  subject: string;
  chip: string;
  meta?: string;
}) => {
  if (minimizeTimer) {
    window.clearTimeout(minimizeTimer);
    minimizeTimer = 0;
  }
  // 同类会话复用，避免同类并行任务（如多图）反复重置
  const sameKind = state.active && state.kind === cfg.kind && state.running;
  state.active = true;
  state.minimized = false;
  state.running = true;
  state.result = "";
  state.resultText = "";
  state.systemStatus = "运行中 RUNNING";
  if (!sameKind) {
    state.kind = cfg.kind;
    state.logs = [];
    state.startedAt = Date.now();
    state.elapsed = 0;
    startElapsed();
  }
  state.title = cfg.title;
  state.subject = cfg.subject;
  state.chip = cfg.chip;
  state.progress = null;
  state.meta = cfg.meta ?? "";
};

const log = (text: string, level: TerminalLevel = "info") => {
  state.logs.push({ id: ++logSeq, time: nowTime(), level, text });
  if (state.logs.length > MAX_LOGS) {
    state.logs.splice(0, state.logs.length - MAX_LOGS);
  }
};

const setProgress = (progress: number | null) => {
  if (progress == null) {
    state.progress = null;
    return;
  }
  state.progress = Math.max(0, Math.min(100, Math.round(progress)));
};

const setMeta = (meta: string) => {
  state.meta = meta;
};

const setSubject = (subject: string) => {
  state.subject = subject;
};

/**
 * 结束会话。success=false 时面板保持展开等待用户处理；
 * success=true 时可选择延时自动最小化。
 */
const endSession = (cfg: {
  result: "success" | "error";
  text: string;
  autoMinimizeMs?: number;
}) => {
  if (!state.active) return;
  state.running = false;
  state.result = cfg.result;
  state.resultText = cfg.text;
  state.systemStatus =
    cfg.result === "success" ? "完成 COMPLETE" : "异常 ERROR";
  if (cfg.result === "success") state.progress = 100;
  if (elapsedTimer) {
    window.clearInterval(elapsedTimer);
    elapsedTimer = 0;
  }
  if (cfg.result === "error") {
    state.minimized = false;
    return;
  }
  if (cfg.autoMinimizeMs && cfg.autoMinimizeMs > 0) {
    minimizeTimer = window.setTimeout(() => {
      if (state.active && state.result === "success") state.minimized = true;
    }, cfg.autoMinimizeMs);
  }
};

const minimize = () => {
  state.minimized = true;
};

const restore = () => {
  state.minimized = false;
};

/** 关闭并清空会话 */
const close = () => {
  clearTimers();
  state.active = false;
  state.minimized = false;
  state.running = false;
  state.kind = "";
  state.logs = [];
  state.progress = null;
  state.result = "";
  state.resultText = "";
  state.meta = "";
};

/** 用时 mm:ss / hh:mm:ss */
const formatElapsed = (s: number) => {
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return h > 0 ? `${pad(h)}:${pad(m)}:${pad(sec)}` : `${pad(m)}:${pad(sec)}`;
};

export function useGenerationTerminal() {
  return {
    terminal: readonly(state) as typeof state,
    startSession,
    endSession,
    log,
    setProgress,
    setMeta,
    setSubject,
    minimize,
    restore,
    close,
    formatElapsed,
  };
}
