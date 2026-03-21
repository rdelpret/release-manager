terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.17"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

# DNS record for release.rdelpret.com → Worker
resource "cloudflare_dns_record" "release" {
  zone_id = var.zone_id
  name    = "release"
  type    = "CNAME"
  content = "music-release-planner.${var.account_subdomain}.workers.dev"
  proxied = true
  ttl     = 1
}
