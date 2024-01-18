# DataLogger Operator

The `DataLogger` Custom Resource Definition (CRD) is a fictional CRD, that is used to create some other Kubernetes
resources within the cluster. We use [kubebuilder](https://book.kubebuilder.io/) to scaffold the operator and the CRD
definition.

## Assessment

You will find an existing operator in this repository. This operator will take a DataLogger CRD (like in the
[example](example.yaml)) and will create resources according to that.

Your task is to implement all missing reconcile methods to deploy all Kubernetes resources needed for a DataLogger.

You'll have to ...

1. implement the `Namespace` reconciler that will create a new namespace for each DataLogger CRD
2. implement the `Deployment` reconciler that will create/update a Deployment with the
   [`kennethreitz/httpbin`](https://github.com/postmanlabs/httpbin) image
3. implement the `Service` reconciler, that allows connections to our Deployment
4. make the used Port configurable by the DataLogger CRD by adding a `Port` field to the CRD (Hint: you'll have to
   update the CRD definition in the go code and do some stuff using the kubebuilder tools)
5. implement the `finalizer` Method in the datalogger reconciler, as we want to make sure, that all resources are
   deleted when the DataLogger CRD is deleted using the Kubernetes Garbage Collector
6. find and implement a sufficient way to test your reconcilers. You should have at least a 70%
   coverage of your written code. Writing tests for code not written by you is optional.

## Development

### Prerequisites

In order to solve the assessment, we advise the following tools to be available
on your local client:

- [Go](https://golang.org/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubebuilder](https://book.kubebuilder.io/quick-start#installation)

### Initialize a kind cluster

```bash
kind create cluster --name datalogger-operator
```

When the kind cluster is available, you can install the CRD definitions using our make target:

```bash
make install
```

This command will generate the CRD manifests from our go-code in `api/v1/*.go` and install them in the cluster.

### Run the operator

As we currently do not have a docker image for the operator, you can run the operator locally. The running operator
will use your current kubeconfig to connect to your local running kind-cluster

```bash
make run
```

### Create a DataLogger

We have an example DataLogger in this directory which can be applied to the cluster:

```bash
kubectl apply -f example.yaml
```

Run that command while your operator instance is running, and you should see that the operator creates the necessary
resources using the reconcile loops.

### Run the tests

We use a make target to run the tests:

```bash
make test
```
