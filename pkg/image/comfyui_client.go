package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ComfyUIClient 调用本地 ComfyUI API 生图
type ComfyUIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	OutputDir  string // 图片输出目录
}

// ComfyUIPrompt 提交给 ComfyUI 的 workfow
type ComfyUIPrompt struct {
	Prompt map[string]interface{} `json:"prompt"`
}

// ComfyUIQueueResponse 提交 prompt 后返回
type ComfyUIQueueResponse struct {
	PromptID   string                 `json:"prompt_id"`
	Number     int                    `json:"number"`
	NodeErrors map[string]interface{} `json:"node_errors,omitempty"`
}

// ComfyUIHistoryItem 轮询任务状态时获取
// Note: Newer ComfyUI versions return "prompt" as an array, not a map.
// We use json.RawMessage to accept either, but don't deserialize it.
type ComfyUIHistoryItem struct {
	Prompt  json.RawMessage        `json:"prompt"`
	Outputs map[string]interface{} `json:"outputs"`
	Status  *struct {
		Completed bool   `json:"completed"`
		StatusStr string `json:"status_str,omitempty"`
	} `json:"status,omitempty"`
}

// SDXLPromptBuilder 从 JSON 工作流模板构建 workflow
type SDXLPromptBuilder struct {
	Checkpoint string
	Positive   string
	Negative   string
	Width      int
	Height     int
	Steps      int
	CfgScale   float64
	Seed       int64
	BatchSize  int
	Prefix     string
}

// templateData 传递给 JSON 模板的参数
type templateData struct {
	MODEL      string
	POSITIVE   string
	NEGATIVE   string
	WIDTH      int
	HEIGHT     int
	STEPS      int
	CFG        float64
	SEED       int64
	BATCH_SIZE int
	PREFIX     string
}

// LoadWorkflowTemplate 从 JSON 文件加载工作流模板
func LoadWorkflowTemplate(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read workflow template: %w", err)
	}
	return string(data), nil
}

// defaultWorkflowJSON 默认内嵌的工作流 JSON（当文件加载失败时回退）
const defaultWorkflowJSON = `{
  "4": {
    "class_type": "CheckpointLoaderSimple",
    "inputs": { "ckpt_name": "__MODEL__" }
  },
  "5": {
    "class_type": "EmptyLatentImage",
    "inputs": {
      "width": "__WIDTH__",
      "height": "__HEIGHT__",
      "batch_size": "__BATCH_SIZE__"
    }
  },
  "6": {
    "class_type": "CLIPTextEncode",
    "inputs": { "text": "__POSITIVE__", "clip": ["4", 1] }
  },
  "7": {
    "class_type": "CLIPTextEncode",
    "inputs": { "text": "__NEGATIVE__", "clip": ["4", 1] }
  },
  "3": {
    "class_type": "KSampler",
    "inputs": {
      "seed": "__SEED__", "steps": "__STEPS__", "cfg": "__CFG__",
      "sampler_name": "dpmpp_2m", "scheduler": "karras", "denoise": 1.0,
      "model": ["4", 0], "positive": ["6", 0], "negative": ["7", 0],
      "latent_image": ["5", 0]
    }
  },
  "8": {
    "class_type": "VAEDecode",
    "inputs": { "samples": ["3", 0], "vae": ["4", 2] }
  },
  "9": {
    "class_type": "SaveImage",
    "inputs": { "filename_prefix": "__PREFIX__", "images": ["8", 0] }
  }
}`

// Build 使用字符串替换占位符，生成最终 workflow 的 map[string]interface{}
//
// 占位符约定：
//
//	__MODEL__, __POSITIVE__, __NEGATIVE__, __PREFIX__ → 字符串（替换后带引号）
//	__WIDTH__, __HEIGHT__, __STEPS__, __CFG__, __SEED__, __BATCH_SIZE__ → 数值（替换后不带引号）
func (b *SDXLPromptBuilder) Build(templateJSON string) (map[string]interface{}, error) {
	// 第一步：替换数值占位符 — 它们在JSON里写成 "__WIDTH__"，需要变成数值
	// 用 "§NUM§VALUE§" 作为桥梁，确保最后JSON unmarshal时是数字类型
	s := templateJSON
	s = strings.ReplaceAll(s, `"__WIDTH__"`, fmt.Sprintf("%d", b.Width))
	s = strings.ReplaceAll(s, `"__HEIGHT__"`, fmt.Sprintf("%d", b.Height))
	s = strings.ReplaceAll(s, `"__STEPS__"`, fmt.Sprintf("%d", b.Steps))
	s = strings.ReplaceAll(s, `"__CFG__"`, fmt.Sprintf("%v", b.CfgScale))
	s = strings.ReplaceAll(s, `"__SEED__"`, fmt.Sprintf("%d", b.Seed))
	s = strings.ReplaceAll(s, `"__BATCH_SIZE__"`, fmt.Sprintf("%d", b.BatchSize))

	// 第二步：替换字符串占位符（保留引号包裹）
	s = strings.ReplaceAll(s, `__MODEL__`, b.Checkpoint)
	s = strings.ReplaceAll(s, `__POSITIVE__`, b.Positive)
	s = strings.ReplaceAll(s, `__NEGATIVE__`, b.Negative)
	s = strings.ReplaceAll(s, `__PREFIX__`, b.Prefix)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, fmt.Errorf("unmarshal generated workflow: %w, raw: %s", err, s[:min(200, len(s))])
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NewComfyUIClient(baseURL string) *ComfyUIClient {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8188"
	}
	outputDir := filepath.Join(os.TempDir(), "wangbugeng-comfyui-outputs")
	os.MkdirAll(outputDir, 0755)
	return &ComfyUIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
		OutputDir: outputDir,
	}
}

