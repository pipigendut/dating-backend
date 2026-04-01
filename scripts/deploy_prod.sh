#!/bin/bash
set -e # Berhenti otomatis jika ada command yang gagal/error

PEM_PATH=$1
SERVER_IP="13.212.117.42"
SERVER_USER="ubuntu"

if [ -z "$PEM_PATH" ]; then
    echo "Error: Please provide the path to your .pem file."
    echo "Usage: ./scripts/deploy_prod.sh /path/to/your/key.pem"
    exit 1
fi

echo "--- 🚀 Preparing Deployment to $SERVER_IP ---"

# 1. Build Local Docker Image
echo "📦 Menarik semua library (vendoring) secara lokal di Macbook Anda (Native ARM) agar tidak nge-crash..."
go mod vendor

echo "🔨 Mem-build Docker Image di komputer lokal (linux/amd64) menggunakan Dockerfile.prod..."
docker buildx build \
  --platform linux/amd64 \
  -f Dockerfile.prod \
  -t dating_backend_prod:latest \
  --load \
  .

echo "📦 Menyimpan Docker Image ke file .tar (bisa memakan waktu beberapa saat)..."
docker save dating_backend_prod:latest -o dating_backend_prod.tar

# Pastikan file tar benar-benar ada sebelum upload
if [ ! -f "dating_backend_prod.tar" ]; then
    echo "❌ Error: File dating_backend_prod.tar gagal dibuat. Proses dibatalkan."
    exit 1
fi

# 2. Sync files to server
echo "📦 Syncing files and Docker image to server..."
rsync -avz -e "ssh -i $PEM_PATH -o StrictHostKeyChecking=no" \
    --exclude '.git' \
    --exclude 'node_modules' \
    --exclude 'main' \
    --exclude 'backend.log' \
    --exclude 'models' \
    --exclude 'lib' \
    ./ $SERVER_USER@$SERVER_IP:~/app/

# 3. Setup Server & Run Docker
echo "🛠 Checking Docker on server and Loading Image..."
ssh -i $PEM_PATH $SERVER_USER@$SERVER_IP << 'EOF'
    # Install Docker if not exists
    if ! [ -x "$(command -v docker)" ]; then
        echo "Installing Docker..."
        sudo apt-get update
        sudo apt-get install -y docker.io
    fi

    # Install Docker Compose V2 Plugin if missing
    if ! docker compose version > /dev/null 2>&1; then
        echo "Installing Docker Compose V2 from GitHub..."
        DOCKER_CONFIG=${DOCKER_CONFIG:-/usr/local/lib/docker}
        sudo mkdir -p $DOCKER_CONFIG/cli-plugins
        sudo curl -SL https://github.com/docker/compose/releases/download/v2.29.1/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose
        sudo chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
    fi

    cd ~/app

    # FIX: jangan pakai $USER (ini penyebab error)
    sudo usermod -aG docker ubuntu || true

    # 3. Setup Nginx Bootstrap (for SSL challenge)
    if [ ! -f "nginx/nginx.conf" ]; then
        echo "🌐 Setting up Nginx bootstrap for SSL..."
        cp nginx/nginx.bootstrap.conf nginx/nginx.conf
    fi

    echo "📦 Me-load Docker image di server..."
    sudo docker load -i dating_backend_prod.tar

    echo "🛑 Menghentikan service Nginx atau Apache di host yang mungkin menggunakan port 80/443..."
    sudo systemctl stop nginx 2>/dev/null || true
    sudo systemctl disable nginx 2>/dev/null || true
    sudo systemctl stop apache2 2>/dev/null || true
    sudo systemctl disable apache2 2>/dev/null || true

    echo "🚀 Starting services..."

    echo "📥 Mendownload model ONNX dan file Lib (.so) langsung di server Host (menghemat docker image)..."
    bash ./scripts/setup_ml.sh

    # Gunakan V2: "docker compose" bukan "docker-compose"
    sudo docker compose -f docker-compose.prod.yml up -d
EOF

echo "🧹 Membersihkan file .tar lokal dan server..."
ssh -i $PEM_PATH $SERVER_USER@$SERVER_IP "rm -f ~/app/dating_backend_prod.tar"
rm -f dating_backend_prod.tar

echo "--- ✅ Deployment Sync Finished ---"