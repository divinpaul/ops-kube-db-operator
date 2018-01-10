.PHONY: publish publish-version login test clean

REPO := myobplatform/ops-kube-db-operator
PKG_DIR = github.com/MYOB-Technology/ops-kube-db-operator/pkg
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_TEST_PKGS = $(shell docker-compose run go list ./... |grep -v $(PKG_DIR)/client |grep -v $(PKG_DIR)/signals |grep -v $(PKG_DIR)/apis )

deps:
	docker-compose run dep ensure -v

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

code-gen:
	@docker-compose run console ./bin/update-codegen.sh