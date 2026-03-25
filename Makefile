VERSION ?= 0.0.1

.PHONY: generate build test testacc lint drift-check docs

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config internal/client/generated/oapi-codegen.cfg.yaml \
		interop-api-v3-spec.yaml

build:
	go build -ldflags "-X main.version=$(VERSION)" -o terraform-provider-pdnd ./cmd/terraform-provider-pdnd

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./internal/... -v -run TestAcc -timeout 10m

lint:
	golangci-lint run ./...

drift-check: generate
	git diff --exit-code internal/client/generated/

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
