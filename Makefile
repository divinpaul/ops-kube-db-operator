.PHONY: publish publish-version login test clean teardown

REPO := myobplatform/ops-kube-db-operator
PKG_DIR = github.com/MYOB-Technology/ops-kube-db-operator/pkg
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

build:
	@docker-compose build base

ci: build vendor test

clean:
	docker-compose run --rm base rm -rf vendor

test:
	@docker-compose run --rm go test ./...

vendor:
	docker-compose run --rm dep ensure -v

# gen-mocks:
# 	@docker-compose run --rm go generate ./...
#
# # Share in vendor directory to ensure libraries are available for local development
# vendor-dev:
# 	docker-compose run --rm dep-dev ensure -v
#
#
# ifdef DOCKERHUB_PASSWORD
# login:
# 	@docker login --username ${DOCKERHUB_USERNAME} --password ${DOCKERHUB_PASSWORD}
# endif
#
# publish: publish-version
# 	docker build --build-arg VERSION=latest -t ${REPO}:latest .
# 	docker push ${REPO}:latest
#
# ifdef TRAVIS_TAG
# publish-version:
# 	docker build --build-arg VERSION=${TRAVIS_TAG} -t ${REPO}:${TRAVIS_TAG} .
# 	docker push ${REPO}:${TRAVIS_TAG}
# endif
#
# gofmt:
# 	@echo "+++ Formatting code with Gofmt"
# 	@docker-compose run --rm gofmt -s -w ${GOFILES_NOVENDOR}
#
# goimports:
# 	@echo "+++ Checking imports with go imports"
# 	@docker-compose run --rm goimports -e -l -w ${GOFILES_NOVENDOR}
#

# lint:
# 	@echo "+++ Running gometalinter"
# 	@docker-compose run --rm gometalinter \
# 	--sort path \
# 	--skip=client --skip=apis --skip=signals --skip mocks \
# 	--deadline 300s \
# 	--vendor \
# 	--enable-all \
# 	--disable lll \
# 	--disable test \
# 	--disable testify \
# 	./... --debug

# postgres-exporter-up:
# 	@echo "Deploying exporter..."
# 	@docker-compose run --rm gomplate --file=postgres-exporter.yaml --datasource config=values.yaml --datasource queries=queries.yaml -o output.yaml
# 	@docker-compose run --rm kubectl apply -f output.yaml
#
# postgres-exporter-down:
# 	@echo "Destroying exporter..."
# 	@docker-compose run --rm kubectl delete -f output.yaml

%-dev-red: DB_SUBNET_GROUP := cluster-development-vpc-dbsubnetgroup-1phyat6a25wh
%-dev-red: DB_SECURITY_GROUP_IDS := sg-95ed83f3
%-dev-red: ns := platform-enablement
%-dev-red: postgresName := $(shell docker-compose run --rm kubectl get postgresdbs --all-namespaces | grep platform-enablement | awk {'print $$2'})

watch-%:
	@echo "+++ Running outside cluster"
	@docker-compose run --rm \
		-e DB_SUBNET_GROUP=$(DB_SUBNET_GROUP) \
		-e DB_SECURITY_GROUP_IDS=$(DB_SECURITY_GROUP_IDS) \
		go run *.go -logtostderr=true -v=2

apply-%:
	@echo "Apply new db RDS resources for $(*)"
	@docker-compose run --rm kubectl apply -f /app/yaml/example.yaml

check-%:
	@echo "---- In $(ns)"
	@echo "          "
	@echo "---- check $(ns) CRD status"
	@docker-compose run --rm kubectl get postgresdbs -n $(ns) $(postgresName) -o yaml
	@echo "          "
	@echo "---- check $(ns) secrets"
	@docker-compose run --rm kubectl get secrets -n $(ns) | grep $(postgresName)
	@echo "          "
	@echo "---- check $(ns) deployment"
	@docker-compose run --rm kubectl get deployment -n $(ns) | grep $(postgresName)
	@echo "          "
	@echo "---- check $(ns) services"
	@docker-compose run --rm kubectl get svc -n $(ns) | grep $(postgresName)
	@echo "          "
	@echo "---- check $(ns) configMap"
	@docker-compose run --rm kubectl get configMap -n $(ns) | grep $(postgresName)
	@echo "          "
	@echo "--- In kube-system"
	@docker-compose run --rm kubectl get secrets -n kube-system | grep $(postgresName)

pgweb-%:
	@echo "+++ pgweb ... "
	./hack/access-db $(ns) $(postgresName)

teardown-%:
	@echo "--- Delete $(ns) postgresdbs"
	@docker-compose run --rm kubectl -n $(ns) delete postgresdbs $(postgresName)
	@echo "--- Delete $(ns) secrets"
	@docker-compose run --rm kubectl -n $(ns) delete secrets $(shell docker-compose run --rm kubectl -n platform-enablement get secrets | grep $(postgresName) | awk {'print $$1'})
	@echo "--- Delete $(ns) deployment"
	@docker-compose run --rm kubectl -n $(ns) delete deployment $(postgresName)-metrics-exporter
	@echo "--- Delete $(ns) services"
	@docker-compose run --rm kubectl delete svc -n $(ns) $(postgresName)-metrics-exporter
	@echo "--- Delete $(ns) configMap"
	@docker-compose run --rm kubectl delete configMap -n $(ns) $(postgresName)-metrics-exporter
	@echo "--- Delete kube-system secrets"
	@docker-compose run --rm kubectl -n kube-system delete secrets $(shell docker-compose run --rm kubectl -n kube-system get secrets | grep $(postgresName) | awk {'print $$1'})
