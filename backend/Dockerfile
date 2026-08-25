FROM golang:1.26.2-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o /out/juntly-api ./cmd/api

FROM alpine:3.22

RUN addgroup -S juntly && adduser -S -G juntly juntly

COPY --from=build /out/juntly-api /usr/local/bin/juntly-api

USER juntly
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/juntly-api"]