// getWorkflowDir 返回工作流模板目录，优先项目内 workflows/，回退到 ComfyUI 标准目录
func (c *ComfyUIClient) getWorkflowDir() string {
	candidates := []string{
		filepath.Join("workflows"),
	}
	if comfyDir := getComfyUIDir(); comfyDir != "" {
		candidates = append(candidates, filepath.Join(comfyDir, "workflows"))
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}

func (c *ComfyUIClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
	options := &ImageOptions{
		Width:    1024,
		Height:   576,
		Steps:    20,
		CfgScale: 7.0,
		Seed:     -1,
	}
	for _, opt := range opts {
		opt(options)
	}

	if err := c.checkHealth(); err != nil {
		return nil, fmt.Errorf("ComfyUI not accessible: %w", err)
	}

	checkpoint, err := c.getAvailableCheckpoint()
	if err != nil {
		return nil, fmt.Errorf("no checkpoint available: %w", err)
	}

	seed := options.Seed
	if seed < 0 {
		seed = time.Now().UnixNano()
	}

	negativePrompt := options.NegativePrompt
	if negativePrompt == "" {
		negativePrompt = "nsfw, lowres, bad anatomy, bad hands, text, error, missing fingers, extra digit, fewer digits, cropped, worst quality, low quality, normal quality, jpeg artifacts, signature, watermark, username, blurry, ugly, deformed"
	}

	// Load workflow template from JSON file (or fallback to embedded default)
	workflowDir := c.getWorkflowDir()
	templateJSON := defaultWorkflowJSON
	if workflowDir != "" {
		templatePath := filepath.Join(workflowDir, "sdxl_base.json")
		if loaded, err := LoadWorkflowTemplate(templatePath); err == nil {
			templateJSON = loaded
		} else {
			fmt.Printf("[ComfyUI] Warning: cannot load workflow template %s: %v, using embedded default\n", templatePath, err)
		}
	}

	builder := &SDXLPromptBuilder{
		Checkpoint: checkpoint,
		Positive:   prompt,
		Negative:   negativePrompt,
		Width:      options.Width,
		Height:     options.Height,
		Steps:      options.Steps,
		CfgScale:   options.CfgScale,
		Seed:       seed,
		BatchSize:  1,
		Prefix:     "wangbugeng_drama",
	}

	workflow, err := builder.Build(templateJSON)
	if err != nil {
		return nil, fmt.Errorf("build workflow: %w", err)
	}

	promptID, err := c.queuePrompt(workflow)
	if err != nil {
		return nil, fmt.Errorf("queue prompt failed: %w", err)
	}

	result, err := c.waitForCompletion(promptID)
	if err != nil {
		return nil, fmt.Errorf("image generation failed: %w", err)
	}

	return result, nil
}

func (c *ComfyUIClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/history/%s", c.BaseURL, taskID))
	if err != nil {
		return nil, fmt.Errorf("query task status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &ImageResult{
			TaskID:    taskID,
			Status:    "processing",
			Completed: false,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query task status returned %d", resp.StatusCode)
	}

	var history map[string]ComfyUIHistoryItem
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}

	item, ok := history[taskID]
	if !ok {
		return &ImageResult{
			TaskID:    taskID,
			Status:    "processing",
			Completed: false,
		}, nil
	}

	if item.Status != nil && item.Status.Completed {
		imageURL, err := c.extractImageFromOutput(item.Outputs)
		if err != nil {
			return nil, fmt.Errorf("extract image: %w", err)
		}
		return &ImageResult{
			TaskID:    taskID,
			Status:    "completed",
			ImageURL:  imageURL,
			Completed: true,
		}, nil
	}

	return &ImageResult{
		TaskID:    taskID,
		Status:    "processing",
		Completed: false,
	}, nil
}

func (c *ComfyUIClient) checkHealth() error {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/system_stats", c.BaseURL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	resp2, err := c.HTTPClient.Get(fmt.Sprintf("%s/object_info", c.BaseURL))
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("ComfyUI not healthy: system_stats=%d, object_info=%d", resp.StatusCode, resp2.StatusCode)
}

func (c *ComfyUIClient) getAvailableCheckpoint() (string, error) {
	// Check ComfyUI model directory directly via common paths
	possibleDirs := []string{
		filepath.Join(getComfyUIDir(), "models", "checkpoints"),
	}

	// Also check well-known default paths
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		possibleDirs = append(possibleDirs,
			filepath.Join(home, "Documents", "ComfyUI", "models", "checkpoints"),
		)
	}

	for _, dir := range possibleDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext == ".safetensors" || ext == ".ckpt" || ext == ".pt" {
				if containsSDXL(name) || containsSDXL(filepath.Base(name)) {
					return name, nil
				}
			}
		}
		// Fallback: pick first checkpoint if no SDXL found
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := filepath.Ext(entry.Name())
			if ext == ".safetensors" || ext == ".ckpt" || ext == ".pt" {
				return entry.Name(), nil
			}
		}
	}

	return "", fmt.Errorf("no checkpoint file found in ComfyUI models/checkpoints directory")
}

