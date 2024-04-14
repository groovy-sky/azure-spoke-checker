terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

locals {  
  global = yamldecode(file("../global.yaml"))  
}  

provider "azurerm" {  
  features {}  
  subscription_id = local.global[terraform.workspace].subscription_id
  tenant_id = local.global.tenant_id
}

variable "resource_type" {
  type = string
}

variable "api_version" {
  type = string
}

module "az-rest-search" {  
  source = "../../modules/az-rest-search"  
  api_ver = var.api_version
  resource_type = var.resource_type  
  subscription_id = local.global[terraform.workspace].subscription_id  
}

output "sub_resource_list" {  
  value = module.az-rest-search.sub_resource_list
}