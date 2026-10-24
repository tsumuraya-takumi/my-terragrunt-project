package helpers

import (
	"context" // 処理のキャンセルやタイムアウトを管理する標準パッケージ
	"testing" // Go のテスト用の標準パッケージ

	"github.com/aws/aws-sdk-go-v2/config"                         // AWS SDK: 認証情報やリージョンの設定を読み込む
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2" // AWS SDK: ALB・ターゲットグループを操作する
	"github.com/stretchr/testify/require"                         // testify: 条件を確認し、失敗したらテストを中止する
)

// NewElbV2Client は ALB / ターゲットグループを操作するための「AWS への窓口（クライアント）」を作って返す
// EC2 用の窓口は Terratest に aws.NewEc2Client が用意されているが、
// ELBv2（ALB・ターゲットグループ）用は用意されていないため、AWS SDK を使って自分で作る
//
//	t      : テストの状態。エラー時にテストを失敗させるために受け取る
//	region : 接続する AWS のリージョン
//	戻り値 : ELBv2 の窓口（* はポインタ。窓口の実体の場所を返している）
func NewElbV2Client(t *testing.T, region string) *elasticloadbalancingv2.Client {
	// AWS に接続するための設定を読み込む
	// 認証情報は AWS CLI と同じ場所（環境変数 → ~/.aws/credentials → ~/.aws/config の順）から探される
	// context.Background() は「特にキャンセルなどの制御はしない」ことを表す決まり文句
	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))

	// 読み込みに失敗していたら（認証情報が見つからないなど）、ここでテストを中止する
	// Go には try/catch がないため、エラーは戻り値で受け取ってすぐ確認するのが基本
	require.NoError(t, err, "AWS の設定を読み込めませんでした")

	// 読み込んだ設定を使って ELBv2 の窓口を作り、呼び出し元に返す
	return elasticloadbalancingv2.NewFromConfig(awsCfg)
}