func getComfyUIDir() string {
	// Try to determine ComfyUI root from common locations
	if dir := os.Getenv("COMFYUI_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "Documents", "ComfyUI"),
		filepath.Join(home, "ComfyUI"),
		"/opt/ComfyUI",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "models")); err == nil {
			return c
		}
	}
	return ""
}

func (c *ComfyUIClient) findCheckpointFiles() (string, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/view?filename=checkpoints", c.BaseURL))
	if err != nil {
		return "", fmt.Errorf("cannot find checkpoints via API: %w", err)
	}
	resp.Body.Close()
	return "", fmt.Errorf("no checkpoint found via API")
}

func (c *ComfyUIClient) queuePrompt(workflow map[string]interface{}) (string, error) {
	reqBody := map[string]interface{}{
		"prompt": workflow,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal prompt: %w", err)
	}

	resp, err := c.HTTPClient.Post(fmt.Sprintf("%s/prompt", c.BaseURL), "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("send prompt: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ComfyUI prompt error (status %d): %s", resp.StatusCode, string(body))
	}

	var queueResp ComfyUIQueueResponse
	if err := json.Unmarshal(body, &queueResp); err != nil {
		return "", fmt.Errorf("parse queue response: %w, body: %s", err, string(body))
	}

	if queueResp.PromptID == "" {
		return "", fmt.Errorf("no prompt_id in response: %s", string(body))
	}

	if len(queueResp.NodeErrors) > 0 {
		nodeErrJSON, _ := json.Marshal(queueResp.NodeErrors)
		return "", fmt.Errorf("workflow node errors: %s", string(nodeErrJSON))
	}

	return queueResp.PromptID, nil
}

func (c *ComfyUIClient) waitForCompletion(promptID string) (*ImageResult, error) {
	maxAttempts := 120
	pollInterval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(pollInterval)

		result, err := c.GetTaskStatus(promptID)
		if err != nil {
			fmt.Printf("[ComfyUI] Poll error: %v (attempt %d/%d)\n", err, i+1, maxAttempts)
			continue
		}
		if result.Completed {
			return result, nil
		}
		fmt.Printf("[ComfyUI] Waiting... prompt_id=%s (attempt %d/%d)\n", promptID, i+1, maxAttempts)
	}

	return nil, fmt.Errorf("timeout: ComfyUI image generation took too long for prompt %s", promptID)
}

func (c *ComfyUIClient) extractImageFromOutput(outputs map[string]interface{}) (string, error) {
	for nodeID, output := range outputs {
		nodeOutput, ok := output.(map[string]interface{})
		if !ok {
			continue
		}
		images, ok := nodeOutput["images"].([]interface{})
		if !ok {
			continue
		}
		if len(images) == 0 {
			continue
		}
		img, ok := images[0].(map[string]interface{})
		if !ok {
			continue
		}
		filename, _ := img["filename"].(string)
		subfolder, _ := img["subfolder"].(string)
		imgType, _ := img["type"].(string)

		if filename != "" {
			imageURL := fmt.Sprintf("%s/view?filename=%s&subfolder=%s&type=%s",
				c.BaseURL, filename, subfolder, imgType)

			localPath := filepath.Join(c.OutputDir, filename)
			if err := c.downloadFile(imageURL, localPath); err == nil {
				fmt.Printf("[ComfyUI] Image saved locally: %s\n", localPath)
			}

			return imageURL, nil
		}
		_ = nodeID
	}
	return "", fmt.Errorf("no image found in ComfyUI output")
}

func (c *ComfyUIClient) downloadFile(url, dest string) error {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func containsSDXL(name string) bool {
	sdxlKeywords := []string{"sdxl", "sd_xl", "SDXL", "stable diffusion xl", "animagine", "noobai", "realvis"}
	for _, kw := range sdxlKeywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (c *ComfyUIClient) uploadImage(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy image: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/upload/image", c.BaseURL), &buf)
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}

	return result.Name, nil
}
