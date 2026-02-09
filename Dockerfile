FROM golang:1.24.13-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /gwaihir \
    ./cmd/gwaihir

FROM scratch

COPY --from=builder /gwaihir /gwaihir

EXPOSE 8080

ENTRYPOINT ["/gwaihir"]
