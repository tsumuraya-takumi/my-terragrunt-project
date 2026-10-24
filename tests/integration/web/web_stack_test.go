package web

import (
	"context" // 処理のキャンセルやタイムアウトを管理する標準パッケージ
	"fmt"     // 文字列の組み立て（fmt.Sprintf）やエラーの作成（fmt.Errorf）に使う標準パッケージ
	"testing" // Go のテスト用の標準パッケージ
	"time"    // 時間（待ち時間の 10*time.Second など）を扱う標準パッケージ

	// AWS SDK（AWS を操作するための公式ライブラリ）
	// Terratest にも aws という名前のパッケージがあって名前がぶつかるため、awssdk という別名を付けている
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"                                     // EC2（インスタンスなど）を操作する
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"                  // ELBv2（ALB・ターゲットグループ）を操作する
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types" // ELBv2 で使う型（ターゲットの状態など）

	"github.com/gruntwork-io/terratest/modules/aws" // Terratest: AWS 関連の便利な関数
	// Terratest: HTTP アクセスの便利な関数（パッケージ名にハイフンがあるため http_helper という別名を付けている）
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/retry"     // Terratest: 成功するまで処理を繰り返す（リトライ）
	"github.com/gruntwork-io/terratest/modules/terraform" // Terratest: terraform コマンドを Go から実行する

	"github.com/stretchr/testify/assert"  // testify: 確認して、失敗しても次の確認に進む
	"github.com/stretchr/testify/require" // testify: 確認して、失敗したらそのテストを中止する

	"github.com/tsumuraya-takumi/my-terragrunt-project/test/helpers" // 自分たちで作った共通処理
)

// ASG で起動するインスタンス数
// min / desired / max をすべてこの値にして台数を固定し、テストの途中で台数が変わらないようにする
const desiredCapacity = 2

