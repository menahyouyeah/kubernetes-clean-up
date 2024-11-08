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
func (ce CommandExecutor) getOldResources(namespace, resourceTypeFlag string) ([]string, error) {
	// Step 1. Get a list of resource types to query
	resourceTypes, err := ce.resourceTypesToQuery(resourceTypeFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to get a list of resources types to query, err: %w", err)
	}

	// Get a list of all resources on the cluster.
	allResources, err := ce.getResources(false, namespace, resourceTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get a list of resources on the cluster, err: %w", err)
	}

	// Get a list of resources that were deployed as part of the latest release on the cluster.
	currentResources, err := ce.getResources(true, namespace, resourceTypes)
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
func kubectlGetArgs(includeReleaseLabel bool, resourceType string, nspace string) []string {
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
		fmt.Sprintf("--namespace=%s", nspace),
		resourceType,
	}
}

// get kubernetes resources
func (ce CommandExecutor) getResources(includeReleaseLabel bool, namespace string, resourceTypes []string) ([]string, error) {

	var resources []string
	for _, r := range resourceTypes {
		beep, err := ce.getResourcesPerType(includeReleaseLabel, namespace, r)
		if err != nil {
			return nil, fmt.Errorf("attempting to get resource type \"%v\" resulted in err: %w", r, err)
		}
		resources = append(resources, beep...)
	}
	return resources, nil
}

func (ce CommandExecutor) resourceTypesToQuery(resourceType string) ([]string, error) {
	var resourceTypes []string
	if resourceType != "" {
		resourceTypes = strings.Split(resourceType, ",")
	} else {
		apiResourcesArgs := apiResourcesQueryArgs()
		output, err := ce.execCommand(apiResourcesArgs)
		if err != nil {
			return nil, fmt.Errorf("failed to execute kubectl api-resources command: %w", err)
		}
		temp := strings.Split(output, "\n")
		// Delete the empty line at the end
		resourceTypes = slices.DeleteFunc(temp, func(e string) bool {
			return e == ""
		})
	}
	return resourceTypes, nil
}

func (ce CommandExecutor) getResourcesPerType(includeReleaseLabel bool, namespace string, resourceType string) ([]string, error) {
	var resources []string
	namespaces := strings.Split(namespace, ",")
	for _, n := range namespaces {
		args := kubectlGetArgs(includeReleaseLabel, resourceType, n)
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
