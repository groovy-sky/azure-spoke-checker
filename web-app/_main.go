package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

// Login to Azure, using different kind of methods - credentials, managed identity
func azureLogin() (cred *azidentity.ChainedTokenCredential, err error) {
	// Create credentials using Managed Identity, Azure CLI, Environment variables
	manCred, _ := azidentity.NewManagedIdentityCredential(nil)
	cliCred, _ := azidentity.NewAzureCLICredential(nil)
	envCred, _ := azidentity.NewEnvironmentCredential(nil)
	// If connection to 169.254.169.254 - skip Managed Identity Credentials
	if _, tcpErr := net.Dial("tcp", "169.254.169.254:80"); tcpErr != nil {
		cred, err = azidentity.NewChainedTokenCredential([]azcore.TokenCredential{cliCred, envCred}, nil)
	} else {
		cred, err = azidentity.NewChainedTokenCredential([]azcore.TokenCredential{manCred, cliCred, envCred}, nil)
	}

	return cred, err
}

func defaultHandler(w http.ResponseWriter, r *http.Request, login *azidentity.ChainedTokenCredential) {
	w.Write([]byte(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Azure Spoke Checker</title>
		<link rel="icon" type="image/x-icon" href="https://www.svgrepo.com/download/473315/network.svg">
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.min.css">
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css">
		<style>
			/* Your custom CSS styles here */
			form {
				margin: 20px;
				padding: 10px;
				display: inline-block;
			}
			big {
				margin: 5px;
				align: center;
				style: bold;
			}
			fieldset {
				border: none;
				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
				margin-bottom: 10px;
				padding: 5px;
				width: 20%; 
			}
			fieldset:nth-child(odd) {
				background-color: #f5f5f5; /* Lighter color for odd fieldsets */
			}
			fieldset:nth-child(even) {
				background-color: #ebebeb; /* Darker color for even fieldsets */
			}
		</style>
	</head>
	<body>
		<fieldset>
			<form action="whitelistip" method="POST">
			<legend>Whitelist IP for PaaS</legend>
				<label for="resid">PaaS resource's ID:</label>
				<input type="text" id="resid" name="resid" value="" required /><br>
				<input type="checkbox" id="debug" name="debug" value="debug" checked><br>
				<label for="debug">Print result</label><br>
				<input type="submit" value="Submit">
			</form>
		</fieldset>
	</body>
	</html>
	`))
}

func main() {
	login, err := azureLogin()
	if err != nil {
		log.Fatal("[ERR] : Failed to login:\n", err)
	}
	httpInvokerPort, exists := os.LookupEnv("HTTP_PORT")
	if exists {
		log.Println("HTTP_PORT: " + httpInvokerPort)
	} else {
		httpInvokerPort = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defaultHandler(w, r, login)
	})
	log.Println("[INF] Listening on port", httpInvokerPort)
	log.Fatal(http.ListenAndServe(":"+httpInvokerPort, mux))
}
