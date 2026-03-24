# Native ONNX Runtime Library Setup

This directory holds the native ONNX Runtime shared library required by `onnxruntime_go`.

## Download

Go to: https://github.com/microsoft/onnxruntime/releases

Find the latest stable release and download:

| Platform | File to download | Rename to |
|---|---|---|
| Linux x86_64 | `onnxruntime-linux-x64-<version>.tgz` | `libonnxruntime.so` |
| macOS arm64 | `onnxruntime-osx-arm64-<version>.tgz` | `libonnxruntime.dylib` |
| macOS x86_64 | `onnxruntime-osx-x86_64-<version>.tgz` | `libonnxruntime.dylib` |

## Quick Install (Linux)

```bash
ORT_VERSION=1.21.0
wget https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-linux-x64-${ORT_VERSION}.tgz
tar xzf onnxruntime-linux-x64-${ORT_VERSION}.tgz
cp onnxruntime-linux-x64-${ORT_VERSION}/lib/libonnxruntime.so.${ORT_VERSION} ./lib/libonnxruntime.so
```

## Quick Install (macOS)

```bash
ORT_VERSION=1.21.0
wget https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-osx-arm64-${ORT_VERSION}.tgz
tar xzf onnxruntime-osx-arm64-${ORT_VERSION}.tgz
cp onnxruntime-osx-arm64-${ORT_VERSION}/lib/libonnxruntime.${ORT_VERSION}.dylib ./lib/libonnxruntime.dylib
```

## .env Configuration

```env
FACE_VERIFICATION_PROVIDER=onnx
ONNX_MODEL_PATH=./models/arcface_resnet50.onnx
ONNXRUNTIME_SHARED_LIBRARY_PATH=./lib/libonnxruntime.so   # or .dylib on macOS
```
