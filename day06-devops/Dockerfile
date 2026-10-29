# syntax=docker/dockerfile:1
# build context = ../day05-go-service (ดู docker-compose.yml)
# ต้องรัน `go get github.com/jackc/pgx/v5 && go mod tidy` ใน day05-go-service ก่อน เพื่อให้มี go.sum

# --- stage 1: build ---
FROM golang:1.24 AS build
WORKDIR /src
# copy go.mod/go.sum ก่อน เพื่อให้ layer ของ dependency ถูก cache จนกว่า go.mod จะเปลี่ยน
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -tags pgx -ldflags="-s -w" -o /out/api ./cmd/api

# --- stage 2: runtime (ไม่มี shell, ไม่มี package manager, รันเป็น non-root) ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
