# Documentation Technique - AgentMeshKafka

> **Version :** 2.0.0 | **Derniere mise a jour :** Janvier 2026

## Vue d'Ensemble

Ce dossier contient la **documentation technique complete** du projet AgentMeshKafka. La documentation est organisee de maniere progressive, du concept general aux specifications detaillees.

---

## Structure de la Documentation

```
/docs
  ├── 00-Readme.md                  # Index et navigation (ce fichier)
  ├── 01-Architecture.md            # Resume des decisions architecturales (3 ADRs cles)
  ├── 01-ArchitectureDecisions.md   # ADRs detailles (7 decisions techniques)
  ├── 02-DataContracts.md           # Schemas Avro et Topologie Kafka
  ├── 03-AgentSpecs.md              # Specifications des 3 agents
  ├── 04-Setup.md                   # Guide d'installation par phase
  ├── 04-EvaluationStrategie.md     # Diamant de l'evaluation (L1-L4)
  ├── 05-ThreatModel.md             # Modele de menaces et AgentSec
  ├── 06-Plan.md                    # Roadmap et plan d'implementation
  ├── 07-Constitution.md            # Code de conduite et standards
  └── 08-Presentation.pdf           # Slides de presentation (optionnel)
```

---

## Guide de Lecture Recommande

| Objectif | Documents a Lire | Temps Estime |
|----------|------------------|--------------|
| **Comprendre le projet** | 00-Readme → 01-Architecture | 15 min |
| **Demarrer rapidement** | 04-Setup → phase0/README.md | 10 min |
| **Concevoir et etendre** | 01-ArchitectureDecisions → 02-DataContracts → 03-AgentSpecs | 45 min |
| **Tester et valider** | 04-EvaluationStrategie → phase3/README.md | 20 min |
| **Securiser** | 05-ThreatModel → 07-Constitution | 30 min |
| **Planifier** | 06-Plan | 15 min |

---

# AgentMeshKafka

**Implementation d'un Maillage Agentique (Agentic Mesh) resilient propulse par Apache Kafka et les pratiques AgentOps.**

## A propos du projet

**AgentMeshKafka** est un projet academique visant a demontrer la faisabilite et la robustesse de l'**Entreprise Agentique**. Contrairement aux approches monolithiques ou aux chatbots isoles, ce projet implemente une architecture decentralisee ou des agents autonomes collaborent de maniere asynchrone pour executer des processus metiers complexes.

### Concepts Cles

| Concept | Description | Phase |
|---------|-------------|-------|
| **Decouplage Temporel & Spatial** | Backbone evenementiel Kafka | Phase 1+ |
| **AgentOps & Fiabilite** | Diamant de l'Evaluation (L1-L4) | Phase 3+ |
| **Gouvernance des Donnees** | Schema Registry + Avro | Phase 4 |
| **RAG (Retrieval-Augmented Generation)** | ChromaDB pour politiques de credit | Phase 2+ |
| **AgentSec** | Protection contre prompt injection | Toutes |

---

## Architecture du Systeme

L'architecture repose sur **trois piliers fondamentaux**, inspires par la biologie organisationnelle :

### 1. Le Systeme Nerveux (Communication)

> Le coeur du systeme n'est pas l'IA, mais le flux de donnees.

| Aspect | Detail |
|--------|--------|
| **Technologie** | Apache Kafka (KRaft mode) |
| **Patterns** | Event Sourcing, CQRS |
| **Role** | Persistance immuable, communication asynchrone |
| **Topics** | 3 topics principaux (application, risk, decision) |

### 2. Le Cerveau (Cognition)

Les agents sont des entites autonomes utilisant le pattern **ReAct** (Reason + Act), propulses par **Anthropic Claude**.

| Agent | Role | Modele LLM | Temperature |
|-------|------|------------|-------------|
| **Intake** | Validation et normalisation | Claude 3.5 Haiku | 0.0 |
| **Risk** | Analyse de risque + RAG | Claude Sonnet 4 / Opus 4.5 | 0.2 |
| **Decision** | Decision finale | Claude 3.5 Sonnet | 0.1 |

### 3. Le Systeme Immunitaire (Securite & Gouvernance)

| Composant | Fonction |
|-----------|----------|
| **AgentSec** | Validation entrees/sorties, detection prompt injection |
| **Data Contracts** | Schemas Avro stricts (Phase 4) |
| **Zero Trust** | Agents communiquent uniquement via Kafka |

---

## Progression des Phases

Le projet est organise en **5 phases progressives** :

```
Phase 0          Phase 1          Phase 2          Phase 3          Phase 4
   MVP      →     Kafka      →      RAG       →     Tests     →   Production

 Scripts        Events         Knowledge        Quality        Governance
 simples       async          augmented        assurance      complete
```

| Phase | Complexite | Temps Setup | Infrastructure |
|-------|------------|-------------|----------------|
| **0** | * | < 5 min | Aucune |
| **1** | ** | ~15 min | Kafka |
| **2** | *** | ~20 min | Kafka + ChromaDB |
| **3** | *** | ~10 min | Kafka + ChromaDB |
| **4** | **** | ~30 min | Stack complete |

**Pour demarrer :** Consultez [04-Setup.md](./04-Setup.md) ou [../QUICKSTART.md](../QUICKSTART.md).

---

## Scenario de Demonstration

Le projet simule un processus de **Traitement de Demande de Pret Bancaire** :

