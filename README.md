# Go AI Proxy

A high-performance Go proxy for Azure OpenAI and AWS Bedrock services, featuring:
- **Unified Interface**: Route requests to both providers.
- **Security**: JWT Authentication.
- **Observability**: Full request/response logging to MongoDB (including streams).
- **Streaming Support**: Handles server-sent events (SSE) properly.

## Prerequisites
- Go 1.21+
- MongoDB instance

## Setup

1. **Install Dependencies**
   ```bash
   go mod tidy
   ```

2. **Configuration**
   Copy `.env.example` to `.env` and fill in your details:
   ```bash
   cp .env.example .env
   ```
   
   Variables:
   - `JWT_SECRET`: Secret key for signing/verifying tokens.
   - `MONGO_URI`: MongoDB connection string.
   - `AZURE_OPENAI_ENDPOINT`: Your Azure OpenAI base URL (e.g., `https://my-resource.openai.azure.com/`).
   - `AZURE_OPENAI_API_KEY`: Your Azure API Key.
   - `AWS_REGION`: AWS Region (e.g., `us-east-1`).
   - `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`: AWS Credentials (or use `~/.aws/credentials`).

3. **Run Locally**
   ```bash
   go run main.go
   ```

## Docker

You can run the entire stack (Proxy + MongoDB) using Docker Compose.

1. **Configure .env** (as above)
2. **Start Services**
   ```bash
   docker-compose up -d --build
   ```
   The server will be available at `http://localhost:8080`.

## Usage

### Authentication
All requests require a `Authorization: Bearer <token>` header. The token must be signed with `JWT_SECRET`.

### Azure OpenAI
Forward requests to Azure by prefixing with `/azure`.
endpoint: `POST /azure/openai/deployments/{deployment}/chat/completions?api-version=2023-05-15`

### AWS Bedrock
Supported actions: `invoke`, `invoke-with-stream`, `converse`, `converse-stream`.
endpoint: `POST /bedrock/{action}/{modelId}`

Example: `POST /bedrock/invoke/anthropic.claude-v2`
