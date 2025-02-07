FROM python:3.9-alpine
WORKDIR /app

ARG HOTPOD_VERSION=latest

RUN wget https://github.com/niradler/hotpod/releases/download/${HOTPOD_VERSION}/hotpod-linux-amd64 -O /app/hotpod && \
    chmod +x /app/hotpod

CMD ["./hotpod", "-command", "python -m http.server 8000", "-port", "8080", "-keepAlive", "-host", ""]