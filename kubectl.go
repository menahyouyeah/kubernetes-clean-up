package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Get current  resources
// 1) Create the query
// 2) Execute query
// 3) Save results

const (
	cloudDeployPrefix = "deploy.cloud.google.com/"
	releaseEnvKey     = "CLOUD_DEPLOY_RELEASE"
	projectEnvKey     = "CLOUD_DEPLOY_PROJECT"
	locationEnvKey    = "CLOUD_DEPLOY_LOCATION"
	pipelineEnvKey    = "CLOUD_DEPLOY_DELIVERY_PIPELINE"
	targetEnvKey      = "CLOUD_DEPLOY_TARGET"
)

// createQueryArgs creates the kubectl args to get the list of resources on the cluster.
func createQueryArgs(includeReleaseLabel bool) []string {
	var labels []string
	if includeReleaseLabel {
		labels = append(labels, fmt.Sprintf("%srelease-id=%s", cloudDeployPrefix, os.Getenv(releaseEnvKey)))
	}
	labels = append(labels, fmt.Sprintf("%sdelivery-pipeline-id=%s", cloudDeployPrefix, os.Getenv(pipelineEnvKey)))
	labels = append(labels, fmt.Sprintf("%starget-id=%s", cloudDeployPrefix, os.Getenv(targetEnvKey)))
	labels = append(labels, fmt.Sprintf("%slocation=%s", cloudDeployPrefix, os.Getenv(locationEnvKey)))
	labels = append(labels, fmt.Sprintf("%sproject-id=%s", cloudDeployPrefix, os.Getenv(projectEnvKey)))

	labelsFormatted := strings.Join(labels, ",")
	labelArg := fmt.Sprintf("-l %s", labelsFormatted)
	outputNameArg := "-o name"

	return []string{
		"api-resources",
		"--verbs=list",
		outputNameArg,
		"|",
		"xargs",
		"-n 1",
		"kubectl",
		"get",
		outputNameArg,
		labelArg,
	}
}

// CommandExecutor contains command execution information.
type CommandExecutor struct {
	// BinPath is the path the binary being used for the command (e.g. the path
	// to the kubectl binary if the kubectl command is to be used).
	binPath string
}

// CreateCommandExecutor returns a CommandExecutor for the given binary
func CreateCommandExecutor(binPath string) *CommandExecutor {
	ce := &CommandExecutor{
		binPath: binPath,
	}
	return ce
}

// execCommand runs the given command and returns the output.
func (ce CommandExecutor) execCommand(args []string) (string, error) {
	fmt.Printf("Running the following command: %s %s\n", ce.binPath, args)
	cmd := exec.Command(ce.binPath, args...)
	// By default set locations to standard error and output (visible in cloud build logs)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Write error output to two locations simultaneously. Unless disabled by a CommandOption, this will allow
	// us to see the error output as the execution is happening (by writing to standard error) and also allow us
	// to gather all stderr at the end (by also writing to var stderr).
	var stderr bytes.Buffer
	errWriter := io.MultiWriter(&stderr, cmd.Stderr)
	cmd.Stderr = errWriter

	// Write output to two locations simultaneously. Unless disabled by a CommandOption, this will allow
	// us to see the output as the execution is happening (by writing to standard out) and also allow us
	// to gather all stdout at the end (by also writing to var stdout).
	var stdout bytes.Buffer
	outWriter := io.MultiWriter(&stdout, cmd.Stdout)
	cmd.Stdout = outWriter

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command, err is %w", err)
	}

	// Wait for everything to finish.
	if err := cmd.Wait(); err != nil {
		// Read the stdErr output
		errorOutput := stderr.Bytes()
		fullErr := fmt.Errorf("error running command: %w\n%s", err, errorOutput)
		return "", fullErr
	}
	return stdout.String(), nil
}

// diffSlices returns the elements in slice1 that aren't in slice2.
func diffSlices(slice1, slice2 []string) []string {
	var diff []string
	set := make(map[string]bool)

	for _, val := range slice2 {
		set[val] = true
	}

	for _, val := range slice1 {
		// If they're not in slice 2
		if !set[val] {
			diff = append(diff, val)
		}
	}

	return diff
}

// Get a list of resources that aren't in the current set of resources.
func (ce CommandExecutor) getOldResources() ([]string, error) {
	// Get a list of all resources on the cluster.
	argsAll := createQueryArgs(false)
	output, err := ce.execCommand(argsAll)
	if err != nil {
		return nil, err
	}
	allResources := strings.Split(output, "\n")

	// Get a list of resources that were deployed as part of the latest release on the cluster.
	argsCurrent := createQueryArgs(true)
	outputCurrent, err := ce.execCommand(argsCurrent)
	if err != nil {
		return nil, err
	}
	currentResources := strings.Split(outputCurrent, "\n")

	return diffSlices(allResources, currentResources), nil
}

// deleteResources deletes the resources given.
func (ce CommandExecutor) deleteResources(resources []string) error {
	// Loop over and delete one by one
	for _, resource := range resources {
		args := []string{"delete", resource}
		_, err := ce.execCommand(args)
		if err != nil {
			fmt.Printf("Attempting to delete resource %v resulted in err %v", resource, err)
			return err
		}
	}
	return nil
}
