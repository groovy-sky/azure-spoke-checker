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

const HTMLHeader = `
<!DOCTYPE html>
<html>
<head>
	<title>Azure Spoke Checker</title>
	<link rel="icon" type="image/x-icon" href="https://raw.githubusercontent.com/groovy-sky/azure-spoke-checker/main/logo.svg">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.min.css">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css">
	<style>
		/* Your custom CSS styles here */
		table {  
			border-collapse: collapse;  
			margin: 20px; 
		} 
		th, td {
			border: 1px solid #dddddd;
			text-align: left;
			padding: 8px;
		}
		th {
			background-color: #f2f2f2;
		}
		tr:nth-child(even) {
			background-color: #f2f2f2;
		}
		tr:nth-child(odd) {
			background-color: #e6e6e6;
		}
	</style>
</head>
`

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
	AddressSpaces []string      `json:"address_spaces"`
	DefaultUDR    string        `json:"default_udr"`
	PeeringState  []VNetPeering `json:"peerings"`
	DNS           []string      `json:"dns"`
}

type VNetPeering struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Properties struct {
		PeeringState         string `json:"peeringState"`
		PeeringSyncLevel     string `json:"peeringSyncLevel"`
		RemoteVirtualNetwork struct {
			ID string `json:"id"`
		} `json:"remoteVirtualNetwork"`
	} `json:"properties"`
	Type string `json:"type"`
}

type ReportTable struct {
	PeeringConnected  bool
	PeeringSynced     bool
	CustomNSGRules    bool
	CustomUDR         bool
	SubnetsWithoutUDR bool
	DNSmismatch       bool
}

func reportWithSummary(report ReportTable) (table, summary string) {

	if report.PeeringConnected {
		table += "<tr><td>✅</td><td>VNet Peering is connected.</td></tr>"
	} else {
		table += "<tr><td>⛔</td><td>Connectivity to Hub is broken or doesn't exsist.</td></tr>"
		summary += "No active connection, between Hub and Spoke, were found. Connect VNet to Hub. <br>"
		return table, summary
	}
	if report.PeeringSynced {
		table += "<tr><td>✅</td><td>VNet Peering is fully synchronized.</td></tr>"
	} else {
		table += "<tr><td>⛔</td><td>Spoke is not fully synchronized with Hub.</td></tr>"
		summary += "Spoke's address space was modified. Sync the virtual network peer to Hub VNet.<br>"
	}
	if report.CustomNSGRules {
		table += "<tr><td>🔔</td><td>Some Spoke subnets uses non-empty NSGs.</td></tr>"
	}
	if report.CustomUDR {
		table += "<tr><td>🔔</td><td>Some Spoke subnets uses non-default UDR.</td></tr>"
	}
	if report.SubnetsWithoutUDR {
		table += "<tr><td>🔔</td><td>Some Spoke subnets don't have UDR association.</td></tr>"
	}
	if report.DNSmismatch {
		table += "<tr><td>🔔</td><td>Spoke VNet uses non-default DNS.</td></tr>"
	}

	if summary == "" {
		summary = "No significant issues were observed."
		if report.CustomNSGRules || report.CustomUDR || report.DNSmismatch || report.SubnetsWithoutUDR {
			summary += "If you have any connectivity issues - try to check sections with 🔔 symbol."
		}
	}

	return table, summary
}

