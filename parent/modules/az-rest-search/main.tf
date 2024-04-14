terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

provider "azapi" {
}

variable "api_ver" {
  type = string
}

variable "resource_type" {
  type = string
}

variable "subscription_id" {
  type = string
}

variable "resource_group" {
  type = string
  default = null
}

data "azapi_resource_list" "by_sub" {
  type                   = "${var.resource_type}@${var.api_ver}"
  parent_id              = "/subscriptions/${var.subscription_id}"
  response_export_values = ["*"]
  count = var.resource_group == null ? 1 : 0
}

data "azapi_resource_list" "by_rg" {
  type                   = "${var.resource_type}@${var.api_ver}"
  parent_id              = "/subscriptions/${var.subscription_id}/resourceGroups/${var.resource_group}"
  response_export_values = ["*"]
  count = var.resource_group == null ? 0 : 1
}

output "sub_resource_list" {
  value = data.azapi_resource_list.by_sub.*.output 
}

output "rg_resource_list" {
  value = data.azapi_resource_list.by_rg.*.output
}