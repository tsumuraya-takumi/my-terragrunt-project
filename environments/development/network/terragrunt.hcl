locals {
  params = yamldecode(file("${get_terragrunt_dir()}/network.yaml"))
}

include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/network"
}

inputs = {
  environment  = include.root.locals.environment
  project_name = include.root.locals.project_name
  vpc_cidr     = local.params.vpc_cidr
  azs          = local.params.azs
}
