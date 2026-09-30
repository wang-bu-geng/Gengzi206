package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// SafeParseAIJSON 安全地解析AI返回的JSON，处理常见的格式问题
// 包括：
// 1. 移除Markdown代码块标记
// 2. 提取JSON对象
// 3. 清理多余的空白和换行
// 4. 尝试修复截断的JSON
// 5. 提供详细的错误信息
func SafeParseAIJSON(aiResponse string, v interface{}) error {
	if aiResponse == "" {
		return fmt.Errorf("AI返回内容为空")
	}

	// 0. 净化非法 UTF-8 字节（个别模型会在长文本中输出损坏的汉字编码，
	// 会导致 Go JSON 解析器直接报 "invalid UTF-8 in string"）
	aiResponse = strings.ToValidUTF8(aiResponse, "\ufffd")

	// 1. 移除可能的Markdown代码块标记
	cleaned := strings.TrimSpace(aiResponse)
	// 移除开头的 ```json 或 ```
	cleaned = regexp.MustCompile("(?m)^```json\\s*").ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile("(?m)^```\\s*").ReplaceAllString(cleaned, "")
	// 移除结尾的 ```
	cleaned = regexp.MustCompile("(?m)```\\s*$").ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	// 2. 提取JSON (支持对象 {} 和数组 [])
	var jsonMatch string

	// 优先尝试提取完整的JSON（对象或数组）
	// 先尝试对象格式
	if strings.HasPrefix(cleaned, "{") {
		jsonRegex := regexp.MustCompile(`(?s)\{.*\}`)
		jsonMatch = jsonRegex.FindString(cleaned)
	}

	// 如果没找到对象，尝试数组格式
	if jsonMatch == "" && strings.HasPrefix(cleaned, "[") {
		jsonRegex := regexp.MustCompile(`(?s)\[.*\]`)
		jsonMatch = jsonRegex.FindString(cleaned)
	}

	// 如果还是没找到，尝试从中间提取
	if jsonMatch == "" {
		// 尝试对象
		objRegex := regexp.MustCompile(`(?s)\{.*\}`)
		jsonMatch = objRegex.FindString(cleaned)

		// 如果对象没找到，尝试数组
		if jsonMatch == "" {
			arrRegex := regexp.MustCompile(`(?s)\[.*\]`)
			jsonMatch = arrRegex.FindString(cleaned)
		}
	}

	if jsonMatch == "" {
		return fmt.Errorf("响应中未找到有效的JSON对象或数组，原始响应: %s", truncateString(aiResponse, 200))
	}

	// 3. 尝试解析JSON
	err := json.Unmarshal([]byte(jsonMatch), v)
	if err == nil {
		return nil // 解析成功
	}

	// 4. 如果解析失败，尝试修复截断的JSON
	fixedJSON := attemptJSONRepair(jsonMatch)
	if fixedJSON != jsonMatch {
		if err := json.Unmarshal([]byte(fixedJSON), v); err == nil {
			return nil // 修复后解析成功
		}
	}

	// 5. 检测是否是响应被截断导致的问题
	if isTruncated(jsonMatch) {
		return fmt.Errorf(
			"AI响应可能被截断，导致JSON不完整。\n请尝试：\n1. 增加maxTokens参数\n2. 简化输入内容\n3. 使用更强大的模型\n\n原始错误: %s\n响应长度: %d\n响应末尾: %s",
			err.Error(),
			len(jsonMatch),
			truncateString(jsonMatch[maxInt(0, len(jsonMatch)-200):], 200),
		)
	}

	// 6. 提供详细的错误上下文
	if jsonErr, ok := err.(*json.SyntaxError); ok {
		errorPos := int(jsonErr.Offset)
		start := maxInt(0, errorPos-100)
		end := minInt(len(jsonMatch), errorPos+100)

		context := jsonMatch[start:end]
		marker := strings.Repeat(" ", errorPos-start) + "^"

		return fmt.Errorf(
			"JSON解析失败: %s\n错误位置附近:\n%s\n%s",
			jsonErr.Error(),
			context,
			marker,
		)
	}

	return fmt.Errorf("JSON解析失败: %w\n原始响应: %s", err, truncateString(jsonMatch, 300))
}

