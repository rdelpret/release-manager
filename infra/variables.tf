variable "cloudflare_api_token" { sensitive = true }
variable "zone_id" { description = "Cloudflare zone ID for rdelpret.com" }
variable "account_id" { description = "Cloudflare account ID" }
variable "account_subdomain" { description = "Cloudflare workers subdomain (e.g. robbie-delprete)" }
