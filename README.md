# Azure Spoke Checker  

![](logo.svg)

## Introduction
 
Azure Spoke Checker is a web application, built using Golang, designed to streamline and automate the validation of a Spoke Virtual Network (VNet) within a Hub-Spoke topology in Microsoft Azure. For those unfamiliar with the term, a Hub-Spoke topology is a network configuration where a central 'Hub' VNet communicates with multiple 'Spoke' VNets, akin to the hub and spokes of a bicycle wheel.

## Overview

The application simplifies the process of network validation by carrying out the following operations:

* It collects a user-provided Spoke VNet ID through a web form. Think of this as the specific address of the Spoke within the Azure environment.
* Using the Terraform webapp, it then fetches in-depth information about the Spoke VNet. This includes details about its subnets (smaller divisions within the VNet), any associated User Defined Routes (UDRs, or custom traffic routes), Network Security Groups (NSGs, rules that allow or deny network traffic), DNS settings, and its peering status with the Hub VNet.
* The application processes this data and generates an easy-to-understand, detailed report.
* The report checks if the Spoke is correctly connected to the Hub, whether each subnet is linked to the right UDR, if NSGs are correctly configured, and if the DNS is on an approved IP list.

## Real-World Application

For example, in a real-world scenario, a network administrator could use Azure Spoke Checker to automate manual checks of the network, saving valuable time and reducing the chance of human error. It could also help ensure that the network is secure and compliant with company policies by checking UDRs, NSGs, and DNS settings. If there are any connectivity or configuration issues, the tool can aid in troubleshooting by identifying potential problems. Finally, it provides a useful report that documents the Spoke VNet's configuration, useful for both current and future reference.
Purpose
 
Azure Spoke Checker was designed to:

* Increase efficiency and accuracy by automating manual checks of network configurations, saving time and reducing errors.
* Promote compliance and security by verifying important network settings such as UDRs, NSGs, and DNS.
* Assist in troubleshooting and diagnostics by uncovering potential issues with connectivity or configuration.
* Provide reporting and documentation of the Spoke VNet's configuration for reference and auditing purposes.

## Related materials

* https://github.com/Azure/terraform/tree/master
* [https://github.com/Azure/terraform-azurerm-subnets](https://github.com/Azure/terraform-azurerm-subnets)
* [Azure Data Labs - Modules](https://github.com/Azure/azure-data-labs-modules?tab=readme-ov-file)
* [Terraform module to deploy Azure DevOps self-hosted agents running on Azure Container Instance](https://github.com/Azure/terraform-azurerm-aci-devops-agent)
* [Data source for AzureRM provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config)
* https://learn.microsoft.com/en-us/azure/container-apps/managed-identity
* https://stackoverflow.com/questions/75771529/trying-to-authenticate-with-azure-using-user-managed-identity-fails-with-401