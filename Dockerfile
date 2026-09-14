FROM golang:1.24-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Cache dependencies dulu
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN go build --ldflags "-s -w -extldflags -static" -o main .

# Final image — minimal
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/main /app/
COPY --from=builder /build/public/ /app/public/
COPY --from=builder /build/resources/ /app/resources/

EXPOSE 3000

ENTRYPOINT ["/app/main"]
