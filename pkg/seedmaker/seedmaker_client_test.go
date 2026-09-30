package seedmaker_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/wangbugeng/wangbugeng-drama/pkg/image"
	sm "github.com/wangbugeng/wangbugeng-drama/pkg/seedmaker"
	"github.com/wangbugeng/wangbugeng-drama/pkg/video"
)

const testPNGDataURI = "data:image/png;base64," +
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+M8AAAMBAQDJ/pLvAAAAAElFTkSuQmCC"

type mockServer struct {
	t      *testing.T
	faceOK bool
	mu     sync.Mutex
	last   map[string]any
	srv    *httptest.Server
}

func newMockServer(t *testing.T, faceOK bool) *mockServer {
	m := &mockServer{t: t, faceOK: faceOK, last: map[string]any{}}
	mux := http.NewServeMux()

	mux.HandleFunc("/uploads", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		if _, err := base64.StdEncoding.DecodeString(req["base64"]); err != nil {
			t.Errorf("uploads base64 invalid: %v", err)
		}
		if req["mimeType"] != "image/png" {
			t.Errorf("uploads mime = %q", req["mimeType"])
		}
		writeData(w, map[string]any{"key": "up-key-1", "url": "https://cdn.example.com/u1.png", "mimeType": "image/png"})
	})

	mux.HandleFunc("/face/assets", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		if req["key"] != "up-key-1" {
			t.Errorf("face assets key = %q", req["key"])
		}
		if faceOK {
			writeData(w, map[string]any{"assetId": "a1", "volcAssetId": "asset://face-a1"})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "FACE_DETECT_FAILED", "message": "no face"},
		})
	})

	mux.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		m.capture(r)
		w.WriteHeader(http.StatusCreated)
		writeData(w, map[string]any{"id": "img-123", "status": "pending"})
	})

	mux.HandleFunc("/images/img-123", func(w http.ResponseWriter, r *http.Request) {
		writeData(w, map[string]any{
			"id": "img-123", "status": "succeeded",
			"persistedImageUrls": []string{"https://cdn.example.com/out1.png", "https://cdn.example.com/out2.png"},
		})
	})

	mux.HandleFunc("/videos", func(w http.ResponseWriter, r *http.Request) {
		m.capture(r)
		w.WriteHeader(http.StatusCreated)
		writeData(w, map[string]any{"id": "vid-1", "status": "pending"})
	})

	mux.HandleFunc("/videos/vid-1", func(w http.ResponseWriter, r *http.Request) {
		writeData(w, map[string]any{
			"id": "vid-1", "status": "succeeded",
			"persistedVideoUrl": "https://cdn.example.com/v1.mp4", "videoDuration": 5,
		})
	})

	m.srv = httptest.NewServer(mux)
	return m
}

func (m *mockServer) capture(r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req map[string]any
	_ = json.Unmarshal(body, &req)
	m.mu.Lock()
	m.last = req
	m.mu.Unlock()
}

func (m *mockServer) close() { m.srv.Close() }

