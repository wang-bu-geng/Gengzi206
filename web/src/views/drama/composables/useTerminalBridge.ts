import { watch, type ComputedRef, type Ref } from "vue";
import { useGenerationTerminal } from "@/composables/useGenerationTerminal";

/**
 * 生成终端桥接层：只观察既有任务状态（视频/分镜/图片），
 * 派生出终端会话与日志。不改变任何任务执行逻辑。
 * 所有日志均来自真实轮询数据或真实状态翻转，无伪造遥测。
 */

interface BridgeSources {
  episodeNumber: number;
  currentEpisode: ComputedRef<any>;
  // 视频
  batchGeneratingVideo: Ref<boolean>;
  batchVideoTotal: Ref<number>;
  batchVideoDone: Ref<number>;
  overallVideoProgress: ComputedRef<number>;
  videoProgress: Ref<Record<number, number>>;
  generatingVideo: Ref<Record<string, boolean>>;
  failedVideoShots: Ref<Record<number, string>>;
  // 分镜拆分
  generatingShots: Ref<boolean>;
  taskProgress: Ref<number>;
  taskMessage: Ref<string>;
  // 图片
  batchGeneratingCharacters: Ref<boolean>;
  batchGeneratingScenes: Ref<boolean>;
  generatingCharacterImages: Ref<Record<number, boolean>>;
  generatingSceneImages: Ref<Record<string, boolean>>;
  selectedCharacterIds: Ref<number[]>;
  selectedSceneIds: Ref<number[]>;
}

const pad2 = (n: number | string) => String(n).padStart(2, "0");

