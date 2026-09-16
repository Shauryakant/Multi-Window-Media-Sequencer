# Multi-stage Docker build for Go backend (<50MB final image)
FROM golang:alpine AS builder

WORKDIR /app

# Download dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Tidy dependencies and build optimized binary
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

# Minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080
ENV PORT=8080
ENV SYNC_DURATION_SECONDS=5

CMD ["./server"]
