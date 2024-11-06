package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	cloudDeployPrefix = "deploy.cloud.google.com/"
	releaseEnvKey     = "CLOUD_DEPLOY_RELEASE"
	projectEnvKey     = "CLOUD_DEPLOY_PROJECT"
	locationEnvKey    = "CLOUD_DEPLOY_LOCATION"
	pipelineEnvKey    = "CLOUD_DEPLOY_DELIVERY_PIPELINE"
	targetEnvKey      = "CLOUD_DEPLOY_TARGET"
	outputNameArg     = "-o name"
)

// Get a list of resources that aren't in the current set of resources.
func (ce CommandExecutor) getOldResources() ([]string, error) {
	// Get a list of all resources on the cluster.
	allResources, err := ce.getResources(false)
	if err != nil {
		return nil, fmt.Errorf("TESTING failed to get a list of resources on the cluster, err: %w", err)
	}

	// Get a list of resources that were deployed as part of the latest release on the cluster.
	currentResources, err := ce.getResources(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get a list of current resources on the cluster, err: %w", err)
	}

	return diffSlices(allResources, currentResources), nil
}

func apiResourcesQueryArgs() []string {
	return []string{
		"api-resources",
		"--verbs=list",
		"-o",
		"name",
	}
}

// kubectlGetArgs returns the get args.
func kubectlGetArgs(includeReleaseLabel bool, resourceType string) []string {
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

	return []string{
		"get",
		"-o",
		"name",
		labelArg,
		resourceType,
	}
}

// get kubernetes resources
func (ce CommandExecutor) getResources(includeReleaseLabel bool) ([]string, error) {
	apiResourcesArgs := apiResourcesQueryArgs()
	// Not sure what the output looks like...long list? separated by new lines?
	output, err := ce.execCommand(apiResourcesArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to execute kubectl api-resources command: %w", err)
	}
	temp := strings.Split(output, "\n")
	// Delete the empty line at the end
	boop := slices.DeleteFunc(temp, func(e string) bool {
		return e == ""
	})

	// Loop over and get resources
	// What if multiple resources
	// What if not found
	var resources []string
	for _, resourceType := range boop {
		args := kubectlGetArgs(includeReleaseLabel, resourceType)
		output, err := ce.execCommand(args)
		if err != nil {
			return nil, fmt.Errorf("attempting to get resource type \"%v\" resulted in err: %w", resourceType, err)
		}
		if output != "" {
			// Separate out by line break
			temp := strings.Split(output, "\n")
			slices.DeleteFunc(temp, func(e string) bool {
				return e == ""
			})
			resources = append(resources, temp...)
		}
	}
	return resources, nil
}

// deleteResources deletes the resources given.
func (ce CommandExecutor) DeleteResources(resources []string) error {
	// Loop over and delete one by one
	fmt.Printf("Beginning to delete resources, there are %d resources to delete\n", len(resources))
	for _, resource := range resources {
		args := []string{"delete", resource}
		_, err := ce.execCommand(args)
		if err != nil {
			return fmt.Errorf("attempting to delete resource %v resulted in err: %w", resource, err)
		}
	}
	return nil
}

// // Get a list of resources that aren't in the current set of resources.
// func (ce CommandExecutor) getOldResources() ([]string, error) {
// 	// Get a list of all resources on the cluster.
// 	allResources, err := ce.getResources(false)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get a list of resources that were deployed as part of the latest release on the cluster.
// 	currentResources, err := ce.getResources(true)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return diffSlices(allResources, currentResources), nil
// }

// // deleteResources deletes the resources given.
// func (ce CommandExecutor) deleteResources(resources []string) error {
// 	// Loop over and delete one by one
// 	for _, resource := range resources {
// 		args := []string{"delete", resource}
// 		_, err := ce.execCommand(args)
// 		if err != nil {
// 			fmt.Printf("Attempting to delete resource %v resulted in err %v", resource, err)
// 			return err
// 		}
// 	}
// 	return nil
// }

// func apiResourcesQueryArgs() []string {
// 	return []string{
// 		"api-resources",
// 		"--verbs=list",
// 		outputNameArg,
// 	}
// }

