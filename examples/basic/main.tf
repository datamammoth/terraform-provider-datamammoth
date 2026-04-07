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

data "datamammoth_zones" "all" {}

resource "datamammoth_server" "web" {
  hostname   = "web-01"
  product_id = "prod_vps_medium"
  image_id   = "ubuntu-22.04"
  zone_id    = data.datamammoth_zones.all.zones[0].id
}

output "server_ip" {
  value = datamammoth_server.web.ip_address
}
