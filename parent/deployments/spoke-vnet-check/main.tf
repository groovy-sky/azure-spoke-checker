terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

variable "resource_id" {
  type = string
}

variable "api_version" {
  type = string
}

variable "hub_vnet_id" {
  type = string
}

locals {  
  vnet_name = jsondecode(module.spoke_vnet_check.resource_information[0]).name
  peerings = jsondecode(module.spoke_vnet_check.resource_information[0]).properties.virtualNetworkPeerings
  address_space = jsondecode(module.spoke_vnet_check.resource_information[0]).properties.addressSpace.addressPrefixes
  subnets = jsondecode(module.spoke_vnet_check.resource_information[0]).properties.subnets
  hub_peer = [for item in local.peerings : item if lower(item.id) == lower(var.hub_vnet_id)]
  hub_peer_state = [for item in local.hub_peer : item.properties.peeringState]
  hub_peer_sync = [for item in local.hub_peer : item.properties.peeringSyncLevel]
}  

module "spoke_vnet_check" {  
  source = "../../modules/az-rest-call"  
  api_ver = var.api_version
  resource_id = var.resource_id
}

output "peerings" {
  value = jsondecode(module.spoke_vnet_check.resource_information[0]).name
}