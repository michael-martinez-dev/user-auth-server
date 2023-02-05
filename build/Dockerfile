FROM golang:1.18-alpine as build

RUN apk update && apk add make git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

ARG ENV_FILE

COPY ./main.go ./main.go
COPY ./api ./api
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./${ENV_FILE} ./.env

COPY ./Makefile ./Makefile
RUN make local.build.lin

FROM alpine:latest

WORKDIR /

ARG ENV_FILE

COPY --from=build /app/bin/server /server
COPY  ./${ENV_FILE} /.env

EXPOSE 9090

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

ENTRYPOINT ["/server"]
