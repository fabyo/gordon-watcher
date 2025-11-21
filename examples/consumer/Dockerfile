FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY examples/consumer/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /consumer .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /consumer .

CMD ["./consumer"]
