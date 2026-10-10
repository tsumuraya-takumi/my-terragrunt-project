MODULES := network security server alb

# fmt-% などのパターンルールは .PHONY に含めると無視されるため、ここには入れない
.PHONY: fmt-hcl fmt-hcl-check fmt-all validate-all test-all check-all

# Terragruntコマンド
fmt-hcl:
	terragrunt hcl fmt

fmt-hcl-check:
	terragrunt hcl fmt --check

# Terraformコマンド（例: make fmt-network / make validate-server / make test-network）
fmt-%:
	cd modules/$* && terraform fmt -recursive

validate-%:
	cd modules/$* && terraform init -backend=false -input=false > /dev/null && terraform validate

# tests/ 配下に *.tftest.hcl があるモジュールのみ実行
test-%:
	@if ls modules/$*/tests/*.tftest.hcl > /dev/null 2>&1; then \
		cd modules/$* && terraform test -var-file=tests/common.tfvars; \
	else \
		echo "skip: modules/$* にテストがありません"; \
	fi

check-%: fmt-% validate-% test-%
	@:

fmt-all: $(addprefix fmt-,$(MODULES))
validate-all: $(addprefix validate-,$(MODULES))
test-all: $(addprefix test-,$(MODULES))
check-all: fmt-hcl-check $(addprefix check-,$(MODULES))
