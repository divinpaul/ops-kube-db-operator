# Kubernetes RDS Operator

[![Build Status](https://travis-ci.org/MYOB-Technology/ops-kube-db-operator.svg?branch=master)](https://travis-ci.org/MYOB-Technology/ops-kube-db-operator)

Operator to control RDS DBs in AWS, uses Config Maps for dafault configuration and Secrets for DB parameters.

## Installation

```bash
glide install
```

## Running from source

* Set required AWS config in the configmap

```bash
 kubectl apply -f yaml/config-map.yaml
```

* authenticate to kubes
* authenticate to AWS
* run it locally

```bash
go run *.go -kubeconfig ~/.kube/config
```

## Auto Generating Client with Kubernetes code-generator

* Make sure to `go get -d k8s.io/code-generator`
* If that is failing, try to get it from github.com/sttts/code-generator

```bash
❯ cd $GOPATH/src/k8s.io/code-generator
❯ ./generate-groups.sh all github.com/MYOB-Technology/ops-kube-db-operator/pkg/client github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis "db:v1alpha1" --go-header-file ./hack/boilerplate.go.txt
Generating deepcopy funcs
Generating clientset for db:v1alpha1 at github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset
Generating listers for db:v1alpha1 at github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/listers
Generating informers for db:v1alpha1 at github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers
```

> If it is failing with the following, make sure to clone the latest version of github.com/kubernetes/gengo into the `vendor/k8s.io/gengo` folder
>
> ```bash
> cmd/client-gen/main.go:34:29: "k8s.io/code-generator/vendor/k8s.io/gengo/args".Default().WithoutDefaultFlagParsing undefined (type *"k8s.io/code-generator/vendor/k8s.io/gengo/args".GeneratorArgs has no field or method WithoutDefaultFlagParsing)
> ```

## ROADMAP

* Update name to fit nautical theme in Kubernetes, I propose `Stow` which means: `to store, or to put away`

* Add unit tests around controller
  * at the moment Dataform is tested which is the library that interacts with RDS however the controller side of things is not, we need to be able to test most of the interactions with the API Server so that we catch regressions and problems in future. Some examples can be found in [Atlassian/Smith](https://github.com/atlassian/smith/blob/9b053cff9f69b1a3c75d18c43d0673dcdc76e015/pkg/controller/controller_test.go)

* Investigate upgrade paths
  * Need to find out how to correctly go from v1alpha1 to v1beta1 to beta to release. Unfortunately there does not seem to be a lot of documentation in this area and #sig-api-machinery in Kubernetes' Slack has not been much help.

* For MVP there is no leader election set up so only 1 replica of the application can run at a time. Kubernetes has leader election functionality in their libraries so for resiliency there should be a way to run > 1.

* Reconciliation loop should be updated so that Instances without matching CRDs should be scheduled for deletion. At the moment the event loop deletes an instance when the event actually happens.

* Deletion process should not rely on Secret and should be self sufficient.
