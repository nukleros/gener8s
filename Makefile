# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

TEST_OUTPUT ?= "/tmp/gener8s-test.go"
TEST_MANIFEST ?= "sample/deploy.yaml"

build:
	go build ./cmd/gener8s

install: build
	mv gener8s $(GOBIN)

test:
	go test ./cmd/gener8s -args -manifest=../../$(TEST_MANIFEST) -output=$(TEST_OUTPUT)

test.run: test
	go run $(TEST_OUTPUT)

test.verify:
	kubectl get deploy

test.clean:
	kubectl delete -f $(TEST_MANIFEST)

