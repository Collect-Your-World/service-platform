FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Install swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger API documentation
RUN swag init --pd -g app/cmd/api/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -s' -o /app/server ./app/cmd/api/main.go

# using debug to have ssh ability
FROM gcr.io/distroless/static-debian12:debug AS runner

WORKDIR /app

COPY --from=builder /app/server /app/server

COPY config ./config

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/server"]
