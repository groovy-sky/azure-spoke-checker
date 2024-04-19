# terraform-azure-structure

export TF_VAR_spoke_vnet_id="/subscriptions/f406059a-f933-45e0-aefe-e37e0382d5de/resourceGroups/spoke-vnet/providers/Microsoft.Network/virtualNetworks/spoke-vnet"
export TF_VAR_api_version="2023-09-01"
export TF_VAR_hub_vnet_id="/subscriptions/f406059a-f933-45e0-aefe-e37e0382d5de/resourceGroups/spoke-vnet/providers/Microsoft.Network/virtualNetworks/spoke-vnet/virtualNetworkPeerings/hub-vnet"
export TF_CONF_PATH="../terraform-workflows/deployments/spoke-vnet-check/"

* https://github.com/Azure/terraform/tree/master
* [https://github.com/Azure/terraform-azurerm-subnets](https://github.com/Azure/terraform-azurerm-subnets)
* [Azure Data Labs - Modules](https://github.com/Azure/azure-data-labs-modules?tab=readme-ov-file)
* [Terraform module to deploy Azure DevOps self-hosted agents running on Azure Container Instance](https://github.com/Azure/terraform-azurerm-aci-devops-agent)
* [Data source for AzureRM provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config)