export function useTerminalBridge(s: BridgeSources) {
  const t = useGenerationTerminal();

  // ── 本地快照：图片 composable 直接改对象键，watch 拿不到旧值 ──
  const charSnap: Record<number, boolean> = {};
  const sceneSnap: Record<string, boolean> = {};
  let singleVideoSnap: Record<string, boolean> = {};
  /** 分镜进度里程碑，避免轮询重复刷日志 */
  let shotMilestones: Record<number, number> = {};
  let storyboardBaseline = 0;
  let imageBaseline = { chars: 0, scenes: 0 };

  const episodeLabel = () => `第${pad2(s.episodeNumber)}集`;

  const findShot = (shotId: number) =>
    (s.currentEpisode.value?.storyboards || []).find((x: any) => Number(x.id) === shotId);

  const shotTag = (shotId: number) => {
    const shot = findShot(shotId);
    const no = shot?.storyboard_number != null ? pad2(shot.storyboard_number) : shotId;
    return shot?.title ? `SHOT_${no}「${shot.title}」` : `SHOT_${no}`;
  };

  const hasImage = (entity: any) =>
    !!(entity && (entity.image_url || entity.local_path));

  // ───────────────────────── 视频：批量 ─────────────────────────
  watch(s.batchGeneratingVideo, (running) => {
    if (running) {
      shotMilestones = {};
      t.startSession({
        kind: "video-batch",
        title: "正在生成",
        subject: episodeLabel(),
        chip: "批量渲染 BATCH_RENDER",
        meta: "任务节点调度中…",
      });
      t.log("> 批量视频渲染任务已提交 // 等待节点调度");
      t.setProgress(0);
      if (s.batchVideoTotal.value > 0) {
        t.log(`> BATCH_SIZE: ${s.batchVideoTotal.value} 个分镜进入队列`);
      }
    } else if (t.terminal.kind === "video-batch" && t.terminal.running) {
      finalizeVideoBatch();
    }
  });

  watch(s.batchVideoTotal, (n) => {
    if (n > 0 && t.terminal.kind === "video-batch" && t.terminal.running) {
      t.log(`> BATCH_SIZE: ${n} 个分镜进入队列`);
    }
  });

  const finalizeVideoBatch = () => {
    const shots = s.currentEpisode.value?.storyboards || [];
    const failures = Object.entries(s.failedVideoShots.value).filter(([, msg]) => !!msg);
    const doneCount = shots.filter((x: any) => hasImage(x) || x.video_url).length;
    const tracked = s.batchVideoTotal.value || Object.keys(s.videoProgress.value).length;

    // 提交阶段异常（没有任何任务被追踪）不能报成功
    if (tracked === 0) {
      t.log("> BATCH_ABORTED // 渲染任务未成功提交", "error");
      t.endSession({ result: "error", text: "批量渲染未启动，可重试" });
      return;
    }

    const okCount = Math.max(0, tracked - failures.length);

    if (failures.length > 0) {
      failures.slice(0, 8).forEach(([id, msg]) => {
        t.log(`> ${shotTag(Number(id))} 渲染失败 // ${msg}`, "error");
      });
      t.log(
        `> BATCH END // 成功 ${okCount} / 失败 ${failures.length} // 当前成片 ${doneCount} 条`,
        "warning",
      );
      t.setMeta(`成功 ${okCount} / 失败 ${failures.length}`);
      t.endSession({ result: "error", text: `${failures.length} 个分镜渲染失败，可在分镜列表单独重试` });
    } else {
      t.log(`> BATCH_COMPILATION_SUCCESS // ${tracked} 个分镜成片已回传`, "success");
      t.setMeta(`已完成 ${tracked}/${tracked} 个分镜`);
      t.endSession({ result: "success", text: `批量渲染完成 · ${tracked} 个分镜`, autoMinimizeMs: 6000 });
    }
  };

  // ───────────────────────── 视频：单个 ─────────────────────────
  watch(
    s.generatingVideo,
    (map) => {
      if (s.batchGeneratingVideo.value) return; // 批量走批量会话
      const activeId = Object.entries(map).find(([, v]) => v)?.[0];
      const prevActive = Object.entries(singleVideoSnap).find(([, v]) => v)?.[0];

      if (activeId && !prevActive) {
        // 新的单分镜任务
        shotMilestones = {};
        t.startSession({
          kind: "video-single",
          title: "正在生成",
          subject: `分镜 #${activeId}`,
          chip: "单镜渲染 SHOT_RENDER",
          meta: "任务节点调度中…",
        });
        t.log(`> ${shotTag(Number(activeId))} 渲染任务已提交`);
        t.setProgress(0);
      } else if (!activeId && prevActive && t.terminal.kind === "video-single" && t.terminal.running) {
        // 单分镜任务结束
        const shotId = Number(prevActive);
        const err = s.failedVideoShots.value[shotId];
        const shot = findShot(shotId);
        if (err) {
          t.log(`> ${shotTag(shotId)} 渲染失败 // ${err}`, "error");
          t.endSession({ result: "error", text: "该分镜渲染失败，可重试" });
        } else if (shot && (shot.video_url || hasImage(shot))) {
          t.log(`> ${shotTag(shotId)} RENDER_SUCCESS // 成片已回传`, "success");
          t.endSession({ result: "success", text: "分镜渲染完成", autoMinimizeMs: 6000 });
        } else {
          // 提交失败等异常：无失败记录但也没有成片
          t.log(`> ${shotTag(shotId)} 任务异常结束 // 未取得成片`, "error");
          t.endSession({ result: "error", text: "渲染任务未完成，可重试" });
        }
      }
      singleVideoSnap = { ...map };
    },
    { deep: true },
  );

  // ───────────────────────── 视频进度（批量 + 单镜） ─────────────────────────
  watch(
    s.videoProgress,
    (map) => {
      const kind = t.terminal.kind;
      if (!t.terminal.running || (kind !== "video-batch" && kind !== "video-single")) return;

      for (const [idStr, p] of Object.entries(map)) {
        const id = Number(idStr);
        const prev = shotMilestones[id] ?? -1;
        if (p <= 0) continue;

        if (prev < 0) {
          t.log(`> ${shotTag(id)} 进入渲染管线 // ${p}%`);
        } else if (prev < 50 && p >= 50 && p < 100) {
          t.log(`> ${shotTag(id)} 渲染过半 // ${p}%`);
        } else if (p >= 100 && prev >= 0 && kind === "video-batch") {
          // 仅批量会话逐镜播报完成；单镜会话在收尾时统一输出（避免重复）。
          // prev<0 时说明 100 是上一任务遗留的旧值，新会话尚未轮询到该镜。
          const err = s.failedVideoShots.value[id];
          if (!err) t.log(`> ${shotTag(id)} 成片回传完成`, "success");
          // 失败行在会话收尾时统一输出（此时错误信息可能尚未落到快照）
        }
        shotMilestones[id] = p;
      }

      t.setProgress(s.overallVideoProgress.value);
      if (kind === "video-batch") {
        const done = Object.values(map).filter((v) => v >= 100).length;
        const total = s.batchVideoTotal.value || Object.keys(map).length;
        t.setMeta(`已完成 ${done}/${total} 个分镜 // 平均进度 ${s.overallVideoProgress.value}%`);
      } else {
        const activeId = Object.entries(s.generatingVideo.value).find(([, v]) => v)?.[0];
        if (activeId) t.setMeta(`分镜 #${activeId} // 进度 ${map[Number(activeId)] ?? 0}%`);
      }
    },
    { deep: true },
  );

  // ───────────────────────── 分镜拆分 ─────────────────────────
  watch(s.generatingShots, (running) => {
    if (running) {
      storyboardBaseline = s.currentEpisode.value?.storyboards?.length || 0;
      t.startSession({
        kind: "storyboard",
        title: "正在拆分",
        subject: `${episodeLabel()} · 分镜结构`,
        chip: "结构解析 STRUCTURE_PARSE",
        meta: "异步任务已创建，轮询任务状态…",
      });
      t.log("> 分镜拆分异步任务已创建 // 2s 轮询");
      t.setProgress(0);
    } else if (t.terminal.kind === "storyboard" && t.terminal.running) {
      const n = s.currentEpisode.value?.storyboards?.length || 0;
      if (n > storyboardBaseline) {
        t.log(`> STRUCTURE_BUILD_SUCCESS // 分镜 ${n} 条已入库`, "success");
        t.setMeta(`分镜入库 // ${n} 条`);
        t.endSession({ result: "success", text: `分镜拆分完成 · ${n} 条`, autoMinimizeMs: 6000 });
      } else {
        t.log("> STRUCTURE_BUILD_FAILED // 未取得新分镜，请查看消息提示后重试", "error");
        t.endSession({ result: "error", text: "分镜拆分未完成，可重试" });
      }
    }
  });

  watch(s.taskProgress, (p) => {
    if (t.terminal.kind === "storyboard" && t.terminal.running) {
      t.setProgress(p);
      t.setMeta(`任务进度 ${p}% // ${s.taskMessage.value}`);
    }
  });

  let lastTaskMessage = "";
  watch(s.taskMessage, (msg) => {
    if (t.terminal.kind !== "storyboard" || !t.terminal.running) return;
    const text = msg?.trim();
    if (text && text !== lastTaskMessage) {
      t.log(`> ${text}`);
      lastTaskMessage = text;
    }
  });

  // ───────────────────────── 图片（无百分比，真实状态轮询） ─────────────────────────
  const ensureImageSession = () => {
    if (t.terminal.active && t.terminal.running) {
      // 视频/分镜优先级更高：图片不抢占
      if (t.terminal.kind !== "image") return false;
      return true;
    }
    imageBaseline = {
      chars: (s.currentEpisode.value?.characters || []).filter(hasImage).length,
      scenes: (s.currentEpisode.value?.scenes || []).filter(hasImage).length,
    };
    t.startSession({
      kind: "image",
      title: "正在生成",
      subject: "角色 / 场景资产",
      chip: "图像合成 IMAGE_SYNTH",
      meta: "该任务类型不提供百分比，按真实状态轮询…",
    });
    t.setProgress(null);
    return true;
  };

  watch(s.batchGeneratingCharacters, (v) => {
    if (v && ensureImageSession()) {
      t.log(`> 角色批量图像任务已提交 // SELECTED: ${s.selectedCharacterIds.value.length}`);
    }
  });

  watch(s.batchGeneratingScenes, (v) => {
    if (v && ensureImageSession()) {
      t.log(`> 场景批量图像任务已提交 // SELECTED: ${s.selectedSceneIds.value.length}`);
    }
  });

  const findChar = (id: number) =>
    (s.currentEpisode.value?.characters || []).find((c: any) => Number(c.id) === id);
  const findScene = (id: string) =>
    (s.currentEpisode.value?.scenes || []).find((x: any) => String(x.id) === id);

  watch(
    s.generatingCharacterImages,
    (map) => {
      for (const [idStr, v] of Object.entries(map)) {
        const id = Number(idStr);
        const prev = charSnap[id];
        if (v && !prev && ensureImageSession()) {
          const c = findChar(id);
          t.log(`> CHAR_${id}${c?.name ? `「${c.name}」` : ""} 进入图像队列 // 轮询中`);
        } else if (!v && prev && t.terminal.kind === "image") {
          const c = findChar(id);
          if (hasImage(c)) {
            t.log(`> CHAR_${id}${c?.name ? `「${c.name}」` : ""} 图像成片已回传`, "success");
          } else {
            t.log(`> CHAR_${id}${c?.name ? `「${c.name}」` : ""} 任务结束但未取得成片`, "warning");
          }
        }
        charSnap[id] = v;
      }
      maybeCloseImageSession();
    },
    { deep: true },
  );

  watch(
    s.generatingSceneImages,
    (map) => {
      for (const [id, v] of Object.entries(map)) {
        const prev = sceneSnap[id];
        if (v && !prev && ensureImageSession()) {
          const sc = findScene(id);
          t.log(`> SCENE_${id}${sc?.location ? `「${sc.location}」` : ""} 进入图像队列 // 轮询中`);
        } else if (!v && prev && t.terminal.kind === "image") {
          const sc = findScene(id);
          if (hasImage(sc)) {
            t.log(`> SCENE_${id}${sc?.location ? `「${sc.location}」` : ""} 图像成片已回传`, "success");
          } else {
            t.log(`> SCENE_${id}${sc?.location ? `「${sc.location}」` : ""} 任务结束但未取得成片`, "warning");
          }
        }
        sceneSnap[id] = v;
      }
      maybeCloseImageSession();
    },
    { deep: true },
  );

  const maybeCloseImageSession = () => {
    if (t.terminal.kind !== "image" || !t.terminal.running) return;
    if (s.batchGeneratingCharacters.value || s.batchGeneratingScenes.value) return;
    const charBusy = Object.values(s.generatingCharacterImages.value).some(Boolean);
    const sceneBusy = Object.values(s.generatingSceneImages.value).some(Boolean);
    if (charBusy || sceneBusy) return;

    const chars = s.currentEpisode.value?.characters || [];
    const scenes = s.currentEpisode.value?.scenes || [];
    const charImgs = chars.filter(hasImage).length;
    const sceneImgs = scenes.filter(hasImage).length;
    const newCount =
      Math.max(0, charImgs - imageBaseline.chars) +
      Math.max(0, sceneImgs - imageBaseline.scenes);

    if (newCount > 0) {
      t.log(`> IMAGE_SYNTH_SUCCESS // 本轮新增成片 ${newCount} 个`, "success");
      t.setMeta(`角色图 ${charImgs}/${chars.length} // 场景图 ${sceneImgs}/${scenes.length}`);
      t.endSession({ result: "success", text: `图像生成完成 · 新增 ${newCount} 个成片`, autoMinimizeMs: 6000 });
    } else {
      t.log("> IMAGE_SYNTH_END // 本轮未取得新成片", "warning");
      t.setMeta(`角色图 ${charImgs}/${chars.length} // 场景图 ${sceneImgs}/${scenes.length}`);
      t.endSession({ result: "success", text: "图像任务队列已结束", autoMinimizeMs: 6000 });
    }
  };
}
