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
	<link rel="icon" type="image/x-icon" href="https://www.svgrepo.com/download/473315/network.svg">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.min.css">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css">
	<style>
		/* Your custom CSS styles here */
		table {
			border-collapse: collapse;
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
	DNSmatch          bool
}

func reportWithSummary(report ReportTable) (table, summary string) {
	if report.PeeringConnected {
		table += "<tr><td>✅</td><td>Spoke is connected to Hub.</td></tr>"
		summary += "Spoke is connected.<br>"
	} else {
		table += "<tr><td>⛔</td><td>Connectivity to Hub is broken or doesn't exsist.</td></tr>"
		summary += "Connectivity is broken - please contact Network Hub team about.<br>"
		return table, summary
	}
	if report.PeeringSynced {
		table += "<tr><td>✅</td><td>Spoke is fully synchronized with Hub.</td></tr>"
		summary += "Spoke is fully synchronized.<br>"
	} else {
		table += "<tr><td>⛔</td><td>Spoke is not fully synchronized with Hub.</td></tr>"
		summary += "Spoke should be synchronized - please use self-service form for re-configuration and re-run the check.<br>"
	}
	if report.CustomNSGRules {
		table += "<tr><td>🔔</td><td>There are NSGs(one or many) with custom rules.</td></tr>"
		summary += "Don't use custom rules in NSG - please remove them and re-run the check or contact Cloud team.<br>"
	}
	if report.CustomUDR || report.SubnetsWithoutUDR {
		table += "<tr><td>⛔</td><td>Some subnets has incorrect UDR assocation(non-default).</td></tr>"
		summary += "Subnets should have default UDR - please use self-service form for re-configuration and re-run the check.<br>"
	}

	return table, summary
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
		report.DNSmatch = true
	} else {
		for _, dns := range vnetInfo.DNS {
			for _, allowed := range allowedDNS {
				if dns == allowed {
					report.DNSmatch = true
				}
			}
		}
	}

	// Draws HTML table with the results
	table, summary := reportWithSummary(report)
	result = HTMLHeader + `
	<body>
		<table>	
			<tr>
				<th colspan="2">Spoke VNet Report</th>
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
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defaultHandler(w, r)
	})
	log.Println("[INF] Listening on port", httpInvokerPort)
	log.Fatal(http.ListenAndServe(":"+httpInvokerPort, mux))
}
