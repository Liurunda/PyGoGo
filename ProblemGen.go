package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"net/http"
	"strings"
)

// SongInfo 对应输入歌曲信息的结构
type SongInfo struct {
	Name   string   `json:"name"`
	Singer string   `json:"singer"`
	Lyric  []string `json:"lyric"`
}

// QuizData 对应 Gemini 返回的结构化题目
type QuizData struct {
	SongName    string            `json:"song_name"`
	QuizContent string            `json:"quiz_content"`
	Options     map[string]string `json:"options"`
	Answer      string            `json:"answer"`
	Explanation string            `json:"explanation"`
}

// Gemini API 响应的嵌套结构定义
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}


var (
	API_KEY = os.Getenv("GEMINI_API_KEY")
	API_URL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + API_KEY
)

// GenerateStructuredQuiz 请求 Gemini 并强制返回 JSON 格式
func GenerateStructuredQuiz(songInfo SongInfo) (*QuizData, error) {
	fullLyric := strings.Join(songInfo.Lyric, "\n")

	// 构建 Prompt
	promptText := fmt.Sprintf(`
你是一个音乐题目生成器。请根据提供的歌词生成一道题目。

选出一段有代表性的歌词片段，挖掉一句关键歌词，用“【待填空】”代替。

必须返回一个 JSON 对象，结构如下：
{
  "song_name": "歌名",
  "quiz_content": "带有【待填空】标记的歌词片段，包含前后的几句歌词",
  "options": {
    "A": "选项内容",
    "B": "选项内容",
    "C": "选项内容",
    "D": "选项内容",
    "E": "选项内容"
  },
  "answer": "正确选项的字母",
  "explanation": "简短的题目解析, 分析比较每个选项"
}

【歌词信息】：
歌手：%s
歌名：%s
歌词全文：
%s
`, songInfo.Singer, songInfo.Name, fullLyric)

	// 构建请求 Payload
	payload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{"text": promptText},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":        0.7,
			"response_mime_type": "application/json", // 关键配置
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload failed: %v", err)
	}

	// 发送 POST 请求
	resp, err := http.Post(API_URL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析 Gemini API 的原始响应
	var geminiRes GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiRes); err != nil {
		return nil, fmt.Errorf("decode gemini response failed: %v", err)
	}

	if len(geminiRes.Candidates) == 0 || len(geminiRes.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	// 提取 text 字符串（它本身是一个 JSON 字符串）
	quizJsonStr := geminiRes.Candidates[0].Content.Parts[0].Text

	// 将字符串转换为 QuizData 结构体
	var quizData QuizData
	if err := json.Unmarshal([]byte(quizJsonStr), &quizData); err != nil {
		return nil, fmt.Errorf("unmarshal quiz data failed: %v", err)
	}

	return &quizData, nil
}
