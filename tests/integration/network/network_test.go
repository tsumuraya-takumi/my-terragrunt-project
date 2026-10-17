package network

import (
	"context"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsumuraya-takumi/my-terragrunt-project/test/helpers"
)

func TestNetworkModule(t *testing.T) {
	terraformOptions := &terraform.Options{
		TerraformDir: helpers.CopyModule(t, "network"),
		Vars: map[string]interface{}{
			"environment":  environment,
			"project_name": cfg.ProjectName,
			"vpc_cidr":     "10.0.0.0/16",
			"azs":          []string{"ap-northeast-1a", "ap-northeast-1c"},
		},
		// network モジュールには provider ブロックがないため、リージョンは環境変数で渡す
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": cfg.AWSRegion,
		},
	}

	// テストが途中で失敗しても必ず削除する（apply より前に defer で登録する）
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	publicSubnetIDs := terraform.OutputMap(t, terraformOptions, "public_subnet_ids")
	privateSubnetIDs := terraform.OutputMap(t, terraformOptions, "private_subnet_ids")

	// Public Subnet が AWS 上で実際に「public」（IGW へのルートを持つ）と判定されること
	t.Run("public_subnets_are_public", func(t *testing.T) {
		require.Len(t, publicSubnetIDs, 2, "Public Subnet が2つ作成されていません")

		for az, subnetID := range publicSubnetIDs {
			assert.True(t, aws.IsPublicSubnet(t, subnetID, cfg.AWSRegion),
				"Public Subnet(%s) がパブリックになっていません", az)
		}
	})

	// Private Subnet が「public」と判定されないこと
	t.Run("private_subnets_are_not_public", func(t *testing.T) {
		require.Len(t, privateSubnetIDs, 2, "Private Subnet が2つ作成されていません")

		for az, subnetID := range privateSubnetIDs {
			assert.False(t, aws.IsPublicSubnet(t, subnetID, cfg.AWSRegion),
				"Private Subnet(%s) がパブリックになっています", az)
		}
	})

	// NAT Gateway が AZ ごとに作成され、使える状態（available）になっていること
	t.Run("nat_gateways_are_available", func(t *testing.T) {
		ec2Client := aws.NewEc2Client(t, cfg.AWSRegion)

		output, err := ec2Client.DescribeNatGateways(context.Background(), &ec2.DescribeNatGatewaysInput{
			Filter: []ec2types.Filter{
				{Name: awssdk.String("vpc-id"), Values: []string{vpcID}},
			},
		})
		require.NoError(t, err)
		require.Len(t, output.NatGateways, 2, "NAT Gateway が AZ ごとに作成されていません")

		for _, nat := range output.NatGateways {
			assert.Equal(t, ec2types.NatGatewayStateAvailable, nat.State,
				"NAT Gateway(%s) が available になっていません", awssdk.ToString(nat.NatGatewayId))
		}
	})
}
