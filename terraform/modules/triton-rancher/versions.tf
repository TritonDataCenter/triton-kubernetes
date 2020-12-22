
terraform {
  required_version = ">= 0.13"

  required_providers {
    triton = {
      source = "joyent/triton"
      version = "0.8.1"
    }
  }
}
