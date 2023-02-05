FROM golang:1.18-alpine as build

RUN apk update && apk add make git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./main.go ./main.go
COPY ./api ./api
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./.env ./.env

COPY ./Makefile ./Makefile
RUN make dockerfile

FROM alpine:latest

WORKDIR /

COPY --from=build /app/bin/server /server
COPY ./.env /.env

EXPOSE 8080

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

ENTRYPOINT ["/server"]