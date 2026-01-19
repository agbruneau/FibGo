# CHANGELOG - EDA-Lab

Toutes les modifications notables de ce projet seront documentées dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.0.0/),
et ce projet adhère au [Versioning Sémantique](https://semver.org/lang/fr/).

---

## [Non publié]

### Ajouté
- PDR.MD : Spécifications complètes du projet (17 sections)
- PLAN.MD : Plan d'implémentation détaillé (8 phases, 19 étapes, 51 sous-étapes)
- TODO.MD : Liste de tâches exhaustive (~400 tâches)
- CLAUDE.MD : Guide pour Claude Code
- CHANGELOG.MD : Journal des modifications

### Structure du projet
- 7 microservices Go planifiés (bancaire, assurance-personne, assurance-dommage, client-360, orchestrator, simulator, gateway)
- Interface Web React + React Flow
- Stack observabilité (Prometheus, Grafana, Jaeger, Loki)
- Infrastructure Docker Compose (Kafka KRaft, Schema Registry, PostgreSQL)

---

## [0.1.0] - 2026-01-19 - Phase de spécification

### Documentation initiale

#### PDR.MD (Product Definition Record)
- Section 1 : Vue d'ensemble et objectifs d'apprentissage
- Section 2 : Domaines métier simulés (Bancaire, Assurance Personne, Assurance Dommage)
- Section 3 : Catalogue des 8 patrons EDA
- Section 4 : Catalogue des 24 événements
- Section 5 : Architecture technique et stack
- Section 6 : Convention de nommage Kafka (avec règles de correspondance)
- Section 7 : Interface utilisateur
- Section 8 : Configuration des scénarios YAML
- Section 9 : Chaos Engineering (5 scénarios)
- Section 10 : Structure monorepo
- Section 11 : Stratégie de tests (TDD)
- Section 12 : Gestion des erreurs et résilience (retry, DLQ, idempotence, circuit breaker)
- Section 13 : Documentation (ADR)
- Section 14 : Environnement d'exécution
- Section 15 : Roadmap des itérations
- Section 16 : Glossaire
- Section 17 : Références

#### PLAN.MD
- Phase 0 : Infrastructure Docker Compose
- Phase 1 : Fondations Go (packages partagés)
- Phase 2 : Schémas Avro
- Phase 3 : Service Simulator
- Phase 4 : Service Bancaire
- Phase 5 : Service Gateway
- Phase 6 : Observabilité
- Phase 7 : Web UI React
- Phase 8 : Intégration et tests de performance

#### TODO.MD
- Tâches détaillées pour chaque sous-étape
- Checklist de validation MVP
- Tableau de progression

### Optimisations appliquées (2026-01-19)
- Ajout de Schema Registry explicite dans CLAUDE.MD
- Clarification Jaeger/Loki prévu pour Itération 2 dans PDR.MD
- Ajout du workspace Go (`go.work`) pour monorepo
- Ajout de la configuration CORS pour Gateway
- Ajout de tous les topics Kafka bancaires (7 topics)
- Ajout section "Gestion des erreurs et résilience" (Section 12)
- Unification du nommage événements/topics avec règles explicites
- Ajout des tests de performance (Phase 8.1.3)
- Renumérotation des sections PDR.MD (12-17)

---

## Roadmap

### MVP (Itération 1) - En cours
- [ ] Patron Pub/Sub
- [ ] Domaine Bancaire (7 événements)
- [ ] Services : Simulator, Bancaire, Gateway
- [ ] Observabilité : Prometheus + Grafana
- [ ] Web UI minimale

### Itération 2 - Planifié
- [ ] Domaines Assurance Personne + Dommage
- [ ] Event Sourcing
- [ ] Jaeger + Loki

### Itération 3 - Planifié
- [ ] CQRS
- [ ] Client 360
- [ ] Chaos Engineering

### Itérations suivantes - Planifié
- [ ] Saga Choreography
- [ ] Saga Orchestration
- [ ] Event Streaming (Kafka Streams)
- [ ] Dead Letter Queue
- [ ] Outbox Pattern

---

## Conventions de commit

```
feat: Nouvelle fonctionnalité
fix: Correction de bug
docs: Documentation
style: Formatage (pas de changement de code)
refactor: Refactorisation
test: Ajout de tests
chore: Maintenance
```

---

**Maintenu par :** Architecte de domaine - Interopérabilité des systèmes
**Outil :** Claude Code (claude.ai/code)
