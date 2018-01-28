.PHONY: publish publish-version login test clean

REPO := myobplatform/ops-kube-db-operator
PKG_DIR = github.com/MYOB-Technology/ops-kube-db-operator/pkg
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_TEST_PKGS = $(shell docker-compose run go list ./... |grep -v $(PKG_DIR)/client |grep -v $(PKG_DIR)/signals |grep -v $(PKG_DIR)/apis )

ci: vendor test

test:
	@docker-compose run --rm go test ${GO_TEST_PKGS}

vendor:
	docker-compose run --rm dep ensure -v

clean:
	docker-compose run --rm base rm -rf vendor

ifdef DOCKERHUB_PASSWORD
login:
	@docker login --username ${DOCKERHUB_USERNAME} --password ${DOCKERHUB_PASSWORD}
endif

publish: publish-version
	docker build --build-arg VERSION=latest -t ${REPO}:latest .
	docker push ${REPO}:latest

ifdef TRAVIS_TAG
publish-version:
	docker build --build-arg VERSION=${TRAVIS_TAG} -t ${REPO}:${TRAVIS_TAG} .
	docker push ${REPO}:${TRAVIS_TAG}
endif

gofmt:
	@echo "+++ Formatting code with Gofmt"
	@docker-compose run --rm gofmt -s -w ${GOFILES_NOVENDOR}

goimports:
	@echo "+++ Checking imports with go imports"
	@docker-compose run --rm goimports -e -l -w ${GOFILES_NOVENDOR}

lint:
	@echo "+++ Running gometalinter"
	@docker-compose run --rm gometalinter \
	--sort path \
	--skip=client --skip=apis --skip=signals \
	--deadline 300s \
	--vendor \
	--enable-all \
	--disable lll \
	--disable test \
	--disable testify \
	./... --debug

run:
	@echo "+++ Running outside cluster"
	@docker-compose run --rm go run *.go -logtostderr=true -v=2

postgres-exporter-up:
	@echo "Deploying exporter..."
	@docker-compose run --rm gomplate --file=postgres-exporter.yaml --datasource config=values.yaml --datasource queries=queries.yaml -o output.yaml
	@docker-compose run --rm kubectl apply -f output.yaml

postgres-exporter-down:
	@echo "Destroying exporter..."
	@docker-compose run --rm kubectl delete -f output.yaml
