#!/bin/bash

# Initialize Go module and install dependencies

echo "Initializing Go module..."

# Initialize the module
go mod init github.com/CS6650-Distributed-Systems/album-store-plus

# Add required dependencies
echo "Adding dependencies..."

# AWS SDK
go get github.com/aws/aws-sdk-go@latest
go get github.com/aws/aws-lambda-go@latest

# HTTP router and middleware
go get github.com/go-chi/chi/v5@latest
go get github.com/go-chi/cors@latest

# MySQL driver
go get github.com/go-sql-driver/mysql@latest

# Utils
go get github.com/google/uuid@latest
go get go.uber.org/zap@latest
go get github.com/nfnt/resize@latest

# Clean up dependencies
echo "Tidying up dependencies..."
go mod tidy

echo "Go module initialization complete"