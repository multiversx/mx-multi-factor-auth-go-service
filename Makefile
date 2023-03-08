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

debugger := $(shell which dlv)

cmd_dir = cmd/multi-factor-auth
binary = tcs
db_setup = mongodb

build:
	cd ${cmd_dir} && \
		go build -o ${binary}

run: build
	cd ${cmd_dir} && \
		./${binary} --log-level="*:DEBUG"

debug: build
	cd ${cmd_dir} && \
		${debugger} exec ./${binary} -- --log-level="*:DEBUG"

# Run local instance with Docker
image = "multi-factor-auth"
image_tag = "latest"
container_name = multi-factor-auth

dockerfile = Dockerfile

docker-build:
	docker build \
		-t ${image}:${image_tag} \
		-f ${dockerfile} \
		.

network_type = host
ifeq (${db_setup},mongodb)
	network_type = docker_mongo
endif

docker-run:
	docker run  \
		-it \
		--network ${network_type} \
		-p 8080:8080 \
		--name ${container_name} \
		${image}:${image_tag}

docker-new: docker-build docker-run

docker-start:
	docker start ${container_name}

docker-stop:
	docker stop ${container_name}

docker-logs:
	docker logs -f ${container_name}

docker-rm: docker-stop
	docker rm ${container_name}

# #########################
# Cluster setup
# #########################

compose_file = docker/mongodb-cluster.yml
ifeq ($(db_setup),redis)
	compose_file = docker/redis-cluster.yml
endif

compose-new:
	docker-compose -f ${compose_file} up -d

compose-start:
	docker-compose -f ${compose_file} start

compose-stop:
	docker-compose -f ${compose_file} stop

compose-rm:
	docker-compose -f ${compose_file} down
