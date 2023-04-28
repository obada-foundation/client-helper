PROJECT = obada/client-helper
COMMIT_BRANCH ?= develop
PROJECT_IMAGE = $(PROJECT):$(COMMIT_BRANCH)
PROJECT_RELEASE_IMAGE = $(PROJECT):main
PROJECT_TAG_IMAGE = $(PROJECT):$(COMMIT_TAG)
GITHUB_CLONE = git@github.com:obada-foundation

SHELL := /bin/sh
.DEFAULT_GOAL := help

docker: docker/build

docker/build:
	docker build -t $(PROJECT_IMAGE) -f docker/Dockerfile .

openapi/python/clone: ## Clone github.com/obada-foundation/client-api-library-python if it does not exists
	if [ ! -d "./client-api-library-python" ]; then git clone -b main $(GITHUB_CLONE)/client-api-library-python ./client-api-library-python; fi

openapi/gen: openapi/python/gen openapi/csharp/gen openapi/node/gen openapi/go/gen openapi/php/gen

openapi/python/gen: openapi/python/clone
	rm -rf $$(pwd)/client-api-library-python/*
	docker run --rm \
		-v $$(pwd)/openapi:/local \
		-v $$(pwd)/client-api-library-python:/src \
		openapitools/openapi-generator-cli generate \
		-i /local/spec.openapi.yaml \
		-g python \
		-o /src \
		-c /local/clients/python/config.yaml

openapi/csharp/clone: ## Clone github.com/obada-foundation/client-api-library-csharp if it does not exists
	if [ ! -d "./client-api-library-csharp" ]; then git clone -b main $(GITHUB_CLONE)/client-api-library-csharp ./client-api-library-csharp; fi

openapi/csharp/gen: openapi/csharp/clone
	rm -rf $$(pwd)/client-api-library-csharp/*
	docker run --rm \
		-v $$(pwd)/openapi:/local \
		-v $$(pwd)/client-api-library-csharp:/src \
		openapitools/openapi-generator-cli generate \
		-i /local/spec.openapi.yaml \
		-g csharp \
		-o /src \
		-c /local/clients/csharp/config.yaml

openapi/node/clone: ## Clone github.com/obada-foundation/client-api-library-node if it does not exists
	if [ ! -d "./client-api-library-node" ]; then git clone -b main $(GITHUB_CLONE)/client-api-library-node ./client-api-library-node; fi

openapi/node/gen: openapi/node/clone
	rm -rf $$(pwd)/client-api-library-node/*
	docker run --rm \
		-v $$(pwd)/openapi:/local \
		-v $$(pwd)/client-api-library-node:/src \
		openapitools/openapi-generator-cli generate \
		-i /local/spec.openapi.yaml \
		-g typescript-node \
		-o /src \
		-c /local/clients/node/config.yaml

openapi/go/clone: ## Clone github.com/obada-foundation/client-api-library-go if it does not exists
	if [ ! -d "./client-api-library-go" ]; then git clone -b main $(GITHUB_CLONE)/client-api-library-go ./client-api-library-go; fi

openapi/go/gen: openapi/go/clone
	rm -rf $$(pwd)/client-api-library-go/*
	docker run --rm \
		-v $$(pwd)/openapi:/local \
		-v $$(pwd)/client-api-library-go:/src \
		openapitools/openapi-generator-cli generate \
		-i /local/spec.openapi.yaml \
		-g go \
		-o /src \
		-c /local/clients/go/config.yaml

openapi/php/clone: ## Clone github.com/obada-foundation/client-api-library-php if it does not exists
	if [ ! -d "./client-api-library-php" ]; then git clone -b main $(GITHUB_CLONE)/client-api-library-php ./client-api-library-php; fi

openapi/php/gen: openapi/php/clone
	rm -rf $$(pwd)/client-api-library-php/*
	docker run --rm \
		-v $$(pwd)/openapi:/local \
		-v $$(pwd)/client-api-library-php:/src \
		openapitools/openapi-generator-cli generate \
		-i /local/spec.openapi.yaml \
		-g php \
		-o /src \
		-c /local/clients/php/config.yaml

openapi/go:

ci/openapi/lint: ## Checks /openapi/spec.openapi.yaml file is follows OpenApi standard schema
	docker run \
                --rm \
		-v $$(pwd)/openapi:/openapi/ \
		wework/speccy lint /openapi/spec.openapi.yaml

ci/coverage:
	go test $(cd src && go list ./... | grep -v /vendor/) -v -coverprofile .testCoverage.txt

ci/test:
	go test -v ./... -cover

ci/lint:
	golangci-lint --config .golangci.yml run --print-issued-lines --out-format=github-actions ./...

run:
	go run main.go \
		server \
		--db-path ../data

vendor: vendor/tidy
	go mod vendor

vendor/tidy:
	go mod tidy

help: ## Show this help.
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "%-30s %s\n" $$help_command $$help_info ; \
	done

