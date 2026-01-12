# Plan d'Implementation (Roadmap)

> **Version :** 1.0.0 | **Statut :** Approuve | **Derniere revision :** Janvier 2026
>
> **Documents connexes :** [05-ThreatModel.md](./05-ThreatModel.md) | [07-Constitution.md](./07-Constitution.md)

Ce document presente le **plan d'implementation detaille** du projet AgentMeshKafka, organise en phases progressives. Chaque phase est autonome et livrable.

---

## 1. Vue d'Ensemble du Projet

### 1.1 Objectif Final

Demonstrer la faisabilite d'une **architecture agentique enterprise** (Agentic Mesh) utilisant :
- Communication asynchrone via Kafka
- Gouvernance des donnees via Avro/Schema Registry
- Intelligence augmentee via RAG
- Validation rigoureuse via le Diamant d'Evaluation

### 1.2 Progression des Phases

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROGRESSION DES PHASES                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Phase 0     Phase 1      Phase 2      Phase 3      Phase 4     │
│    MVP    →   Kafka    →    RAG     →   Tests   →  Production   │
│                                                                  │
│  [Scripts]   [Events]    [Knowledge]  [Quality]   [Governance]  │
│                                                                  │
│  Complexite: * → ** → *** → *** → ****                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Phase 0 : MVP Fonctionnel

### 2.1 Objectifs

- Demontrer le flux complet Intake → Risk → Decision
- Valider l'approche ReAct avec Claude
- Etablir les modeles de donnees de base

### 2.2 Livrables

| Livrable | Description | Statut |
|----------|-------------|--------|
| `phase0/main.py` | Orchestrateur sequentiel | Termine |
| `phase0/agents/*.py` | 3 agents Python simples | Termine |
| `phase0/models.py` | Modeles Pydantic | Termine |
| `phase0/README.md` | Documentation | Termine |

### 2.3 Criteres d'Acceptation

- [ ] Execution complete en < 30 secondes
- [ ] Sortie structuree (APPROVED/REJECTED/MANUAL_REVIEW)
- [ ] Score de risque entre 0 et 100
- [ ] Cout par execution < $0.10

### 2.4 Dependances

- Python 3.10+
- Cle API Anthropic
- Aucune infrastructure

---

## 3. Phase 1 : Communication Evenementielle

### 3.1 Objectifs

- Decouplage temporel et spatial via Kafka
- Agents independants et scalables
- Communication asynchrone

### 3.2 Livrables

| Livrable | Description | Statut |
|----------|-------------|--------|
| `docker-compose.yml` | Kafka KRaft | Termine |
| `src/shared/kafka_client.py` | Wrapper Producer/Consumer | Termine |
| `scripts/init_kafka.py` | Creation des topics | Termine |
| `config.yaml` | Configuration centralisee | Termine |
| Agents adaptes | Consommation/Production Kafka | Termine |

### 3.3 Architecture Technique

```
┌──────────────┐     ┌──────────────────────┐     ┌──────────────┐
│ Intake Agent │────>│ finance.loan.app.v1  │<────│  Risk Agent  │
└──────────────┘     └──────────────────────┘     └──────────────┘
                                                          │
                     ┌──────────────────────┐             │
                     │ risk.scoring.result  │<────────────┘
                     └──────────────────────┘
                                │
                     ┌──────────────────────┐     ┌──────────────┐
                     │ finance.loan.decision│<────│Decision Agent│
                     └──────────────────────┘     └──────────────┘
```

### 3.4 Criteres d'Acceptation

- [ ] Agents demarrent independamment
- [ ] Messages persistes dans Kafka
- [ ] Replay possible depuis le debut
- [ ] Consumer Groups fonctionnels

### 3.5 Dependances

- Phase 0 complete
- Docker & Docker Compose

---

## 4. Phase 2 : Intelligence Augmentee (RAG)

### 4.1 Objectifs

- Enrichir l'Agent Risk avec une base de connaissances
- Implementer la recherche semantique
- Valider les decisions contre les politiques

### 4.2 Livrables

| Livrable | Description | Statut |
|----------|-------------|--------|
| ChromaDB | Ajout au docker-compose | Termine |
| `src/shared/rag_client.py` | Client ChromaDB | Termine |
| `scripts/ingest_policies.py` | Ingestion documents | Termine |
| `data/credit_policy.md` | Politique de credit | Termine |
| Risk Agent RAG | Integration recherche | Termine |

### 4.3 Flux RAG

