# Kubernetes RDS Operator

[![Build Status](https://travis-ci.org/MYOB-Technology/ops-kube-db-operator.svg?branch=master)](https://travis-ci.org/MYOB-Technology/ops-kube-db-operator)

Operator to control RDS DBs in AWS, uses Config Maps for dafault configuration and Secrets for DB parameters.

## Installation

To install the controller in your cluster make sure to apply the CRD first and then create a deployment with the appropriate images:

```bash
# apply the crd
❯ kubectl apply -f yaml/crd.yaml
# apply the settings config-map (make sure to edit it with settings to suit you)
❯ kubectl apply -f yaml/config-map.yaml
# now create a deployment
❯ kubectl apply -f yaml/deployment.yaml
```

## Usage

Once the controller is running, users can create RDS Postgres DBs with the following yml:

```bash
❯ cat db.yml
apiVersion: myob.com/v1alpha1
kind: PostgresDB
metadata:
  name: example-db
  namespace: my-namespace
spec:
  size: "db.t2.small"
  storage: "10"
  iops: "1000"

❯ kubectl apply -f db.yml
```

Once this yml is applied, an RDS instance will be created. Note that it takes up to 10 minutes for RDS Instances to be ready so to check the status of the instance the user can check the `.status.ready` field on the resource:

```bash
❯ kubectl get postgresdb example-db -o yaml
apiVersion: myob.com/v1alpha1
kind: PostgresDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"myob.com/v1alpha1","kind":"PostgresDB","metadata":{"annotations":{},"name":"example-db","namespace":"kube-system"},"spec":{"size":"db.t2.small","storage":"10"}}
  clusterName: ""
  creationTimestamp: 2017-11-13T04:15:34Z
  generation: 0
  name: example-db
  namespace: kube-system
  resourceVersion: "29305859"
  selfLink: /apis/myob.com/v1alpha1/namespaces/kube-system/postgresdbs/example-db
  uid: 4b5c5df7-c829-11e7-9341-06163b58e928
spec:
  size: db.t2.small
  storage: 10
status:
  arn: arn:aws:rds:ap-southeast-2:693429498512:db:example-db-4b5c5df7-c829-11e7-9341-06163b58e928
  ready: available

# or more directly
❯ kubectl get postgresdb example-db -o go-template='{{.status.ready}}'
available

# The credentials to the DB can be found in kubernetes secrets which will be created for you
# note that the values are base64 encoded.
❯ kubectl get secrets example-db -o yaml
apiVersion: v1
data:
  dbname: dGVzdC1leGFtcGxlLWRiLTItM2QyZWYwYjMtYzhjOC0xMWU3LWI4OGItMDJhNGU3Nzc5MWI0
  endpoint: dGVzdC1leGFtcGxlLWRiLTItM2QyZWYwYjMtYzhjOC0xMWU3LWI4OGItMDJhNGU3Nzc5MWI0LmNidWp2Y2R5MGh3aC5hcC1zb3V0aGVhc3QtMi5yZHMuYW1hem9uYXdzLmNvbTo1NDMy
  password: Qk9PdFRzPVpLbWNSKnJRWEZDKmcoaFklNE92cEhiTlQ=
  username: cG5odmhqa2FtaXlldW5ydA==
kind: Secret
metadata:
  creationTimestamp: 2017-11-13T23:13:20Z
  name: example-db
  namespace: kube-system
  resourceVersion: "29307136"
  selfLink: /api/v1/namespaces/kube-system/secrets/example-db
  uid: 3d36c40d-c8c8-11e7-b88b-02a4e77791b4
type: Opaque
```

## Running from source

* Authenticate to Kubernetes
* Authenticate to AWS
* Set required AWS config in the configmap

```bash
❯ kubectl apply -f yaml/config-map.yaml
```

* Install dependencies with glide

```bash
❯ glide install
```

* run it

```bash
❯ go run *.go -kubeconfig ~/.kube/config
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
