package main

import (
	"os"
	"time"

	"ai-eino-interview-agent/internal/config"
)

func getTestConfig() *config.Config {
	return &config.Config{
		Embedding: config.EmbeddingConfig{
			APIKey:     "c1c8f7ce-266f-4af5-a832-9a8457f36e74",
			Model:      "doubao-embedding-text-240715",
			BaseURL:    "https://ark.cn-beijing.volces.com/api/v3/",
			Region:     getEnvOrDefault("EMBEDDING_REGION", "cn-beijing"),
			Timeout:    30 * time.Second,
			RetryTimes: 3,
			Dimensions: 2560,
		},
		DocumentSplitter: config.SplitterConfig{
			ChunkSize:   500,
			OverlapSize: 50,
			Separators:  []string{"\n\n", "\n", " "},
			KeepType:    0,
		},
		Milvus: config.MilvusConfig{
			Address:        getEnvOrDefault("MILVUS_ADDRESS", "localhost:19530"),
			CollectionName: "feishu_docs_20251126", // 全新 collection 名称，2560 维向量
			DatabaseName:   "default",              // 修改为 Milvus 默认数据库
			MetricType:     "COSINE",
			Username:       getEnvOrDefault("MILVUS_USERNAME", "minioadmin"),
			Password:       getEnvOrDefault("MILVUS_PASSWORD", "minioadmin"),
			TopK:           5,
			ConnectTimeout: 10 * time.Second,
			SearchTimeout:  30 * time.Second,
		},
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