// // kubectlGetArgs returns the get args.
// func kubectlGetArgs(includeReleaseLabel bool) []string {
// 	var labels []string
// 	if includeReleaseLabel {
// 		labels = append(labels, fmt.Sprintf("%srelease-id=%s", cloudDeployPrefix, os.Getenv(releaseEnvKey)))
// 	}
// 	labels = append(labels, fmt.Sprintf("%sdelivery-pipeline-id=%s", cloudDeployPrefix, os.Getenv(pipelineEnvKey)))
// 	labels = append(labels, fmt.Sprintf("%starget-id=%s", cloudDeployPrefix, os.Getenv(targetEnvKey)))
// 	labels = append(labels, fmt.Sprintf("%slocation=%s", cloudDeployPrefix, os.Getenv(locationEnvKey)))
// 	labels = append(labels, fmt.Sprintf("%sproject-id=%s", cloudDeployPrefix, os.Getenv(projectEnvKey)))

// 	labelsFormatted := strings.Join(labels, ",")
// 	labelArg := fmt.Sprintf("-l %s", labelsFormatted)

// 	return []string{
// 		"get",
// 		outputNameArg,
// 		labelArg,
// 	}
// }

// // get kubernetes resources
// func (ce CommandExecutor) getResources(includeReleaseLabel bool) ([]string, error) {
// 	apiResources := apiResourcesQueryArgs()
// 	// Execute the API resources command to get the list of supported API resources
// 	apiResourceCmd := exec.Command(ce.binPath, apiResources...)
// 	// Get the API resource command's stdout and attach it to the kubectl get commands stdin.
// 	apiOutputPipe, err := apiResourceCmd.StdoutPipe()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer apiOutputPipe.Close()

// 	// kubectl get command
// 	args := kubectlGetArgs(includeReleaseLabel)
// 	kubectlGetCmd := exec.Command(ce.binPath, args...)
// 	kubectlGetCmd.Stdin = apiOutputPipe

// 	// Now run the api resource command
// 	apiResourceCmd.Start()

// 	// Run and get the output of kubectl get
// 	result, err := kubectlGetCmd.Output()
// 	if err != nil {
// 		return nil, err
// 	}
// 	output := string(result)
// 	var resources []string
// 	if output != "" {
// 		// Separate out by line break
// 		temp := strings.Split(output, "\n")
// 		resources = append(resources, temp...)
// 		// for _, r := range temp {
// 		// 	resources = append(resources, r)
// 		// }
// 	}
// 	// Loop over and get resources
// 	// What if multiple resources
// 	// What if not found
// 	// for _, resource := range apiResources {
// 	// 	args := kubectlGetArgs(includeReleaseLabel, resource)
// 	// 	output, err := ce.execCommand(args)
// 	// 	if err != nil {
// 	// 		fmt.Printf("Attempting to get resource %v resulted in err %v", resource, err)
// 	// 		return nil, err
// 	// 	}
// 	// 	if output != "" {
// 	// 		// Separate out by line break
// 	// 		temp := strings.Split(output, "\n")
// 	// 	}
// 	// 	resources = append(resources, output)
// 	// }
// 	return resources, nil
// }

// // CommandExecutor contains command execution information.
// // type CommandExecutor struct {
// // 	// BinPath is the path the binary being used for the command (e.g. the path
// // 	// to the kubectl binary if the kubectl command is to be used).
// // 	binPath string
// // }

// // CreateCommandExecutor returns a CommandExecutor for the given binary
// // func CreateCommandExecutor(binPath string) *CommandExecutor {
// // 	ce := &CommandExecutor{
// // 		binPath: binPath,
// // 	}
// // 	return ce
// // }

// // execCommand runs the given command and returns the output.
// func (ce CommandExecutor) execCommand(args []string) (string, error) {
// 	fmt.Printf("Running the following command: %s %s\n", ce.binPath, args)
// 	cmd := exec.Command(ce.binPath, args...)
// 	// By default set locations to standard error and output (visible in cloud build logs)
// 	cmd.Stderr = os.Stderr
// 	cmd.Stdout = os.Stdout

// 	// Write error output to two locations simultaneously. Unless disabled by a CommandOption, this will allow
// 	// us to see the error output as the execution is happening (by writing to standard error) and also allow us
// 	// to gather all stderr at the end (by also writing to var stderr).
// 	var stderr bytes.Buffer
// 	errWriter := io.MultiWriter(&stderr, cmd.Stderr)
// 	cmd.Stderr = errWriter

