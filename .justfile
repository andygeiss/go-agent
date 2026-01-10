set dotenv-load := true
set shell := ["bash", "-cu"]

# ======================================
# Variables
# ======================================
# User running the command (defaults to 'user')

app_user := env("USER", "user")

# Docker image name: {user}/{app}:latest

app_image := app_user + "/" + env("APP_SHORTNAME", "app") + ":latest"

# ======================================
# Aliases - Quick shortcuts
# ======================================

alias b := build
alias u := up
alias d := down
alias t := test

# ======================================
# Bench - Run performance benchmarks
# ======================================
# Runs all benchmarks in pkg/agent with memory allocation stats
# Use for performance regression testing and optimization
#
# Usage:
#   just bench                    # Run all benchmarks
#   go test -bench=. -benchmem -count=5 ./pkg/agent/  # Multiple iterations

bench:
    @echo "Running benchmarks..."
    @go test -bench=. -benchmem ./pkg/agent/

# ======================================
# Build - Create Docker image
# ======================================
# Builds the application container image.
#
# Notes:
# - This template uses `podman build` for image builds.
# - `up`/`down` use `docker-compose` to run the dev stack.
# - If you prefer Docker for builds, replace `podman build` with `docker build`.

build:
    @echo "Building image: {{ app_image }}"
    @podman build \
      -t {{ app_image }} \
      -f Dockerfile .

# ======================================
# Down - Stop Docker Compose services
# ======================================
# Stops and removes all containers defined in docker-compose.yml
# Loads environment from .env file (required for variable interpolation)
# Note: `.env` is expected to be a local file (copy from `.env.example`).

down:
    @docker-compose --env-file .env down

# ======================================
# Fmt - Format Go code
# ======================================
# Formats Go source files using golangci-lint formatters
# Modifies files in place
#
# Requirements:
# - `golangci-lint` must be installed (brew install golangci-lint)

fmt:
    @golangci-lint fmt ./...

# ======================================
# Lint - Run golangci-lint
# ======================================
# Runs golangci-lint to check code quality and style (read-only)
# Uses default configuration or .golangci.yml if present
#
# Requirements:
# - `golangci-lint` must be installed (brew install golangci-lint)

lint:
    @golangci-lint run ./...

# ======================================
# Profile - CPU profiling for PGO
# ======================================
# Runs go test benchmarks with CPU profiling for Profile-Guided Optimization
# Generates cpuprofile.pprof and cpuprofile.svg in the repo root
#
# Requirements:
# - `go` must be on PATH
# - Graphviz (`dot`) must be installed for SVG generation (brew install graphviz)
#
# Usage:
#   just profile              # Run benchmarks and generate profile
#   go build -pgo=cpuprofile.pprof ./cmd/cli  # Build with PGO
#
# Output:
# - cpuprofile.pprof: CPU profile for PGO builds
# - cpuprofile.svg: Visual flame graph of CPU usage

profile:
    @echo "Running benchmarks with CPU profiling..."
    @go test -bench=. -benchtime=10s -cpuprofile=cpuprofile.pprof ./cmd/cli/...
    @echo "Generating SVG visualization..."
    @go tool pprof -svg cpuprofile.pprof > cpuprofile.svg
    @echo "Profile written to cpuprofile.pprof"
    @echo "SVG written to cpuprofile.svg"
    
# ======================================
# Run - Execute CLI application locally
# ======================================
# Builds the image then runs the CLI binary from cmd/cli/main.go

run:
    @go run ./cmd/cli/main.go

# ======================================
# Setup - Install dependencies
# ======================================
# Installs required development tools via Homebrew (macOS/Linux)
# Required: just (command runner), golangci-lint (linting)

setup:
    @echo "Installing dependencies via Homebrew..."
    @brew install golangci-lint just
    @echo "Setup complete! You can now use 'just' commands."

# ======================================
# Test - Run unit tests with coverage
# ======================================
# Runs all tests in internal/ with coverage profiling
# Outputs coverage percentage and generates coverage.pprof

test:
    @echo "Running Go tests..."
    @go test -v -coverprofile=coverage.pprof ./internal/...
    @echo "total coverage: $(go tool cover -func=coverage.pprof | grep total | awk '{print $3}')"

# ======================================
# Test Integration - Run integration tests
# ======================================
# Runs integration tests that require external services (e.g., LM Studio)
# These tests are tagged with //go:build integration and are skipped by default
#
# Requirements:
# - LM Studio (or OpenAI-compatible server) must be running locally (default: http://localhost:1234)
# - Set OPENAI_CHAT_URL, OPENAI_CHAT_MODEL, OPENAI_EMBED_URL, OPENAI_EMBED_MODEL in .env or environment
#
# Usage:
#   just test-integration                    # Run all integration tests
#   just test-integration ./internal/...     # Run integration tests in specific path

test-integration *ARGS='./internal/...':
    @echo "Running integration tests..."
    @echo "OPENAI_CHAT_URL=${OPENAI_CHAT_URL:-http://localhost:1234}"
    @echo "OPENAI_CHAT_MODEL=${OPENAI_CHAT_MODEL:-not set}"
    @echo "OPENAI_EMBED_URL=${OPENAI_EMBED_URL:-http://localhost:1234}"
    @echo "OPENAI_EMBED_MODEL=${OPENAI_EMBED_MODEL:-not set}"
    @echo ""
    @go test -tags=integration -v {{ ARGS }}

# ======================================
# Up - Start Docker Compose services
# ======================================
# Builds image and starts all services
# Steps:
#   1. Build Docker image
#   2. Start all services defined in docker-compose.yml with .env variables
#
# Notes:
# - `.env` is a local-only file (copy from *.example).

up: build
    @docker-compose --env-file .env up -d
