.PHONY: all generate build test lint install setup service frontend integration-test docs clean release docker-amd64 docker-arm64 docker-manifest

all: build

generate:
	cd protocol/proto && buf generate --template ../buf.gen.yaml

build: generate service frontend

service:
	$(MAKE) -wC service build

frontend:
	$(MAKE) -wC frontend build

test: generate
	$(MAKE) -wC service test
	$(MAKE) -wC frontend test

lint:
	$(MAKE) -wC service lint
	$(MAKE) -wC frontend lint

install:
	cd service && go mod download
	cd frontend && npm install

setup: install generate
	@echo "Setup complete. Run 'make dev-service' and 'make dev-frontend'."

dev-service:
	$(MAKE) -wC service run

dev-frontend:
	$(MAKE) -wC frontend run

integration-test: service
	cd integration-tests && npm install && node run-tests.js

docs:
	mkdocs build

clean:
	rm -rf service/faridoon-service frontend/dist frontend/gen service/gen

RELEASE_VERSION ?= development

docker-amd64:
	docker buildx build --platform linux/amd64 -t ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-amd64 -f Dockerfile --output type=docker --load .
	docker push ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-amd64

docker-arm64:
	docker buildx build --platform linux/arm64 -t ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-arm64 -f Dockerfile --output type=docker --load .
	docker push ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-arm64

docker-manifest-latest:
	docker manifest create ghcr.io/jamesread/faridoon:latest \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-arm64 \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-amd64
	docker manifest annotate ghcr.io/jamesread/faridoon:latest \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-amd64 --os linux --arch amd64
	docker manifest annotate ghcr.io/jamesread/faridoon:latest \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-arm64 --os linux --arch arm64
	docker manifest push ghcr.io/jamesread/faridoon:latest

docker-manifest-release-version:
	docker manifest create ghcr.io/jamesread/faridoon:${RELEASE_VERSION} \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-amd64 \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-arm64
	docker manifest annotate ghcr.io/jamesread/faridoon:${RELEASE_VERSION} \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-amd64 --os linux --arch amd64
	docker manifest annotate ghcr.io/jamesread/faridoon:${RELEASE_VERSION} \
		ghcr.io/jamesread/faridoon:${RELEASE_VERSION}-arm64 --os linux --arch arm64
	docker manifest push ghcr.io/jamesread/faridoon:${RELEASE_VERSION}

docker-manifest: docker-manifest-latest docker-manifest-release-version

release: docker-amd64 docker-arm64 docker-manifest
