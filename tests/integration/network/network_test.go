package network

import (
	"context" // 処理のキャンセルやタイムアウトを管理する標準パッケージ
	"testing" // Go のテスト用の標準パッケージ

	// AWS SDK（AWS を操作するための公式ライブラリ）
	// Terratest にも aws という名前のパッケージがあって名前がぶつかるため、awssdk という別名を付けている
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"                // EC2（VPC・サブネット・NAT など）を操作する
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types" // EC2 で使う型（Filter・NAT の状態など）

	"github.com/gruntwork-io/terratest/modules/aws"       // Terratest: AWS 関連の便利な関数
	"github.com/gruntwork-io/terratest/modules/terraform" // Terratest: terraform コマンドを Go から実行する

	"github.com/stretchr/testify/assert"  // testify: 確認して、失敗しても次の確認に進む
	"github.com/stretchr/testify/require" // testify: 確認して、失敗したらそのテストを中止する

	"github.com/tsumuraya-takumi/my-terragrunt-project/test/helpers" // 自分たちで作った共通処理
)

// TestNetworkModule は network モジュールを本物の AWS に作り、使える状態になっているかを確認する
// 関数名が Test で始まり、引数が (t *testing.T) の関数が「テスト」として go test に実行される
//
// 流れ:
//
//	① terraform に渡す設定を用意する
//	② 後片付け（destroy）を予約する
//	③ terraform init / apply で本物の AWS にリソースを作る
//	④ 作ったリソースの ID を受け取る
//	⑤ AWS に問い合わせて3項目を確認する
//	⑥ テスト関数が終わると、②で予約した destroy が実行される
func TestNetworkModule(t *testing.T) {
	// ------------------------
	// ① terraform に渡す設定を用意する
	// ------------------------

	// & は「この設定の実体の場所（ポインタ）を渡す」という意味。Terratest の関数がこの形で受け取る
	terraformOptions := &terraform.Options{
		// どの Terraform コードを使うか
		// modules/network を一時フォルダにコピーして使う（テスト同士で terraform.tfstate が混ざらないようにするため）
		TerraformDir: helpers.CopyModule(t, "network"),

		// モジュールの変数（variables.tf）に渡す値
		// dev 環境では environments/development/network/terragrunt.hcl の inputs が担当している部分
		Vars: map[string]interface{}{
			"environment":  environment,     // main_test.go で定義した "it-network"
			"project_name": cfg.ProjectName, // .env の PROJECT_NAME
			"vpc_cidr":     "10.0.0.0/16",
			"azs":          []string{"ap-northeast-1a", "ap-northeast-1c"},
		},

		// terraform 実行時に設定する環境変数
		// network モジュールには provider ブロックがない（dev 環境では Terragrunt が自動で作っている）ため、
		// どのリージョンに作るかを環境変数で渡す
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": cfg.AWSRegion,
		},
	}

	// ------------------------
	// ② 後片付けを予約する
	// ------------------------

	// defer は「この関数（TestNetworkModule）が終わるときに必ず実行する」という予約
	// apply より前に書いておくことで、apply の途中で失敗しても作りかけのリソースが削除され、課金が続かない
	defer terraform.Destroy(t, terraformOptions)

	// ------------------------
	// ③ 本物の AWS にリソースを作る
	// ------------------------

	// terraform init → terraform apply を順に実行する（失敗したらテストは失敗し、②の destroy が走る）
	terraform.InitAndApply(t, terraformOptions)

	// ------------------------
	// ④ 作ったリソースの ID を受け取る
	// ------------------------

	// terraform output の値を取り出す（modules/network/outputs.tf で定義しているもの）
	// := は「変数を新しく作って値を入れる」書き方
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")                           // 例: "vpc-0abc..."
	publicSubnetIDs := terraform.OutputMap(t, terraformOptions, "public_subnet_ids")   // 例: {"ap-northeast-1a": "subnet-0aaa...", ...}
	privateSubnetIDs := terraform.OutputMap(t, terraformOptions, "private_subnet_ids") // 例: {"ap-northeast-1a": "subnet-0bbb...", ...}

	// ------------------------
	// ⑤ AWS に問い合わせて確認する
	// ------------------------
	// t.Run("名前", func(t *testing.T) { ... }) は、テストの中の小テスト（サブテスト）
	// 1つが失敗しても残りは実行され、結果にはサブテストごとの成功・失敗が表示される

	// 確認1: Public Subnet が AWS 上で実際に「public」（IGW へのルートを持つ）と判定されること
	t.Run("public_subnets_are_public", func(t *testing.T) {
		// まず Public Subnet が2つあることを確認する（なければ以降の確認は意味がないので require で中止）
		require.Len(t, publicSubnetIDs, 2, "Public Subnet が2つ作成されていません")

		// for 〜 range で map から1つずつ取り出す（az にキー＝AZ名、subnetID に値＝サブネットID が入る）
		for az, subnetID := range publicSubnetIDs {
			// aws.IsPublicSubnet は、AWS に問い合わせて
			// 「そのサブネットのルートテーブルにインターネットゲートウェイへの経路があるか」を調べ、true/false を返す
			// assert.True は結果が true であることを確認する（%s の部分には az の値が入る）
			assert.True(t, aws.IsPublicSubnet(t, subnetID, cfg.AWSRegion),
				"Public Subnet(%s) がパブリックになっていません", az)
		}
	})

	// 確認2: Private Subnet が「public」と判定されないこと（インターネットから直接見える状態になっていない）
	t.Run("private_subnets_are_not_public", func(t *testing.T) {
		require.Len(t, privateSubnetIDs, 2, "Private Subnet が2つ作成されていません")

		for az, subnetID := range privateSubnetIDs {
			// 確認1と同じ関数を使い、今度は false（public ではない）であることを確認する
			assert.False(t, aws.IsPublicSubnet(t, subnetID, cfg.AWSRegion),
				"Private Subnet(%s) がパブリックになっています", az)
		}
	})

	// 確認3: NAT Gateway が AZ ごとに作成され、使える状態（available）になっていること
	t.Run("nat_gateways_are_available", func(t *testing.T) {
		// EC2 サービスに問い合わせるための窓口（クライアント）を作る
		ec2Client := aws.NewEc2Client(t, cfg.AWSRegion)

		// 「このVPCにある NAT Gateway の一覧」を AWS に問い合わせる
		// AWS CLI の aws ec2 describe-nat-gateways --filter Name=vpc-id,Values=<VPC ID> と同じ処理
		output, err := ec2Client.DescribeNatGateways(context.Background(), &ec2.DescribeNatGatewaysInput{
			Filter: []ec2types.Filter{
				// awssdk.String は文字列をポインタに変換する（AWS SDK はポインタで受け取る項目が多い）
				{Name: awssdk.String("vpc-id"), Values: []string{vpcID}},
			},
		})

		// 問い合わせ自体が失敗していないかを確認する（失敗していたら中止）
		require.NoError(t, err)

		// NAT Gateway が2つ（AZ ごとに1つ）あることを確認する
		require.Len(t, output.NatGateways, 2, "NAT Gateway が AZ ごとに作成されていません")

		// 1つずつ取り出して、状態が available であることを確認する
		// （作成に失敗すると failed などになる。_ は使わない値（何番目か）を捨てている）
		for _, nat := range output.NatGateways {
			// assert.Equal(t, 期待する値, 実際の値, メッセージ)
			// awssdk.ToString はポインタの文字列を普通の文字列に戻す
			assert.Equal(t, ec2types.NatGatewayStateAvailable, nat.State,
				"NAT Gateway(%s) が available になっていません", awssdk.ToString(nat.NatGatewayId))
		}
	})

	// ------------------------
	// ⑥ ここで関数が終わり、②で予約した terraform destroy が実行される
	// ------------------------
}
