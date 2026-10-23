# ---- Etapa 1: compilar el binario de Go ----
FROM golang:1.26-alpine AS build
WORKDIR /src
# Primero solo los módulos: así Docker reutiliza la caché si el código cambia pero las dependencias no.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api .

# ---- Etapa 2: imagen final pequeña ----
FROM alpine:3.21
# ca-certificates: para conexiones TLS (p. ej. a la base de datos); tzdata: zonas horarias.
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
COPY --from=build /out/api /app/api
USER app
# Render inyecta la variable PORT (por defecto 10000); la API ya la lee. En local usa 3000.
EXPOSE 3000
CMD ["/app/api"]
