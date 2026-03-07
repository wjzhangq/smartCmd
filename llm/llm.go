package llm

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smartCmd/config"
	"strings"
	"text/template"
	"time"
)

//go:embed prompt/system.txt
var systemPromptTemplate string

//go:embed prompt/alternative.txt
var alternativePromptTemplate string

//go:embed prompt/analyze.txt
var analyzePromptTemplate string

// Message 表示聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 表示聊天请求
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse 表示聊天响应
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Client LLM 客户端
type Client struct {
	config *config.Config
	client *http.Client
}

// NewClient 创建新的 LLM 客户端
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat 发送聊天请求
func (c *Client) Chat(messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.config.Model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API 错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API 返回空响应")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// TestConnection 测试 LLM 连接
func (c *Client) TestConnection() (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: "你是什么模型？请用一句话回答。",
		},
	}

	return c.Chat(messages)
}

// CommandResult 命令解析结果
type CommandResult struct {
	Intent  string `json:"intent"`
	Command string `json:"command"`
	Exists  bool   `json:"exists"`
	Note    string `json:"note,omitempty"`
}

// ParseCommand 解析用户命令
func (c *Client) ParseCommand(userInput, systemInfo, shellType string) (*CommandResult, error) {
	// 使用模板渲染系统提示
	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析系统提示模板失败: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]string{
		"Shell":       shellType,
		"SystemInfo":  systemInfo,
		"ExtraPrompt": c.config.ExtraPrompt,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("渲染系统提示失败: %w", err)
	}

	systemPrompt := buf.String()

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userInput},
	}

	response, err := c.Chat(messages)
	if err != nil {
		return nil, err
	}

	// 清理响应，移除可能的代码块标记
	response = cleanJSONResponse(response)

	var result CommandResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析命令结果失败: %w\n原始响应: %s", err, response)
	}

	return &result, nil
}

// RequestAlternative 请求替代命令
func (c *Client) RequestAlternative(originalCmd, errorMsg, systemInfo, shellType string, attempt int) (*CommandResult, error) {
	// 使用模板渲染替代命令提示
	tmpl, err := template.New("alternative").Parse(alternativePromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析替代命令模板失败: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"Shell":      shellType,
		"SystemInfo": systemInfo,
		"Command":    originalCmd,
		"Error":      errorMsg,
		"Attempt":    attempt,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("渲染替代命令提示失败: %w", err)
	}

	systemPrompt := buf.String()

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "请提供替代命令"},
	}

	response, err := c.Chat(messages)
	if err != nil {
		return nil, err
	}

	response = cleanJSONResponse(response)

	var result CommandResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析命令结果失败: %w", err)
	}

	return &result, nil
}

// AnalyzeOutput 分析命令输出
func (c *Client) AnalyzeOutput(command, stdout, stderr string) (string, error) {
	// 使用模板渲染分析提示
	tmpl, err := template.New("analyze").Parse(analyzePromptTemplate)
	if err != nil {
		return "", fmt.Errorf("解析分析模板失败: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]string{
		"Command": command,
		"Stdout":  truncateOutput(stdout, 4000),
		"Stderr":  truncateOutput(stderr, 1000),
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染分析提示失败: %w", err)
	}

	prompt := buf.String()

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	return c.Chat(messages)
}

// cleanJSONResponse 清理 JSON 响应
func cleanJSONResponse(response string) string {
	// 移除可能的 markdown 代码块标记
	result := bytes.TrimPrefix([]byte(response), []byte("```json\n"))
	result = bytes.TrimPrefix(result, []byte("```\n"))
	result = bytes.TrimSuffix(result, []byte("\n```"))
	result = bytes.TrimSpace(result)
	return string(result)
}

// truncateOutput 截断输出
func truncateOutput(output string, maxLen int) string {
	output = strings.TrimSpace(output)
	if len(output) <= maxLen {
		return output
	}
	return output[:maxLen] + "\n... (输出已截断)"
}
