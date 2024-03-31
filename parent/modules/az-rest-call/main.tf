terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

provider "azapi" {
}

variable "resource_id" {
  type = string
}

variable "api_ver" {
  type = string
}

variable "method" {
  type = string
  default = "GET"
}

locals {
  res_id_split = split("/", var.resource_id)
  res_id_split_len = length(local.res_id_split)
  res_type = "${local.res_id_split[local.res_id_split_len - 3]}/${local.res_id_split[local.res_id_split_len - 2]}"
  subscription_id = local.res_id_split[2]
}


resource "azapi_resource_action" "resource_action" {
  type = "${local.res_type}@${var.api_ver}"
  resource_id = var.resource_id
  method = var.method
}

data "azapi_resource" "resource" {
  type = "${local.res_type}@${var.api_ver}"
  resource_id = var.resource_id
  response_export_values = ["*"]
}


data "azapi_resource_list" "resource_list" {
  type                   = "${local.res_type}@${var.api_ver}"
  parent_id              = "/subscriptions/${local.subscription_id}"
  response_export_values = ["*"]
}

output "resource_list" {
  value = jsondecode(data.azapi_resource_list.resource_list.output)
}
