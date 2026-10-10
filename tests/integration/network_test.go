package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestNetworkModule(t *testing.T) {
	terraformOptions := &terraform.Options{
		TerraformDir: "../../modules/network",
		Vars: map[string]interface{}{
			"environment":  "test",
			"project_name": "my-project",
			"vpc_cidr":     "10.0.0.0/16",
			"azs":          []string{"ap-northeast-1a", "ap-northeast-1c"},
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	assert.NotEmpty(t, vpcID, "VPC IDが取得できませんでした")
}
