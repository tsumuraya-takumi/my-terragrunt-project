# modules/network/Makefile

fmt:
	terraform fmt -recursive

validate:
	terraform validate

test:
	terraform test -var-file=tests/common.tfvars

check: fmt validate test