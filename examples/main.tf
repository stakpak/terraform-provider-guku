terraform {
  required_providers {
    guku = {
      source = "devopzilla/guku"
    }
  }
}

provider "guku" {
  username = var.username
  password = var.password
}

data "guku_platform" "secrets" {
  platform_id      = "platform-secrets"
  platform_version = "v1alpha2"
}

data "guku_platform" "security" {
  platform_id      = "platform-security"
  platform_version = "v1beta3-aws"
}

resource "guku_cluster" "demo" {
  name        = "demo"
  api_version = "1.21"
  server      = var.cluster_server
  context = jsonencode({
    AWS_REGION     = var.aws_region
    AWS_ACCOUNT_ID = var.aws_account
  })
  ca    = var.cluster_ca
  token = var.cluster_token
}

resource "guku_platform_binding" "demo_secrets" {
  cluster_id         = guku_cluster.demo.id
  platform_id        = data.guku_platform.secrets.platform_id
  platform_version   = data.guku_platform.secrets.platform_version
  platform_config_id = data.guku_platform.secrets.configs["default-aws"].id
}

resource "guku_platform_binding" "demo_security" {
  cluster_id         = guku_cluster.demo.id
  platform_id        = data.guku_platform.security.platform_id
  platform_version   = data.guku_platform.security.platform_version
  platform_config_id = data.guku_platform.security.configs["demo-2"].id

  depends_on = [
    guku_platform_binding.demo_secrets
  ]
}
