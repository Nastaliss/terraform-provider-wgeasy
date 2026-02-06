terraform {
  required_providers {
    wgeasy = {
      source = "Nastaliss/wgeasy"
    }
  }
}

provider "wgeasy" {
  endpoint = "http://localhost:51821"
  username = "admin"
  password = "secret"
}
