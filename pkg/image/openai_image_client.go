package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CompatibleImageClient 是 OpenAI 兼容格式的图片生成客户端
// 适用于 OpenAI、智谱AI、火山引擎、ChatFire 等兼容 OpenAI API 格式的提供商
type CompatibleImageClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	Endpoint   string
	Provider   Provider
	HTTPClient *http.Client
}

// 保留 OpenAIImageClient 类型别名，兼容旧代码
type OpenAIImageClient = CompatibleImageClient

type DALLERequest struct {
	Model   string   `json:"model"`
	Prompt  string   `json:"prompt"`
	Size    string   `json:"size,omitempty"`
	Quality string   `json:"quality,omitempty"`
	N       int      `json:"n"`
	Image   []string `json:"image,omitempty"`
}

type DALLEResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
}

// NewCompatibleImageClient 创建 OpenAI 兼容格式的图片客户端
func NewCompatibleImageClient(cfg *ClientConfig) *CompatibleImageClient {
	provider := cfg.Provider
	if provider == "" {
		provider = ProviderCustom
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "/images/generations"
	}

	return &CompatibleImageClient{
		BaseURL:  cfg.BaseURL,
		APIKey:   cfg.APIKey,
		Model:    cfg.Model,
		Endpoint: endpoint,
		Provider: provider,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// NewOpenAIImageClient 兼容旧代码的构造函数
func NewOpenAIImageClient(baseURL, apiKey, model, endpoint string) *CompatibleImageClient {
	return NewCompatibleImageClient(&ClientConfig{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
		Provider: ProviderOpenAI,
	})
}

// logPrefix 返回带 provider 名称的日志前缀
func (c *CompatibleImageClient) logPrefix() string {
	return fmt.Sprintf("[%s Image]", c.Provider.DisplayName())
}

func (c *CompatibleImageClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
	options := &ImageOptions{
		Size:    "1024x1024",
		Quality: "standard",
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}

	// 如果设置了 Width/Height，转换为 Size 字符串
	if options.Width > 0 && options.Height > 0 {
		options.Size = fmt.Sprintf("%dx%d", options.Width, options.Height)
	}

	// 确保尺寸符合要求：宽高均在 512-2880 之间，为 16 的整数倍
	options.Size = NormalizeSize(options.Size)

	reqBody := DALLERequest{
		Model:   model,
		Prompt:  prompt,
		Size:    options.Size,
		Quality: options.Quality,
		N:       1,
		Image:   options.ReferenceImages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + c.Endpoint
	fmt.Printf("%s Request URL: %s\n", c.logPrefix(), url)
	fmt.Printf("%s Model: %s, Size: %s\n", c.logPrefix(), model, options.Size)
	fmt.Printf("%s Request Body: %s\n", c.logPrefix(), string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s API error (status %d): %s\n", c.logPrefix(), resp.StatusCode, string(body))
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("%s Response: %s\n", c.logPrefix(), string(body))

	var result DALLEResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w, body: %s", err, string(body))
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image generated, response: %s", string(body))
	}

	return &ImageResult{
		Status:    "completed",
		ImageURL:  result.Data[0].URL,
		Completed: true,
	}, nil
}

func (c *CompatibleImageClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	return nil, fmt.Errorf("not supported for %s", c.Provider.DisplayName())
}
