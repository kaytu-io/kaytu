build:
	go build -ldflags "-X github.com/kaytu-io/kaytu/pkg/version.VERSION=99.99.99" -o cli .

release:
	git tag v1.0.0
	git push origin v1.0.0

# removes the dist folder for a clean build
clean:
	rm -rf dist

# cleans the dist directory and generates build using the goreleaser's configuration for current GOOS and GOARCH
goreleaser: clean
	REPOSITORY_NAME="kaytu" REPOSITORY_OWNER="kaytu-io" goreleaser build --snapshot --single-target
