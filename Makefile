.PHONY: publish publish-version login test clean

REPO := myobplatform/ops-kube-db-operator
PKG_DIR = github.com/MYOB-Technology/ops-kube-db-operator/pkg
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_TEST_PKGS = $(shell docker-compose run go list ./... |grep -v $(PKG_DIR)/client |grep -v $(PKG_DIR)/signals |grep -v $(PKG_DIR)/apis )

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