func contains(s []string, e string) bool {
	// Checks if a string slice contains a specific string
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkSpokeVNet(spokeId, hubId string) (result string) {
	var report ReportTable
	var nsgInfo []NSGInfo
	var subnetsInfo []SubnetsInfo
	var vnetInfo VNetInfo
	var peeringState, peeringSyncLevel string

	log.Printf("[INF] Checking Spoke VNet %s and its connectivity to %s\n", spokeId, hubId)

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
		log.Printf("[ERR] tfexec instance creation: %s", err)
		return "Something went wrong. Please try again later."
	}

	ctx := context.Background()

	// Initialize Terraform
	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		log.Printf("[ERR] terraform init: %s", err)
		return "Something went wrong. Please try again later."
	}

	defer tf.Destroy(ctx, tfexec.Var("spoke_vnet_id="+spokeId), tfexec.Var("hub_vnet_id="+hubId))

	// Apply the Terraform configuration
	err = tf.Apply(ctx, tfexec.Var("spoke_vnet_id="+spokeId), tfexec.Var("hub_vnet_id="+hubId))
	if err != nil {
		log.Printf("[ERR] terraform apply: %s", err)
		return "Couln't obtain the information about the Spoke VNet. Please check resource ID and try again."
	}

	// Get the outputs
	outputs, err := tf.Output(ctx)
	if err != nil {
		log.Printf("[ERR] terraform output: %s", err)
		return "Something went wrong. Please try again later."
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

	// Search a peering to hub and get its state
	for _, peering := range vnetInfo.PeeringState {
		if strings.ToLower(peering.Properties.RemoteVirtualNetwork.ID) == strings.ToLower(hubId) {
			peeringState = strings.ToLower(peering.Properties.PeeringState)
			peeringSyncLevel = strings.ToLower(peering.Properties.PeeringSyncLevel)
		}
	}

	if peeringState == "connected" {
		report.PeeringConnected = true
	}

	if peeringSyncLevel == "fullyinsync" {
		report.PeeringSynced = true
	}

	// Checks that there is no NSG with more than 0 rules
	for _, nsg := range nsgInfo {
		if nsg.TotalRules > 0 {
			fmt.Printf("[INF] NSG %s has %d rules\n", nsg.NSGName, nsg.TotalRules)
			report.CustomNSGRules = true
		}
	}

	// Checks that each subnet has UDR from vnetInfo.DefaultUDR
	for _, subnet := range subnetsInfo {
		if subnet.UDR != vnetInfo.DefaultUDR {
			fmt.Printf("[INF] Subnet %s has UDR %s\n", subnet.Name, subnet.UDR)
			report.CustomUDR = true
		}
	}

	defaultDNS, exists := os.LookupEnv("DEFAULT_DNS")
	var allowedDNS []string

	if exists {
		allowedDNS = strings.Split(defaultDNS, ",")
	}

	if len(allowedDNS) == len(vnetInfo.DNS) {
		fmt.Println("[INF] DNS Address:", allowedDNS, "VNet DNS:", vnetInfo.DNS, len(allowedDNS))
	}

	// Checks that Spoke DNS IPs (if DEFAULT_DNS is defined) are in the allowed DNS list or (if DEFAULT_DNS is not defined) is empty list
	if len(vnetInfo.DNS) == 0 {
		if len(allowedDNS) > 0 {
			report.DNSmismatch = true
		}
	} else {
		// Walks through Spoke DNS IPs and checks if any of them is not in the allowed DNS list
		for _, dns := range vnetInfo.DNS {
			if !contains(allowedDNS, dns) {
				report.DNSmismatch = true
			}
		}

	}

	// Draws HTML table with the results
	table, summary := reportWithSummary(report)
	result = HTMLHeader + `
	<body>
		<table>	
			<tr>
				<th colspan="2" align="center"> <center> Azure Spoke Checker's report </center> </th>
			</tr>
				` + table + `
				<tr><td> Summary </td> <td width="600">
			 	` + summary + `
				</td></tr>
		</table>
	</body>
	</html>
	`
	return result

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
		w.Write([]byte(HTMLHeader + `
		<body>
		<table>
		<tr>
				<th colspan="2" align="center"> <center>Spoke VNet ID: </center> </th>
			</tr>
			<tr>
				<th colspan="2" align="center"> 
			<fieldset>
			<br>
				<form action="/" method="POST">
					<input type="text" id="resid" name="vnetid" value="" required /><br>
					<center><input type="submit" value="Submit"></center>
				</form>
			</fieldset>
			</th>
			</tr>
		</table>
		</body>
		</html>
		`))
	case "POST":
		r.ParseForm()
		vnetID := sanitazeInput(r.FormValue("vnetid"))
		hubID, exists := os.LookupEnv("HUB_VNET_ID")
		log.Println("VNet ID:", vnetID)
		if validateResID(vnetID) && exists {
			response := checkSpokeVNet(vnetID, hubID)
			w.Write([]byte(response))
		} else {
			w.Write([]byte("Invalid VNet ID has been provided."))
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func main() {

	httpInvokerPort, exists := os.LookupEnv("HTTP_PORT")
	if exists {
		log.Println("HTTP_PORT: " + httpInvokerPort)
	} else {
		httpInvokerPort = "8080"
	}

	_, exists = os.LookupEnv("HUB_VNET_ID")
	if !exists {
		log.Fatal("[ERR] HUB_VNET_ID is not defined.")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defaultHandler(w, r)
	})
	log.Println("[INF] Listening on port", httpInvokerPort)
	log.Fatal(http.ListenAndServe(":"+httpInvokerPort, mux))
}
