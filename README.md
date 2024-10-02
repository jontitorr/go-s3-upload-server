# S3 Upload Service

This service provides functionality for uploading files to Amazon S3 buckets. It is built with Go and uses the Gin web framework.

## Features

- Upload files to specified S3 buckets
- Configurable through environment variables
- Docker support for easy deployment

## Requirements

- Go 1.22+
- Gin web framework
- AWS SDK for Go
- Godotenv

## Setup

1. Clone the repository
2. Create a `.env` file in the root directory with the content according to [Configuration](#configuration):

3. Fill in the appropriate values for each environment variable in the `.env` file

## Usage

### Running locally

Ensure you have Go installed and the `.env` file is properly configured, then run:

```shell
go run main.go
```

### Using Docker

Build and run the Docker image:

```shell
docker build -t s3-upload-service . docker run -p 8080:8080 --env-file .env s3-upload-service
```

## Configuration

The service is configured using environment variables. Make sure to set the following in your `.env` file:

- `API_KEY`: Your API key for authentication
- `CDN_BASE_URL`: Base URL for your CDN
- `ALLOWED_ORIGINS`: Comma-separated list of allowed origins for CORS
- `S3_BUCKET`: Name of your S3 bucket
- `S3_ENDPOINT`: S3 endpoint URL
- `S3_KEY`: Your S3 access key
- `S3_SECRET`: Your S3 secret key
- `S3_REGION`: AWS region for your S3 bucket
- `IP_WHITELIST`: Comma-separated list of IP addresses to allow access to the service (optional)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Credits

This project is built with:

- [Go](https://golang.org/)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [AWS SDK for Go](https://aws.amazon.com/sdk-for-go/)
- [Godotenv](https://github.com/joho/godotenv)
