# my-terragrunt-project/Makefile

test-network:
	cd modules/network && terraform test -var-file=tests/common.tfvars

test-security:
	cd modules/security && terraform test -var-file=tests/common.tfvars

test-server:
	cd modules/server && terraform test -var-file=tests/common.tfvars

test-all: test-network test-security test-server