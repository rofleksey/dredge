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

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/dredge /app/dredge
COPY config.yaml /app/config.yaml
EXPOSE 8080
CMD ["/app/dredge"]
