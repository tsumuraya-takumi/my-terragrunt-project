locals {
  params = yamldecode(file("${get_terragrunt_dir()}/route53.yaml"))
}

include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/route53"
}

inputs = {
  environment  = include.root.locals.environment
  project_name = include.root.locals.project_name
  domain_name  = local.params.domain_name
}