func writeData(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func TestSharedResolveFaceAndFallback(t *testing.T) {
	m := newMockServer(t, true)
	defer m.close()
	c := sm.New(m.srv.URL, "k")

	url, err := c.ResolveReferenceAsFace(testPNGDataURI)
	if err != nil {
		t.Fatalf("resolve face: %v", err)
	}
	if url != "asset://face-a1" {
		t.Fatalf("want asset://face-a1, got %q", url)
	}

	// 缓存：第二次不再请求（结果一致即可）
	url2, _ := c.ResolveReferenceAsFace(testPNGDataURI)
	if url2 != url {
		t.Fatalf("cache mismatch: %q vs %q", url2, url)
	}

	got, err := c.ResolveReference("https://x.com/a.png")
	if err != nil || got != "https://x.com/a.png" {
		t.Fatalf("remote resolve: %q %v", got, err)
	}

	mf := newMockServer(t, false)
	defer mf.close()
	cf := sm.New(mf.srv.URL, "k")
	url3, err := cf.ResolveReferenceAsFace(testPNGDataURI)
	if err != nil {
		t.Fatalf("face failure should fallback, got err %v", err)
	}
	if url3 != "https://cdn.example.com/u1.png" {
		t.Fatalf("fallback url = %q", url3)
	}
}

func TestImageClientEditModeUploadRefAndPoll(t *testing.T) {
	m := newMockServer(t, true)
	defer m.close()
	ic := image.NewSeedMakerImageClient(m.srv.URL, "k", "doubao-seedream-5-0-flash-260915")

	res, err := ic.GenerateImage("赛博少女，霓虹灯",
		image.WithReferenceImages([]string{testPNGDataURI}),
		image.WithDimensions(1440, 2560),
	)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if res.TaskID != "img-123" || res.Completed {
		t.Fatalf("initial result = %+v", res)
	}

	if m.last["mode"] != "image_edit" {
		t.Errorf("mode = %v, want image_edit", m.last["mode"])
	}
	if m.last["ratio"] != "9:16" {
		t.Errorf("ratio = %v, want 9:16", m.last["ratio"])
	}
	content, _ := m.last["content"].([]any)
	foundRef, foundSettings := false, false
	for _, p := range content {
		pm := p.(map[string]any)
		if pm["type"] == "image_settings" {
			foundSettings = true
			if pm["size"] != "2K" {
				t.Errorf("size = %v", pm["size"])
			}
		}
		if iu, ok := pm["image_url"].(map[string]any); ok && iu["url"] == "https://cdn.example.com/u1.png" {
			foundRef = true
		}
	}
	if !foundRef {
		t.Errorf("content 缺少普通上传 URL 参考: %v", content)
	}
	if !foundSettings {
		t.Errorf("content 缺少 image_settings")
	}

	done, err := ic.GetTaskStatus("img-123")
	if err != nil || !done.Completed || done.ImageURL != "https://cdn.example.com/out1.png" {
		t.Fatalf("poll = %+v err=%v", done, err)
	}
}

func TestImageClientTextToImageDefaults(t *testing.T) {
	m := newMockServer(t, true)
	defer m.close()
	ic := image.NewSeedMakerImageClient(m.srv.URL, "k", "doubao-seedream-5-0-260128")
	if _, err := ic.GenerateImage("远景城市天际线，黄昏"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if m.last["mode"] != "text_to_image" {
		t.Errorf("mode = %v", m.last["mode"])
	}
	if m.last["ratio"] != "16:9" {
		t.Errorf("default ratio = %v, want 16:9", m.last["ratio"])
	}
}

func TestProviderRegistration(t *testing.T) {
	if image.NormalizeProvider("seedmaker") != image.ProviderSeedMaker {
		t.Fatal("NormalizeProvider(seedmaker) 未注册")
	}
	c, err := image.NewClient(&image.ClientConfig{
		Provider: image.ProviderSeedMaker,
		BaseURL:  "https://example.test/api/v1",
		APIKey:   "k",
		Model:    "m",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, ok := c.(*image.SeedMakerImageClient); !ok {
		t.Fatalf("工厂返回类型 %T, 想要 *SeedMakerImageClient", c)
	}
}

func TestVideoModesAndPoll(t *testing.T) {
	cases := []struct {
		name       string
		imageURL   string
		firstFrame string
		lastFrame  string
		refs       []string
		wantMode   string
		wantRoles  []string
	}{
		{"text_to_video", "", "", "", nil, "text_to_video", nil},
		{"image_first_frame", "https://x.com/f.png", "", "", nil, "image_first_frame", []string{"first_frame"}},
		{"first_last", "", "https://x.com/a.png", "https://x.com/b.png", nil,
			"image_first_last_frame", []string{"first_frame", "last_frame"}},
		{"multimodal", "", "", "", []string{"https://x.com/r1.png"},
			"multimodal_reference", []string{"reference_image"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newMockServer(t, true)
			defer m.close()
			vc := video.NewSeedMakerVideoClient(m.srv.URL, "k", "ep-20260625152214-vh6v2")

			var opts []video.VideoOption
			if tc.firstFrame != "" {
				opts = append(opts, video.WithFirstFrame(tc.firstFrame))
			}
			if tc.lastFrame != "" {
				opts = append(opts, video.WithLastFrame(tc.lastFrame))
			}
			if tc.refs != nil {
				opts = append(opts, video.WithReferenceImages(tc.refs))
			}

			res, err := vc.GenerateVideo(tc.imageURL, "镜头缓慢推进", opts...)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if res.TaskID != "vid-1" || res.Completed {
				t.Fatalf("initial = %+v", res)
			}
			if m.last["mode"] != tc.wantMode {
				t.Errorf("mode = %v want %v; body=%v", m.last["mode"], tc.wantMode, m.last)
			}
			if m.last["resolution"] != "720p" || m.last["duration"].(float64) != 5 || m.last["ratio"] != "16:9" {
				t.Errorf("默认参数错误: res=%v dur=%v ratio=%v", m.last["resolution"], m.last["duration"], m.last["ratio"])
			}
			content, _ := m.last["content"].([]any)
			var roles []string
			for _, p := range content {
				if r := p.(map[string]any)["role"]; r != nil {
					roles = append(roles, r.(string))
				}
			}
			if len(roles) != len(tc.wantRoles) {
				t.Fatalf("roles = %v want %v", roles, tc.wantRoles)
			}
			for i := range roles {
				if roles[i] != tc.wantRoles[i] {
					t.Errorf("role[%d] = %q want %q", i, roles[i], tc.wantRoles[i])
				}
			}

			done, err := vc.GetTaskStatus("vid-1")
			if err != nil || !done.Completed || done.VideoURL != "https://cdn.example.com/v1.mp4" || done.Duration != 5 {
				t.Fatalf("poll = %+v err=%v", done, err)
			}
		})
	}
}
