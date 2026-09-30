<template>
  <!--
    创作风格线性图标集
    统一规范：24x24 网格 / 1.7 描边 / 圆角端点 / currentColor 单色
    像素艺术按题材需要使用实心方块，其余全部线性描边
  -->
  <svg
    class="style-icon-svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.7"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <!-- 吉卜力风：小屋与树 -->
    <g v-if="name === 'ghibli'">
      <path d="M4.5 12.2 10 6.8l5.5 5.4" />
      <path d="M6.2 11.2V19.5h7.6v-8.3" />
      <path d="M9.3 19.5v-3.8h2.5v3.8" />
      <path d="M15.6 19.5v-4.6" />
      <circle cx="18.6" cy="11.4" r="2.3" />
      <path d="M3.5 19.5h17" />
    </g>

    <!-- 国风动漫：灯笼 -->
    <g v-else-if="name === 'guoman'">
      <path d="M12 3.4v2.2" />
      <path d="M10 5.6h4" />
      <path d="M12 6C8.8 6 8.6 12 12 12.4c3.4-.4 3.2-6.4 0-6.4Z" />
      <path d="M12 6.2v6" />
      <path d="M9.8 6.9c-.45 2-.45 3 0 4.6" />
      <path d="M14.2 6.9c.45 2 .45 3 0 4.6" />
      <path d="M10 12.6h4" />
      <path d="M12 14.4v3.2" />
      <circle cx="12" cy="18.3" r="0.7" />
    </g>

    <!-- 国风3D：双剑交叉 -->
    <g v-else-if="name === 'guoman3d'">
      <path d="M5.2 18.8 19 5" />
      <path d="M18.8 18.8 5 5" />
      <path d="M6.3 15.3 8.7 17.7" />
      <path d="M17.7 15.3 15.3 17.7" />
      <circle cx="5.2" cy="18.8" r="0.9" />
      <circle cx="18.8" cy="18.8" r="0.9" />
    </g>

    <!-- 3D 动漫：游戏手柄 -->
    <g v-else-if="name === '3d-anime'">
      <path
        d="M7.5 9.5h9a4.5 4.5 0 0 1 4.5 4.5v.8a3.2 3.2 0 0 1-5.8 1.9l-.6-.7h-5.2l-.6.7a3.2 3.2 0 0 1-5.8-1.9v-.8a4.5 4.5 0 0 1 4.5-4.5Z"
      />
      <path d="M8 12.1v3.1" />
      <path d="M6.45 13.65h3.1" />
      <circle cx="15.9" cy="12.7" r="0.6" fill="currentColor" stroke="none" />
      <circle cx="17.5" cy="14.1" r="0.6" fill="currentColor" stroke="none" />
    </g>

    <!-- 温馨少女：花 -->
    <g v-else-if="name === 'warm-girl'">
      <circle cx="12" cy="8.6" r="1.9" />
      <circle cx="15.2" cy="11" r="1.9" />
      <circle cx="14" cy="14.8" r="1.9" />
      <circle cx="10" cy="14.8" r="1.9" />
      <circle cx="8.8" cy="11" r="1.9" />
      <circle cx="12" cy="12.3" r="1.15" />
      <path d="M12 16.7v3.4" />
      <path d="M12 18.7c1.6-.2 2.4.6 2.6 1.8-1.4.1-2.4-.6-2.6-1.8Z" />
    </g>

    <!-- 都市精致：城市天际线 -->
    <g v-else-if="name === 'urban'">
      <path d="M4 20v-9h4v9" />
      <path d="M9 20V7h5v13" />
      <path d="M15 20v-6.5h5V20" />
      <path d="M11.5 7V5" />
      <path d="M3 20h18" />
      <path d="M5.2 13.2h1.6M5.2 15.7h1.6" />
      <path d="M10.5 9.2h2M10.5 11.5h2M10.5 13.8h2" />
      <path d="M16.4 15.8h2.2M16.4 17.8h2.2" />
    </g>

    <!-- 复古动画：录像带 -->
    <g v-else-if="name === 'nostalgia'">
      <rect x="3" y="7.5" width="18" height="9" rx="1.6" />
      <path d="M3 13.4h18" />
      <circle cx="8.5" cy="10.7" r="1.7" />
      <circle cx="8.5" cy="10.7" r="0.5" fill="currentColor" stroke="none" />
      <circle cx="15.5" cy="10.7" r="1.7" />
      <circle cx="15.5" cy="10.7" r="0.5" fill="currentColor" stroke="none" />
      <path d="M10 16.5h4" />
    </g>

    <!-- 像素艺术：像素爱心（实心方块是像素语言的一部分） -->
    <g v-else-if="name === 'pixel'" stroke="none" fill="currentColor">
      <rect x="6" y="4.8" width="2.4" height="2.4" />
      <rect x="8.4" y="4.8" width="2.4" height="2.4" />
      <rect x="13.2" y="4.8" width="2.4" height="2.4" />
      <rect x="15.6" y="4.8" width="2.4" height="2.4" />
      <rect x="3.6" y="7.2" width="16.8" height="2.4" />
      <rect x="3.6" y="9.6" width="16.8" height="2.4" />
      <rect x="6" y="12" width="12" height="2.4" />
      <rect x="8.4" y="14.4" width="7.2" height="2.4" />
      <rect x="10.8" y="16.8" width="2.4" height="2.4" />
    </g>

    <!-- Q版3D：小熊玩偶 -->
    <g v-else-if="name === 'chibi3d'">
      <circle cx="7.6" cy="8" r="1.9" />
      <circle cx="16.4" cy="8" r="1.9" />
      <circle cx="12" cy="13" r="6" />
      <circle cx="9.7" cy="12" r="0.7" fill="currentColor" stroke="none" />
      <circle cx="14.3" cy="12" r="0.7" fill="currentColor" stroke="none" />
      <ellipse cx="12" cy="14.6" rx="2" ry="1.5" />
      <circle cx="12" cy="14.1" r="0.55" fill="currentColor" stroke="none" />
    </g>

    <!-- 废土朋克：残阳、枯树与沙丘 -->
    <g v-else-if="name === 'wasteland'">
      <circle cx="8" cy="9" r="2.6" />
      <path d="M8 4.8v1.2M11.4 9h1.2M4.6 9h1.2M10.4 6.6l.8-.8M5.6 6.6l-.8-.8" />
      <path d="M2.5 20c2.5-3.5 5.5-3.5 8 0" />
      <path d="M9 20c3-4.2 6.5-4.2 12.5 0" />
      <path d="M16.5 19.5V9.2" />
      <path d="M16.5 12.2 18.7 10.4M16.5 13.6l-2-1.4" />
    </g>

    <!-- 自定义：魔法棒 -->
    <g v-else-if="name === 'custom'">
      <path d="M4.5 19.5 15.8 8.2" />
      <path d="M14.6 3.8 15.9 7l3.2 1.3-3.2 1.3-1.3 3.2-1.3-3.2-3.2-1.3 3.2-1.3Z" />
      <path d="M18.8 14.2l.6 1.6 1.6.6-1.6.6-.6 1.6-.6-1.6-1.6-.6 1.6-.6Z" />
    </g>

    <!-- 写实风格：相机 -->
    <g v-else-if="name === 'realistic'">
      <path
        d="M4 8.5h3l1.4-2.2h7.2L17 8.5h3a1.5 1.5 0 0 1 1.5 1.5v7.5a1.5 1.5 0 0 1-1.5 1.5H4a1.5 1.5 0 0 1-1.5-1.5V10A1.5 1.5 0 0 1 4 8.5Z"
      />
      <circle cx="12" cy="13.6" r="3.4" />
      <circle cx="12" cy="13.6" r="1.6" />
      <circle cx="17.6" cy="10.7" r="0.6" fill="currentColor" stroke="none" />
    </g>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  name: string
}>()
</script>

<style scoped>
.style-icon-svg {
  width: 1em;
  height: 1em;
  display: block;
}
</style>
