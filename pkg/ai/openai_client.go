package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CompatibleClient 是 OpenAI 兼容格式的通用客户端
// 适用于 OpenAI、智谱AI、火山引擎、ChatFire 等兼容 OpenAI API 格式的提供商
type CompatibleClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	Endpoint   string
	Provider   Provider
	HTTPClient *http.Client
}

// 保留 OpenAIClient 类型别名，兼容旧代码
type OpenAIClient = CompatibleClient

type ChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // 纯文本为 string；多模态时为 []ChatContentPart
}

// ChatContentPart OpenAI 兼容多模态消息片段
type ChatContentPart struct {
	Type     string        `json:"type"` // text / image_url
	Text     string        `json:"text,omitempty"`
	ImageURL *ChatImageRef `json:"image_url,omitempty"`
}

// ChatImageRef 多模态图片引用，URL 支持 http(s) 链接与 data: URI
type ChatImageRef struct {
	URL string `json:"url"`
}

// NewTextPart 构造文本片段
func NewTextPart(text string) ChatContentPart {
	return ChatContentPart{Type: "text", Text: text}
}

// NewImagePart 构造图片片段
func NewImagePart(url string) ChatContentPart {
	return ChatContentPart{Type: "image_url", ImageURL: &ChatImageRef{URL: url}}
}

