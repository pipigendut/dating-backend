FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder
ARG TARGETOS TARGETARCH

WORKDIR /app

# Enable workaround for Mac M1/M2 Rosetta crashes
ENV GODEBUG=asyncpreemptoff=1

# Install cross-compiler untuk CGO ke AMD64, serta tools lain
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    bash \
    gcc-x86-64-linux-gnu \
    g++-x86-64-linux-gnu \
    libc6-dev-amd64-cross \
    && rm -rf /var/lib/apt/lists/*

# Copy script downloader
COPY scripts/setup_ml.sh ./scripts/setup_ml.sh

# Force setup_ml.sh untuk pakai arsitektur target, bukan host (karena host arm64 tapi butuh lib amd64)
ENV FORCE_OS=Linux
ENV FORCE_ARCH=x86_64
RUN bash ./scripts/setup_ml.sh

# Alih-alih `go mod download` yang nge-crash di emulator,
# kita copy langsung seluruh kode dan folder `vendor/` dari lokal Mac Anda.
COPY . .

# Build binary menggunakan cross compiler TANPA emulator
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ go build -mod=vendor -o main ./cmd/app/main.go

# ---------------------------------------------------
# STAGE 2: Minimalist Runtime Image (Debian/Ubuntu)
# ---------------------------------------------------
FROM debian:bookworm-slim

WORKDIR /app

# Dependency runtime ONNX
RUN apt-get update && apt-get install -y \
  ca-certificates \
  libgomp1 \
  && rm -rf /var/lib/apt/lists/*

# Copy hasil build
COPY --from=builder /app/main .
COPY --from=builder /app/lib ./lib
COPY --from=builder /app/models ./models

# Set path library
ENV LD_LIBRARY_PATH=/app/lib:$LD_LIBRARY_PATH

EXPOSE 8080

CMD ["./main"]