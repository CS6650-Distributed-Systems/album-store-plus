.PHONY: build clean test run-command run-query run-all lambda localstack terraform-init terraform-plan terraform-apply help

# Basic configuration
BINARY_NAME_COMMAND=command-api
BINARY_NAME_QUERY=query-api
BINARY_NAME_LAMBDA=imageprocessor
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")
LAMBDA_DIR=lambda/imageprocessor

# AWS configuration
AWS_REGION=us-east-1
S3_BUCKET=album-store-images
LAMBDA_FUNCTION=album-image-processor
SNS_TOPIC=album-events
SQS_QUEUE=album-events-queue
DYNAMODB_TABLE=albums

# Docker & LocalStack configuration
LOCALSTACK_CONTAINER=album-store-localstack
LOCAL_PORT=4566

# Build commands
build: build-command build-query build-lambda

build-command:
	@echo "Building command API..."
	go build -o bin/$(BINARY_NAME_COMMAND) ./cmd/commandapi

build-query:
	@echo "Building query API..."
	go build -o bin/$(BINARY_NAME_QUERY) ./cmd/queryapi

build-lambda:
	@echo "Building Lambda function..."
	cd $(LAMBDA_DIR) && GOOS=linux GOARCH=amd64 go build -o ../../bin/$(BINARY_NAME_LAMBDA)
	cd bin && zip $(BINARY_NAME_LAMBDA).zip $(BINARY_NAME_LAMBDA)

# Clean built binaries
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the applications
run-command: build-command
	@echo "Running command API..."
	./bin/$(BINARY_NAME_COMMAND)

run-query: build-query
	@echo "Running query API..."
	./bin/$(BINARY_NAME_QUERY)

run-all: build
	@echo "Running all services..."
	./bin/$(BINARY_NAME_COMMAND) & ./bin/$(BINARY_NAME_QUERY)

# LocalStack setup and management
localstack-start:
	@echo "Starting LocalStack..."
	docker run -d --name $(LOCALSTACK_CONTAINER) \
		-p $(LOCAL_PORT):4566 \
		-e SERVICES=s3,dynamodb,lambda,sns,sqs \
		-e DEFAULT_REGION=$(AWS_REGION) \
		-e LAMBDA_EXECUTOR=docker \
		-e LAMBDA_REMOTE_DOCKER=true \
		-e DEBUG=1 \
		localstack/localstack:latest

localstack-stop:
	@echo "Stopping LocalStack..."
	docker stop $(LOCALSTACK_CONTAINER)
	docker rm $(LOCALSTACK_CONTAINER)

# Initialize AWS resources in LocalStack
localstack-init: build-lambda
	@echo "Creating S3 bucket..."
	aws --endpoint-url=http://localhost:$(LOCAL_PORT) s3 mb s3://$(S3_BUCKET)
	
	@echo "Creating DynamoDB table..."
	aws --endpoint-url=http://localhost:$(LOCAL_PORT) dynamodb create-table \
		--table-name $(DYNAMODB_TABLE) \
		--attribute-definitions AttributeName=id,AttributeType=S \
		--key-schema AttributeName=id,KeyType=HASH \
		--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
		
	@echo "Creating SNS topic..."
	aws --endpoint-url=http://localhost:$(LOCAL_PORT) sns create-topic \
		--name $(SNS_TOPIC)
		
	@echo "Creating SQS queue..."
	aws --endpoint-url=http://localhost:$(LOCAL_PORT) sqs create-queue \
		--queue-name $(SQS_QUEUE)
		
	@echo "Deploying Lambda function..."
	aws --endpoint-url=http://localhost:$(LOCAL_PORT) lambda create-function \
		--function-name $(LAMBDA_FUNCTION) \
		--runtime go1.x \
		--handler $(BINARY_NAME_LAMBDA) \
		--zip-file fileb://bin/$(BINARY_NAME_LAMBDA).zip \
		--role arn:aws:iam::000000000000:role/lambda-role
		
	@echo "LocalStack resources initialized."

# Terraform commands
terraform-init:
	@echo "Initializing Terraform..."
	cd infra/terraform && terraform init

terraform-plan:
	@echo "Planning Terraform infrastructure..."
	cd infra/terraform && terraform plan

terraform-apply:
	@echo "Applying Terraform changes..."
	cd infra/terraform && terraform apply

# Setup project dependencies and environment
setup:
	@echo "Installing dependencies..."
	go mod tidy
	
	@echo "Creating necessary directories..."
	mkdir -p bin logs configs

# Generate configuration file
config-gen:
	@echo "Generating default configuration..."
	cp configs/config.example.json configs/config.json

# Create initialization script
init-script:
	chmod +x setup_project.sh
	./setup_project.sh

# Help command
help:
	@echo "Album Store Plus - Makefile commands:"
	@echo "make build           - Build all components"
	@echo "make build-command   - Build only command API"
	@echo "make build-query     - Build only query API"
	@echo "make build-lambda    - Build only Lambda function"
	@echo "make clean           - Remove built binaries"
	@echo "make test            - Run tests"
	@echo "make run-command     - Run command API"
	@echo "make run-query       - Run query API"
	@echo "make run-all         - Run both APIs"
	@echo "make localstack-start - Start LocalStack container"
	@echo "make localstack-stop  - Stop LocalStack container"
	@echo "make localstack-init  - Initialize AWS resources in LocalStack"
	@echo "make terraform-init   - Initialize Terraform"
	@echo "make terraform-plan   - Plan Terraform infrastructure"
	@echo "make terraform-apply  - Apply Terraform changes"
	@echo "make setup            - Setup project dependencies"
	@echo "make config-gen       - Generate default configuration"
	@echo "make init-script      - Run initialization script"