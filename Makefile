.PHONY: proto clean test docker-build docker-up docker-down

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p api/proto/racing
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto
	@echo "Proto generation complete!"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f api/proto/racing/*.pb.go
	@echo "Clean complete!"

# Run all tests
test:
	go test ./... -v

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker build -f docker/Dockerfile.engine -t racing-engine:latest .
	docker build -f docker/Dockerfile.ai -t racing-ai:latest .
	docker build -f docker/Dockerfile.gateway -t racing-gateway:latest .
	@echo "Docker build complete!"

# Start all services with docker-compose
docker-up:
	docker-compose up -d

# Stop all services
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f

# Rebuild and restart services
docker-restart: docker-down docker-build docker-up
