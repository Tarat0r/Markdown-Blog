terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean" 
      version = "~> 2.0"
    }
  }

  required_version = ">= 1.0.0"
}


provider "digitalocean" {
  token = var.do_token
}

variable "do_token" {}
variable "ssh_key_name" {}
variable "ssh_key_id" {}

resource "digitalocean_droplet" "web" {
  name   = "docker-droplet"
  region = "fra1"
  size   = "s-1vcpu-512mb-10gb"  # Самый дешёвый размер
  image  = "ubuntu-24-04-x64"

  ssh_keys = [var.ssh_key_id]

  user_data = <<-EOF
              #!/bin/bash
              apt-get update
              apt-get install -y docker.io docker-compose
              systemctl enable docker
              systemctl start docker
            EOF
}

output "droplet_ip" {
  value = digitalocean_droplet.web.ipv4_address
}
