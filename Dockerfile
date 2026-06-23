FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN GOPRIVATE="github.com/galileostd/*" go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cosmonaut ./cmd/server

FROM gcr.io/distroless/static-debian12

COPY --from=builder /cosmonaut /cosmonaut

EXPOSE 8080 8081 9090

ENTRYPOINT ["/cosmonaut"]