#!/bin/bash
set -e

echo "🔧 Setting up Gordon Watcher dependencies..."

# Remove old files
rm -f go.mod go.sum

# Initialize module
echo "📦 Initializing Go module..."
go mod init github.com/fabyo/gordon-watcher

# Add dependencies
echo "📥 Adding dependencies..."
go get github.com/fsnotify/fsnotify@v1.7.0
go get github.com/google/uuid@v1.6.0
go get github.com/prometheus/client_golang@v1.18.0
go get github.com/rabbitmq/amqp091-go@v1.9.0
go get github.com/redis/go-redis/v9@v9.4.0
go get github.com/spf13/viper@v1.18.2
go get go.opentelemetry.io/otel@v1.22.0
go get go.opentelemetry.io/otel/exporters/jaeger@v1.17.0
go get go.opentelemetry.io/otel/sdk@v1.22.0
go get go.opentelemetry.io/otel/trace@v1.22.0
go get golang.org/x/time@v0.5.0

# Tidy up
echo "🧹 Tidying up..."
go mod tidy

# Download everything
echo "⬇️  Downloading modules..."
go mod download

echo ""
echo "✅ Dependencies installed successfully!"
echo ""
echo "Next steps:"
echo "  1. make build"
echo "  2. make dev"
echo ""