type ChatCompletionRequest struct {
	Model               string        `json:"model"`
	Messages            []ChatMessage `json:"messages"`
	Temperature         float64       `json:"temperature,omitempty"`
	MaxTokens           *int          `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
	TopP                float64       `json:"top_p,omitempty"`
	Stream              bool          `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ImageGenerationRequest struct {
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

type ImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewCompatibleClient 创建 OpenAI 兼容格式的客户端
func NewCompatibleClient(cfg *ClientConfig) *CompatibleClient {
	provider := cfg.Provider
	if provider == "" {
		provider = ProviderCustom
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "/chat/completions"
	}

	return &CompatibleClient{
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

// NewOpenAIClient 兼容旧代码的构造函数
func NewOpenAIClient(baseURL, apiKey, model, endpoint string) *CompatibleClient {
	return NewCompatibleClient(&ClientConfig{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
		Provider: ProviderOpenAI,
	})
}

// logPrefix 返回带 provider 名称的日志前缀
func (c *CompatibleClient) logPrefix() string {
	return fmt.Sprintf("[%s]", c.Provider.DisplayName())
}

func (c *CompatibleClient) ChatCompletion(messages []ChatMessage, options ...func(*ChatCompletionRequest)) (*ChatCompletionResponse, error) {
	req := &ChatCompletionRequest{
		Model:    c.Model,
		Messages: messages,
	}

	for _, option := range options {
		option(req)
	}

	return c.sendChatRequest(req)
}

func (c *CompatibleClient) sendChatRequest(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	resp, err := c.doChatRequest(req)
	if err == nil {
		return resp, nil
	}

	if shouldRetryWithMaxCompletionTokens(err, req) {
		tokens := *req.MaxTokens
		retryReq := *req
		retryReq.MaxTokens = nil
		retryReq.MaxCompletionTokens = &tokens
		fmt.Printf("%s retrying with max_completion_tokens=%d\n", c.logPrefix(), tokens)
		return c.doChatRequest(&retryReq)
	}

	return nil, err
}

func (c *CompatibleClient) doChatRequest(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("%s Failed to marshal request: %v\n", c.logPrefix(), err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.BaseURL + c.Endpoint

	fmt.Printf("%s Sending request to: %s\n", c.logPrefix(), url)
	fmt.Printf("%s Model=%s, Endpoint=%s\n", c.logPrefix(), c.Model, c.Endpoint)
	requestPreview := string(jsonData)
	if len(jsonData) > 300 {
		requestPreview = string(jsonData[:300]) + "..."
	}
	fmt.Printf("%s Request body: %s\n", c.logPrefix(), requestPreview)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s Failed to create request: %v\n", c.logPrefix(), err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		fmt.Printf("%s HTTP request failed: %v\n", c.logPrefix(), err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("%s Received response with status: %d\n", c.logPrefix(), resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s Failed to read response body: %v\n", c.logPrefix(), err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s API error (status %d): %s\n", c.logPrefix(), resp.StatusCode, string(body))
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	bodyPreview := string(body)
	if len(body) > 500 {
		bodyPreview = string(body[:500]) + "..."
	}
	fmt.Printf("%s Response body: %s\n", c.logPrefix(), bodyPreview)

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		errorPreview := string(body)
		if len(body) > 200 {
			errorPreview = string(body[:200])
		}
		fmt.Printf("%s Failed to parse response: %v\n", c.logPrefix(), err)
		return nil, fmt.Errorf("failed to unmarshal response: %w, body preview: %s", err, errorPreview)
	}

	fmt.Printf("%s Successfully parsed response, choices count: %d\n", c.logPrefix(), len(chatResp.Choices))

	if len(chatResp.Choices) == 0 {
		fmt.Printf("%s No choices in response\n", c.logPrefix())
		return nil, fmt.Errorf("no choices in response")
	}

	if len(chatResp.Choices) > 0 {
		finishReason := chatResp.Choices[0].FinishReason
		content := chatResp.Choices[0].Message.Content
		usage := chatResp.Usage

		fmt.Printf("%s finish_reason=%s, content_length=%d\n", c.logPrefix(), finishReason, len(content))

		if finishReason == "content_filter" {
			return nil, fmt.Errorf("AI内容被安全过滤器拦截，可能因为：\n1. 请求内容触发了安全策略\n2. 生成的内容包含敏感信息\n3. 建议：调整输入内容或联系API提供商调整过滤策略")
		}

		if usage.TotalTokens == 0 && finishReason != "stop" {
			return nil, fmt.Errorf("AI返回内容为空 (finish_reason: %s)，可能的原因：\n1. 内容被过滤\n2. Token限制\n3. API异常", finishReason)
		}
	}

	return &chatResp, nil
}

func WithTemperature(temp float64) func(*ChatCompletionRequest) {
	return func(req *ChatCompletionRequest) {
		req.Temperature = temp
	}
}

func WithMaxTokens(tokens int) func(*ChatCompletionRequest) {
	return func(req *ChatCompletionRequest) {
		req.MaxTokens = &tokens
	}
}

func WithTopP(topP float64) func(*ChatCompletionRequest) {
	return func(req *ChatCompletionRequest) {
		req.TopP = topP
	}
}

func (c *CompatibleClient) GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	messages := []ChatMessage{}

	if systemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: prompt,
	})

	resp, err := c.ChatCompletion(messages, options...)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateTextWithImages 发送带参考图片的多模态对话（OpenAI 兼容 vision 协议）
func (c *CompatibleClient) GenerateTextWithImages(prompt string, systemPrompt string, images []string, options ...func(*ChatCompletionRequest)) (string, error) {
	parts := []ChatContentPart{NewTextPart(prompt)}
	for _, img := range images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		parts = append(parts, NewImagePart(img))
	}

	messages := []ChatMessage{}
	if systemPrompt != "" {
		messages = append(messages, ChatMessage{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, ChatMessage{Role: "user", Content: parts})

	resp, err := c.ChatCompletion(messages, options...)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *CompatibleClient) GenerateImage(prompt string, size string, n int) ([]string, error) {
	// 使用配置的 endpoint，如果为空或为文本端点则回退到图片生成端点
	imageEndpoint := c.Endpoint
	if imageEndpoint == "" || imageEndpoint == "/chat/completions" {
		imageEndpoint = "/images/generations"
	}

	url := c.BaseURL + imageEndpoint

	reqBody := ImageGenerationRequest{
		Prompt: prompt,
		N:      n,
		Size:   size,
		Model:  c.Model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var imgResp ImageGenerationResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		return nil, err
	}

	var urls []string
	for _, data := range imgResp.Data {
		if data.URL != "" {
			urls = append(urls, data.URL)
		} else if data.B64JSON != "" {
			urls = append(urls, "data:image/png;base64,"+data.B64JSON)
		}
	}

	return urls, nil
}

func (c *CompatibleClient) TestConnection() error {
	fmt.Printf("%s TestConnection: BaseURL=%s, Model=%s, Endpoint=%s\n", c.logPrefix(), c.BaseURL, c.Model, c.Endpoint)

	messages := []ChatMessage{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	_, err := c.ChatCompletion(messages, WithMaxTokens(50))
	if err != nil {
		fmt.Printf("%s TestConnection failed: %v\n", c.logPrefix(), err)
	} else {
		fmt.Printf("%s TestConnection succeeded\n", c.logPrefix())
	}
	return err
}

func shouldRetryWithMaxCompletionTokens(err error, req *ChatCompletionRequest) bool {
	if err == nil || req == nil || req.MaxTokens == nil || req.MaxCompletionTokens != nil {
		return false
	}

	msg := err.Error()
	if strings.Contains(msg, "Unsupported parameter: 'max_tokens'") {
		return true
	}
	if strings.Contains(msg, "max_tokens is not supported") {
		return true
	}
	if strings.Contains(msg, "max_completion_tokens") {
		return true
	}
	return false
}
