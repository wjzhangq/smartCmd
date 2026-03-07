package config

import "os"

// 编译时注入的默认值
var (
	DefaultBaseURL = ""
	DefaultAPIKey  = ""
	DefaultModel   = ""
)

// Config 存储 LLM 配置
type Config struct {
	BaseURL     string
	APIKey      string
	Model       string
	ExtraPrompt string
}

// Load 加载配置，环境变量优先于编译时默认值
func Load() *Config {
	cfg := &Config{
		BaseURL:     getEnvOrDefault("LLM_BASE_URL", DefaultBaseURL),
		APIKey:      getEnvOrDefault("LLM_API_KEY", DefaultAPIKey),
		Model:       getEnvOrDefault("LLM_MODEL", DefaultModel),
		ExtraPrompt: os.Getenv("LLM_PROMPT"),
	}
	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsValid 检查配置是否有效
func (c *Config) IsValid() bool {
	return c.BaseURL != "" && c.APIKey != "" && c.Model != ""
}
