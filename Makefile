.PHONY: publish publish-version login test clean

REPO := myobplatform/ops-kube-db-operator

test: vendor
	docker-compose run go test ./...

vendor:
	docker-compose run dep ensure -v

clean:
	docker-compose run base rm -rf vendor

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
