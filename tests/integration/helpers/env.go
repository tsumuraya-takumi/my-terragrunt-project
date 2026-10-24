// Package helpers は結合テストの各シナリオ（network/・web/ など）で共通して使う処理をまとめる
// ここにはテストそのものは書かず、テストから呼び出される「道具」だけを置く
package helpers

import (
	"log"           // ログ（実行中のメッセージ）を表示する標準パッケージ
	"os"            // 環境変数の読み取りなど、OS の機能を使う標準パッケージ
	"path/filepath" // ファイルパスを組み立てる標準パッケージ
	"runtime"       // 実行中のプログラムの情報（このファイルの場所など）を取る標準パッケージ

	"github.com/joho/godotenv" // .env ファイルを読み込んで環境変数に設定する外部ライブラリ
)

// Config は .env（または環境変数）から読み込んだ設定値を入れておく「箱」
// type 〜 struct は「構造体」という、複数の値をまとめて持つ型を定義する書き方
// 名前が大文字で始まるものは、他のパッケージ（network/・web/）からも使える
type Config struct {
	AWSRegion   string // テストでリソースを作る AWS のリージョン（例: ap-northeast-1）
	ProjectName string // リソース名の先頭に付けるプロジェクト名（例: my-project）
}

// LoadConfig は tests/integration/.env を読み込み、設定値を Config に入れて返す
// 各フォルダの TestMain（テストの最初に1回だけ呼ばれる関数）から呼び出される
func LoadConfig() Config {
	// runtime.Caller(0) は「今実行しているこのファイル」のパスを返す
	// 戻り値は4つあるが、使うのは2番目（ファイルパス）だけなので、残りは _ で捨てる
	_, thisFile, _, _ := runtime.Caller(0)

	// go test は各フォルダ（network/ など）に移動してから実行されるため、
	// 単に ".env" と書くと network/.env を探してしまう
	// そこで、このファイル（helpers/env.go）の場所を基準に、1つ上（..）の .env を指定する
	//   helpers/env.go → helpers/ → ..（tests/integration/）→ .env
	envPath := filepath.Join(filepath.Dir(thisFile), "..", ".env")

	// .env を読み込み、中身を環境変数として設定する
	// ファイルがなくてもテストは止めず、メッセージだけ表示する（環境変数を直接設定している場合もあるため）
	if err := godotenv.Load(envPath); err != nil {
		log.Printf(".env が読み込めませんでした（環境変数をそのまま使います）: %v", err)
	}

	// 環境変数から値を取り出して Config に詰めて返す
	// 環境変数が設定されていなければ、第2引数の初期値が使われる
	return Config{
		AWSRegion:   getEnv("AWS_REGION", "ap-northeast-1"),
		ProjectName: getEnv("PROJECT_NAME", "my-project"),
	}
}

// getEnv は環境変数 key の値を返す。未設定（空）なら defaultValue を返す
// 名前が小文字で始まるので、helpers パッケージの中でしか使えない
func getEnv(key, defaultValue string) string {
	// os.Getenv で環境変数を読む。値が空でなければそれを返す
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