// IsLikelyTruncated 判断 AI 响应是否疑似因输出 token 上限被截断。
// 判据：JSON 未以闭合括号结尾、括号/引号不配对（调用方可据此提高 max_tokens 重试）。
// 注意：孤立的非法 UTF-8 字节不算截断，由 SafeParseAIJSON 净化处理。
func IsLikelyTruncated(aiResponse string) bool {
	t := strings.TrimSpace(aiResponse)
	if t == "" {
		return false
	}
	// 净化非法字节后再做结构判断：孤立的替换字符只代表坏字，不等于截断；
	// 截断的可靠信号是 JSON 结构不闭合（括号/引号不配对、末尾无闭合括号）
	t = strings.ToValidUTF8(t, "\ufffd")

	// 与 SafeParseAIJSON 相同的围栏清理
	t = regexp.MustCompile("(?m)^```json\\s*").ReplaceAllString(t, "")
	t = regexp.MustCompile("(?m)^```\\s*").ReplaceAllString(t, "")
	t = regexp.MustCompile("(?m)```\\s*$").ReplaceAllString(t, "")
	t = strings.TrimSpace(t)

	// 定位首个 JSON 起点（[ 或 {，取更靠前的）
	idxBracket := strings.Index(t, "[")
	idxBrace := strings.Index(t, "{")
	start := -1
	switch {
	case idxBracket >= 0 && (idxBrace < 0 || idxBracket <= idxBrace):
		start = idxBracket
	case idxBrace >= 0:
		start = idxBrace
	}
	if start < 0 {
		return false
	}
	sub := t[start:]

	// 截到最后一个闭合括号为止（容忍 JSON 之后的解释性文字）
	endBracket := strings.LastIndex(sub, "]")
	endBrace := strings.LastIndex(sub, "}")
	end := endBracket
	if endBrace > end {
		end = endBrace
	}
	if end >= 0 {
		sub = sub[:end+1]
	}

	return isTruncated(sub)
}

// bareStringValueRe 匹配 "key": 裸值, 形式中未加引号的字符串值。
// 值首字符排除：引号、空白、数字、负号、括号、true/false/null 的首字母，
// 避免误伤已加引号值、数字、布尔、null、对象和数组。
var bareStringValueRe = regexp.MustCompile(`("(?:[^"\\]|\\.)*"\s*:\s*)([^",\[\]{}0-9\s\-tfn][^,\n}]*)([,\n}])`)

// quoteBareStringValues 给 AI 漏写引号的裸字符串值补引号，
// 如 "shot_type":远景, → "shot_type":"远景",
func quoteBareStringValues(s string) string {
	return bareStringValueRe.ReplaceAllStringFunc(s, func(m string) string {
		groups := bareStringValueRe.FindStringSubmatch(m)
		val := strings.TrimSpace(groups[2])
		if val == "" || val == "true" || val == "false" || val == "null" {
			return m
		}
		// 值内的裸双引号转义
		val = strings.ReplaceAll(val, `"`, `\"`)
		return groups[1] + `"` + val + `"` + groups[3]
	})
}

// attemptJSONRepair 尝试修复常见的JSON问题
func attemptJSONRepair(jsonStr string) string {
	trimmed := strings.TrimSpace(jsonStr)

	// 1. 给未加引号的裸字符串值补引号（如 "shot_type":远景,）
	trimmed = quoteBareStringValues(trimmed)

	// 2. 处理字符串值中未转义的双引号
	// AI可能在字符串值中使用未转义的引号，如："result": "手机显示着"2006年7月3日""
	// 需要找到字符串内容中的未转义引号并转义它们
	trimmed = fixUnescapedQuotesInStrings(trimmed)

	// 2. 检查是否有未闭合的引号
	if strings.Count(trimmed, `"`)%2 != 0 {
		// 有奇数个引号，尝试补全最后一个引号
		trimmed += `"`
	}

	// 3. 统计括号
	openBraces := strings.Count(trimmed, "{")
	closeBraces := strings.Count(trimmed, "}")
	openBrackets := strings.Count(trimmed, "[")
	closeBrackets := strings.Count(trimmed, "]")

	// 4. 处理多余的闭合括号（从末尾移除）
	// 这是 AI 生成 JSON 时常见的问题
	for closeBrackets > openBrackets && len(trimmed) > 0 {
		// 从末尾向前查找多余的 ]
		lastIdx := strings.LastIndex(trimmed, "]")
		if lastIdx >= 0 {
			trimmed = trimmed[:lastIdx] + trimmed[lastIdx+1:]
			closeBrackets--
		} else {
			break
		}
	}

	for closeBraces > openBraces && len(trimmed) > 0 {
		// 从末尾向前查找多余的 }
		lastIdx := strings.LastIndex(trimmed, "}")
		if lastIdx >= 0 {
			trimmed = trimmed[:lastIdx] + trimmed[lastIdx+1:]
			closeBraces--
		} else {
			break
		}
	}

	// 重新统计括号（因为可能已修改）
	openBraces = strings.Count(trimmed, "{")
	closeBraces = strings.Count(trimmed, "}")
	openBrackets = strings.Count(trimmed, "[")
	closeBrackets = strings.Count(trimmed, "]")

	// 5. 补全未闭合的数组
	for i := 0; i < openBrackets-closeBrackets; i++ {
		trimmed += "]"
	}

	// 6. 补全未闭合的对象
	for i := 0; i < openBraces-closeBraces; i++ {
		trimmed += "}"
	}

	return trimmed
}

