
locals {
  aws_region   = "ap-northeast-1"
  project_name = "my-terragrunt-project"
}

remote_state {
  backend = "s3"

  config = {
    bucket       = "terraform-state-tsumuraya"
    key          = "${path_relative_to_include()}/terraform.tfstate"
    region       = "ap-northeast-1"
    encrypt      = true
    kms_key_id   = "arn:aws:kms:ap-northeast-1:381492180439:key/1e1dc460-fc41-4fb0-9ff0-6c9fb0416dc5"
    use_lockfile = true
  }

  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
}

generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"

  contents = <<EOF
provider "aws" {
  region = "${local.aws_region}"

  default_tags {
    tags = {
      Project = "${local.project_name}"
    }
  }
}
EOF
}

generate "versions" {
  path      = "versions.tf"
  if_exists = "overwrite_terragrunt"

  contents = <<EOF
terraform {
  required_version = "~> 1.15.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}
EOF
}