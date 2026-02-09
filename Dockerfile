FROM golang:1.22.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /gwaihir \
    ./cmd/gwaihir

FROM scratch

COPY --from=builder /gwaihir /gwaihir

EXPOSE 8080

ENTRYPOINT ["/gwaihir"]
