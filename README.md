# terraform-azure-structure

export TF_VAR_spoke_vnet_id="/subscriptions/f406059a-f933-45e0-aefe-e37e0382d5de/resourceGroups/spoke-vnet/providers/Microsoft.Network/virtualNetworks/spoke-vnet"
export TF_VAR_api_version="2023-09-01"
export TF_VAR_hub_vnet_id="/subscriptions/f406059a-f933-45e0-aefe-e37e0382d5de/resourceGroups/spoke-vnet/providers/Microsoft.Network/virtualNetworks/spoke-vnet/virtualNetworkPeerings/hub-vnet"
export TF_CONF_PATH="../terraform-workflows/deployments/spoke-vnet-check/"
export ARM_USE_MSI=true
export ARM_SUBSCRIPTION_ID=f406059a-f933-45e0-aefe-e37e0382d5de
export ARM_TENANT_ID=f35c0d65-3fd8-489a-88fc-95b8cbad0bff
export ARM_CLIENT_ID=3b7a0816-e5b1-4433-b18a-e915047ba845

* https://github.com/Azure/terraform/tree/master
* [https://github.com/Azure/terraform-azurerm-subnets](https://github.com/Azure/terraform-azurerm-subnets)
* [Azure Data Labs - Modules](https://github.com/Azure/azure-data-labs-modules?tab=readme-ov-file)
* [Terraform module to deploy Azure DevOps self-hosted agents running on Azure Container Instance](https://github.com/Azure/terraform-azurerm-aci-devops-agent)
* [Data source for AzureRM provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config)
* https://learn.microsoft.com/en-us/azure/container-apps/managed-identity
* https://stackoverflow.com/questions/75771529/trying-to-authenticate-with-azure-using-user-managed-identity-fails-with-401