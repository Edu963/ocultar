FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY apps/slm-engine/python/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY scripts/serve_privacy_filter.py .

ENV PORT=8086
ENV PRIVACY_FILTER_MODEL_PATH=openai/privacy-filter

CMD ["python", "serve_privacy_filter.py"]
