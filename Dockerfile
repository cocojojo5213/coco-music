# frontend
FROM node:24-bookworm-slim AS web
WORKDIR /web
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# backend
FROM golang:1.22-bookworm AS api
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/coco-music ./cmd/server

# runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=api /out/coco-music /app/coco-music
COPY --from=web /web/dist /app/frontend/dist
ENV ADDR=:18280 \
    STATIC_DIR=/app/frontend/dist \
    PUBLIC_ORIGIN=https://music.52131415.xyz
# COCO_PLAY_BASE / UPSTREAM_PUBLIC must be provided at runtime
EXPOSE 18280
CMD ["/app/coco-music"]
