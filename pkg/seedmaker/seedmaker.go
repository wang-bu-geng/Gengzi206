package seedmaker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultBaseURL SeedMaker 公共 API 根地址
const DefaultBaseURL = "https://seedmaker.tianlangedu.com.cn/api/v1"

// 任务状态（pending → submitted → running → succeeded | failed）
const (
	StatusPending   = "pending"
	StatusSubmitted = "submitted"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

// APIError SeedMaker 错误信封 {"error":{"code","message"}}
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("seedmaker %s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("seedmaker: %s", e.Message)
}

type errorEnvelope struct {
	Error *APIError `json:"error"`
}

// Client SeedMaker API 共享客户端：鉴权、信封解包、素材上传、人脸库入库
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client

	mu    sync.Mutex
	cache map[string]string
}

func New(baseURL, apiKey string) *Client {
	if strings.TrimRight(baseURL, "/") == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		// 单次 HTTP 调用给足时间（上传/入库可能较慢），任务等待走轮询
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		cache:      make(map[string]string),
	}
}

// do 发送请求并解包 {"data": ...}；失败时解析 error 信封
// rateLimitBackoffs 命中限流（429）或服务端临时错误（5xx）时的退避间隔
var rateLimitBackoffs = []time.Duration{5 * time.Second, 15 * time.Second, 30 * time.Second}

func (c *Client) do(method, path string, body any, out any) error {
	// 请求体预先序列化，便于重试时重放
	var rawReq []byte
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		rawReq = raw
	}

	var lastErr error
	for attempt := 0; attempt <= len(rateLimitBackoffs); attempt++ {
		if attempt > 0 {
			wait := rateLimitBackoffs[attempt-1]
			log.Printf("[SeedMaker] rate limited / temporary error, retry %d/%d after %s: %s %s",
				attempt, len(rateLimitBackoffs), wait, method, path)
			time.Sleep(wait)
		}

		var reader io.Reader
		if rawReq != nil {
			reader = bytes.NewReader(rawReq)
		}

		req, err := http.NewRequest(method, c.BaseURL+path, reader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.APIKey)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("send request: %w", err)
			continue // 网络错误同样重试
		}

		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("read response: %w", readErr)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// 429 限流与 5xx 临时错误：退避后重试；其他错误立即返回
			if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) &&
				attempt < len(rateLimitBackoffs) {
				var env errorEnvelope
				if json.Unmarshal(raw, &env) == nil && env.Error != nil {
					lastErr = env.Error
				} else {
					lastErr = fmt.Errorf("seedmaker API error (status %d): %s", resp.StatusCode, truncate(string(raw), 500))
				}
				continue
			}
			var env errorEnvelope
			if json.Unmarshal(raw, &env) == nil && env.Error != nil {
				return env.Error
			}
			return fmt.Errorf("seedmaker API error (status %d): %s", resp.StatusCode, truncate(string(raw), 500))
		}

		if out != nil {
			// 所有响应包在 data 字段内
			var wrap struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(raw, &wrap); err != nil {
				return fmt.Errorf("parse response envelope: %w", err)
			}
			if len(wrap.Data) > 0 {
				if err := json.Unmarshal(wrap.Data, out); err != nil {
					return fmt.Errorf("parse response data: %w", err)
				}
			}
		}
		return nil
	}
	return lastErr
}

// Do 对外暴露的通用 JSON 调用（供图片/视频客户端使用）
func (c *Client) Do(method, path string, body any, out any) error {
	return c.do(method, path, body, out)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

type uploadResponse struct {
	AssetID  string `json:"assetId"`
	URL      string `json:"url"`
	Key      string `json:"key"`
	MIMEType string `json:"mimeType"`
}

type faceAssetResponse struct {
	AssetID     string `json:"assetId"`
	VolcAssetID string `json:"volcAssetId"`
}

// ResolveReference 把任意形态的参考图转换为 SeedMaker 可用的 URL：
//   - http(s) 远程图：直接返回（公开可抓取的前提由调用方保证）
//   - data:URI / 本地路径：上传到 /uploads 换取临时公网 URL
//
// 人物图请用 ResolveReferenceAsFace：火山肖像审核要求人物必须走人脸库 asset:// URI。
func (c *Client) ResolveReference(ref string) (string, error) {
	return c.resolve(ref, false)
}

// ResolveReferenceAsFace 上传后尝试进入人脸库拿 asset:// URI；
// 非人物图入库会失败（场景/道具），自动回退为普通上传 URL。
func (c *Client) ResolveReferenceAsFace(ref string) (string, error) {
	return c.resolve(ref, true)
}

func (c *Client) resolve(ref string, asFace bool) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", nil
	}

	c.mu.Lock()
	if cached, ok := c.cache[ref]; ok {
		c.mu.Unlock()
		return cached, nil
	}
	c.mu.Unlock()

	// 远程图没有 upload key，公开 API 无法对其做人脸入库
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return c.remember(ref, ref), nil
	}

	var b64, mime string
	switch {
	case strings.HasPrefix(ref, "data:"):
		m, data, err := parseDataURI(ref)
		if err != nil {
			return "", err
		}
		b64, mime = data, m
	default:
		raw, err := os.ReadFile(ref)
		if err != nil {
			return "", fmt.Errorf("read local reference %s: %w", ref, err)
		}
		b64 = base64.StdEncoding.EncodeToString(raw)
		mime = mimeByExt(ref)
	}

	var up uploadResponse
	if err := c.do(http.MethodPost, "/uploads", map[string]string{
		"base64":   b64,
		"mimeType": mime,
	}, &up); err != nil {
		return "", fmt.Errorf("upload reference: %w", err)
	}
	if up.URL == "" {
		return "", fmt.Errorf("upload reference: empty url in response")
	}

	if !asFace || up.Key == "" {
		return c.remember(ref, up.URL), nil
	}

	var face faceAssetResponse
	if err := c.do(http.MethodPost, "/face/assets", map[string]string{
		"key": up.Key,
	}, &face); err != nil || face.VolcAssetID == "" {
		// 非人物（场景/道具）或真人图入库失败：素材本身已上传，回退普通 URL
		return c.remember(ref, up.URL), nil
	}
	return c.remember(ref, face.VolcAssetID), nil
}

func (c *Client) remember(ref, resolved string) string {
	c.mu.Lock()
	c.cache[ref] = resolved
	c.mu.Unlock()
	return resolved
}

// ── 工具函数 ──────────────────────────────────────────────

func parseDataURI(dataURI string) (mime, b64 string, err error) {
	const prefix = "data:"
	if !strings.HasPrefix(dataURI, prefix) {
		return "", "", fmt.Errorf("not a data URI")
	}
	comma := strings.Index(dataURI, ",")
	if comma < 0 {
		return "", "", fmt.Errorf("malformed data URI")
	}
	head := dataURI[len(prefix):comma]
	b64 = dataURI[comma+1:]
	mime = "image/png"
	if head != "" {
		parts := strings.Split(head, ";")
		if parts[0] != "" {
			mime = parts[0]
		}
	}
	return mime, b64, nil
}

func mimeByExt(path string) string {
	switch {
	case strings.HasSuffix(strings.ToLower(path), ".jpg"), strings.HasSuffix(strings.ToLower(path), ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(strings.ToLower(path), ".webp"):
		return "image/webp"
	default:
		return "image/png"
	}
}

// IsTerminal 状态是否已结束
func IsTerminal(status string) bool {
	return status == StatusSucceeded || status == StatusFailed
}
