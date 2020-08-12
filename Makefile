.PHONY: all build

WORKSPACE ?= $$(pwd)

GO_PKG_LIST := $(shell go list ./... | grep -v /vendor/ | grep -v /mock | grep -v ./pkg/apic/apiserver/*/definitions/v1alpha \
	| grep -v ./pkg/apic/apiserver/*/management/v1alpha | grep -v ./pkg/apic/unifiedcatalog/models \
	| grep -v ./pkg/apic/apiserver/clients/api/v1)

export GOFLAGS := -mod=vendor

all : clean

clean:
	@echo "Clean complete"

dep-check:
	@go mod verify

resolve-dependencies:
	@echo "Resolving go package dependencies"
	@go mod tidy
	@go mod vendor
	@echo "Package dependencies completed"

dep: resolve-dependencies

test:
	@go vet ${GO_PKG_LIST}
	@go test -short -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST}

test-sonar:
	@go vet ${GO_PKG_LIST}
	@go test -short -coverpkg=./... -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST} -json > ${WORKSPACE}/goreport.json

sonar: test-sonar
	sonar-scanner -X \
		-Dsonar.host.url=http://quality1.ecd.axway.int \
		-Dsonar.language=go \
		-Dsonar.projectName=APIC_AGENTS_SDK \
		-Dsonar.projectVersion=1.0 \
		-Dsonar.projectKey=APIC_AGENTS_SDK \
		-Dsonar.sourceEncoding=UTF-8 \
		-Dsonar.projectBaseDir=${WORKSPACE} \
		-Dsonar.sources=. \
		-Dsonar.tests=. \
		-Dsonar.exclusions=**/mock/**,**/vendor/**,**/definitions/v1alpha1/**,**/management/v1alpha1/**,**/api/v1/** \
		-Dsonar.test.inclusions=**/*test*.go \
		-Dsonar.go.tests.reportPaths=goreport.json \
		-Dsonar.go.coverage.reportPaths=gocoverage.out

lint: ## Lint the files
	@golint -set_exit_status $(shell go list ./... | grep -v /vendor/ | grep -v /mock | grep -v ./pkg/apic/apiserver/models/management | grep -v ./pkg/apic/apiserver/models/definitions | grep -v ./pkg/apic/unifiedcatalog/models)

apiserver_generate: ## generate api server resources
	./scripts/apiserver/apiserver_generate.sh

unifiedcatalog_generate: ## generate unified catalog resources
	./scripts/unifiedcatalog/unifiedcatalog_generate.sh
