# Model & Library Setup for ONNX Face Verification

This directory holds the ONNX model file used for face recognition.

## Download ArcFace ResNet50 ONNX Model

```bash
# Option A — from ONNX Model Zoo (official)
wget https://github.com/onnx/models/raw/main/validated/vision/body_analysis/arcface/model/arcface-resnet50-8.onnx \
  -O arcface_resnet50.onnx

# Option B — if the above is unavailable, use this mirror
pip install onnx
python -c "
import onnx
from onnxmltools import convert_coreml
# or use any pre-trained ArcFace model from insightface
"
```

## Expected Input/Output

| Property | Value |
|---|---|
| Input name | `input.1` |
| Input shape | `[1, 3, 112, 112]` (float32, CHW, normalized [-1,1]) |
| Output name | `683` |
| Output shape | `[1, 512]` (float32 embedding) |

## Download ONNX Runtime Native Library

See `../lib/README.md` for the native `.so` / `.dylib` download instructions.
