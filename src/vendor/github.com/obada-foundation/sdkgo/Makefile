lint:
	golangci-lint --config .golangci.yml run --print-issued-lines --out-format=github-actions ./...

test:
	go test ./... -v

vendor:
	go mod tidy && go mod vendor

coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out