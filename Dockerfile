FROM golang:1.25.3-alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/bin/ozonposts ./cmd/app


FROM alpine:3.21


RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/ozonposts ./ozonposts
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./ozonposts"]
CMD ["--config", "configs/config.yaml"]
