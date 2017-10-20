.PHONY: publish login

REPO := myobplatform/ops-kube-db-operator

ifdef DOCKERHUB_PASSWORD
login:
	@docker login --username ${DOCKERHUB_USERNAME} --password ${DOCKERHUB_PASSWORD}
endif

publish:
	docker build --build-arg VERSION=latest -t ${REPO}:latest .
	docker push ${REPO}:latest

ifdef BUILDKITE_TAG
	docker build --build-arg VERSION=${BUILDKITE_TAG} -t ${REPO}:${BUILDKITE_TAG} .
	docker push ${REPO}:${BUILDKITE_TAG}
endif
