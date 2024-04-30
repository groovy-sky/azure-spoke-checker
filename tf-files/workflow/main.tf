terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

variable "spoke_vnet_id" {
  type = string
}

variable "api_version" {
  type = string
  default = "2023-09-01"
}

variable "hub_vnet_id" {
  type = string
}

locals {  
  spoke_sub_id = module.spoke_vnet_check.parsed_resource_id.subscription_id
  spoke_rg = module.spoke_vnet_check.parsed_resource_id.resource_group
  spoke_name = jsondecode(module.spoke_vnet_check.resource_information[0]).name
  spoke_default_udr = "${local.spoke_name}-udr"
  peerings = try(jsondecode(module.spoke_vnet_check.resource_information[0]).properties.virtualNetworkPeerings,null)
  vnet_ips = jsondecode(module.spoke_vnet_check.resource_information[0]).properties.addressSpace.addressPrefixes
  subnets = try(jsondecode(module.spoke_vnet_check.resource_information[0]).properties.subnets, null)
  subnets_nsg = [for s in local.subnets : try(s.properties.networkSecurityGroup.id, null)]
  depends-on = [module.spoke_vnet_check]
}  

// Search for default UDR in the same resource group as Spoke VNet
module "spoke_vnet_udr_search" {
  source = "../modules/az-rest-search"
  subscription_id = local.spoke_sub_id
  resource_group = local.spoke_rg
  resource_type = "Microsoft.Network/routeTables"
  api_ver = "2023-11-01"
}

// Obtain the resource information for the Spoke VNet
module "spoke_vnet_check" {  
  source = "../modules/az-rest-call"  
  api_ver = var.api_version
  resource_id = var.spoke_vnet_id
}

module "subnet_nsg_check"{
  source = "../modules/az-rest-call" 
  for_each = toset(compact(local.subnets_nsg))
  api_ver = var.api_version
  resource_id = each.value
}

output "vnet_info" {
  value = {
  default_udr = try(jsondecode(module.spoke_vnet_udr_search.rg_resource_list[0]).value[0].id,null)
  address_spaces = local.vnet_ips
  peerings = local.peerings
  }
}

output "subnets_info" {  
  value = [for s in local.subnets : {  
    name = s.name  
    udr = try(s.properties.routeTable.id, null)
    nsg = try(s.properties.networkSecurityGroup.id, null)
  }]  
}  

output "nsg_info" {  
  value = [for s in module.subnet_nsg_check : {  
    total_rules =  length(jsondecode(s.resource_information[0]).properties.securityRules)
    nsg_name = jsondecode(s.resource_information[0]).name
    nsg_id = jsondecode(s.resource_information[0]).id
    }]  
}  