package web

import (
	"os"
	"testing"

	"github.com/tsumuraya-takumi/my-terragrunt-project/test/helpers"
)

// 他のシナリオとリソース名が重ならないよう、フォルダごとに environment を分ける
const environment = "it-web"

var cfg helpers.Config

// このフォルダの全テストの前に1回だけ実行される
func TestMain(m *testing.M) {
	cfg = helpers.LoadConfig()
	os.Exit(m.Run())
}
