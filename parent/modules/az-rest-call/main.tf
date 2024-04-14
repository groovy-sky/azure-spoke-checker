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

variable "action"{
  type = string
  default = null
}

// Parses the resource_id to get the subscription_id, resource_group, resource_provider, and resource_type
locals {
  res_id_split = split("/", var.resource_id)
  subscription_id = local.res_id_split[2]
  resource_group = local.res_id_split[4]
  res_provider = local.res_id_split[6]
  res_type = local.res_id_split[7]
}

// If action is not null, then perform the action on the resource
// https://registry.terraform.io/providers/Azure/azapi/latest/docs/resources/azapi_resource_action
resource "azapi_resource_action" "res_action" {
  type = "${local.res_provider}/${local.res_type}@${var.api_ver}"
  resource_id = var.resource_id
  method = var.method
  action = var.action
  count = var.action == null ? 0 : 1
}

// If action is null, then get the resource
// https://registry.terraform.io/providers/Azure/azapi/latest/docs/data-sources/azapi_resource
data "azapi_resource" "res_info" {
  type = "${local.res_provider}/${local.res_type}@${var.api_ver}"
  resource_id = var.resource_id
  response_export_values = ["*"]
  count = var.action == null ? 1 : 0
}

output "action_result" {
  value = resource.azapi_resource_action.res_action
}

output "resource_information" {
  value = data.azapi_resource.res_info.*.output
}