```
┌─────────────────────────────────────────────────────────────────┐
│                    FLUX DE TRAITEMENT                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Demande JSON                                                 │
│         │                                                        │
│         ▼                                                        │
│  ┌─────────────┐     finance.loan.application.v1                │
│  │ INTAKE      │ ────────────────────────────────►              │
│  │ Agent       │                                   │             │
│  └─────────────┘                                   ▼             │
│                                             ┌─────────────┐      │
│                                             │    RISK     │      │
│                  risk.scoring.result.v1     │    Agent    │      │
│                  ◄──────────────────────────│  (+ RAG)    │      │
│         │                                   └─────────────┘      │
│         ▼                                                        │
│  ┌─────────────┐                                                │
│  │ DECISION    │     finance.loan.decision.v1                   │
│  │ Agent       │ ────────────────────────────────►              │
│  └─────────────┘                                                │
│                                                                  │
│  Resultat: APPROVED | REJECTED | MANUAL_REVIEW_REQUIRED         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Installation Rapide

### Option A : Demarrage Ultra-Rapide (Phase 0)

```bash
cd phase0
pip install -r requirements.txt
# Creer .env avec ANTHROPIC_API_KEY
python main.py
```

**Temps :** < 5 minutes | **Prerequis :** Python 3.10+, cle API Anthropic

### Option B : Version Evenementielle (Phase 1+)

```bash
cd phase1  # ou phase2, phase3, phase4
docker-compose up -d
pip install -r requirements.txt
python scripts/init_kafka.py
# Lancer les agents dans des terminaux separes
```

**Pour les details :** Consultez [04-Setup.md](./04-Setup.md).

---

## Strategie d'Evaluation (AgentOps)

Le projet applique le **Diamant de l'Evaluation Agentique** :

| Niveau | Type | Description | Phase |
|--------|------|-------------|-------|
| **L1** | Unitaire | Tests Python deterministes | Phase 3+ |
| **L2** | Cognitif | LLM-Juge (factualite, conformite) | Phase 3+ |
| **L3** | Adversite | Prompt injection, red teaming | Phase 4 |
| **L4** | Simulation | 50+ demandes variees | Phase 4 |

```bash
# Lancer les tests
pytest tests/unit/ -v      # L1
pytest tests/evaluation/ -v # L2
```

**Pour les details :** Consultez [04-EvaluationStrategie.md](./04-EvaluationStrategie.md).

---

## Securite (AgentSec)

Le modele de securite est detaille dans [05-ThreatModel.md](./05-ThreatModel.md).

**Principes cles :**

1. **Zero Trust Network** : Agents communiquent uniquement via Kafka
2. **Defense en profondeur** : 6 couches de validation
3. **Delimiteurs XML** : Isolation des donnees utilisateur
4. **Dead Letter Queue** : Quarantaine des messages invalides

---

## Documentation Complete

| Document | Description |
|----------|-------------|
| [01-Architecture.md](./01-Architecture.md) | Resume des 3 ADRs cles |
| [01-ArchitectureDecisions.md](./01-ArchitectureDecisions.md) | 7 ADRs detailles (Kafka, Avro, ReAct, Config, Modeles) |
| [02-DataContracts.md](./02-DataContracts.md) | Schemas Avro et topologie Kafka |
| [03-AgentSpecs.md](./03-AgentSpecs.md) | Specifications cognitives des agents |
| [04-Setup.md](./04-Setup.md) | Guide d'installation par phase |
| [04-EvaluationStrategie.md](./04-EvaluationStrategie.md) | Diamant de l'Evaluation |
| [05-ThreatModel.md](./05-ThreatModel.md) | Modele de menaces STRIDE + AgentSec |
| [06-Plan.md](./06-Plan.md) | Roadmap et plan d'implementation |
| [07-Constitution.md](./07-Constitution.md) | Standards et regles fondamentales |

---

## Ressources Complementaires

### Dans ce Repository

| Ressource | Description |
|-----------|-------------|
| [../README.md](../README.md) | Vue d'ensemble du projet |
| [../QUICKSTART.md](../QUICKSTART.md) | Demarrage rapide Phase 0 |
| [../PHASES.md](../PHASES.md) | Guide de progression entre phases |
| [../notebooks/](../notebooks/) | Tutoriels Jupyter interactifs |
| [../examples/](../examples/) | Scripts d'exemple progressifs |

### Externes

| Ressource | Lien |
|-----------|------|
| Documentation Anthropic | [docs.anthropic.com](https://docs.anthropic.com) |
| Apache Kafka | [kafka.apache.org](https://kafka.apache.org) |
| LangChain | [python.langchain.com](https://python.langchain.com) |
| ChromaDB | [docs.trychroma.com](https://docs.trychroma.com) |

---

## Auteurs et References

Projet realise dans le cadre academique sur l'architecture des systemes agentiques.

| Aspect | Detail |
|--------|--------|
| **Auteur** | Andre-Guy Bruneau |
| **Sujet** | Architecture - Maillage Agentique et AgentOps |
| **Stack IA** | Anthropic Claude (Opus 4.5, Sonnet, Haiku) |
| **Outils** | Claude Code, LangChain, LangGraph |
| **Licence** | MIT |

---

## Navigation

| Ce document | Suivant |
|:-----------:|--------:|
| **Index de la documentation** | [01-Architecture.md](./01-Architecture.md) |
