# Stage 1: Compile the Go application using Go 1.25
FROM golang:1.25-alpine AS builder

# Install build dependencies required for your SQLite driver
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency files first from your acm folder
COPY acm/go.mod acm/go.sum ./
RUN go mod download

# Copy the remaining project application source files
COPY acm/ .

# Build the Go binary executable with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server/main.go

# Stage 2: Create a minimal production runner container image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary from the builder workspace
COPY --from=builder /app/main .

# Copy static frontend assets and HTML views layout files 
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Expose network pipeline
EXPOSE 8080

CMD ["./main"]
