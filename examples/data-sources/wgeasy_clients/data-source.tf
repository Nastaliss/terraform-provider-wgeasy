data "wgeasy_clients" "all" {}

output "all_clients" {
  value = data.wgeasy_clients.all.clients
}
