APP_NAME = user-auth
APP_VERSION = v1.3
APP_BIN = server
CONTAINER_REPO_USER = mixedmachine


.PHONY: local.build.lin local.build.win local.dev \
		pre-build docs lint sec test db \
		docker.dev docker.prod docker.run docker.compose.dev docker.push \
		clean


local.build.win: main.go pkg/* cmd/* pre-build docs
	@go build -o ./bin/$(APP_BIN).exe main.go

local.build.lin:
	@go build -o ./bin/$(APP_BIN) main.go

local.dev: db build 
	@./bin/$(APP_BIN).exe

pre-build:
	@go mod tidy
	@go fmt ./...

docs:
	@swag init -g ./main.go -o ./api

lint:
	@golangci-lint run

sec:
	@gosec ./...

test:
	@go test -v ./...

db:
	@docker compose -f ./build/docker-compose.db.yml up -d

docker.dev:
	@docker build -f ./build/Dockerfile  --build-arg ENV_FILE=.env -t $(CONTAINER_REPO_USER)/$(APP_NAME):latest-dev .
	@docker build -f ./build/Dockerfile  --build-arg ENV_FILE=.env -t $(CONTAINER_REPO_USER)/$(APP_NAME):$(APP_VERSION)-dev .
	@docker image prune -f

docker.prod:
	@docker build -f ./build/Dockerfile  --build-arg ENV_FILE=.env.prod -t $(CONTAINER_REPO_USER)/$(APP_NAME):latest .
	@docker build -f ./build/Dockerfile  --build-arg ENV_FILE=.env.prod -t $(CONTAINER_REPO_USER)/$(APP_NAME):$(APP_VERSION) .
	@docker image prune -f


docker.run: docker.dev
	@docker run -d \
	-p 8080:9090 \
	--name $(APP_NAME) \
	$(APP_NAME):latest-dev

docker.compose.dev: docker.dev
	@docker compose -f ./build/docker-compose.db.yml up --build -d
	@docker compose -f ./build/docker-compose.api.yml up --build -d

docker.push: docker.prod
	@docker push $(CONTAINER_REPO_USER)/$(APP_NAME):latest
	@docker push $(CONTAINER_REPO_USER)/$(APP_NAME):$(APP_VERSION)

clean:
	@rm -f ./bin/$(APP_BIN)
	@docker rm -f $(APP_NAME)
	@docker compose -f ./build/docker-compose.db.yml down
	@docker compose -f ./build/docker-compose.api.yml down
