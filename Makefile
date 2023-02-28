test:
	@echo "  >  Running unit tests"
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic -v ./...

lint-install:
ifeq (,$(wildcard test -f bin/golangci-lint))
	@echo "Installing golint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s
endif

run-lint:
	@echo "Running golint"
	bin/golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 --timeout=2m

lint: lint-install run-lint

# #########################
# Manage TCS locally
# #########################

.PHONY: help build run runb kill debug debug-ath

cmd_dir = cmd/multi-factor-auth
binary = tcs

build:
	cd ${cmd_dir} && \
		go build -o ${binary}

run: build
	cd ${cmd_dir} && \
		./${binary} --log-level="*:DEBUG"


# #########################
# Redis setup
# #########################
compose_file = docker/docker-compose.yml

compose-new:
	docker-compose -f ${compose_file} up -d

compose-start:
	docker-compose -f ${compose_file} start

compose-stop:
	docker-compose -f ${compose_file} stop

compose-rm:
	docker-compose -f ${compose_file} down
