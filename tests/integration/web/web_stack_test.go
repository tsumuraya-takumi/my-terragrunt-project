package web

import (
	"context"
	"fmt"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/gruntwork-io/terratest/modules/aws"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsumuraya-takumi/my-terragrunt-project/test/helpers"
)

// ASG で起動するインスタンス数（min / desired / max を同じにして台数を固定する）
const desiredCapacity = 2

// ALB経由でWebサーバー(httpd)に到達できること
func TestWebStack(t *testing.T) {
	region := cfg.AWSRegion

	// 各モジュールには provider ブロックがないため、リージョンは環境変数で渡す
	envVars := map[string]string{
		"AWS_DEFAULT_REGION": region,
	}

	// ------------------------
	// 1. network → alb → server の順に作成する
	//    削除は defer の逆順（server → alb → network）で行われる
	// ------------------------

	networkOptions := &terraform.Options{
		TerraformDir: helpers.CopyModule(t, "network"),
		Vars: map[string]interface{}{
			"environment":  environment,
			"project_name": cfg.ProjectName,
			"vpc_cidr":     "10.0.0.0/16",
			"azs":          []string{"ap-northeast-1a", "ap-northeast-1c"},
		},
		EnvVars: envVars,
	}
	defer terraform.Destroy(t, networkOptions)
	terraform.InitAndApply(t, networkOptions)

	vpcID := terraform.Output(t, networkOptions, "vpc_id")
	publicSubnetIDs := terraform.OutputMap(t, networkOptions, "public_subnet_ids")
	privateSubnetIDs := terraform.OutputMap(t, networkOptions, "private_subnet_ids")

	albOptions := &terraform.Options{
		TerraformDir: helpers.CopyModule(t, "alb"),
		Vars: map[string]interface{}{
			"environment":       environment,
			"project_name":      cfg.ProjectName,
			"vpc_id":            vpcID,
			"public_subnet_ids": publicSubnetIDs,
		},
		EnvVars: envVars,
	}
	defer terraform.Destroy(t, albOptions)
	terraform.InitAndApply(t, albOptions)

	albDNSName := terraform.Output(t, albOptions, "alb_dns_name")
	albSGID := terraform.Output(t, albOptions, "alb_sg_id")
	targetGroupARN := terraform.Output(t, albOptions, "target_group_arn")

	serverOptions := &terraform.Options{
		TerraformDir: helpers.CopyModule(t, "server"),
		Vars: map[string]interface{}{
			"environment":        environment,
			"project_name":       cfg.ProjectName,
			"vpc_id":             vpcID,
			"private_subnet_ids": privateSubnetIDs,
			"alb_sg_id":          albSGID,
			"target_group_arn":   targetGroupARN,
			"instance_type":      "t3.micro",
			"min_size":           desiredCapacity,
			"desired_capacity":   desiredCapacity,
			"max_size":           desiredCapacity,
		},
		EnvVars: envVars,
	}
	defer terraform.Destroy(t, serverOptions)
	terraform.InitAndApply(t, serverOptions)

	asgName := terraform.Output(t, serverOptions, "asg_name")
	appSGID := terraform.Output(t, serverOptions, "app_sg_id")

	// ------------------------
	// 2. 検証
	// ------------------------

	// ALB の DNS 名にアクセスすると HTTP 200 が返ること
	// （インスタンスの起動・httpd のインストール・ヘルスチェック通過を待つため、最大10分リトライする）
	t.Run("alb_returns_http_200", func(t *testing.T) {
		url := fmt.Sprintf("http://%s", albDNSName)

		http_helper.HttpGetWithRetryWithCustomValidation(t, url, nil, 60, 10*time.Second,
			func(status int, body string) bool {
				return status == 200
			},
		)
	})

	// ターゲットグループのターゲットが desired_capacity 台あり、すべて healthy であること
	t.Run("all_targets_are_healthy", func(t *testing.T) {
		elbClient := helpers.NewElbV2Client(t, region)

		retry.DoWithRetry(t, "全ターゲットが healthy になるのを待つ", 30, 10*time.Second, func() (string, error) {
			output, err := elbClient.DescribeTargetHealth(context.Background(), &elasticloadbalancingv2.DescribeTargetHealthInput{
				TargetGroupArn: awssdk.String(targetGroupARN),
			})
			if err != nil {
				return "", err
			}

			if len(output.TargetHealthDescriptions) != desiredCapacity {
				return "", fmt.Errorf("ターゲット数が %d 台ではありません（現在 %d 台）",
					desiredCapacity, len(output.TargetHealthDescriptions))
			}

			for _, target := range output.TargetHealthDescriptions {
				if target.TargetHealth.State != elbv2types.TargetHealthStateEnumHealthy {
					return "", fmt.Errorf("ターゲット %s が healthy ではありません（現在 %s）",
						awssdk.ToString(target.Target.Id), target.TargetHealth.State)
				}
			}

			return "全ターゲットが healthy です", nil
		})
	})

	// ASG のインスタンスがパブリックIPを持たず、App 用の SG だけが付いていること
	t.Run("instances_are_private_and_use_app_sg", func(t *testing.T) {
		instanceIDs := aws.GetInstanceIdsForAsg(t, asgName, region)
		require.Len(t, instanceIDs, desiredCapacity, "ASG のインスタンス数が %d 台ではありません", desiredCapacity)

		ec2Client := aws.NewEc2Client(t, region)
		output, err := ec2Client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: instanceIDs,
		})
		require.NoError(t, err)

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				instanceID := awssdk.ToString(instance.InstanceId)

				assert.Nil(t, instance.PublicIpAddress,
					"インスタンス %s にパブリックIPが付与されています", instanceID)

				var sgIDs []string
				for _, sg := range instance.SecurityGroups {
					sgIDs = append(sgIDs, awssdk.ToString(sg.GroupId))
				}
				assert.Equal(t, []string{appSGID}, sgIDs,
					"インスタンス %s に App 用の SG だけが付いていません", instanceID)
			}
		}
	})
}
