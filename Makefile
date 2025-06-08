RELEASE_VERSION ?= development

lint: phpcs phpcbf phpstan

phpcs:
	vendor/bin/phpcs src/

phpcbf:
	vendor/bin/phpcbf src/

phpstan:
	vendor/bin/phpstan analyse src/

phpunit:
	vendor/bin/phpunit tests/*

clean:
	rm -rf build

container-image:
	docker kill faridoon || true
	docker rm faridoon && docker rmi faridoon || true
	docker build -t faridoon:latest .

container: container-image
	docker create --name faridoon -p 8080:8080 --env-file=.env.dev faridoon:latest
	docker start faridoon

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

.PHONY: dist clean docker-container-image container
