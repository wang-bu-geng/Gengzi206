package utils

import "testing"

func TestJSONSanityCases(t *testing.T) {
	// 1. 字符串中段含非法 UTF-8 字节但结构完整：应能解析，不应判为截断
	broken := "[{\"action\": \"灯" + string([]byte{0xE9, 0x87}) + "围\"}, {\"n\": 2}]"
	var out []map[string]any
	if err := SafeParseAIJSON(broken, &out); err != nil {
		t.Fatalf("broken utf8 parse: %v", err)
	}
	if IsLikelyTruncated(broken) {
		t.Fatal("structurally complete response with bad bytes must NOT be flagged truncated")
	}
	// 2. 裸字符串值
	bare := "[\n  {\"shot_type\":远景, \"n\": 1}\n]"
	var out2 []map[string]any
	if err := SafeParseAIJSON(bare, &out2); err != nil {
		t.Fatalf("bare value parse: %v", err)
	}
	if out2[0]["shot_type"] != "远景" {
		t.Fatalf("got %v", out2[0]["shot_type"])
	}
	// 3. 真截断（末对象不完整）：应判为截断
	trunc := "[{\"a\": 1}, {\"b\": \"xxx"
	if !IsLikelyTruncated(trunc) {
		t.Fatal("truncated JSON must be flagged")
	}
	// 4. 正常 JSON 不判截断
	if IsLikelyTruncated("[{\"a\": 1}]") {
		t.Fatal("normal JSON flagged")
	}
}
