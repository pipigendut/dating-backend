#!/bin/bash
set -e

echo "========================================="
echo "   Menyiapkan Environment ML untuk Go    "
echo "========================================="

# Direktori tujuan
mkdir -p lib
mkdir -p models

# 1. Download ONNX Runtime berdasarkan OS dan Arsitektur
OS="$(uname -s)"
ARCH="$(uname -m)"
ORT_VERSION="1.17.1" # Versi ONNX Runtime stabil

echo "=> Mendeteksi Sistem Operasi: $OS ($ARCH)"

if [ "$OS" = "Linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        URL="https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-linux-x64-${ORT_VERSION}.tgz"
    elif [ "$ARCH" = "aarch64" ]; then
        URL="https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-linux-aarch64-${ORT_VERSION}.tgz"
    else
        echo "Arsitektur Linux $ARCH tidak terdaftar di script ini."
        exit 1
    fi
elif [ "$OS" = "Darwin" ]; then
    if [ "$ARCH" = "arm64" ]; then
        # Apple Silicon (M1/M2/M3)
        URL="https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-osx-arm64-${ORT_VERSION}.tgz"
    else
        # Intel Mac
        URL="https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-osx-x86_64-${ORT_VERSION}.tgz"
    fi
else
    echo "OS tidak didukung oleh script otomatis ini. Silakan install manual."
    exit 1
fi

echo "=> Mengunduh ONNX Runtime dari GitHub Microsoft..."
curl -L -o ort.tgz "$URL"

echo "=> Mengekstrak Library C/C++..."
tar -xzf ort.tgz

# Temukan folder hasil ekstrak dan copy isinya ke ./lib
EXTRA_FOLDER=$(find . -maxdepth 1 -type d -name "onnxruntime-*" | head -n 1)
cp -r ${EXTRA_FOLDER}/lib/* ./lib/

# Pembersihan file sisa
rm -rf ${EXTRA_FOLDER} ort.tgz

echo "=> ✅ ONNX Runtime Library berhasil disiapkan di folder ./lib!"

# 2. Download Model ArcFace
MODEL_FILE="models/arcface_resnet50.onnx"
if [ ! -f "$MODEL_FILE" ]; then
    MODEL_URL="https://github.com/pipigendut/dating-backend/releases/download/v1.0.0/arcface_resnet50.onnx"
    echo "=> Mengunduh Model ONNX ArcFace dari GitHub Releases..."
    curl -L -o "$MODEL_FILE" "$MODEL_URL"
else
    echo "=> ✅ Model $MODEL_FILE sudah ada. Melewati unduhan."
fi

echo "========================================="
echo "   Environment ML Siap Digunakan!        "
echo "========================================="
