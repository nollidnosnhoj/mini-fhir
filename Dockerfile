FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mini-fhir ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/mini-fhir /app/mini-fhir

EXPOSE 8080

ENTRYPOINT ["/app/mini-fhir"]
