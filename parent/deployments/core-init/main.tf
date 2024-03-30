locals {  
  global = yamldecode(file("../global.yaml"))  
}  
  
provider "azurerm" {  
  features {}  
  subscription_id = local.global[terraform.workspace].subscription_id
  tenant_id = local.global.tenant_id
}