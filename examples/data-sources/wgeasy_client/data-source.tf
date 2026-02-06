data "wgeasy_client" "example" {
  id = 1
}

output "client_name" {
  value = data.wgeasy_client.example.name
}

output "client_ipv4" {
  value = data.wgeasy_client.example.ipv4_address
}
