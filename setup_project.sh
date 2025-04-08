#!/bin/bash

# Setup script for AlbumStore+ project
set -e

echo "Setting up AlbumStore+ project..."

# Create necessary directories
mkdir -p bin logs configs

# Initialize Go module if not already initialized
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init github.com/CS6650-Distributed-Systems/album-store-plus
fi

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Create example configuration file
if [ ! -f "configs/config.example.json" ]; then
    echo "Creating example configuration file..."
    mkdir -p configs
    cat > configs/config.example.json << EOF
{
    "environment": "development",
    "server": {
        "host": "0.0.0.0",
        "port": 8080,
        "timeout": 60
    },
    "database": {
        "host": "localhost",
        "port": "3306",
        "username": "albumuser",
        "password": "albumpass",
        "name": "albumstore",
        "maxOpen": 10,
        "maxIdle": 5,
        "lifetime": 300
    },
    "dynamoDB": {
        "region": "us-east-1",
        "endpoint": "http://localhost:4566",
        "tableName": "albums",
        "readCapacityUnits": 5,
        "writeCapacityUnits": 5
    },
    "s3": {
        "region": "us-east-1",
        "endpoint": "http://localhost:4566",
        "bucketName": "album-store-images"
    },
    "sns": {
        "region": "us-east-1",
        "endpoint": "http://localhost:4566",
        "topicArn": ""
    },
    "sqs": {
        "region": "us-east-1",
        "endpoint": "http://localhost:4566",
        "queueUrl": ""
    },
    "lambda": {
        "region": "us-east-1",
        "endpoint": "http://localhost:4566",
        "functionName": "album-image-processor"
    },
    "logging": {
        "level": "info",
        "format": "console",
        "outputPath": "stdout"
    }
}
EOF
fi

# Copy example configuration to config.json if it doesn't exist
if [ ! -f "configs/config.json" ]; then
    echo "Creating config.json from example..."
    cp configs/config.example.json configs/config.json
fi

# Create Dockerfiles
echo "Creating Dockerfiles..."

# Dockerfile for command API
cat > Dockerfile.command << EOF
FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/command-api ./cmd/commandapi

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/command-api .
COPY --from=builder /app/configs/config.json ./configs/config.json

EXPOSE 8080

CMD ["./command-api"]
EOF

# Dockerfile for query API
cat > Dockerfile.query << EOF
FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/query-api ./cmd/queryapi

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/query-api .
COPY --from=builder /app/configs/config.json ./configs/config.json

EXPOSE 8081

CMD ["./query-api"]
EOF

# Dockerfile for AWS initialization
cat > Dockerfile.init << EOF
FROM amazon/aws-cli:latest

WORKDIR /app

COPY ./scripts/init-aws.sh .
RUN chmod +x init-aws.sh

CMD ["./init-aws.sh"]
EOF

# Create AWS initialization script
mkdir -p scripts
cat > scripts/init-aws.sh << EOF
#!/bin/bash

echo "Initializing AWS resources..."

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
sleep 10

# Create S3 bucket
echo "Creating S3 bucket..."
aws --endpoint-url=\${AWS_ENDPOINT} s3 mb s3://album-store-images

# Create DynamoDB table
echo "Creating DynamoDB table..."
aws --endpoint-url=\${AWS_ENDPOINT} dynamodb create-table \
    --table-name albums \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

# Create SNS topic
echo "Creating SNS topic..."
TOPIC_ARN=\$(aws --endpoint-url=\${AWS_ENDPOINT} sns create-topic \
    --name album-events \
    --output text \
    --query 'TopicArn')

echo "SNS Topic ARN: \$TOPIC_ARN"

# Create SQS queue
echo "Creating SQS queue..."
QUEUE_URL=\$(aws --endpoint-url=\${AWS_ENDPOINT} sqs create-queue \
    --queue-name album-events-queue \
    --output text \
    --query 'QueueUrl')

echo "SQS Queue URL: \$QUEUE_URL"

# Get queue ARN
QUEUE_ARN=\$(aws --endpoint-url=\${AWS_ENDPOINT} sqs get-queue-attributes \
    --queue-url \$QUEUE_URL \
    --attribute-names QueueArn \
    --output text \
    --query 'Attributes.QueueArn')

echo "SQS Queue ARN: \$QUEUE_ARN"

# Subscribe SQS to SNS topic
echo "Subscribing SQS to SNS topic..."
aws --endpoint-url=\${AWS_ENDPOINT} sns subscribe \
    --topic-arn \$TOPIC_ARN \
    --protocol sqs \
    --notification-endpoint \$QUEUE_ARN

echo "AWS resources initialization complete."
EOF

chmod +x scripts/init-aws.sh

echo "Setup complete! You can now build and run the project."
echo "To start the development environment, run: docker-compose up -d"