package main

import (
	"fmt"
	"os"
)

const (
	releaseEnvKey  = "CLOUD_DEPLOY_RELEASE"
	projectEnvKey  = "CLOUD_DEPLOY_PROJECT"
	locationEnvKey = "CLOUD_DEPLOY_LOCATION"
	pipelineEnvKey = "CLOUD_DEPLOY_DELIVERY_PIPELINE"
)

func main() {
	if err := do(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}

func do() error {
	return nil
}

// Step 1. Get current  resources
// Part a - prepare for a query
// Execute query
// Save results
func envValue(envKey string) string {
	return os.Getenv(envKey)
}
