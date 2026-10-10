include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/security"
}
inputs = {
  environment  = include.root.locals.environment
  project_name = include.root.locals.project_name
}
