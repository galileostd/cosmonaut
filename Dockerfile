FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git nodejs npm

WORKDIR /build

# Copia todo o projeto
COPY . .

# Build do frontend
WORKDIR /build/ui
RUN npm ci --silent
RUN npm run build

# Copia o build para o diretório do embed
WORKDIR /build
RUN mkdir -p internal/ui/build \
 && cp -r ui/build/. internal/ui/build/

# Apenas para debug (remover depois)
RUN find internal/ui/build

# Dependências Go
RUN GOPRIVATE="github.com/galileostd/*" go mod download

# Build do servidor
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cosmonaut ./cmd/server

FROM gcr.io/distroless/static-debian12

COPY --from=builder /cosmonaut /cosmonaut

EXPOSE 8080 8081 9090

ENTRYPOINT ["/cosmonaut"]