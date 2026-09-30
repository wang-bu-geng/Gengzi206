package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GeminiClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	Endpoint   string
	HTTPClient *http.Client
}

type GeminiTextRequest struct {
	Contents          []GeminiContent    `json:"contents"`
	SystemInstruction *GeminiInstruction `json:"systemInstruction,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *GeminiBlob `json:"inlineData,omitempty"`
}

// GeminiBlob Gemini 内联图片数据
type GeminiBlob struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"base64Data"`
}

type GeminiInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiTextResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		Index         int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func NewGeminiClient(baseURL, apiKey, model, endpoint string) *GeminiClient {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	if endpoint == "" {
		endpoint = "/v1beta/models/{model}:generateContent"
	}
	if model == "" {
		model = "gemini-3-pro"
	}
	return &GeminiClient{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (c *GeminiClient) GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	// 构建请求体
	reqBody := GeminiTextRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{{Text: prompt}},
				Role:  "user",
			},
		},
	}

	// 使用 systemInstruction 字段处理系统提示
	if systemPrompt != "" {
		reqBody.SystemInstruction = &GeminiInstruction{
			Parts: []GeminiPart{{Text: systemPrompt}},
		}
	}

	return c.sendTextRequest(reqBody)
}

// GenerateTextWithImages 发送带参考图片的多模态对话（Gemini inlineData 协议）
func (c *GeminiClient) GenerateTextWithImages(prompt string, systemPrompt string, images []string, options ...func(*ChatCompletionRequest)) (string, error) {
	parts := []GeminiPart{{Text: prompt}}
	for _, img := range images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		blob, err := c.resolveImageBlob(img)
		if err != nil {
			return "", fmt.Errorf("resolve image for gemini: %w", err)
		}
		parts = append(parts, GeminiPart{InlineData: blob})
	}

	reqBody := GeminiTextRequest{
		Contents: []GeminiContent{
			{Parts: parts, Role: "user"},
		},
	}
	if systemPrompt != "" {
		reqBody.SystemInstruction = &GeminiInstruction{
			Parts: []GeminiPart{{Text: systemPrompt}},
		}
	}

	return c.sendTextRequest(reqBody)
}

// resolveImageBlob 将 data: URI 或 http(s) 链接转换为 Gemini inlineData
func (c *GeminiClient) resolveImageBlob(ref string) (*GeminiBlob, error) {
	if strings.HasPrefix(ref, "data:") {
		// data:image/png;base64,xxxx
		header, data, ok := strings.Cut(ref, ",")
		if !ok {
			return nil, fmt.Errorf("invalid data uri")
		}
		mime := "image/png"
		if semi := strings.Index(header, ";"); semi > 5 {
			mime = header[5:semi]
		}
		return &GeminiBlob{MIMEType: mime, Data: data}, nil
	}

	resp, err := c.HTTPClient.Get(ref)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch image status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return nil, err
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" || strings.Contains(mime, "text") {
		mime = http.DetectContentType(body)
	}
	if idx := strings.Index(mime, ";"); idx > 0 {
		mime = mime[:idx]
	}
	return &GeminiBlob{MIMEType: mime, Data: base64.StdEncoding.EncodeToString(body)}, nil
}

// sendTextRequest 执行 Gemini generateContent 请求并解析文本结果
func (c *GeminiClient) sendTextRequest(reqBody GeminiTextRequest) (string, error) {
	model := c.Model

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Gemini: Failed to marshal request: %v\n", err)
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// 替换端点中的 {model} 占位符
	endpoint := c.BaseURL + c.Endpoint
	endpoint = strings.ReplaceAll(endpoint, "{model}", model)
	url := fmt.Sprintf("%s?key=%s", endpoint, c.APIKey)

	// 打印请求信息（隐藏 API Key）
	safeURL := strings.Replace(url, c.APIKey, "***", 1)
	fmt.Printf("Gemini: Sending request to: %s\n", safeURL)
	requestPreview := string(jsonData)
	if len(jsonData) > 300 {
		requestPreview = string(jsonData[:300]) + "..."
	}
	fmt.Printf("Gemini: Request body: %s\n", requestPreview)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Gemini: Failed to create request: %v\n", err)
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("Gemini: Executing HTTP request...\n")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("Gemini: HTTP request failed: %v\n", err)
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Gemini: Received response with status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Gemini: Failed to read response body: %v\n", err)
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Gemini: API error (status %d): %s\n", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 打印响应体用于调试
	bodyPreview := string(body)
	if len(body) > 500 {
		bodyPreview = string(body[:500]) + "..."
	}
	fmt.Printf("Gemini: Response body: %s\n", bodyPreview)

	var result GeminiTextResponse
	if err := json.Unmarshal(body, &result); err != nil {
		errorPreview := string(body)
		if len(body) > 200 {
			errorPreview = string(body[:200])
		}
		fmt.Printf("Gemini: Failed to parse response: %v\n", err)
		return "", fmt.Errorf("parse response: %w, body preview: %s", err, errorPreview)
	}

	fmt.Printf("Gemini: Successfully parsed response, candidates count: %d\n", len(result.Candidates))

	if len(result.Candidates) == 0 {
		fmt.Printf("Gemini: No candidates in response\n")
		return "", fmt.Errorf("no candidates in response")
	}

	if len(result.Candidates[0].Content.Parts) == 0 {
		fmt.Printf("Gemini: No parts in first candidate\n")
		return "", fmt.Errorf("no parts in response")
	}

	responseText := result.Candidates[0].Content.Parts[0].Text
	fmt.Printf("Gemini: Generated text: %s\n", responseText)

	return responseText, nil
}

func (c *GeminiClient) GenerateImage(prompt string, size string, n int) ([]string, error) {
	return nil, fmt.Errorf("GenerateImage not implemented for Gemini client")
}

func (c *GeminiClient) TestConnection() error {
	fmt.Printf("Gemini: TestConnection called with BaseURL=%s, Model=%s, Endpoint=%s\n", c.BaseURL, c.Model, c.Endpoint)
	_, err := c.GenerateText("Hello", "")
	if err != nil {
		fmt.Printf("Gemini: TestConnection failed: %v\n", err)
	} else {
		fmt.Printf("Gemini: TestConnection succeeded\n")
	}
	return err
}
