resource "wgeasy_client" "example" {
  name = "my-laptop"
}

resource "wgeasy_client" "custom" {
  name = "custom-client"

  allowed_ips        = ["0.0.0.0/0"]
  server_allowed_ips = ["10.8.0.0/24"]
  dns                = ["1.1.1.1", "8.8.8.8"]
  mtu                = 1420
  persistent_keepalive = 25
  enabled            = true
}
