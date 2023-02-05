APP_NAME = user-auth
APP_VERSION = v1.2
APP_BIN = server
CONTAINER_REPO_USER = mixedmachine


.PHONY: doc db dev pipeline docker docker.run docker.push docker.compose 


build: main.go pkg/* cmd/* docs
	@go mod tidy
	@go build -o ./bin/$(APP_BIN).exe main.go

lint:
	@golangci-lint run

sec:
	@gosec ./...

test:
	@go test -v ./...

docs:
	@swag init -g ./main.go -o ./api

db:
	@docker compose -f ./build/docker-compose.db.yml up -d

dev: db build 
	@./bin/$(APP_BIN).exe

pipeline:
	@go fmt ./...
	@golangci-lint run

dockerfile:
	@go build -o ./bin/$(APP_BIN) main.go

docker: docs test lint sec
	@docker build -f ./build/Dockerfile -t $(CONTAINER_REPO_USER)/$(APP_NAME):latest .
	@docker build -f ./build/Dockerfile -t $(CONTAINER_REPO_USER)/$(APP_NAME):$(APP_VERSION) .

docker.run: docker
	@docker run -d \
	-p 8080:8080 \
	--env-file .env.local \
	--name $(APP_NAME) \
	$(APP_NAME):latest

docker.push: docker
	@docker push $(CONTAINER_REPO_USER)/$(APP_NAME):latest
	@docker push $(CONTAINER_REPO_USER)/$(APP_NAME):$(APP_VERSION)

docker.compose.dev: docker
	@docker compose -f ./build/docker-compose.db.yml up --build -d
	@docker compose -f ./build/docker-compose.api.yml up --build -d

clean:
	@rm -f ./bin/$(APP_BIN)
	@docker rm -f $(APP_NAME)
	@docker compose -f ./build/docker-compose.db.yml down
	@docker compose -f ./build/docker-compose.api.yml down
