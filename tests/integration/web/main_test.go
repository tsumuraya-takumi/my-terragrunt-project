// web パッケージ: network → alb → server を実際に作り、ALB 経由で Web サーバーに届くかを確認する結合テスト
// Go ではフォルダごとに別のパッケージになる（このフォルダのファイルはすべて package web と書く）
package web

import (
	"os"      // プログラムの終了（os.Exit）などに使う標準パッケージ
	"testing" // Go のテスト用の標準パッケージ

	// 自分たちで作った共通処理（tests/integration/helpers/）
	"github.com/tsumuraya-takumi/my-terragrunt-project/test/helpers"
)

// テストで作るリソース名に使う environment の値
// 名前は「my-project-it-web-alb」のようになり、dev 環境（development）や network/ のテスト（it-network）と重ならない
// const は「後から変更できない値（定数）」を定義する書き方
const environment = "it-web"

// .env から読み込んだ設定値（リージョン・プロジェクト名）
// TestMain で値を入れ、このフォルダのすべてのテストから cfg.AWSRegion のように使う
var cfg helpers.Config

// TestMain は、このフォルダ（パッケージ）のテストを実行する前に1回だけ呼ばれる特別な関数
// 関数名は必ず TestMain、引数は *testing.M にする決まり
func TestMain(m *testing.M) {
	// 前処理: .env を読み込んで設定値を用意する
	cfg = helpers.LoadConfig()

	// m.Run() でこのフォルダのテスト（TestWebStack など）を実行し、
	// その結果（成功なら 0、失敗なら 1）でプログラムを終了する
	os.Exit(m.Run())
}
