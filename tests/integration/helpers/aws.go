package helpers

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/stretchr/testify/require"
)

// NewElbV2Client は ALB / ターゲットグループ操作用のクライアントを返す
// （Terratest には ELBv2 用のヘルパーがないため、AWS SDK を直接使う）
func NewElbV2Client(t *testing.T, region string) *elasticloadbalancingv2.Client {
	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	require.NoError(t, err, "AWS の設定を読み込めませんでした")

	return elasticloadbalancingv2.NewFromConfig(awsCfg)
}
