package image

import "fmt"

// Provider 标识图片 AI 服务提供商类型
type Provider string

const (
	ProviderOpenAI      Provider = "openai"
	ProviderZhipu       Provider = "zhipu"
	ProviderGemini      Provider = "gemini"
	ProviderVolcEngine  Provider = "volcengine"
	ProviderChatFire    Provider = "chatfire"
	ProviderComfyUI     Provider = "comfyui"
	ProviderSeedMaker   Provider = "seedmaker"
	ProviderCustom      Provider = "custom"
)

// DisplayName 返回提供商的显示名称
func (p Provider) DisplayName() string {
	switch p {
	case ProviderOpenAI:
		return "OpenAI"
	case ProviderZhipu:
		return "智谱AI"
	case ProviderGemini:
		return "Gemini"
	case ProviderVolcEngine:
		return "火山引擎"
	case ProviderChatFire:
		return "ChatFire"
	case ProviderComfyUI:
		return "ComfyUI"
	case ProviderSeedMaker:
		return "SeedMaker"
	case ProviderCustom:
		return "自定义"
	default:
		return string(p)
	}
}

// NormalizeProvider 将各种别名统一为标准 Provider
func NormalizeProvider(provider string) Provider {
	switch provider {
	case "openai", "dalle":
		return ProviderOpenAI
	case "zhipu", "zhipuai", "glm":
		return ProviderZhipu
	case "gemini", "google":
		return ProviderGemini
	case "volcengine", "volces", "doubao":
		return ProviderVolcEngine
	case "chatfire":
		return ProviderChatFire
	case "comfyui":
		return ProviderComfyUI
	case "seedmaker", "seed":
		return ProviderSeedMaker
	case "":
		return ProviderCustom
	default:
		return ProviderCustom
	}
}

// IsOpenAICompatible 判断 provider 是否使用 OpenAI 兼容格式
func (p Provider) IsOpenAICompatible() bool {
	switch p {
	case ProviderOpenAI, ProviderZhipu, ProviderVolcEngine, ProviderChatFire, ProviderCustom:
		return true
	default:
		return false
	}
}

type ImageClient interface {
	GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error)
	GetTaskStatus(taskID string) (*ImageResult, error)
}

type ImageResult struct {
	TaskID    string
	Status    string
	ImageURL  string
	Width     int
	Height    int
	Error     string
	Completed bool
}

type ImageOptions struct {
	NegativePrompt  string
	Size            string
	Quality         string
	Style           string
	Steps           int
	CfgScale        float64
	Seed            int64
	Model           string
	Width           int
	Height          int
	ReferenceImages []string // 参考图片URL列表
}

type ImageOption func(*ImageOptions)

func WithNegativePrompt(prompt string) ImageOption {
	return func(o *ImageOptions) {
		o.NegativePrompt = prompt
	}
}

func WithSize(size string) ImageOption {
	return func(o *ImageOptions) {
		o.Size = size
	}
}

func WithQuality(quality string) ImageOption {
	return func(o *ImageOptions) {
		o.Quality = quality
	}
}

func WithStyle(style string) ImageOption {
	return func(o *ImageOptions) {
		o.Style = style
	}
}

func WithSteps(steps int) ImageOption {
	return func(o *ImageOptions) {
		o.Steps = steps
	}
}

func WithCfgScale(scale float64) ImageOption {
	return func(o *ImageOptions) {
		o.CfgScale = scale
	}
}

func WithSeed(seed int64) ImageOption {
	return func(o *ImageOptions) {
		o.Seed = seed
	}
}

func WithModel(model string) ImageOption {
	return func(o *ImageOptions) {
		o.Model = model
	}
}

func WithDimensions(width, height int) ImageOption {
	return func(o *ImageOptions) {
		o.Width = width
		o.Height = height
	}
}

func WithReferenceImages(images []string) ImageOption {
	return func(o *ImageOptions) {
		o.ReferenceImages = images
	}
}

// ClientConfig 图片客户端配置
type ClientConfig struct {
	BaseURL       string
	APIKey        string
	Model         string
	Endpoint      string
	QueryEndpoint string
	Provider      Provider
}

// NewClient 根据 provider 创建对应的图片客户端
func NewClient(cfg *ClientConfig) (ImageClient, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = ProviderCustom
	}

	switch provider {
	case ProviderGemini:
		return NewGeminiImageClient(cfg.BaseURL, cfg.APIKey, cfg.Model, cfg.Endpoint), nil
	case ProviderVolcEngine:
		return NewVolcEngineImageClient(cfg.BaseURL, cfg.APIKey, cfg.Model, cfg.Endpoint, cfg.QueryEndpoint), nil
	case ProviderComfyUI:
		return NewComfyUIClient(cfg.BaseURL), nil
	case ProviderSeedMaker:
		return NewSeedMakerImageClient(cfg.BaseURL, cfg.APIKey, cfg.Model), nil
	default:
		// OpenAI 兼容格式：openai, zhipu, chatfire, custom 等
		return NewCompatibleImageClient(cfg), nil
	}
}

// GetDefaultEndpoint 根据 provider 返回默认的图片生成端点
func GetDefaultEndpoint(provider Provider) string {
	switch provider {
	case ProviderGemini:
		return "/v1beta/models/{model}:generateContent"
	case ProviderVolcEngine:
		return "/api/v3/images/generations"
	case ProviderComfyUI:
		return "/prompt"
	case ProviderSeedMaker:
		return "/images"
	default:
		return "/images/generations"
	}
}

// NormalizeSize 确保尺寸符合 API 要求：
// 宽高均在 512-2880 之间，为 16 的整数倍，最大像素数不超过 2^21
func NormalizeSize(size string) string {
	if size == "" {
		return "1024x1024"
	}

	parts := splitSize(size)
	if len(parts) != 2 {
		return "1024x1024"
	}

	w := parts[0]
	h := parts[1]

	// 限制范围 512-2880
	w = clamp(w, 512, 2880)
	h = clamp(h, 512, 2880)

	// 对齐到 16 的整数倍
	w = w / 16 * 16
	h = h / 16 * 16

	// 确保对齐后不小于 512
	if w < 512 {
		w = 512
	}
	if h < 512 {
		h = 512
	}

	// 检查最大像素数不超过 2^21 = 2097152
	maxPixels := 2097152
	for w*h > maxPixels {
		if w >= h {
			w -= 16
		} else {
			h -= 16
		}
	}

	return fmt.Sprintf("%dx%d", w, h)
}

func splitSize(size string) []int {
	var w, h int
	for i, c := range size {
		if c == 'x' || c == 'X' {
			fmt.Sscanf(size[:i], "%d", &w)
			fmt.Sscanf(size[i+1:], "%d", &h)
			return []int{w, h}
		}
	}
	return nil
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
