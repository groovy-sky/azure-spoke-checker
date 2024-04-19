package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func main() {
	// httpInvokerPort, exists := os.LookupEnv("HTTP_PORT")

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
	err = tf.Apply(context.Background())
	if err != nil {
		log.Fatalf("error running terraform apply: %s", err)
	}

	// Get the outputs
	outputs, err := tf.Output(context.Background())
	if err != nil {
		log.Fatalf("error getting outputs: %s", err)
	}

	for name, value := range outputs {
		fmt.Printf("%s:%s\n", name, string(value.Value))
	}

	//fmt.Println(string(outputs["vnet_info"].Value))
}
