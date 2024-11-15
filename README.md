# Kubernetes Resource Clean Up
This contains a sample container that can be used to clean up Kubernetes
resources that were deployed by Cloud Deploy. It should be used as a post-deploy
hook (a configuration example is provided below).

At a high level, the sample image:
1. Gets a list of kubernetes resources that were deployed by Cloud Deploy in the 
   current release
2. Gets a list of all kubernetes resources that were deployed by Cloud Deploy
   on the cluster
3. Does a diff and deletes any resources that were not deployed as part of the
   current release (i.e. deletes all the old resources).

# Prerequisites
1. You will need to build the image and push it to a repository, accessible by
Cloud Build.
2. There is a sample clouddeploy.yaml, kubernetes.yaml and skaffold.yaml file in
the config-sample directory. Either use those and replace the %PROJECT_ID%, 
%REGION%, and %IMAGE% values or update your existing Cloud Deploy config
file to reference a post-deploy hook, and your Skaffold file to then reference
the image you built.

# Building and pushing the image to a repo
1. In the directory of this README, run the following command to build the image:

```
docker build --tag <REPO-TAG> . 
```

For example, if you're pushing to an Artifact Registry with:
* region=us-central1
* project=my-project
* docker repo=my-repo
* you'd like to name the image my-image

The command would look like this:

```
docker build --tag us-central1-docker.pkg.dev/my-project/my-repo/my-image .
```

2. After the build is complete, push the image to the repository:

```
docker push <REPO-PATH>
```

Sticking with the example above, the command would be:

```
docker push us-central1-docker.pkg.dev/minnah-easymode-testing/clean-up/anothertest
```

# Update your config or use the sample configs

Within the `config-sample` folder, there are sample YAMLs.
1. `clouddeploy.yaml`: Defines a single [GKE Target](https://cloud.google.com/deploy/docs/deploy-app-gke).
Defines a Delivery Pipeline that references that GKE Target and specifies a post-deploy action `postdeploy-action`.
1. `kubernetes.yaml`: Defines an Deployment and Service that will be applied to the cluster.
1. `skaffold.yaml`: Defines a custom action `postdeploy-action` which is referenced in the clouddeploy.yaml. 
Within that customAction stanza there is a reference to the image that was
built above. 

# Register your pipeline and targets with Cloud Deploy

```
gcloud deploy apply --file=clouddeploy.yaml --region=REGION --project=PROJECT_ID
```

# Create a release and at the end the post-deploy hook will run

Create a release and after the release has been deployed to the cluster, the
post-deploy hook will run and delete any old resources that were previously 
deploy by Cloud Deploy. If you're using the sample, the command would look 
something like this:

```
gcloud deploy releases create my-release --project=PROJECT_ID --region=REGION --delivery-pipeline=my-pipeline    --images=my-app-image=gcr.io/google-containers/nginx@sha256:f49a843c290594dcf4d193535d1f4ba8af7d56cea2cf79d1e9554f077f1e7aaa
```

Note: Unless you've used Cloud Deploy before to deploy to that cluster, nothing
will be deleted. If create a second release `my-release2`, then the post-deploy
hook will actually do something and delete any resources that were deployed as
part of `my-release`. 