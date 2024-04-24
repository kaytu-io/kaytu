build:
	go build -o cli .

release:
	git tag v1.0.0
	git push origin v1.0.0