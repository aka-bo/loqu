
# Image URL to use all building/pushing image targets
IMG ?= 795669731331.dkr.ecr.us-west-2.amazonaws.com/utils/loqu:0.0.2

all: build

# Run tests
test: fmt vet
	go test ./{cmd,pkg}/... -coverprofile cover.out

# Build loqu binary
build: fmt vet
	go build -o bin/loqu main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: build
	./bin/loqu serve $(filter-out $@,$(MAKECMDGOALS))

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Build the docker image
docker-build: test build
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

%:
    @:
