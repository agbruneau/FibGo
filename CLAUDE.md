# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**EDA-Lab** is an academic application for simulating interoperability in an enterprise ecosystem based on Event Driven Architecture (EDA). The project is in pre-development phase - see `PDR.MD` for complete specifications.

## Tech Stack (Planned)

- **Backend**: Go
- **Message Broker**: Confluent Platform Community (KRaft mode)
- **Schema Registry**: Confluent Schema Registry
- **Serialization**: Apache Avro (with Schema Registry for governance)
- **Database**: PostgreSQL
- **Frontend**: React + React Flow
- **Observability**: Prometheus, Grafana, Jaeger, Loki
- **Containerization**: Docker Compose (Windows 11 / WSL2)

## Architecture

Monorepo structure with 7 Go microservices:
- `bancaire`, `assurance-personne`, `assurance-dommage` (domain services)
- `client-360`, `orchestrator`, `simulator`, `gateway` (cross-cutting services)

Events flow through Kafka topics following pattern: `<domain>.<entity>.<action>` (e.g., `bancaire.compte.ouvert`)

## Development Commands (To Be Implemented)

```bash
# Infrastructure
docker-compose -f infra/docker-compose.yml up -d
docker-compose -f infra/docker-compose.yml down

# Go services (per service)
cd services/<service-name>
go build ./cmd/...
go test ./...
go test -race ./...

# Single test
go test -v -run TestName ./path/to/package

# Frontend
cd web-ui
npm install
npm run dev
npm test
```

## Key Conventions

- Avro schemas centralized in `schemas/` directory
- Scenario configurations in YAML format under `scenarios/`
- Tests use `testcontainers-go` for integration tests with real Kafka/PostgreSQL
- OpenTelemetry for distributed tracing across services

## Iterative Development

MVP (Iteration 1) focuses on: Pub/Sub pattern, Bancaire domain only, basic observability, minimal Web UI.
Subsequent iterations add: Event Sourcing, CQRS, Sagas, Chaos Engineering (see PDR.MD section 15).

## Project Documentation

| File | Description |
|------|-------------|
| `PDR.MD` | Product Definition Record - Complete specifications |
| `PLAN.MD` | Implementation plan with 51 sub-steps |
| `TODO.MD` | Detailed task checklist |
| `CHANGELOG.MD` | Project changelog |
