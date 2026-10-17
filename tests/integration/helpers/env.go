// Package helpers は結合テストの各シナリオで共通して使う処理をまとめる
package helpers

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// Config は .env（または環境変数）から読み込んだ設定値
type Config struct {
	AWSRegion   string
	ProjectName string
}

// LoadConfig は tests/integration/.env を読み込み、設定値を返す
func LoadConfig() Config {
	// go test は各フォルダに移動して実行されるため、
	// このファイル（helpers/env.go）の場所から見て1つ上の .env を指定する
	_, thisFile, _, _ := runtime.Caller(0)
	envPath := filepath.Join(filepath.Dir(thisFile), "..", ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Printf(".env が読み込めませんでした（環境変数をそのまま使います）: %v", err)
	}

	return Config{
		AWSRegion:   getEnv("AWS_REGION", "ap-northeast-1"),
		ProjectName: getEnv("PROJECT_NAME", "my-project"),
	}
}

// getEnv は環境変数を取得する。未設定なら初期値を返す
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