// TestWebStack は network → alb → server を本物の AWS に作り、ALB 経由で Web サーバー（httpd）に届くかを確認する
// 関数名が Test で始まり、引数が (t *testing.T) の関数が「テスト」として go test に実行される
//
// 流れ:
//
//  1. network → alb → server の順に作る（前のモジュールの output を次のモジュールの変数に渡す）
//  2. 3項目を確認する
//  3. テスト関数が終わると、server → alb → network の順に destroy される
func TestWebStack(t *testing.T) {
	// .env から読み込んだリージョンを、短い名前の変数に入れておく
	region := cfg.AWSRegion

	// 各モジュールには provider ブロックがない（dev 環境では Terragrunt が自動で作っている）ため、
	// どのリージョンに作るかを環境変数で渡す（3つのモジュールで共通）
	envVars := map[string]string{
		"AWS_DEFAULT_REGION": region,
	}

	// ========================================================
	// 1. network → alb → server の順に作成する
	//
	//    defer は「この関数が終わるときに実行する」予約で、後から予約したものから順に実行される
	//    そのため、作成と逆の順番（server → alb → network）で削除される
	//    （ALB や EC2 が残ったままでは VPC を削除できないので、この順番が必要）
	// ========================================================

	// ------------------------
	// 1-1. network（VPC・サブネット・NAT）
	// ------------------------

	networkOptions := &terraform.Options{
		// modules/network を一時フォルダにコピーして使う（network/ のテストと terraform.tfstate が混ざらないようにするため）
		TerraformDir: helpers.CopyModule(t, "network"),
		// モジュールの変数（variables.tf）に渡す値
		Vars: map[string]interface{}{
			"environment":  environment,     // main_test.go で定義した "it-web"
			"project_name": cfg.ProjectName, // .env の PROJECT_NAME
			"vpc_cidr":     "10.0.0.0/16",
			"azs":          []string{"ap-northeast-1a", "ap-northeast-1c"},
		},
		EnvVars: envVars,
	}
	// 作成より前に削除を予約しておく（作成の途中で失敗しても、作りかけのリソースが削除される）
	defer terraform.Destroy(t, networkOptions)
	// terraform init → terraform apply を実行する
	terraform.InitAndApply(t, networkOptions)

	// alb と server に渡すため、network の output を受け取る
	vpcID := terraform.Output(t, networkOptions, "vpc_id")
	publicSubnetIDs := terraform.OutputMap(t, networkOptions, "public_subnet_ids")   // ALB を置くサブネット
	privateSubnetIDs := terraform.OutputMap(t, networkOptions, "private_subnet_ids") // EC2 を置くサブネット

	// ------------------------
	// 1-2. alb（ALB・ターゲットグループ・リスナー・ALB 用 SG）
	// ------------------------

	albOptions := &terraform.Options{
		TerraformDir: helpers.CopyModule(t, "alb"),
		// network の output を alb の変数に渡す
		// dev 環境では terragrunt.hcl の dependency "network" が担当している部分
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

	// 検証と server に渡すため、alb の output を受け取る
	albDNSName := terraform.Output(t, albOptions, "alb_dns_name")         // アクセス先（例: my-project-it-web-alb-xxx.ap-northeast-1.elb.amazonaws.com）
	albSGID := terraform.Output(t, albOptions, "alb_sg_id")               // App 用 SG で「ここからの通信だけ許可」に使う
	targetGroupARN := terraform.Output(t, albOptions, "target_group_arn") // ASG のインスタンスを登録する先

	// ------------------------
	// 1-3. server（App 用 SG・Launch Template・ASG）
	// ------------------------

	serverOptions := &terraform.Options{
		TerraformDir: helpers.CopyModule(t, "server"),
		// network と alb の output を server の変数に渡す
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

	// 検証に使うため、server の output を受け取る
	asgName := terraform.Output(t, serverOptions, "asg_name")  // インスタンスの一覧を取るのに使う
	appSGID := terraform.Output(t, serverOptions, "app_sg_id") // インスタンスに付いている SG と比べるのに使う

	// ========================================================
	// 2. 検証
	//    t.Run("名前", func(t *testing.T) { ... }) は、テストの中の小テスト（サブテスト）
	//    1つが失敗しても残りは実行され、結果にはサブテストごとの成功・失敗が表示される
	// ========================================================

	// ------------------------
	// 確認1: ALB の DNS 名にアクセスすると HTTP 200 が返ること
	// ------------------------
	// apply が終わった直後は、インスタンスの起動・httpd のインストール・ヘルスチェックの通過が済んでいない
	// そのため、すぐには 200 が返らないので、10秒おきに最大60回（最大10分）アクセスし直す
	t.Run("alb_returns_http_200", func(t *testing.T) {
		// fmt.Sprintf で文字列を組み立てる（%s の部分に albDNSName が入る）
		url := fmt.Sprintf("http://%s", albDNSName)

		// HttpGetWithRetryWithCustomValidation(t, URL, TLS設定, 最大回数, 待ち時間, 判定する関数)
		//   TLS設定 : HTTPS ではないので nil（なし）
		//   判定する関数 : レスポンスのステータスと本文を受け取り、true を返したら成功
		//   最大回数まで繰り返しても true にならなければ、テストは失敗する
		http_helper.HttpGetWithRetryWithCustomValidation(t, url, nil, 60, 10*time.Second,
			// 名前のない関数（無名関数）をその場で定義して渡している
			func(status int, body string) bool {
				return status == 200
			},
		)
	})

	// ------------------------
	// 確認2: ターゲットグループのターゲットが desiredCapacity 台あり、すべて healthy であること
	// ------------------------
	// ALB はヘルスチェックに合格したインスタンス（healthy）にだけ通信を送る
	// 今回の障害（User Data の不備・SG の付け忘れ）は、ここが unhealthy になることで見つけられる
	t.Run("all_targets_are_healthy", func(t *testing.T) {
		// ELBv2（ALB・ターゲットグループ）に問い合わせるための窓口を作る（helpers/aws.go）
		elbClient := helpers.NewElbV2Client(t, region)

		// retry.DoWithRetry(t, 説明, 最大回数, 待ち時間, 処理)
		//   処理がエラーを返したら、待ち時間をおいてやり直す（10秒おきに最大30回 = 最大5分）
		//   エラーなし（nil）を返したら成功して終了する
		//   最大回数まで繰り返してもエラーのままなら、テストは失敗する
		retry.DoWithRetry(t, "全ターゲットが healthy になるのを待つ", 30, 10*time.Second, func() (string, error) {
			// ターゲットグループに登録されているターゲットの状態を問い合わせる
			// AWS CLI の aws elbv2 describe-target-health --target-group-arn <ARN> と同じ処理
			output, err := elbClient.DescribeTargetHealth(context.Background(), &elasticloadbalancingv2.DescribeTargetHealthInput{
				TargetGroupArn: awssdk.String(targetGroupARN),
			})
			if err != nil {
				// 問い合わせ自体に失敗したら、エラーを返してリトライさせる
				return "", err
			}

			// ターゲットの数が desiredCapacity 台でなければ、まだ揃っていないのでリトライさせる
			if len(output.TargetHealthDescriptions) != desiredCapacity {
				return "", fmt.Errorf("ターゲット数が %d 台ではありません（現在 %d 台）",
					desiredCapacity, len(output.TargetHealthDescriptions))
			}

			// ターゲットを1台ずつ取り出し、healthy でないものがあればリトライさせる
			// （起動直後は initial、ヘルスチェックに失敗していると unhealthy になる）
			for _, target := range output.TargetHealthDescriptions {
				if target.TargetHealth.State != elbv2types.TargetHealthStateEnumHealthy {
					return "", fmt.Errorf("ターゲット %s が healthy ではありません（現在 %s）",
						awssdk.ToString(target.Target.Id), target.TargetHealth.State)
				}
			}

			// ここまで来たら全台 healthy。エラーなし（nil）を返して成功させる
			return "全ターゲットが healthy です", nil
		})
	})

	// ------------------------
	// 確認3: ASG のインスタンスがパブリックIPを持たず、App 用の SG だけが付いていること
	// ------------------------
	// インスタンスはインターネットから直接アクセスされない（ALB 経由だけ）状態であるべき
	// また、App 用 SG ではなく VPC の default SG が付いていると ALB からの通信が届かない（今回の障害②）
	t.Run("instances_are_private_and_use_app_sg", func(t *testing.T) {
		// ASG が起動したインスタンスの ID 一覧を取得する
		instanceIDs := aws.GetInstanceIdsForAsg(t, asgName, region)
		// 台数が desiredCapacity 台であることを確認する（違えば以降の確認は意味がないので中止）
		require.Len(t, instanceIDs, desiredCapacity, "ASG のインスタンス数が %d 台ではありません", desiredCapacity)

		// EC2 に問い合わせるための窓口を作り、インスタンスの詳細情報を取得する
		// AWS CLI の aws ec2 describe-instances --instance-ids <ID...> と同じ処理
		ec2Client := aws.NewEc2Client(t, region)
		output, err := ec2Client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: instanceIDs,
		})
		require.NoError(t, err)

		// 結果は「Reservation（起動リクエストの単位）」の中に「Instance」が入る2段構造なので、for を2重にして取り出す
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				instanceID := awssdk.ToString(instance.InstanceId)

				// パブリックIPが付いていない（nil = 値がない）ことを確認する
				assert.Nil(t, instance.PublicIpAddress,
					"インスタンス %s にパブリックIPが付与されています", instanceID)

				// インスタンスに付いている SG の ID を集めてリストにする
				// var sgIDs []string は空の文字列リストを用意する書き方、append でリストの末尾に追加する
				var sgIDs []string
				for _, sg := range instance.SecurityGroups {
					sgIDs = append(sgIDs, awssdk.ToString(sg.GroupId))
				}

				// 付いている SG が「App 用 SG の1つだけ」であることを確認する
				// （default SG が付いていたり、他の SG が混ざっていたりすると失敗する）
				assert.Equal(t, []string{appSGID}, sgIDs,
					"インスタンス %s に App 用の SG だけが付いていません", instanceID)
			}
		}
	})

	// ========================================================
	// 3. ここで関数が終わり、予約した destroy が server → alb → network の順に実行される
	// ========================================================
}
