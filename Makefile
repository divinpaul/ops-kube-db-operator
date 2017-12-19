.PHONY: publish login

REPO := myobplatform/ops-kube-db-operator

test: deps
	docker-compose run go test ./...

deps:
	docker-compose run dep ensure -v

ifdef DOCKERHUB_PASSWORD
login:
	@docker login --username ${DOCKERHUB_USERNAME} --password ${DOCKERHUB_PASSWORD}
endif

publish:
	docker build --build-arg VERSION=latest -t ${REPO}:latest .
	docker push ${REPO}:latest

ifdef TRAVIS_TAG
	docker build --build-arg VERSION=${TRAVIS_TAG} -t ${REPO}:${TRAVIS_TAG} .
	docker push ${REPO}:${TRAVIS_TAG}
endif