```
┌─────────────────────────────────────────────────────────────┐
│                      FLUX RAG                                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. INGESTION                                                │
│     Document → Chunking → Embedding → ChromaDB              │
│                                                              │
│  2. RECHERCHE                                                │
│     Query → Embedding → Similarite → Top-K Documents        │
│                                                              │
│  3. AUGMENTATION                                             │
│     Prompt = System + Retrieved Docs + User Data            │
│                                                              │
│  4. GENERATION                                               │
│     LLM → Reponse informee par les politiques               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 4.4 Criteres d'Acceptation

- [ ] Documents ingerables via script
- [ ] Recherche retourne des resultats pertinents
- [ ] Agent Risk cite les politiques consultees
- [ ] Latence RAG < 500ms

### 4.5 Dependances

- Phase 1 complete
- Modele d'embeddings (sentence-transformers)

---

## 5. Phase 3 : Tests et Validation

### 5.1 Objectifs

- Implementer le Diamant d'Evaluation (L1-L2)
- Valider la qualite des decisions
- Etablir une baseline de performance

### 5.2 Livrables

| Livrable | Description | Statut |
|----------|-------------|--------|
| `tests/unit/*.py` | Tests unitaires (L1) | Termine |
| `tests/evaluation/*.py` | Tests cognitifs (L2) | Termine |
| `pytest.ini` | Configuration pytest | Termine |
| Golden Dataset | Cas de test references | Termine |

### 5.3 Niveaux de Test

```
┌─────────────────────────────────────────────────────────────┐
│                 DIAMANT D'EVALUATION                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│                        L4                                    │
│                     Simulation                               │
│                    /          \                              │
│                   /            \                             │
│                 L3              L3                           │
│              Adversite       Red Team                        │
│                 |              |                             │
│                 └──────┬───────┘                             │
│                        │                                     │
│                       L2                                     │
│                   Cognitif                                   │
│                   (LLM-Juge)                                │
│                        │                                     │
│                       L1                                     │
│                    Unitaire                                  │
│                   (Deterministe)                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 5.4 Criteres d'Acceptation

- [ ] Couverture L1 > 80%
- [ ] Score factualite L2 > 8.0/10
- [ ] Tous les tests passent en CI
- [ ] Temps d'execution < 5 minutes

### 5.5 Dependances

- Phase 2 complete
- pytest et plugins

---

## 6. Phase 4 : Version Production

### 6.1 Objectifs

- Gouvernance des donnees via Schema Registry
- Observabilite complete
- Tests avances (L3-L4)

### 6.2 Livrables

| Livrable | Description | Statut |
|----------|-------------|--------|
| Schema Registry | Ajout au docker-compose | Termine |
| `schemas/*.avsc` | Schemas Avro | Termine |
| `scripts/register_schemas.py` | Enregistrement schemas | Termine |
| Kafka Avro | Serialisation Avro | Termine |
| OpenTelemetry | Tracing distribue | En cours |
| Control Center | Monitoring UI | Optionnel |
| Tests L3 | Adversite | En cours |
| Tests L4 | Simulation | Planifie |

### 6.3 Architecture Production

```
┌─────────────────────────────────────────────────────────────┐
│                ARCHITECTURE PHASE 4                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐                  │
│  │ Intake  │    │  Risk   │    │Decision │   AGENTS         │
│  └────┬────┘    └────┬────┘    └────┬────┘                  │
│       │              │              │                        │
│       ▼              ▼              ▼                        │
│  ┌─────────────────────────────────────┐                    │
│  │         SCHEMA REGISTRY             │   GOUVERNANCE      │
│  │      (Validation Avro)              │                    │
│  └─────────────────────────────────────┘                    │
│       │              │              │                        │
│       ▼              ▼              ▼                        │
│  ┌─────────────────────────────────────┐                    │
│  │           KAFKA CLUSTER             │   COMMUNICATION    │
│  │    (3 topics, replication)          │                    │
│  └─────────────────────────────────────┘                    │
│       │              │              │                        │
│       ▼              ▼              ▼                        │
│  ┌─────────────────────────────────────┐                    │
│  │         OBSERVABILITE               │   MONITORING       │
│  │  (OpenTelemetry + Prometheus)       │                    │
│  └─────────────────────────────────────┘                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 6.4 Criteres d'Acceptation

- [ ] Schemas valides par Schema Registry
- [ ] Traces distribuees fonctionnelles
- [ ] Metriques Prometheus exposees
- [ ] Tests L3 implementes
- [ ] Documentation complete

### 6.5 Dependances

- Phase 3 complete
- Confluent Schema Registry

---

## 7. Jalons et Milestones

| Milestone | Description | Date Cible | Statut |
|-----------|-------------|------------|--------|
| **M1** | Phase 0 - MVP Demo | - | Termine |
| **M2** | Phase 1 - Kafka operationnel | - | Termine |
| **M3** | Phase 2 - RAG fonctionnel | - | Termine |
| **M4** | Phase 3 - Tests L1-L2 | - | Termine |
| **M5** | Phase 4 - Schema Registry | - | Termine |
| **M6** | Phase 4 - Observabilite | - | En cours |
| **M7** | Phase 4 - Tests L3-L4 | - | Planifie |
| **M8** | Documentation complete | - | En cours |

---

## 8. Risques et Mitigations

| Risque | Probabilite | Impact | Mitigation |
|--------|-------------|--------|------------|
| Latence LLM elevee | Moyenne | Moyen | Timeouts + Retry |
| Cout API excessif | Moyenne | Moyen | Modeles adaptes par tache |
| Hallucination agent | Elevee | Eleve | Validation post-LLM |
| Schema drift | Faible | Critique | Schema Registry strict |
| Prompt injection | Elevee | Critique | AgentSec multicouche |

---

## 9. Prochaines Etapes

### Court Terme
1. Finaliser l'observabilite (OpenTelemetry)
2. Implementer les tests d'adversite (L3)
3. Completer la documentation

### Moyen Terme
1. Simulation d'ecosysteme (L4)
2. Rate limiting production
3. mTLS pour Kafka

### Long Terme
1. Multi-tenant support
2. Federation d'agents
3. Apprentissage continu

---

## Navigation

| Precedent | Index | Suivant |
|:---|:---:|---:|
| [05-ThreatModel.md](./05-ThreatModel.md) | [Documentation](./00-Readme.md) | [07-Constitution.md](./07-Constitution.md) |
