# Album Store Plus

A distributed application for album management with CQRS architecture using AWS services.

## Architecture

Album Store Plus is built using the Command Query Responsibility Segregation (CQRS) pattern with event sourcing. The system is separated into two main components:

- **Command API**: Handles write operations (creating albums, registering likes/dislikes)
- **Query API**: Handles read operations (retrieving albums, searching, getting statistics)

### Key Components:

- **Command API** writes to MySQL (RDS) and publishes events to SNS
- **Query API** reads from DynamoDB for fast queries
- **Lambda** processes images triggered by events
- **SQS** subscribes to events for asynchronous processing
- **S3** stores album images

## Project Structure

```
AlbumStore/
├── cmd/                    # Application entry points
│   ├── commandapi/         # Command API service
│   └── queryapi/           # Query API service
├── internal/               # Internal packages
│   ├── model/              # Domain models
│   │   ├── command/        # Command models
│   │   └── query/          # Query models
│   ├── repository/         # Data access layer
│   │   ├── rds/            # RDS access (write operations)
│   │   └── dynamodb/       # DynamoDB access (read operations)
│   ├── service/            # Service layer
│   │   ├── command/        # Command services
│   │   └── query/          # Query services
│   ├── eventbus/           # Event handling
│   │   ├── sns/            # SNS publisher
│   │   └── sqs/            # SQS subscriber
│   └── storage/            # Storage services
│       └── s3/             # S3 file storage
├── pkg/                    # Public packages
│   ├── aws/                # AWS clients
│   ├── config/             # Configuration
│   ├── logging/            # Logging
│   └── middleware/         # HTTP middleware
├── lambda/                 # Lambda functions
│   └── imageprocessor/     # Image processing function
├── infra/                  # Infrastructure code
│   └── terraform/          # Terraform configuration
├── configs/                # Configuration files
├── scripts/                # Utility scripts
└── docker-compose.yml      # Local development Docker configuration
```

## Key Features

1. **Real-time Album Analytics** - `GET /api/v1/albums/{id}/stats`
   - Returns engagement metrics like view counts, likes/dislikes ratio, and popularity ranking
   - Demonstrates efficient read operations using DynamoDB's fast query capabilities

2. **Advanced Search with Faceted Filtering** - `GET /api/v1/albums/search`
   - Powerful search with multiple filter parameters
   - Showcases DynamoDB's secondary indexes for optimized queries

3. **Batch Album Processing** - `POST /api/v1/albums/batch`
   - Allows multiple album uploads at once
   - Demonstrates event-driven architecture through SNS/SQS
   - Triggers asynchronous image processing via Lambda

## Getting Started

### Prerequisites

- Go 1.19 or later
- Docker and Docker Compose
- AWS CLI (for local development with LocalStack)
- Make

### Local Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/CS6650-Distributed-Systems/album-store-plus.git
   cd album-store-plus
   ```

2. Run the setup script:
   ```bash
   chmod +x setup_project.sh
   ./setup_project.sh
   ```

3. Start the development environment:
   ```bash
   docker-compose up -d
   ```

4. Build and run the services:
   ```bash
   make build
   make run-command  # In one terminal
   make run-query    # In another terminal
   ```

## API Endpoints

### Command API

- `POST /api/v1/albums` - Create a new album
- `POST /api/v1/albums/batch` - Create multiple albums
- `POST /api/v1/review/{like|dislike}/{albumID}` - Like or dislike an album

### Query API

- `GET /api/v1/albums/{albumID}` - Get album by ID
- `GET /api/v1/albums/{albumID}/stats` - Get album statistics
- `GET /api/v1/albums/search` - Search albums with filtering
- `GET /api/v1/albums/{albumID}/image` - Get album image

## Running Tests

```bash
make test
```

## Deployment

### AWS Deployment with Terraform

1. Initialize Terraform:
   ```bash
   make terraform-init
   ```

2. Plan the infrastructure:
   ```bash
   make terraform-plan
   ```

3. Apply the changes:
   ```bash
   make terraform-apply
   ```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Commit your changes: `git commit -m "Add some feature"`
4. Push to the branch: `git push origin feature/your-feature-name`
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.