#!/bin/bash
appName="spoke-vnet-checker"
rgName="$appName-rg"
location="westeurope"
appIdentity="$appName-system-identity"

hubVnetId="/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/hub-vnet-rg/providers/Microsoft.Network/virtualNetworks/hub-vnet"

az group create --name $rgName --location $location

# Registry deployment
registryName=$(echo $appName | awk '{print tolower($0)}' | md5sum | cut -d" " -f1)
registryDns="$registryName.azurecr.io"
registryId=$(az acr create --resource-group $rgName \
--name $registryName --sku Standard \
--admin-enabled true \
--location $location \
--query id --output tsv)

# https://learn.microsoft.com/en-us/azure/container-registry/anonymous-pull-access
az acr update --name $registryName --anonymous-pull-enabled

# Identity creation
# https://learn.microsoft.com/en-us/azure/app-service/tutorial-custom-container?tabs=azure-cli&pivots=container-linux
principalId=$(az identity create --name $appIdentity --resource-group $rgName --query principalId --output tsv)
az role assignment create --assignee $principalId --scope $registryId --role "AcrPull"

# Lists available subcriptions and tries to assing Reader role to each of subscription for the managed identity
# ignore errors if the managed identity already has the role assigned
subIds=$(az account list --query "[].id" -o tsv)
for subId in $subIds
do
    az role assignment create --assignee $principalId --scope /subscriptions/$subId --role "Reader" || true
done

# Login Azure Registry
registryToken=$(az acr login --name $registryName --expose-token --output tsv --query accessToken )        
registryImage="$appName:latest"

# Login Azure Docker Registry
docker login $registryDns -u  00000000-0000-0000-0000-000000000000 -p $registryToken

# Build Docker image
echo "## Building Docker image"
#az acr build --resource-group $rgName --registry $registryName --image $registryImage .

# Deploys the app
echo "## Initial App deployment"
deployResult=$(az deployment group create --resource-group $rgName --parameters appName=$appName appImage="$registryDns/$appName:latest" hubVnetID=$hubVnetId manageIdentityName=$appIdentity --template-file arm-template/azuredeploy.json)

# Getting apps managed identity id
appResId=$(echo $deployResult | jq -r '.properties.outputs.containerAppId.value')

# Getting apps managed identity id
echo "## Assigning Contributor role to the App"
identityId=$(az resource show --id $appResId --query "identity.principalId" -o tsv)
