.PHONY: all build test clean sync-work

all: sync-work provision-llama build test

sync-work:
	go work sync

provision-llama:
	bash scripts/provision_llama.sh

build:
	CGO_ENABLED=1 go build ./internal/pii/...
	CGO_ENABLED=1 go build ./services/refinery/...
	CGO_ENABLED=1 go build ./apps/proxy/...
	CGO_ENABLED=1 go build ./apps/sombra/...
	CGO_ENABLED=1 go build ./enterprise/refinery-extensions/...

test:
	CGO_ENABLED=1 go test ./internal/pii/...
	CGO_ENABLED=1 go test ./services/refinery/...
	CGO_ENABLED=1 go test ./apps/proxy/...
	CGO_ENABLED=1 go test ./apps/sombra/...
	CGO_ENABLED=1 go test ./enterprise/refinery-extensions/...

clean:
	go clean -cache
	rm -rf dist/

release:
	@echo "Building release artifacts..."
	bash tools/scripts/scripts/build_release.sh
	@echo "Tagging v1.0.0..."
	git tag -a v1.0.0 -m "Release v1.0.0"
	@echo "Pushing tag to origin..."
	git push origin v1.0.0
	@echo "Done! Please create the release on GitHub: https://github.com/Edu963/ocultar/releases/new?tag=v1.0.0"
	@echo "Make sure to attach the .zip and .tar.gz files from the dist/ folder to the release!"


