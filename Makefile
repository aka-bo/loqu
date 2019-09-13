
# Image URL to use all building/pushing image targets
IMG ?= 795669731331.dkr.ecr.us-east-1.amazonaws.com/appsol/loqu:0.0.1

all: server

# Run tests
test: fmt vet
	go test ./{cmd,pkg}/... -coverprofile cover.out

# Build manager binary
server: fmt vet
	go build -o bin/loqu main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: server
	./bin/loqu serve

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
