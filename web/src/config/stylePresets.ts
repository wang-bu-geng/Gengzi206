export const CUSTOM_STYLE_ID = 'custom'

export interface StylePreset {
  id: string
  name: string
  description: string
  imagePrompt: string
  videoPrompt: string
  characterPrompt: string
  scenePrompt: string
  colorPalette?: string
  custom?: boolean
}

export const stylePresets: StylePreset[] = [
  {
    id: 'ghibli',
    name: '吉卜力风',
    description: '宫崎骏风格，手绘水彩质感，温馨治愈',
    imagePrompt: 'Studio Ghibli style, hand-drawn animation, soft watercolor backgrounds, detailed nature, warm lighting, whimsical atmosphere, detailed linework, dreamy quality',
    videoPrompt: 'Ghibli style animation, hand-drawn aesthetic, warm gentle motion, whimsical atmosphere, detailed backgrounds',
    characterPrompt: 'Ghibli style character, expressive design, hand-drawn quality, warm colors, relatable features',
    scenePrompt: 'Ghibli style background, lush nature, detailed environment, warm lighting, hand-painted quality, magical atmosphere',
  },
  {
    id: 'guoman',
    name: '国风动漫',
    description: '新国风数字艺术，东方幻想，华丽特效',
    imagePrompt: 'Modern Chinese illustration style, oriental fantasy, flowing lines, luminous elements, rich details, atmospheric lighting, traditional meets modern',
    videoPrompt: 'Chinese style animation, oriental fantasy, flowing motion, luminous aesthetics, epic atmosphere',
    characterPrompt: 'Chinese style character, elegant design, flowing clothing, oriental aesthetics, detailed costume',
    scenePrompt: 'Chinese style background, traditional architecture, flowing clouds, atmospheric lighting, oriental aesthetics',
  },
  {
    id: 'guoman3d',
    name: '国风3D',
    description: '高精细3D仙侠风格，PBR渲染，东方美学',
    imagePrompt: 'High-fidelity 3D rendering, Chinese fantasy style, PBR materials, detailed silk and leather textures, cinematic lighting, elegant oriental design',
    videoPrompt: '3D Chinese fantasy animation, PBR rendering, cinematic lighting, detailed costumes, epic atmosphere',
    characterPrompt: '3D realistic character, Chinese fantasy style, detailed clothing, cinematic lighting, elegant proportions',
    scenePrompt: '3D Chinese fantasy environment, detailed architecture, cinematic lighting, epic scale, oriental atmosphere',
  },
  {
    id: '3d-anime',
    name: '3D 动漫',
    description: '精致的3D卡通渲染风格，适合奇幻、冒险题材',
    imagePrompt: '3D anime style, cel-shaded, vibrant colors, clean lines, studio quality, 3D rendered, anime aesthetic, smooth gradients, detailed textures',
    videoPrompt: '3D anime style animation, smooth motion, cel-shaded rendering, vibrant colors, high quality 3D animation',
    characterPrompt: '3D anime character design, cel-shaded, clean proportions, expressive features, detailed hair and clothing, studio quality',
    scenePrompt: '3D anime background, cel-shaded, detailed environment, immersive atmosphere, studio quality rendering',
  },
  {
    id: 'warm-girl',
    name: '温馨少女',
    description: '柔和暖色调，治愈系画风，适合日常、恋爱题材',
    imagePrompt: 'warm and cozy aesthetic, soft pastel colors, gentle lighting, dreamy atmosphere, shoujo style, delicate features, warm color palette, soft focus, romantic mood',
    videoPrompt: 'warm cozy animation style, soft pastel colors, gentle motion, dreamy atmosphere, healing aesthetic',
    characterPrompt: 'shoujo style character, soft features, warm colors, gentle expression, delicate design, romantic aesthetic',
    scenePrompt: 'warm cozy background, soft lighting, pastel colors, dreamy atmosphere, healing aesthetic, detailed environment',
  },
  {
    id: 'urban',
    name: '都市精致',
    description: '现代韩漫风格，锐利线条，高冷精致',
    imagePrompt: 'Modern webtoon art style, crisp line art, clean aesthetic, muted urban tones, neon accents, hard cel-shading, rim lighting',
    videoPrompt: 'Modern webtoon style animation, clean lines, urban aesthetic, sharp design, moody atmosphere',
    characterPrompt: 'Modern webtoon character, crisp line art, stylish outfit, urban aesthetic, clean design',
    scenePrompt: 'Modern city background, urban night scene, neon lights, clean aesthetic, moody atmosphere',
  },
  {
    id: 'nostalgia',
    name: '复古动画',
    description: '90年代复古赛璐珞风格，胶片颗粒感，怀旧氛围',
    imagePrompt: '90s retro anime style, film grain, chromatic aberration, soft lines, muted pastel tones, dreamy atmosphere, nostalgic cel-shading',
    videoPrompt: 'Retro 90s anime style, film grain, soft motion, nostalgic atmosphere, cel-shaded animation',
    characterPrompt: '90s retro anime character, cel-shaded, soft lines, nostalgic design, muted colors',
    scenePrompt: 'Retro anime background, 90s style, film grain, nostalgic atmosphere, dreamy lighting',
  },
  {
    id: 'pixel',
    name: '像素艺术',
    description: '复古像素风格，怀旧游戏感，适合轻松、休闲题材',
    imagePrompt: 'pixel art style, retro game aesthetic, 16-bit style, pixelated details, vibrant limited color palette, clean pixel edges, nostalgic video game art',
    videoPrompt: 'pixel art animation, retro game style, 16-bit aesthetic, smooth pixel motion, nostalgic feel',
    characterPrompt: 'pixel art character, sprite design, retro game style, limited colors, clean pixel outlines',
    scenePrompt: 'pixel art background, retro game environment, 16-bit style, limited color palette, detailed pixel design',
  },
  {
    id: 'chibi3d',
    name: 'Q版3D',
    description: '3D盲盒手办风格，Q版二头身，超可爱',
    imagePrompt: '3D blind box toy art style, chibi proportions, plastic and resin texture, rounded shapes, PBR rendering, cute design, ambient occlusion',
    videoPrompt: '3D chibi animation, toy-like style, cute proportions, smooth motion, high quality rendering',
    characterPrompt: '3D chibi character, toy-like style, big head small body, cute design, plastic texture',
    scenePrompt: '3D chibi environment, miniature style, cute design, toy-like quality, soft lighting',
  },
  {
    id: 'wasteland',
    name: '废土朋克',
    description: '末世废土风格，硬核线条，复古印刷感',
    imagePrompt: 'hard-edged line art, grainy textures, limited color palette, wasteland aesthetic, high contrast side lighting, moebius style, post-apocalyptic',
    videoPrompt: 'Wasteland style animation, hard lines, grainy texture, post-apocalyptic atmosphere, dramatic lighting',
    characterPrompt: 'Wasteland character, hard line art, rugged design, post-apocalyptic style, limited colors',
    scenePrompt: 'Wasteland environment, post-apocalyptic, grainy texture, hard lines, dramatic lighting',
  },
  {
    id: 'realistic',
    name: '写实风格',
    description: '真实感强，接近真人实拍效果，适合悬疑、都市题材',
    imagePrompt: 'photorealistic, hyperrealistic, cinematic lighting, detailed textures, realistic proportions, lifelike features, professional photography, depth of field',
    videoPrompt: 'cinematic realistic video, photorealistic, lifelike motion, cinematic lighting, high production value',
    characterPrompt: 'realistic character, lifelike features, detailed skin texture, natural proportions, photorealistic rendering',
    scenePrompt: 'realistic environment, detailed textures, cinematic lighting, photorealistic, lifelike atmosphere',
  },
  {
    id: CUSTOM_STYLE_ID,
    name: '自定义',
    description: '用自己的话描述想要的画风、色彩与质感，图片和视频都会遵循',
    imagePrompt: '',
    videoPrompt: '',
    characterPrompt: '',
    scenePrompt: '',
    custom: true,
  },
]

export const getStylePresetById = (id: string): StylePreset | undefined => {
  return stylePresets.find((s) => s.id === id)
}

export const getStylePrompt = (styleId: string, type: 'image' | 'video' | 'character' | 'scene' = 'image'): string => {
  const preset = getStylePresetById(styleId)
  if (!preset) return ''
  
  switch (type) {
    case 'video':
      return preset.videoPrompt
    case 'character':
      return preset.characterPrompt
    case 'scene':
      return preset.scenePrompt
    default:
      return preset.imagePrompt
  }
}
