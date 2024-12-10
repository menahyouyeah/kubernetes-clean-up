package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
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

const (
	// The name of the post-deploy hook cleanup sample, this is passed back to
	// Cloud Deploy as metadata in the deploy results, mainly to keep track of
	// how many times the sample is getting used.
	cleanupSampleName = "clouddeploy-cleanup-sample"
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
	// Step 1. Run gcloud get-credentials to set up the cluster credentials.
	gkeCluster := os.Getenv("GKE_CLUSTER")
	if err := gcloudClusterCredentials(gkeCluster); err != nil {
		return err
	}

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

// PostDeployHookResult represents the json data in the results file for a
// post deploy hook operation.
type PostDeployHookResult struct {
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UploadResult uploads the provided deploy result to the Cloud Storage path where Cloud Deploy expects it.
// Returns the Cloud Storage URI of the uploaded result.
// CLOUD_DEPLOY_OUTPUT_GCS_PATH
func (d *PostDeployHookResult) UploadResult(ctx context.Context, gcsClient *storage.Client, deployHookResult *PostDeployHookResult) (string, error) {
	uri := os.Getenv("CLOUD_DEPLOY_OUTPUT_GCS_PATH")
	jsonResult, err := json.Marshal(deployHookResult)
	if err != nil {
		return "", fmt.Errorf("error marshalling post deploy hook result: %v", err)
	}
	if err := uploadGCS(ctx, gcsClient, uri, &GCSUploadContent{Data: res}); err != nil {
		return "", err
	}
	return uri, nil
}

// uploadGCS uploads the provided content to the specified Cloud Storage URI.
func uploadGCS(ctx context.Context, gcsClient *storage.Client, gcsURI string, content []byte) error {

	gcsObjURI, err := parseGCSURI(gcsURI)
	if err != nil {
		return err
	}
	w := gcsClient.Bucket(gcsObjURI.bucket).Object(gcsObjURI.name).NewWriter(ctx)
	if _, err := w.Write(content); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// parseGCSURI parses the Cloud Storage URI and returns the corresponding gcsObjectURI.
func parseGCSURI(uri string) (gcsObjectURI, error) {
	var obj gcsObjectURI
	u, err := url.Parse(uri)
	if err != nil {
		return gcsObjectURI{}, fmt.Errorf("cannot parse URI %q: %w", uri, err)
	}
	if u.Scheme != "gs" {
		return gcsObjectURI{}, fmt.Errorf("URI scheme is %q, must be 'gs'", u.Scheme)
	}
	if u.Host == "" {
		return gcsObjectURI{}, errors.New("bucket name is empty")
	}
	obj.bucket = u.Host
	obj.name = strings.TrimLeft(u.Path, "/")
	if obj.name == "" {
		return gcsObjectURI{}, errors.New("object name is empty")
	}
	return obj, nil
}

// gcsObjectURI is used to split the object Cloud Storage URI into the bucket and name.
type gcsObjectURI struct {
	// bucket the GCS object is in.
	bucket string
	// name of the GCS object.
	name string
}
