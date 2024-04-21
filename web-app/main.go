package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type NSGInfo struct {
	NSGID      string `json:"nsg_id"`
	NSGName    string `json:"nsg_name"`
	TotalRules int    `json:"total_rules"`
}

type SubnetsInfo struct {
	Name string `json:"name"`
	NSG  string `json:"nsg"`
	UDR  string `json:"udr"`
}

type VNetInfo struct {
	AddressSpaces []string `json:"address_spaces"`
	DefaultUDR    string   `json:"default_udr"`
	PeeringState  []string `json:"peering_state"`
	PeeringSync   []string `json:"peering_sync"`
}

type AzureInfo struct {
	NSGInfo     []NSGInfo     `json:"nsg_info"`
	SubnetsInfo []SubnetsInfo `json:"subnets_info"`
	VNetInfo    VNetInfo      `json:"vnet_info"`
}

type OutputMeta struct {
	Sensitive bool            `json:"sensitive"`
	Type      json.RawMessage `json:"type"`
	Value     json.RawMessage `json:"value"`
}

func runTerraform(spokeId, hubId string) {
	var nsgInfo []NSGInfo
	var subnetsInfo []SubnetsInfo
	var vnetInfo VNetInfo

	tfConfPath, exists := os.LookupEnv("TF_CONF_PATH")

	if !exists {
		// Getting current directory and use it as default
		tfConfPath, _ = os.Getwd()
	}

	tfBinPath, exists := os.LookupEnv("TF_BIN_PATH")

	if !exists {
		tfBinPath = "terraform"
	}

	tf, err := tfexec.NewTerraform(tfConfPath, tfBinPath)
	if err != nil {
		log.Fatalf("error creating tfexec instance: %s", err)
	}

	// Initialize Terraform
	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running terraform init: %s", err)
	}

	// Apply the Terraform configuration
	err = tf.Apply(context.Background(), tfexec.Var("spoke_vnet_id="+spokeId), tfexec.Var("hub_vnet_id="+hubId))
	if err != nil {
		log.Fatalf("error running terraform apply: %s", err)
	}

	// Get the outputs
	outputs, err := tf.Output(context.Background())
	if err != nil {
		log.Fatalf("error running terraform output: %s", err)
	}

	for key, output := range outputs {
		var value json.RawMessage
		err := json.Unmarshal(output.Value, &value)
		if err != nil {
			log.Fatalf("error decoding output value: %s", err)
		}

		// Parses the outputs and stores them in the corresponding struct using switch
		switch key {
		case "nsg_info":
			var nsgs []NSGInfo
			err := json.Unmarshal(value, &nsgs)
			if err != nil {
				log.Fatalf("error decoding nsg_info: %s", err)
			}
			nsgInfo = append(nsgInfo, nsgs...)

		case "subnets_info":
			var subnets []SubnetsInfo
			err := json.Unmarshal(value, &subnets)
			if err != nil {
				log.Fatalf("error decoding subnets_info: %s", err)
			}
			subnetsInfo = append(subnetsInfo, subnets...)

		case "vnet_info":
			err := json.Unmarshal(value, &vnetInfo)
			if err != nil {
				log.Fatalf("error decoding vnet_info: %s", err)
			}
		}

	}
	fmt.Printf("NSG Info: %v\n", nsgInfo.)
}

func sanitazeInput(input string) string {
	// Removes all non-alphanumeric characters from input
	regex := `[^\w\d\.\/\@\;\_\-]`
	r := regexp.MustCompile(regex)
	return r.ReplaceAllString(input, "")
}

// Checks that input matches Azure Resource Id format
func validateResID(input string) bool {
	// Lowercase input and validates that input is valid Azure Resource Id using regex
	input = strings.ToLower(input)
	regex := `^\/subscriptions\/.{36}\/resourcegroups\/.*\/providers\/[a-zA-Z0-9]*.[a-zA-Z0-9]*\/[a-zA-Z0-9]*\/.*`
	r := regexp.MustCompile(regex)
	return r.MatchString(input)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
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
				<form action="/" method="POST">
					<label for="vnetid">Spoke VNet ID:</label>
					<input type="text" id="resid" name="vnetid" value="" required /><br>
					<input type="submit" value="Submit">
				</form>
			</fieldset>
		</body>
		</html>
		`))
	case "POST":
		r.ParseForm()
		vnetID := sanitazeInput(r.FormValue("vnetid"))
		hubId, exists := os.LookupEnv("HUB_VNET_ID")
		log.Println("VNet ID:", vnetID)
		if validateResID(vnetID) && exists {
			runTerraform(vnetID, hubId)
		}
	}
}

func main() {

	httpInvokerPort, exists := os.LookupEnv("HTTP_PORT")
	if exists {
		log.Println("HTTP_PORT: " + httpInvokerPort)
	} else {
		httpInvokerPort = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defaultHandler(w, r)
	})
	log.Println("[INF] Listening on port", httpInvokerPort)
	log.Fatal(http.ListenAndServe(":"+httpInvokerPort, mux))
}
