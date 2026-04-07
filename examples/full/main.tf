terraform {
  required_providers {
    datamammoth = {
      source  = "datamammoth/datamammoth"
      version = "~> 0.1"
    }
  }
}

provider "datamammoth" {
  api_key = var.dm_api_key
}

variable "dm_api_key" {
  type      = string
  sensitive = true
}

variable "environment" {
  type    = string
  default = "production"
}

# --- Data Sources ---

data "datamammoth_zones" "all" {}

data "datamammoth_products" "vps" {
  category = "vps"
}

# --- Server ---

resource "datamammoth_server" "app" {
  hostname   = "app-${var.environment}-01"
  product_id = "prod_vps_large"
  image_id   = "ubuntu-22.04"
  zone_id    = data.datamammoth_zones.all.zones[0].id
}

resource "datamammoth_server" "db" {
  hostname   = "db-${var.environment}-01"
  product_id = "prod_vps_xlarge"
  image_id   = "ubuntu-22.04"
  zone_id    = data.datamammoth_zones.all.zones[0].id
}

# --- Snapshots ---

resource "datamammoth_snapshot" "app_baseline" {
  server_id = datamammoth_server.app.id
  name      = "app-baseline-${formatdate("YYYY-MM-DD", timestamp())}"
}

resource "datamammoth_snapshot" "db_baseline" {
  server_id = datamammoth_server.db.id
  name      = "db-baseline-${formatdate("YYYY-MM-DD", timestamp())}"
}

# --- Webhooks ---

resource "datamammoth_webhook" "slack_notifications" {
  url    = "https://hooks.slack.com/services/T00/B00/xxx"
  events = ["server.created", "server.deleted", "server.error", "snapshot.completed"]
  active = true
}

resource "datamammoth_webhook" "monitoring" {
  url    = "https://monitoring.example.com/webhooks/datamammoth"
  events = ["server.error", "server.unreachable"]
  secret = var.webhook_secret
  active = true
}

variable "webhook_secret" {
  type      = string
  sensitive = true
  default   = ""
}

# --- Outputs ---

output "app_server_ip" {
  value = datamammoth_server.app.ip_address
}

output "db_server_ip" {
  value = datamammoth_server.db.ip_address
}

output "available_zones" {
  value = [for z in data.datamammoth_zones.all.zones : z.name]
}

output "vps_products" {
  value = [for p in data.datamammoth_products.vps.products : {
    id    = p.id
    name  = p.name
    price = p.price_monthly
  }]
}
