package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	namespace = flag.String("namespace", "", "Namespace(s) to filter on when finding resources to delete. "+
		"For multiple namespaces, separate them with a comma. For example --namespace=foo,bar.")
	resourceType = flag.String("resource-type", "", "Resource type(s) to filter on when finding resources to delete. "+
		"If listing multiple resource types, separate them with commas. For example, --resource-type=Deployment,Job. "+
		"You can also qualify the resource type by an API group if you want to specify resources only in a specific "+
		"API group. For example --resource-type=deployments.apps")
)

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
	// Step 1. Get a list of resources to delete.
	kubectlExec := CreateCommandExecutor("kubectl")
	oldResources, err := kubectlExec.resourcesToDelete(*namespace, *resourceType)
	if err != nil {
		return err
	}

	// Step 2. Delete the resources.
	if err := kubectlExec.deleteResources(oldResources); err != nil {
		return err
	}

	return nil
}