// 	// Write output to two locations simultaneously. Unless disabled by a CommandOption, this will allow
// 	// us to see the output as the execution is happening (by writing to standard out) and also allow us
// 	// to gather all stdout at the end (by also writing to var stdout).
// 	var stdout bytes.Buffer
// 	outWriter := io.MultiWriter(&stdout, cmd.Stdout)
// 	cmd.Stdout = outWriter

// 	// Start the command
// 	if err := cmd.Start(); err != nil {
// 		return "", fmt.Errorf("failed to start command, err is %w", err)
// 	}

// 	// Wait for everything to finish.
// 	if err := cmd.Wait(); err != nil {
// 		// Read the stdErr output
// 		errorOutput := stderr.Bytes()
// 		fullErr := fmt.Errorf("error running command: %w\n%s", err, errorOutput)
// 		return "", fullErr
// 	}
// 	return stdout.String(), nil
// }

// // diffSlices returns the elements in slice1 that aren't in slice2.
// func diffSlices(slice1, slice2 []string) []string {
// 	var diff []string
// 	set := make(map[string]bool)

// 	for _, val := range slice2 {
// 		set[val] = true
// 	}

// 	for _, val := range slice1 {
// 		// If they're not in slice 2
// 		if !set[val] {
// 			diff = append(diff, val)
// 		}
// 	}

// 	return diff
// }

// TRASH
// get kubernetes resources
// func (ce CommandExecutor) getResources(includeReleaseLabel bool) ([]string, error) {
// 	apiResources := apiResourcesQueryArgs()
// 	// Execute the API resources command to get the list of supported API resources
// 	apiResourceCmd := exec.Command(ce.binPath, apiResources...)
// 	// Get the API resource command's stdout and attach it to the kubectl get commands stdin.
// 	apiOutputPipe, err := apiResourceCmd.StdoutPipe()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get stdout pipe from kubectl api-resources command: %w", err)
// 	}
// 	defer apiOutputPipe.Close()

// 	// kubectl get command
// 	args := kubectlGetArgs(includeReleaseLabel)
// 	kubectlGetCmd := exec.Command(ce.binPath, args...)
// 	kubectlGetCmd.Stdin = apiOutputPipe

// 	// Write error output to multiple places for api command
// 	var stderr bytes.Buffer
// 	errWriter := io.MultiWriter(&stderr, apiResourceCmd.Stderr)
// 	apiResourceCmd.Stderr = errWriter

// 	// Write error output to multiple places for api command
// 	var stderr2 bytes.Buffer
// 	errWriter2 := io.MultiWriter(&stderr2, kubectlGetCmd.Stderr)
// 	kubectlGetCmd.Stderr = errWriter2

// 	// Now run the api resource command
// 	if err := apiResourceCmd.Start(); err != nil {
// 		return nil, fmt.Errorf("failed to execute kubectl api-resources command: %w", err)
// 	}

// 	// Run and get the output of kubectl get
// 	result, err := kubectlGetCmd.Output()
// 	output := string(result)
// 	if err != nil {
// 		errorOutput := stderr2.Bytes()

// 		return nil, fmt.Errorf("failed to execute kubectl get command: %w, output: %v, full error: %s", err, output, errorOutput)
// 	}

// 	var resources []string
// 	if output != "" {
// 		// Separate out by line break
// 		temp := strings.Split(output, "\n")
// 		resources = append(resources, temp...)
// 		// for _, r := range temp {
// 		// 	resources = append(resources, r)
// 		// }
// 	}
// 	// Loop over and get resources
// 	// What if multiple resources
// 	// What if not found
// 	// for _, resource := range apiResources {
// 	// 	args := kubectlGetArgs(includeReleaseLabel, resource)
// 	// 	output, err := ce.execCommand(args)
// 	// 	if err != nil {
// 	// 		fmt.Printf("Attempting to get resource %v resulted in err %v", resource, err)
// 	// 		return nil, err
// 	// 	}
// 	// 	if output != "" {
// 	// 		// Separate out by line break
// 	// 		temp := strings.Split(output, "\n")
// 	// 	}
// 	// 	resources = append(resources, output)
// 	// }
// 	return resources, nil
// }
