# splunk-operator

This repository is used to build the [Kubernetes operator](https://coreos.com/operators/) for Splunk.

## Vendor Dependencies

This project uses [dep](https://github.com/golang/dep) to manage dependencies. On MacOS, you can install `dep` using Homebrew:

```
$ brew install dep
$ brew upgrade dep
```

On other platforms you can use the `install.sh` script:

```
$ curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```


## Kubernetes Operator SDK

The Kubernetes [Operator SDK](https://github.com/operator-framework/operator-sdk) must also be installed to build this project.

```
$ mkdir -p $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk
$ cd operator-sdk
$ git checkout master
$ make dep
$ make install
```


## Cloning this repository

This repository should be cloned into your `~/go/src/git.splunk.com` directory:
```
$ mkdir -p ~/go/src/git.splunk.com
$ cd ~/go/src/git.splunk.com
$ git clone ssh://git@git.splunk.com:7999/tools/splunk-operator.git
$ cd splunk-operator
```


## Building the operator

You can build the operator by just running `make splunk-operator`.

Other make targets include:

* `make all`: builds the splunk-operator, splunk-dfs and splunk-spark docker images (requires splunk-debian-9)
* `make dep`: checks all vendor dependencies and ensures they are up to date
* `make splunk-operator`: builds the `splunk-operator` docker image
* `make splunk-dfs`: builds the `splunk-dfs` docker image
* `make splunk-spark`: builds the `splunk-spark` docker image
* `make push-operator`: pushes the `splunk-operator` docker image to all `push_targets`
* `make push-dfs`: pushes the `splunk-dfs` docker image to all `push_targets`
* `make push-spark`: pushes the `splunk-spark` docker image to all `push_targets`
* `make push`: pushes all docker images to all `push_targets`
* `make publish-repo`: publishes the `splunk-operator` docker image to repo.splunk.com
* `make publish-playground`: publishes the `splunk-operator` docker image to cloudrepo-docker-playground.jfrog.io
* `make publish`: publishes the `splunk-operator` docker image to all registries
* `make install`: installs required resources in current k8s target cluster
* `make uninstall`: removes required resources from current k8s target cluster
* `make start`: starts splunk operator in current k8s target cluster
* `make stop`: stops splunk operator (including all instances) in current k8s target cluster
* `make rebuild`: rebuilds and reinstalls splunk operator in current k8s target cluster


## Installing Required Resources

The Splunk operator requires that your k8s target cluster have certain resources. These can be installed by running
`make install`.

You can later remove these resources by running
`make uninstall`.


## Building Docker Images

The Splunk operator requires three docker images to be present or available to your k8s target cluster:

* `splunk/splunk`: The default Splunk Enterprise image (available from Docker Hub)
* `splunk-dfs`: The default Splunk Enterprise image used when DFS is enabled
* `splunk-spark`: The default Spark image used by DFS, when enabled
* `splunk-operator`: The Splunk operator image

Building the `splunk-dfs` image requires that you first clone the [docker-splunk repository](https://github.com/splunk/docker-splunk)
and build a `splunk-debian-9` base image using a DFS build of Splunk Enterprise:

```
$ git clone git@github.com:splunk/docker-splunk.git
$ cd docker-splunk
$ make SPLUNK_LINUX_BUILD_URL=http://releases.splunk.com/dl/epic-dfs_builds/7.3.0-20181205-0400/splunk-7.3.0-ce483f77eb7f-Linux-x86_64.tgz SPLUNK_LINUX_FILENAME=splunk-7.3.0-ce483f77eb7f-Linux-x86_64.tgz splunk-debian-9

```
After you have built `splunk-debian-9` you can build build `splunk-dfs` by running `make splunk-dfs`.

The `splunk-spark` image can be built by running `make splunk-spark`.

The `splunk-operator` image can be built by running `make splunk-operator`.


## Publishing and Pushing Docker Images

The Splunk operator requires that your k8s cluster has access to the Docker images listed above.


### Local Clusters

If you are using a local single-node k8s cluster, you only need to build the images. You can skip this section.


### UCP Clusters

* `splunk/splunk`: available from Docker Hub
* `splunk-dfs`: available via `repo.splunk.com`
* `splunk-spark`: available via `repo.splunk.com`

The `splunk-operator` image is published and available via `repo.splunk.com`. You can publish a new local build by
running `make publish-repo`. This will tag the image as `repo.splunk.com/splunk/products/splunk-operator:[COMMIT_ID]`.


### Splunk8s Clusters

* `splunk/splunk`: available from Docker Hub
* `splunk-dfs`: available via `cloudrepo-docker-playground.jfrog.io`
* `splunk-spark`: available via `cloudrepo-docker-playground.jfrog.io`

The `splunk-operator` image is published and available via `cloudrepo-docker-playground.jfrog.io`. You can publish a new
local build by running `make publish-repo`. This will tag the image as
`cloudrepo-docker-playground.jfrog.io/pcp/splunk-operator:[COMMIT_ID]`.


### Other Clusters

You can still run the Splunk operator on clusters that do not have access to Splunk's Docker Registries.
THIS IS PRE-RELEASE SOFTWARE, SO PLEASE BE VERY CAREFUL NOT TO PUBLISH THESE IMAGES TO ANY PUBLIC REGISTRIES.

Provided is a convenient `push` commands that you can use to upload the images directly to each worker node.
Create a file in the top level directory of this repository named `push_targets`. This file
should include every worker node in your K8s cluster, one user@host for each line. For example:

```
ubuntu@myvm1.splunk.com
ubuntu@myvm2.splunk.com
ubuntu@myvm3.splunk.com
```

You can push the `splunk-dfs` image to each of these nodes by running `make push-dfs`.

You can push the `splunk-spark` image to each of these nodes by running `make push-spark`.

You can push the `splunk-operator` image to each of these nodes by running `make push-operator`.

Or, just push all three using `make push`.


## Running the Splunk Operator


### UCP Clusters

You can start the operator by running
`kubectl create -f deploy/operator-repo.yaml`.

You can stop the operator by running
`kubectl create -f deploy/operator-repo.yaml`.


### Splunk8s Clusters

You can start the operator by running
`kubectl create -f deploy/operator-playground.yaml`.

You can stop the operator by running
`kubectl create -f deploy/operator-playground.yaml`.


### Local and Other Clusters

You can start the operator by running
`kubectl create -f deploy/operator.yaml` or `make start` for short.

You can stop the operator by running
`kubectl create -f deploy/operator.yaml` or `make stop` for short.


## Creating Splunk Enterprise Instances

To create a new Splunk Enterprise instance, run `kubectl create -f deploy/crds/enterprise_v1alpha1_splunkenterprise_cr.yaml`.

To remove the instance, run `kubectl delete -f deploy/crds/enterprise_v1alpha1_splunkenterprise_cr.yaml`


## Running Splunk Enterprise with DFS

To create a new Splunk Enterprise instance with DFS (including Spark), run `kubectl create -f deploy/crds/enterprise_v1alpha1_splunkenterprise_cr_dfs.yaml`.

To remove the instance, run `kubectl delete -f deploy/crds/enterprise_v1alpha1_splunkenterprise_cr_dfs.yaml`


## CR Spec

After creating the relevant resources on your Kubernetes cluster, you will now be able to create resources of type **SplunkInstance**

Here is a sample yaml file that can be used to create a **SplunkEnterprise**

```yaml
apiVersion: "enterprise.splunk.com/v1alpha1"
kind: "SplunkEnterprise"
metadata:
	name: "example"
spec:
	config:
		splunkPassword: helloworld
		splunkStartArgs: --accept-license
	topology:
		indexers: 1
		searchHeads: 1
```

### Relevant Parameters

#### Metadata
| Key       | Type   | Description                                                                                                       |
| --------- | ------ | ----------------------------------------------------------------------------------------------------------------- |
| name      | string | Your splunk deployments will be distinguished using this name.                                                    |
| namespace | string | Your splunk deployments will be created in this namespace. You must insure that this namespace exists beforehand. |

#### Spec
| Key                   | Type    | Description                                                                                                                                                           |
| --------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Config**            |         |                                                                                                                                                                       |
| splunkPassword        | string  | The password that can be used to login to splunk instances.                                                                                                           |
| splunkStartArgs       | string  | Arguments to launch each splunk instance with.                                                                                                                        |
| splunkImage           | string  | Docker image to use for Splunk instances (overrides SPLUNK_IMAGE and SPLUNK_DFS_IMAGE environment variables)                                                          |
| sparkImage            | string  | Docker image to use for Spark instances (overrides SPLUNK_SPARK_IMAGE environment variables)                                                                          |
| defaultsConfigMapName | string  | The name of the ConfigMap which stores the splunk defaults data.                                                                                                      |
| enableDFS             | bool    | If this is true, DFS will be installed on **searchHeads** being launched.                                                                                             |
| **Topology**          |         |                                                                                                                                                                       |
| standalones           | integer | The number of standalone instances to launch.                                                                                                                         |
| searchHeads           | integer | The number of search heads to launch. If this number is greater than 1 then a deployer will be launched as well to create a search head cluster.                      |
| indexers              | integer | The number of indexers to launch. When **searchHeads** is defined and **indexers** is defined a **cluster master** is also launched to create a clustered deployment. |
| sparkWorkers          | integer | The number of spark workers to launch. When this is defined, a **spark cluster master** will be launched as well to create a spark cluster.                           |

**Notes**
+ If **searchHeads** is defined then **indexers** must also be defined (and vice versa).
+ If **enableDFS** is defined then **sparkWorkers** must also be defined (and vice versa) or else a DFS search won't work.