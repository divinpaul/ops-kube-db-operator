#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO=github.com/MYOB-Technology/ops-kube-db-operator
HEADER_FILE="${GOPATH}/src/${REPO}/hack/boilerplate.go.txt"


mkdir -p ${GOPATH}/src/k8s.io/kubernetes/hack/boilerplate/
cp ${HEADER_FILE} "${GOPATH}/src/k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt"

pushd "${GOPATH}/src/k8s.io/code-generator"

./generate-groups.sh all \
    ${REPO}/pkg/client \
    ${REPO}/pkg/apis "postgresdb:v1alpha1"

popd

find pkg -type f -exec grep -l myob-technology {} \;|xargs sed -i '.backup' -e 's/myob-technology/MYOB-Technology/g'

find pkg/client/clientset/ -name *.backup -exec rm {} \;
