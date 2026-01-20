# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**EDA-Lab** is an academic simulation of Event Driven Architecture (EDA) for learning EDA patterns in an enterprise ecosystem. The MVP (Itération 1) implements Pub/Sub with a simulated banking domain.

## Terminology

> **Important**: This project uses two numbering systems:
> - **Itérations (1-8)**: EDA patterns from PDR.MD (1=Pub/Sub, 2=Event Sourcing, etc.)
> - **Phases techniques (0-8)**: Technical build steps from PLAN.MD for MVP construction
>
> **Itération 1 (MVP Pub/Sub)** = **Phases techniques 0-8** from PLAN.MD

## Tech Stack

- **Backend**: Go 1.21+
- **Message Broker**: Confluent Platform (Kafka KRaft mode, no ZooKeeper)
- **Schema Registry**: Confluent Schema Registry with Avro
- **Database**: PostgreSQL 16
- **Frontend**: React + Vite + React Flow + Tailwind CSS + Zustand
- **Observability**: Prometheus + Grafana
- **Containerization**: Docker Compose (Windows 11 / WSL2)

## Architecture

Monorepo with 3 Go microservices (MVP):
- `simulator` - Generates fake banking events at configurable rate
- `bancaire` - Consumes events, persists accounts/transactions to PostgreSQL
- `gateway` - REST API proxy + WebSocket hub for real-time UI updates

Events flow: **Simulator → Kafka → Bancaire**, with Gateway streaming to web-ui via WebSocket.

Kafka topic naming: `<domain>.<entity>.<action>` (e.g., `bancaire.compte.ouvert`)

## Development Commands

```bash
# Infrastructure (Kafka, Schema Registry, PostgreSQL, Prometheus, Grafana)
make infra-up              # Start all infrastructure containers
make infra-down            # Stop all containers
make infra-logs            # View container logs
make infra-clean           # Remove volumes and restart fresh
make test-infra            # Validate infrastructure is healthy

# Kafka operations
make kafka-topics                    # List all topics
make kafka-create-topic TOPIC=name   # Create a specific topic
./scripts/create-topics.sh           # Create all MVP topics

# Schema Registry
./scripts/register-schemas.sh        # Register all Avro schemas

# Go services (from service directory)
cd services/<service-name>
go build ./cmd/...
go test ./...
go test -race ./...
go test -v -run TestName ./path/to/package  # Single test

# Frontend
cd web-ui
npm install
npm run dev     # Dev server on :5173
npm run build   # Production build

# Validation
./scripts/validate-mvp.sh   # Full MVP validation (infra + services + tests)
make test-integration       # Go integration tests
```

## Key Conventions

**Go Service Structure**:
```
services/<name>/
├── cmd/<name>/main.go      # Entry point
├── internal/
│   ├── api/                # HTTP handlers
│   ├── domain/             # Entities
│   ├── handler/            # Kafka event handlers
│   ├── repository/         # PostgreSQL persistence
│   ├── generator/          # (simulator only) Fake data generation
│   └── simulation/         # (simulator only) Simulation manager
├── migrations/             # SQL migrations
├── Dockerfile
└── go.mod
```

**Shared packages** in `pkg/`: config, kafka, database, events, observability

**Avro schemas** in `schemas/<domain>/` with namespace `com.edalab.<domain>.events`

**Tests**: Use `testcontainers-go` for integration tests with real Kafka/PostgreSQL. Build tags: `//go:build integration` or `//go:build e2e`

## Project Iterations (EDA Patterns from PDR.MD)

| Itération | Pattern | Status |
|-----------|---------|--------|
| 1 - MVP | Pub/Sub | Code written, validation pending |
| 2 | Event Sourcing | Planned |
| 3 | CQRS | Planned |
| 4 | Saga Choreography | Planned |
| 5 | Saga Orchestration | Planned |
| 6 | Event Streaming | Planned |
| 7 | Dead Letter Queue | Planned |
| 8 | Outbox Pattern | Planned |

To implement a new iteration: `Implémente l'Itération [N] du projet EDA-Lab selon le PDR.MD et AGENT.MD`

**Current status (Itération 1 MVP)**:
- Code: ~90% written
- Tests: Pending execution (`make test-infra`, `make test-integration`)
- Validation: Run `./scripts/validate-mvp.sh` to complete

## Project Documentation

| File | Purpose |
|------|---------|
| `PDR.MD` | Product Definition Record - Itérations (EDA patterns) specifications |
| `PLAN.MD` | Implementation plan - Phases techniques (0-8) with progress tracking |
| `AGENT.MD` | Agent instructions for implementing iterations |
| `docs/ARCHITECTURE.md` | C4 diagrams and technical architecture |
| `docs/adr/` | Architecture Decision Records |
