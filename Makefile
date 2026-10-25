MODULES := network security server alb route53

# fmt-% などのパターンルールは .PHONY に含めると無視されるため、ここには入れない
.PHONY: fmt-hcl fmt-hcl-check fmt-all validate-all test-all check-all it-network it-web it-all

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

# 結合テスト（実際に AWS にリソースを作成するため課金が発生する）
it-network:
	cd tests/integration && go test -v -timeout 30m ./network/...

it-web:
	cd tests/integration && go test -v -timeout 30m ./web/...

# シナリオ同士を並列に実行しない（NAT・EIP が同時に倍作られるのを防ぐ）
it-all:
	cd tests/integration && go test -v -p 1 -timeout 60m ./...
