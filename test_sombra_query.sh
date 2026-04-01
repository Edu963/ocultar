#!/bin/bash
export OCU_MASTER_KEY=1111111111111111111111111111111111111111111111111111111111111111
export SLM_MODEL_PATH=models/mock.gguf
export OCULTAR_DEBUG=true
set -a
source .env
set +a

cd apps/sombra
./sombra > /tmp/sombra.log 2>&1 &
PID=$!
sleep 15

echo "--- Sombra NDA Query ---"
curl -s -X POST http://localhost:8086/query \
  -H "Authorization: Bearer edu" \
  -F "connector=file" \
  -F "model=gemini-flash-latest" \
  -F "source_id=/tmp/saab_rfp.txt" \
  -F "prompt=Based on the provided document, what is the 'Estimated Volume' for 'Production' environment? Show the hosts and CPU-cores."

echo "--- Sombra Log ---"
cat /tmp/sombra.log
kill $PID
