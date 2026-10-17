package helpers

import (
	"path/filepath"
	"runtime"
	"testing"

	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

// CopyModule は modules/<name> を一時フォルダにコピーし、そのパスを返す
// （各テストが別々の .terraform / terraform.tfstate を使うようにして、テスト同士の干渉を防ぐ）
func CopyModule(t *testing.T, name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	modulesDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "modules")

	return test_structure.CopyTerraformFolderToTemp(t, modulesDir, name)
}
