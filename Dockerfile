FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git nodejs npm

WORKDIR /build

COPY . .

WORKDIR /build/ui
RUN npm ci --silent
RUN npm run build

WORKDIR /build
# adapter-static gera em ui/build — copia o CONTEÚDO (barra + ponto)
RUN rm -rf internal/ui/build \
 && mkdir -p internal/ui/build \
 && cp -r ui/build/. internal/ui/build/

# debug: confirma que index.html está na raiz
RUN ls -la internal/ui/build/index.html

RUN GOPRIVATE="github.com/galileostd/*" go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cosmonaut ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /cosmonaut /cosmonaut
EXPOSE 8080 8081 9090
ENTRYPOINT ["/cosmonaut"]
