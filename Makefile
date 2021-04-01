# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

TEST_OUTPUT ?= "/tmp/kocg-test.go"

build:
	go build -o kocg main.go

install: build
	mv kocg $(GOBIN)

test:
	go run test.go -manifest sample/deploy-part.yaml output $(TEST_OUTPUT)

test.run: test
	go run $(TEST_OUTPUT)

test.validate:
	kubectl get deploy

test.clean:
	kubectl delete deploy coredns

