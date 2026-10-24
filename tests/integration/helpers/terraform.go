package helpers

import (
	"path/filepath" // ファイルパスを組み立てる標準パッケージ
	"runtime"       // 実行中のプログラムの情報（このファイルの場所など）を取る標準パッケージ
	"testing"       // Go のテスト用の標準パッケージ

	// Terratest の test-structure パッケージ
	// パッケージ名にハイフン（-）が入っていてそのままでは使えないため、test_structure という別名を付けている
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

// CopyModule は modules/<name> を一時フォルダにコピーし、コピー先のパスを返す
//
// なぜコピーするのか:
//
//	terraform を実行すると、実行したフォルダに terraform.tfstate（何を作ったかの記録）ができる
//	modules/ のコードには backend の設定がない（dev 環境では Terragrunt が自動で作っている）ため、
//	そのまま実行すると記録は modules/network/terraform.tfstate のように1か所に書かれる
//	network/ と web/ のテストが同じ場所で実行すると、お互いのリソースを自分のものと勘違いしてしまう
//	コピーしてテストごとに別の作業場所にすれば、記録もテストごとに分かれて干渉しない
//
//	name   : コピーしたいモジュールの名前（"network"・"alb"・"server"）
//	戻り値 : コピー先のモジュールのパス（例: /var/folders/.../modules/network）
func CopyModule(t *testing.T, name string) string {
	// このファイル（helpers/terraform.go）自身のパスを取得する（使わない戻り値は _ で捨てる）
	_, thisFile, _, _ := runtime.Caller(0)

	// このファイルの場所から modules/ フォルダまでたどる（.. は1つ上のフォルダ）
	//   helpers/ → ..（integration/）→ ..（tests/）→ ..（プロジェクトのルート）→ modules/
	// テストのフォルダではなくこのファイルを基準にするので、どこから呼ばれても同じ modules/ を指せる
	modulesDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "modules")

	// modules/ 全体を一時フォルダにコピーし、その中の <name> のパスを返す
	// .terraform/ や terraform.tfstate、*.tfvars はコピーされない（前回の実行の残りを持ち込まないため）
	return test_structure.CopyTerraformFolderToTemp(t, modulesDir, name)
}
