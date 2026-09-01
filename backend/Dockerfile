FROM golang:1.26.6-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o /out/juntly-api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates && addgroup -S juntly && adduser -S -G juntly juntly

COPY --from=build /out/juntly-api /usr/local/bin/juntly-api

USER juntly
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=5 CMD wget -q -O /dev/null http://127.0.0.1:8080/api/v1/ready || exit 1

ENTRYPOINT ["/usr/local/bin/juntly-api"]
