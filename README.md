# Dredge

[![Go Version](https://img.shields.io/github/go-mod/go-version/rofleksey/dredge)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-available-blue.svg)](Dockerfile)

Dredge is a high-performance Twitch chat monitoring service that captures and analyzes messages from specified channels in real-time.
It identifies messages containing specific keywords or substrings and stores them in a PostgreSQL database for further analysis.

## Features

- **Real-time Twitch Chat Monitoring**: Connects to Twitch IRC and monitors multiple channels simultaneously
- **Keyword/Substring Detection**: Identifies messages containing predefined keywords or substrings
- **Persistent Storage**: Stores all captured messages in PostgreSQL with optimized indexing
- **Automatic Token Refresh**: Handles Twitch OAuth token refresh automatically
- **Sentry Integration**: Comprehensive error tracking and monitoring
- **Telegram Notifications**: Sends error alerts to Telegram
- **Migrations System**: Database schema versioning and migration management
- **Docker Support**: Containerized deployment with optimized Alpine image

## Prerequisites

- Go 1.25+
- PostgreSQL database
- Twitch client_id, client_secret and refresh_token

## Docker image
```bash
rofleksey/dredge:latest
```

## Build

### From Source

```bash
git clone <repository-url>
cd dredge
go mod download
make build
./dredge
```

### Using Docker
```bash
docker build -t dredge .
docker run -v $(pwd)/config.yaml:/opt/config.yaml dredge
```

## Configuration
See config_example.yaml file for an example config.
