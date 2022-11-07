
# Install binaries
install:
	go install cmd/kubeui/kubeui.go

# Run all tests in repository.
test:
	go test ./...