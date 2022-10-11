
# Install binaries
install:
	go install cmd/cxs/cxs.go
	go install cmd/pods/pods.go

# Run all tests in repository.
test:
	go test ./...