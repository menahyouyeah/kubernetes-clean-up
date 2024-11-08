package main

import (
	"os"
	"testing"

	// "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp"
)

func TestCreateQueryArgs(t *testing.T) {
	os.Setenv(releaseEnvKey, "myrelease")
	os.Setenv(pipelineEnvKey, "mypipeline")
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
				"--verbs=list",
				"-o name",
				"|",
				"xargs",
				"-n 1",
				"kubectl",
				"get",
				"-o name",
				"-l deploy.cloud.google.com/delivery-pipeline-id=mypipeline,deploy.cloud.google.com/target-id=mytarget,deploy.cloud.google.com/location=losangeles,deploy.cloud.google.com/project-id=myproject",
			},
		},
		{
			name:                "with release",
			includeReleaseLabel: true,
			wantArgs: []string{
				"api-resources",
				"--verbs=list",
				"-o name",
				"|",
				"xargs",
				"-n 1",
				"kubectl",
				"get",
				"-o name",
				"-l deploy.cloud.google.com/release-id=myrelease,deploy.cloud.google.com/delivery-pipeline-id=mypipeline,deploy.cloud.google.com/target-id=mytarget,deploy.cloud.google.com/location=losangeles,deploy.cloud.google.com/project-id=myproject",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotArgs := kubectlGetArgs(tc.includeReleaseLabel)
			if diff := cmp.Diff(tc.wantArgs, gotArgs); diff != "" {
				t.Errorf("kubectlGetArgs() produced diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestDiffSlices(t *testing.T) {

	for _, tc := range []struct {
		name     string
		slice1   []string
		slice2   []string
		wantDiff []string
	}{
		{
			name:   "no diff",
			slice1: []string{"1", "2", "3"},
			slice2: []string{"1", "2", "3"},
		},
		{
			name:     "no overlap",
			slice1:   []string{"1", "2", "3"},
			slice2:   []string{"4", "5", "6"},
			wantDiff: []string{"1", "2", "3"},
		},
		{
			name:     "some overlap",
			slice1:   []string{"1", "2", "3", "4"},
			slice2:   []string{"1", "4", "5", "6"},
			wantDiff: []string{"2", "3"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotDiff := diffSlices(tc.slice1, tc.slice2)
			if diff := cmp.Diff(tc.wantDiff, gotDiff); diff != "" {
				t.Errorf("diffSlices() produced diff (-want, +got):\n%s", diff)
			}
		})
	}
}
