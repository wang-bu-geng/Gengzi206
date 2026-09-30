package ai

// Provider 标识 AI 服务提供商类型
type Provider string

const (
	ProviderOpenAI      Provider = "openai"
	ProviderZhipu       Provider = "zhipu"
	ProviderGemini      Provider = "gemini"
	ProviderVolcEngine  Provider = "volcengine"
	ProviderChatFire    Provider = "chatfire"
	ProviderCustom      Provider = "custom"
)

// IsOpenAICompatible 判断 provider 是否使用 OpenAI 兼容格式
func (p Provider) IsOpenAICompatible() bool {
	switch p {
	case ProviderOpenAI, ProviderZhipu, ProviderVolcEngine, ProviderChatFire, ProviderCustom:
		return true
	default:
		return false
	}
}

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
	case "":
		return ProviderCustom
	default:
		// 未知 provider 视为自定义，走 OpenAI 兼容格式
		return ProviderCustom
	}
}

// AIClient 定义文本生成客户端接口
type AIClient interface {
	GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error)
	GenerateImage(prompt string, size string, n int) ([]string, error)
	TestConnection() error
}

// VisionCapableClient 表示支持图像输入（视觉理解）的客户端
// 通过类型断言探测能力，不支持的客户端回退为纯文本逻辑
type VisionCapableClient interface {
	// GenerateTextWithImages 发送带参考图片的多模态对话
	// images 元素可以是 data: URI 或 http(s) URL
	GenerateTextWithImages(prompt string, systemPrompt string, images []string, options ...func(*ChatCompletionRequest)) (string, error)
}

// ClientConfig 客户端配置
type ClientConfig struct {
	BaseURL  string
	APIKey   string
	Model    string
	Endpoint string
	Provider Provider
}

// NewClient 根据 provider 创建对应的 AI 客户端
func NewClient(cfg *ClientConfig) AIClient {
	provider := cfg.Provider
	if provider == "" {
		provider = ProviderCustom
	}

	switch provider {
	case ProviderGemini:
		return NewGeminiClient(cfg.BaseURL, cfg.APIKey, cfg.Model, cfg.Endpoint)
	default:
		// OpenAI 兼容格式：openai, zhipu, volcengine, chatfire, custom 等
		return NewCompatibleClient(cfg)
	}
}

// GetDefaultEndpoint 根据 provider 和 serviceType 返回默认 endpoint
func GetDefaultEndpoint(provider Provider, serviceType string) string {
	switch provider {
	case ProviderGemini:
		return "/v1beta/models/{model}:generateContent"
	default:
		// OpenAI 兼容格式
		switch serviceType {
		case "text":
			return "/chat/completions"
		case "image":
			return "/images/generations"
		case "video":
			return "/videos"
		default:
			return "/chat/completions"
		}
	}
}
