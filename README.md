# s3ops

A kubernetes operator to manage S3 resources.

## Description

### Supported resources

- S3Server: The connection to an existing s3 backend
  - Actions: Check online status
- S3Bucket: A bucket in the s3 backend
  - Actions: Create, Delete
- S3Policy: A policy that can be attached to users
  - Actions: Create, Update, Delete
- S3User: A user with access to the s3 backend
  - Actions: Create, Delete, AttachPolicy

### Supported Backends

| Backend | S3Server | S3Bucket | S3Policy | S3User |
|---------|----------|----------|----------|--------|
| Minio   | Yes      | Yes      | Yes      | Yes    |

## Deploy the operator

### Prerequisites

- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Deploy the full operator

> This includes the CRDs, RBAC, and the controller itself.

```sh
kubectl apply -f https://raw.githubusercontent.com/odit-services/s3ops/main/config/deployment/full.yaml
```

## Getting started with the development

### Local prerequisites

- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
# Single Arch
make docker-build docker-push IMG=ghcr.io/odit-services/s3ops:tag

# Multi Arch
make docker-build-multiarch IMG=ghcr.io/odit-services/s3ops:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=ghcr.io/odit-services/s3ops:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=ghcr.io/odit-services/s3ops:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/s3ops/<tag or branch>/dist/install.yaml
```

## Contributing

Feel free to contribute to this project by following the steps below:

1. Fork the repository
2. Create a new branch (git checkout -b feat/some-feature)
3. Make changes
4. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin feat/some-feature)
6. Create a new Pull Request

All new features and bug fixes should have associated tests.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
