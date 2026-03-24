# Gunakan base image golang resmi
FROM golang:1.23-bullseye AS builder

WORKDIR /app

# Install curl dan tools lain yang dibutuhkan
RUN apt-get update && apt-get install -y ca-certificates curl && rm -rf /var/lib/apt/lists/*

# Copy script downloader
COPY scripts/setup_ml.sh ./scripts/setup_ml.sh

# Jalankan script untuk mengunduh ONNX Runtime (linux-x64 .so file)
# dan model ML ke folder ./lib dan ./models
RUN bash ./scripts/setup_ml.sh

# Install Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary dengan mengaktifkan CGO (wajib untuk ONNX Runtime)
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/app/main.go

# ---------------------------------------------------
# STAGE 2: Minimalist Runtime Image (Debian/Ubuntu)
# ---------------------------------------------------
FROM debian:bullseye-slim

WORKDIR /app

# Install dependency C++ yang dibutuhkan ONNX Runtime
RUN apt-get update && apt-get install -y ca-certificates libgomp1 && rm -rf /var/lib/apt/lists/*

# Copy hasil build dari builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/lib ./lib
COPY --from=builder /app/models ./models

# Set ENV agar Go tahu di mana letak library .so ONNX Runtime-nya saat aplikasi berjalan
ENV LD_LIBRARY_PATH=/app/lib:$LD_LIBRARY_PATH

EXPOSE 8080

CMD ["./main"]
