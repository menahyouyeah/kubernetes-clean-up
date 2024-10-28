package main

import (
	"os"
	"testing"

	// "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp"
)

func TestCreateQueryArgs(t *testing.T) {
	os.Setenv(releaseEnvKey, "myrelease")
	os.Setenv(pipelineEnvKey, "mypipeline1")
	os.Setenv(targetEnvKey, "mytarget")
	os.Setenv(projectEnvKey, "myproject")
	os.Setenv(locationEnvKey, "losangeles")

	for _, tc := range []struct {
		name                string
		includeReleaseLabel bool
		wantArgs            []string
	}{
		{
			name:                "no release",
			includeReleaseLabel: false,
			wantArgs: []string{
				"api-resources",
				"verbs=list",
				"-o name",
				"-l deploy.cloud.google.com/delivery-pipeline-id=mypipeline,deploy.cloud.google.com/target-id=mytarget,deploy.cloud.google.com/location=losangeles,deploy.cloud.google.com/project-id=my-project,",
				"|",
				"xargs",
				"-n 1",
				"kubectl",
				"get",
				"-o name",
			},
		},
		{
			name:                "with release",
			includeReleaseLabel: true,
			wantArgs: []string{
				"api-resources",
				"verbs=list",
				"-o name",
				"-l deploy.cloud.google.com/release-id=myrelease,deploy.cloud.google.com/delivery-pipeline-id=mypipeline,deploy.cloud.google.com/target-id=mytarget,deploy.cloud.google.com/location=losangeles,deploy.cloud.google.com/project-id=my-project,",
				"|",
				"xargs",
				"-n 1",
				"kubectl",
				"get",
				"-o name",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotArgs := createQueryArgs(tc.includeReleaseLabel)
			if diff := cmp.Diff(tc.wantArgs, gotArgs); diff != "" {
				t.Errorf("createQueryArgs() produced diff (-want, +got):\n%s", diff)
			}
		})
	}
}
