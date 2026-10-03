# Terragruntコマンド
fmt-hcl:
	terragrunt hcl fmt

fmt-hcl-check:
	terragrunt hcl fmt --check

check-network:
	cd modules/network && make check

check-security:
	cd modules/security && make check

check-server:
	cd modules/server && make check

check-all: fmt-hcl-check check-network check-security check-server


# Terraformコマンド
fmt-network:
	cd modules/network && terraform fmt -recursive

fmt-security:
	cd modules/security && terraform fmt -recursive

fmt-server:
	cd modules/server && terraform fmt -recursive

fmt-all: fmt-network fmt-security fmt-server

validate-network:
	cd modules/network && terraform validate

validate-security:
	cd modules/security && terraform validate

validate-server:
	cd modules/server && terraform validate

validate-all: validate-network validate-security validate-server


# Terraform testコマンド
test-network:
	cd modules/network && terraform test -var-file=tests/common.tfvars

test-security:
	cd modules/security && terraform test -var-file=tests/common.tfvars

test-server:
	cd modules/server && terraform test -var-file=tests/common.tfvars

test-all: test-network test-security test-server