# Terraform Provider for DataMammoth

Manage DataMammoth infrastructure as code with Terraform.

> **Status**: Under development. Not yet published to the Terraform Registry.

## Installation

```hcl
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
```

## Example Usage

### Create a Server

```hcl
resource "datamammoth_server" "web" {
  hostname   = "web-01"
  product_id = "prod_abc"
  image_id   = "img_ubuntu2204"
  region     = "eu-central-1"

  tags = {
    environment = "production"
    team        = "platform"
  }
}

output "server_ip" {
  value = datamammoth_server.web.ip_address
}
```

### Firewall Rules

```hcl
resource "datamammoth_firewall_rule" "ssh" {
  server_id = datamammoth_server.web.id
  protocol  = "tcp"
  port      = 22
  source    = "0.0.0.0/0"
  action    = "allow"
}

resource "datamammoth_firewall_rule" "http" {
  server_id = datamammoth_server.web.id
  protocol  = "tcp"
  port      = 80
  source    = "0.0.0.0/0"
  action    = "allow"
}

resource "datamammoth_firewall_rule" "https" {
  server_id = datamammoth_server.web.id
  protocol  = "tcp"
  port      = 443
  source    = "0.0.0.0/0"
  action    = "allow"
}
```

### DNS Records

```hcl
resource "datamammoth_dns_zone" "example" {
  domain = "example.com"
}

resource "datamammoth_dns_record" "www" {
  zone_id = datamammoth_dns_zone.example.id
  name    = "www"
  type    = "A"
  value   = datamammoth_server.web.ip_address
  ttl     = 300
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `datamammoth_server` | Manage VPS instances |
| `datamammoth_firewall_rule` | Manage firewall rules |
| `datamammoth_dns_zone` | Manage DNS zones |
| `datamammoth_dns_record` | Manage DNS records |
| `datamammoth_ssh_key` | Manage SSH keys |
| `datamammoth_snapshot` | Manage server snapshots |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `datamammoth_products` | List available products |
| `datamammoth_images` | List available OS images |
| `datamammoth_regions` | List available regions |

## Documentation

- [API Reference](https://data-mammoth.com/api-docs/reference)
- [Getting Started Guide](https://data-mammoth.com/api-docs/guides)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
