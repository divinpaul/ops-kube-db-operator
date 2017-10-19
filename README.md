# rds-operator

Operator to control RDS DBs in amazon

## Installation

```bash
glide install
```

## Running from source

* Set required AWS configuration via environment variables (for now)

```
DB_SECURITYGROUPID=sg-2610f740
DB_SUBNETGROUPNAME=test-rds-dataform
DB_ENCRYPTIONKEYARN=arn:aws:kms:ap-southeast-2:12345678901:key/123456789-2f82-467f-b61e-123456789a

export DB_SECURITYGROUPID DB_SUBNETGROUPNAME DB_ENCRYPTIONKEYARN
```
* authenticate to kubes
* authenticate to AWS 
* run it locally
```bash
go run *.go -kubeconfig ~/.kube/config
```

## Auto Generating Client with Kubernetes code-generator

- Make sure to `go get -d k8s.io/code-generator`
- If that is failing, try to get it from github.com/sttts/code-generator

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
