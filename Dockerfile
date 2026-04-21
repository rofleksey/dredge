FROM node:22-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist /src/internal/webui/static
RUN go build -o /out/dredge ./cmd/dredge

FROM alpine
ENV ENVIRONMENT=production
ENV CGO_ENABLED=0
WORKDIR /opt
RUN apk update && \
    apk add --no-cache curl ca-certificates && \
    update-ca-certificates && \
    addgroup -g 65532 -S dredge && \
    adduser -S -u 65532 -G dredge -h /opt -s /sbin/nologin dredge
COPY --from=build /out/dredge /opt/dredge
RUN chown dredge:dredge /opt/dredge
USER dredge:dredge
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["./dredge"]
