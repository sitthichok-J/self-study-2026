# build เป็น static แล้วเสิร์ฟด้วย nginx (ไม่รันเป็น root)
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# VITE_API_BASE ถูกฝังตอน build: ที่อยู่ gateway ที่ browser เรียกถึง
ARG VITE_API_BASE=http://localhost:8000/api
ENV VITE_API_BASE=$VITE_API_BASE
RUN npm run build

FROM nginxinc/nginx-unprivileged:stable-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
# image นี้ (nginx-unprivileged) ใช้ uid 101 อยู่แล้ว ประกาศ USER ให้ชัดเจนและให้ตัวสแกนเห็น
USER 101
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=5 CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