// ExtractJSONFromText 从文本中提取JSON对象或数组
func ExtractJSONFromText(text string) string {
	text = strings.TrimSpace(text)

	// 移除Markdown代码块
	text = regexp.MustCompile("(?m)^```json\\s*").ReplaceAllString(text, "")
	text = regexp.MustCompile("(?m)^```\\s*").ReplaceAllString(text, "")
	text = strings.TrimSpace(text)

	// 查找JSON对象
	if idx := strings.Index(text, "{"); idx != -1 {
		if lastIdx := strings.LastIndex(text, "}"); lastIdx != -1 && lastIdx > idx {
			return text[idx : lastIdx+1]
		}
	}

	// 查找JSON数组
	if idx := strings.Index(text, "["); idx != -1 {
		if lastIdx := strings.LastIndex(text, "]"); lastIdx != -1 && lastIdx > idx {
			return text[idx : lastIdx+1]
		}
	}

	return text
}

// ValidateJSON 验证JSON字符串是否有效
func ValidateJSON(jsonStr string) error {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js)
}

// isTruncated 检测JSON字符串是否可能被截断
func isTruncated(jsonStr string) bool {
	trimmed := strings.TrimSpace(jsonStr)
	if len(trimmed) == 0 {
		return false
	}

	// 检查是否以不完整的字符串结尾（引号未闭合）
	lastChar := trimmed[len(trimmed)-1]
	if lastChar != '}' && lastChar != ']' {
		return true
	}

	// 检查括号是否匹配
	openBraces := strings.Count(trimmed, "{")
	closeBraces := strings.Count(trimmed, "}")
	openBrackets := strings.Count(trimmed, "[")
	closeBrackets := strings.Count(trimmed, "]")

	if openBraces != closeBraces || openBrackets != closeBrackets {
		return true
	}

	// 检查引号是否匹配（简化检查，不考虑转义）
	quoteCount := strings.Count(trimmed, `"`)
	if quoteCount%2 != 0 {
		return true
	}

	return false
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func fixUnescapedQuotesInStrings(jsonStr string) string {
	var result strings.Builder
	inString := false
	escaped := false

	for i, char := range jsonStr {
		if escaped {
			result.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' {
			result.WriteRune(char)
			escaped = true
			continue
		}

		if char == '"' {
			if inString {
				nextIdx := i + 1
				for nextIdx < len(jsonStr) && (jsonStr[nextIdx] == ' ' || jsonStr[nextIdx] == 9 || jsonStr[nextIdx] == 10 || jsonStr[nextIdx] == 13) {
					nextIdx++
				}

				if nextIdx >= len(jsonStr) {
					// 引号是输入末尾的字符串闭合符（如缺少 } 的截断 JSON），应正常闭合而非转义
					inString = false
					result.WriteRune(char)
					continue
				}

				nextChar := jsonStr[nextIdx]
				if nextChar == ',' || nextChar == '}' || nextChar == ']' || nextChar == ':' {
					inString = false
					result.WriteRune(char)
					continue
				}

				result.WriteRune('\\')
				result.WriteRune(char)
				continue
			} else {
				inString = true
				result.WriteRune(char)
				continue
			}
		}

		result.WriteRune(char)
	}

	return result.String()
}
