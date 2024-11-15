package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

var (
	namespace = flag.String("namespace", "", "Namespace(s) to filter on when finding resources to delete. "+
		"For multiple namespaces, separate them with a comma. For example --namespace=foo,bar.")
	resourceType = flag.String("resource-type", "", "Resource type(s) to filter on when finding resources to delete. "+
		"If listing multiple resource types, separate them with commas. For example, --resource-type=Deployment,Job. "+
		"You can also qualify the resource type by an API group if you want to specify resources only in a specific "+
		"API group. For example --resource-type=deployments.apps")
)

// gkeClusterRegex represents the regex that a GKE cluster resource name needs to match.
var gkeClusterRegex = regexp.MustCompile("^projects/([^/]+)/locations/([^/]+)/clusters/([^/]+)$")

func main() {
	flag.Parse()

	if err := do(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
	os.Exit(0)
}

func do() error {
	// Step 1. Run gcloud get-credentials to set up the cluster credentials.
	gkeCluster := os.Getenv("GKE_CLUSTER")
	if err := gcloudClusterCredentials(gkeCluster); err != nil {
		return err
	}
	//gcloud services enable cloudresourcemanager.googleapis.com

	// Step 2. Get a list of resources to delete.
	kubectlExec := CreateCommandExecutor("kubectl")
	oldResources, err := kubectlExec.resourcesToDelete(*namespace, *resourceType)
	if err != nil {
		return err
	}

	// Step 3. Delete the resources.
	if err := kubectlExec.deleteResources(oldResources); err != nil {
		return err
	}

	return nil
}

// gcloudClusterCredentials runs `gcloud container clusters get-crendetials` to set up
// the cluster credentials.
func gcloudClusterCredentials(gkeCluster string) error {
	gcloudExec := CreateCommandExecutor("gcloud")
	m := gkeClusterRegex.FindStringSubmatch(gkeCluster)
	if len(m) == 0 {
		return fmt.Errorf("invalid GKE cluster name: %s", gkeCluster)
	}

	args := []string{"container", "clusters", "get-credentials", m[3], fmt.Sprintf("--region=%s", m[2]), fmt.Sprintf("--project=%s", m[1])}
	_, err := gcloudExec.execCommand(args)
	if err != nil {
		return fmt.Errorf("unable to set up cluster credentials: %w", err)
	}
	return nil
}
