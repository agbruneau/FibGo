# PDR - Product Definition Record
# EDA-Lab : Simulateur d'Architecture Événementielle

**Version :** 1.0
**Date :** 2026-01-19
**Statut :** Approuvé

---

## 1. Vue d'ensemble

### 1.1 Description du projet

**EDA-Lab** est une application académique de simulation d'interopérabilité dans un écosystème d'entreprise fondé sur l'architecture événementielle (EDA - Event Driven Architecture). L'application permet d'apprendre et d'expérimenter l'implémentation et l'exploitation des patrons d'architecture EDA à travers une simulation automatisée de domaines métier interconnectés.

### 1.2 Contexte

| Élément | Description |
|---------|-------------|
| **Type** | Outil d'apprentissage personnel |
| **Utilisateur cible** | Architecte de domaine spécialisé en interopérabilité des systèmes |
| **Approche** | Itérative et incrémentale |
| **Génération** | Claude Code + Google Antigravity |

### 1.3 Objectifs d'apprentissage

| Priorité | Objectif |
|----------|----------|
| **1** | Maîtriser l'implémentation technique des patrons EDA avec Apache Kafka |
| **2** | Comprendre les compromis architecturaux (couplage, latence, résilience) |
| **3** | Visualiser les flux d'événements en temps réel dans une simulation entièrement automatisée |

---

## 2. Domaine métier

### 2.1 Domaines simulés

L'application simule un **écosystème de groupe financier intégré** composé de trois domaines :

| Domaine | Description |
|---------|-------------|
| **Bancaire** | Gestion des comptes, transactions, virements |
| **Assurance de Personne** | Contrats vie, santé, prévoyance |
| **Assurance Dommage** | Contrats auto, habitation, sinistres |

### 2.2 Flux d'interopérabilité

Les trois domaines partagent des événements selon les flux suivants :

```
┌─────────────┐     ┌─────────────────────┐     ┌──────────────────┐
│  BANCAIRE   │◄────┤    CLIENT 360°      ├────►│ ASSURANCE PERS.  │
│             │     │  (Vue unifiée)      │     │                  │
└──────┬──────┘     └─────────────────────┘     └────────┬─────────┘
       │                                                  │
       │            ┌─────────────────────┐              │
       │            │ ASSURANCE DOMMAGE   │              │
       └───────────►│                     │◄─────────────┘
                    └─────────────────────┘
```

| Flux | Description |
|------|-------------|
| **Client 360°** | Vision unifiée du client partagée entre tous les domaines |
| **Paiement de primes** | Prélèvement des primes d'assurance via le compte bancaire |
| **Indemnisation sinistres** | Versement des indemnisations sur le compte bancaire |

---

## 3. Catalogue des patrons EDA

### 3.1 Ordre d'implémentation

| Itération | Patron | Description |
|-----------|--------|-------------|
| **1** | **Producteur/Consommateur (Pub/Sub)** | Pattern de base - publication et consommation d'événements |
| **2** | **Event Sourcing** | Stockage de l'état comme séquence d'événements |
| **3** | **CQRS** | Séparation des modèles de lecture et d'écriture |
| **4** | **Saga Choreography** | Transactions distribuées par événements décentralisés |
| **5** | **Saga Orchestration** | Transactions distribuées par coordinateur central |
| **6** | **Event Streaming** | Traitement continu de flux avec Kafka Streams |
| **7** | **Dead Letter Queue** | Gestion des événements en erreur |
| **8** | **Outbox Pattern** | Garantie de livraison transactionnelle |

---

## 4. Catalogue des événements

### 4.1 Domaine Bancaire

| Événement | Description |
|-----------|-------------|
| `CompteOuvert` | Ouverture d'un nouveau compte bancaire |
| `CompteFerme` | Fermeture d'un compte bancaire |
| `DepotEffectue` | Dépôt d'argent sur un compte |
| `RetraitEffectue` | Retrait d'argent d'un compte |
| `VirementEmis` | Émission d'un virement sortant |
| `VirementRecu` | Réception d'un virement entrant |
| `PaiementPrimeEffectue` | Paiement d'une prime d'assurance |

### 4.2 Domaine Assurance de Personne

| Événement | Description |
|-----------|-------------|
| `ContratSouscrit` | Souscription d'un contrat vie/santé/prévoyance |
| `ContratResilie` | Résiliation d'un contrat |
| `BeneficiaireModifie` | Modification du bénéficiaire d'un contrat |
| `PrimeAppelee` | Appel de prime émis |
| `PrimeEncaissee` | Prime encaissée avec succès |
| `SinistreVieDepose` | Déclaration d'un sinistre vie |
| `IndemnisationVieEmise` | Émission d'une indemnisation vie |

### 4.3 Domaine Assurance Dommage

| Événement | Description |
|-----------|-------------|
| `ContratAutoSouscrit` | Souscription d'un contrat auto |
| `ContratHabitationSouscrit` | Souscription d'un contrat habitation |
| `ContratResilie` | Résiliation d'un contrat dommage |
| `SinistreAutoDepose` | Déclaration d'un sinistre auto |
| `SinistreHabitationDepose` | Déclaration d'un sinistre habitation |
| `ExpertisePlanifiee` | Planification d'une expertise |
| `IndemnisationDommageEmise` | Émission d'une indemnisation dommage |

### 4.4 Domaine Client 360

| Événement | Description |
|-----------|-------------|
| `ProfilClientCree` | Création d'un profil client unifié |
| `ProfilClientMisAJour` | Mise à jour du profil client |
| `VueUnifieeGeneree` | Génération de la vue 360° du client |

---

## 5. Architecture technique

### 5.1 Stack technologique

| Composant | Technologie | Justification |
|-----------|-------------|---------------|
| **Message Broker** | Confluent Platform Community (KRaft) | Standard entreprise, Schema Registry intégré |
| **Sérialisation** | Apache Avro + Schema Registry | Gouvernance des schémas, évolution contrôlée |
| **Langage backend** | Go | Performance, légèreté des conteneurs |
| **Persistance** | PostgreSQL | Polyvalent, ACID, adapté pour Event Sourcing et Outbox |
| **Frontend** | React + React Flow | Écosystème mature, visualisation de flux interactive |
| **Conteneurisation** | Docker Compose | Simplicité, environnement local tout-en-un |

### 5.2 Stack d'observabilité

| Fonction | Outil | Description |
|----------|-------|-------------|
| **Métriques** | Prometheus + Grafana | Dashboards personnalisables, métriques Kafka |
| **Traçage distribué** | Jaeger | Suivi des événements à travers les services (OpenTelemetry) |
| **Logs centralisés** | Loki + Grafana | Agrégation des logs, intégré à Grafana |

### 5.3 Architecture des services

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              EDA-LAB                                     │
├─────────────────────────────────────────────────────────────────────────┤
│  SERVICES MÉTIER                                                        │
│  ┌─────────────┐  ┌──────────────────┐  ┌──────────────────┐           │
│  │  bancaire   │  │ assurance-pers.  │  │ assurance-domm.  │           │
│  └─────────────┘  └──────────────────┘  └──────────────────┘           │
├─────────────────────────────────────────────────────────────────────────┤
│  SERVICES TRANSVERSES                                                   │
│  ┌─────────────┐  ┌──────────────────┐  ┌──────────────────┐           │
│  │ client-360  │  │   orchestrator   │  │    simulator     │           │
│  └─────────────┘  └──────────────────┘  └──────────────────┘           │
│                   ┌──────────────────┐                                  │
│                   │     gateway      │  (API REST + WebSocket)          │
│                   └──────────────────┘                                  │
├─────────────────────────────────────────────────────────────────────────┤
│  INFRASTRUCTURE                                                         │
│  ┌───────────┐ ┌─────────────┐ ┌────────────┐ ┌─────────────────┐      │
│  │   Kafka   │ │  Schema Reg │ │ PostgreSQL │ │ Prometheus/Graf │      │
│  │  (KRaft)  │ │  (Confluent)│ │            │ │ Jaeger/Loki     │      │
│  └───────────┘ └─────────────┘ └────────────┘ └─────────────────┘      │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.4 Description des services

| Service | Responsabilité |
|---------|----------------|
| `bancaire` | Gestion des comptes, transactions, virements |
| `assurance-personne` | Gestion des contrats et sinistres vie/santé/prévoyance |
| `assurance-dommage` | Gestion des contrats et sinistres auto/habitation |
| `client-360` | Agrégation de la vue client unifiée |
| `orchestrator` | Coordination des sagas et transactions distribuées |
| `simulator` | Génération automatique d'événements fictifs |
| `gateway` | API REST pour le contrôle + WebSocket pour temps réel |

---

## 6. Convention de nommage Kafka

### 6.1 Format des topics

**Pattern :** `<domaine>.<entite>.<action>`

**Convention de correspondance Événement → Topic :**

| Événement (PascalCase) | Topic (kebab-case) |
|------------------------|-------------------|
| `CompteOuvert` | `bancaire.compte.ouvert` |
| `DepotEffectue` | `bancaire.depot.effectue` |
| `PaiementPrimeEffectue` | `bancaire.paiement-prime.effectue` |

**Règles :**
- Les mots composés dans l'entité utilisent le tiret (`paiement-prime`)
- L'action est toujours en minuscules (`ouvert`, `effectue`, `emis`)
- Le domaine correspond au service émetteur

### 6.2 Catalogue des topics

| Domaine | Topics |
|---------|--------|
| **Bancaire** | `bancaire.compte.ouvert`, `bancaire.compte.ferme`, `bancaire.depot.effectue`, `bancaire.retrait.effectue`, `bancaire.virement.emis`, `bancaire.virement.recu`, `bancaire.paiement-prime.effectue` |
| **Assurance Personne** | `assurance-personne.contrat.souscrit`, `assurance-personne.contrat.resilie`, `assurance-personne.beneficiaire.modifie`, `assurance-personne.prime.appelee`, `assurance-personne.prime.encaissee`, `assurance-personne.sinistre.depose`, `assurance-personne.indemnisation.emise` |
| **Assurance Dommage** | `assurance-dommage.contrat-auto.souscrit`, `assurance-dommage.contrat-habitation.souscrit`, `assurance-dommage.contrat.resilie`, `assurance-dommage.sinistre-auto.depose`, `assurance-dommage.sinistre-habitation.depose`, `assurance-dommage.expertise.planifiee`, `assurance-dommage.indemnisation.emise` |
| **Client 360** | `client.profil.cree`, `client.profil.mis-a-jour`, `client.vue-unifiee.generee` |
| **Système** | `system.dlq`, `system.saga.events` |

---

## 7. Interface utilisateur

### 7.1 Web UI - Contrôle de simulation

| Fonctionnalité | Description |
|----------------|-------------|
| **Start/Stop** | Boutons pour démarrer et arrêter la simulation |
| **Sélection de scénario** | Choix parmi les scénarios YAML prédéfinis |
| **Paramètres dynamiques** | Ajustement du débit, durée, etc. à la volée |
| **Déclenchement chaos** | Boutons pour activer les scénarios de défaillance |

### 7.2 Dashboard de visualisation (React Flow)

| Fonctionnalité | Description |
|----------------|-------------|
| **Graphe des services** | Visualisation des nodes (services) et edges (flux) |
| **Animation temps réel** | Événements circulant entre services avec animation |
| **Compteurs live** | Nombre d'événements par topic |
| **État des services** | Indicateurs de santé (vert/orange/rouge) |

### 7.3 Grafana Dashboards

| Dashboard | Métriques |
|-----------|-----------|
| **Kafka Overview** | Messages/sec, lag consommateurs, partitions |
| **Services Health** | CPU, mémoire, latence par service |
| **Tracing** | Durée des transactions, erreurs |
| **Business Metrics** | Événements par domaine, flux d'interopérabilité |

---

## 8. Configuration des scénarios

### 8.1 Format de configuration

Les scénarios sont définis en **YAML** et modifiables via l'interface Web.

### 8.2 Structure d'un scénario

```yaml
scenario:
  name: "flux-paiement-prime"
  description: "Simulation du flux de paiement de prime d'assurance"
  duration: 300s

  producers:
    - domain: bancaire
      event: compte-ouvert
      rate: 10/s
      data_generator: fake-client

    - domain: assurance-personne
      event: contrat-souscrit
      rate: 5/s
      depends_on: bancaire.compte.ouvert

    - domain: assurance-personne
      event: prime-appelee
      rate: 5/s
      delay: 10s

  consumers:
    - service: bancaire
      topics:
        - assurance-personne.prime.appelee
      action: paiement-prime

  assertions:
    - topic: bancaire.paiement.prime
      min_count: 100
      max_latency: 500ms

  chaos:
    - type: consumer-lag
      target: service-bancaire
      trigger_at: 60s
      duration: 30s
```

### 8.3 Scénarios prédéfinis

| Scénario | Description | Flux exercé |
|----------|-------------|-------------|
| `flux-client-360` | Création de clients multi-domaines | Vision 360° |
| `flux-paiement-prime` | Appel et paiement de primes | Paiement primes |
| `flux-sinistre-auto` | Déclaration et indemnisation sinistre | Indemnisation |
| `flux-complet` | Cycle de vie complet client | Tous les flux |
| `stress-test` | Charge élevée pour tester les limites | Performance |

---

## 9. Chaos Engineering

### 9.1 Scénarios de défaillance

| Type | Description | Apprentissage visé |
|------|-------------|-------------------|
| **Consumer Lag** | Ralentir un consommateur pour accumuler du retard | Observer le backpressure, métriques de lag |
| **Service Down** | Arrêter un service temporairement | Tester la résilience, retry, Dead Letter Queue |
| **Network Latency** | Injecter de la latence réseau entre services | Impact sur les timeouts, comportement des sagas |
| **Duplicate Events** | Produire des événements en double | Valider l'idempotence des consommateurs |
| **Schema Evolution** | Modifier un schéma Avro en cours de simulation | Tester la compatibilité avant/arrière |

### 9.2 Déclenchement

- **Via Interface Web** : Boutons dédiés pour chaque type de chaos
- **Via YAML** : Configuration dans les scénarios avec `trigger_at` et `duration`
- **Via API REST** : Endpoints pour automatisation

---

## 10. Structure du repository

### 10.1 Organisation (Monorepo)

```
eda-lab/
├── services/
│   ├── bancaire/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── assurance-personne/
│   ├── assurance-dommage/
│   ├── client-360/
│   ├── orchestrator/
│   ├── simulator/
│   └── gateway/
│
├── schemas/
│   ├── bancaire/
│   │   ├── compte-ouvert.avsc
│   │   ├── virement-emis.avsc
│   │   └── ...
│   ├── assurance-personne/
│   ├── assurance-dommage/
│   └── client/
│
├── web-ui/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── hooks/
│   ├── package.json
│   └── Dockerfile
│
├── infra/
│   ├── docker-compose.yml
│   ├── docker-compose.observability.yml
│   ├── kafka/
│   ├── grafana/
│   │   └── dashboards/
│   ├── prometheus/
│   │   └── prometheus.yml
│   ├── jaeger/
│   └── loki/
│
├── scenarios/
│   ├── flux-client-360.yaml
│   ├── flux-paiement-prime.yaml
│   ├── flux-sinistre-auto.yaml
│   └── ...
│
├── docs/
│   ├── README.md
│   ├── INSTALLATION.md
│   ├── ARCHITECTURE.md
│   ├── patterns/
│   │   ├── 01-pub-sub.md
│   │   ├── 02-event-sourcing.md
│   │   └── ...
│   ├── adr/
│   │   ├── 001-choix-kafka.md
│   │   ├── 002-choix-avro.md
│   │   └── ...
│   └── scenarios/
│
├── scripts/
│   ├── start.sh
│   ├── stop.sh
│   ├── reset.sh
│   └── create-topics.sh
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── .gitignore
├── Makefile
└── README.md
```

---

## 11. Stratégie de tests

### 11.1 Niveaux de tests

| Niveau | Outils | Couverture |
|--------|--------|------------|
| **Unitaires** | Go testing, testify | Logique métier, transformations d'événements, validations |
| **Intégration** | testcontainers-go | Kafka réel, PostgreSQL réel en containers éphémères |
| **E2E** | Scripts Go + assertions | Scénarios complets de bout en bout |

### 11.2 Couverture cible

| Composant | Couverture minimale |
|-----------|---------------------|
| Services métier | 80% |
| Orchestrator | 90% |
| Simulator | 70% |
| Gateway | 75% |

---

## 12. Gestion des erreurs et résilience

### 12.1 Stratégie de retry

| Composant | Stratégie | Configuration |
|-----------|-----------|---------------|
| **Producer Kafka** | Retry avec backoff exponentiel | max_retries=3, initial_backoff=100ms, max_backoff=10s |
| **Consumer Kafka** | Retry avant DLQ | max_attempts=3, delay=1s |
| **Appels HTTP** | Retry avec circuit breaker | max_retries=3, timeout=5s |
| **Database** | Retry connexion | max_retries=5, backoff=2s |

### 12.2 Dead Letter Queue (DLQ)

| Élément | Configuration |
|---------|---------------|
| **Topic DLQ** | `system.dlq` |
| **Format** | Message original + métadonnées d'erreur |
| **Rétention** | 7 jours |
| **Monitoring** | Alerte si > 10 messages/minute |

**Structure message DLQ :**
```json
{
  "original_topic": "bancaire.compte.ouvert",
  "original_message": { ... },
  "error": {
    "type": "DESERIALIZATION_ERROR",
    "message": "Invalid Avro schema",
    "timestamp": "2024-01-01T00:00:00Z",
    "attempt_count": 3
  },
  "metadata": {
    "service": "bancaire",
    "consumer_group": "bancaire-group"
  }
}
```

### 12.3 Idempotence

| Événement | Clé d'idempotence | Stratégie |
|-----------|-------------------|-----------|
| CompteOuvert | `compte_id` | INSERT ... ON CONFLICT DO NOTHING |
| DepotEffectue | `event_id` | Table de tracking des événements traités |
| VirementEmis | `event_id` | Table de tracking des événements traités |

### 12.4 Circuit Breaker

| Service | Seuil ouverture | Timeout demi-ouvert | Seuil fermeture |
|---------|-----------------|---------------------|-----------------|
| Gateway → Simulator | 5 erreurs/30s | 30s | 3 succès consécutifs |
| Gateway → Bancaire | 5 erreurs/30s | 30s | 3 succès consécutifs |

---

## 13. Documentation

### 13.1 Structure de la documentation

| Document | Contenu |
|----------|---------|
| `README.md` | Vue d'ensemble, quickstart, liens |
| `INSTALLATION.md` | Prérequis, installation pas à pas |
| `ARCHITECTURE.md` | Diagrammes C4, décisions techniques |
| `patterns/*.md` | Guide détaillé par patron EDA |
| `adr/*.md` | Architecture Decision Records |
| `scenarios/*.md` | Documentation des scénarios |

### 13.2 Architecture Decision Records (ADR)

| ADR | Décision |
|-----|----------|
| ADR-001 | Choix de Confluent Platform Community pour Kafka |
| ADR-002 | Choix d'Avro pour la sérialisation |
| ADR-003 | Choix de Go comme langage principal |
| ADR-004 | Choix de PostgreSQL comme base de données |
| ADR-005 | Choix de React + React Flow pour le frontend |
| ADR-006 | Architecture monorepo |

---

## 14. Environnement d'exécution

### 14.1 Prérequis

| Élément | Version minimale |
|---------|------------------|
| **OS** | Windows 11 |
| **Docker Desktop** | 4.x avec WSL2 |
| **Docker Compose** | v2.x |
| **Go** | 1.21+ |
| **Node.js** | 20 LTS |

### 14.2 Ressources recommandées

| Ressource | Minimum | Recommandé |
|-----------|---------|------------|
| **RAM** | 8 Go | 16 Go |
| **CPU** | 4 cores | 8 cores |
| **Disque** | 20 Go | 50 Go |

---

## 15. Roadmap des itérations

### 15.1 MVP (Itération 1)

| Composant | Statut |
|-----------|--------|
| Patron Pub/Sub | ✅ Inclus |
| Domaine Bancaire (3-4 événements) | ✅ Inclus |
| Service Simulator | ✅ Inclus |
| Infrastructure Kafka + Schema Registry | ✅ Inclus |
| Observabilité de base (Prometheus + Grafana) | ✅ Inclus |
| _Note: Jaeger et Loki prévus pour Itération 2_ | |
| Web UI minimale (Start/Stop + React Flow) | ✅ Inclus |
| Tests unitaires + intégration | ✅ Inclus |
| Documentation de base | ✅ Inclus |

### 15.2 Itération 2

| Composant | Statut |
|-----------|--------|
| Domaines Assurance Personne + Dommage | 🔜 Planifié |
| Event Sourcing | 🔜 Planifié |
| Jaeger (traçage distribué) | 🔜 Planifié |
| Loki (logs centralisés) | 🔜 Planifié |

### 15.3 Itération 3

| Composant | Statut |
|-----------|--------|
| CQRS | 🔜 Planifié |
| Client 360 | 🔜 Planifié |
| Chaos Engineering | 🔜 Planifié |

### 15.4 Itérations suivantes

| Composant | Statut |
|-----------|--------|
| Saga Choreography | 🔜 Planifié |
| Saga Orchestration | 🔜 Planifié |
| Event Streaming (Kafka Streams) | 🔜 Planifié |
| Dead Letter Queue | 🔜 Planifié |
| Outbox Pattern | 🔜 Planifié |

---

## 16. Glossaire

| Terme | Définition |
|-------|------------|
| **EDA** | Event Driven Architecture - Architecture pilotée par les événements |
| **Pub/Sub** | Pattern Publish/Subscribe - Publication et abonnement aux événements |
| **Event Sourcing** | Pattern où l'état est dérivé d'une séquence d'événements |
| **CQRS** | Command Query Responsibility Segregation - Séparation lecture/écriture |
| **Saga** | Pattern de gestion des transactions distribuées |
| **DLQ** | Dead Letter Queue - File pour les événements en erreur |
| **Outbox** | Pattern garantissant la livraison transactionnelle des événements |
| **Schema Registry** | Service de gestion et validation des schémas Avro |
| **KRaft** | Mode Kafka sans ZooKeeper |

---

## 17. Références

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Confluent Platform](https://docs.confluent.io/)
- [Apache Avro Specification](https://avro.apache.org/docs/current/spec.html)
- [React Flow](https://reactflow.dev/)
- [Grafana Documentation](https://grafana.com/docs/)
- [OpenTelemetry](https://opentelemetry.io/)
- [testcontainers-go](https://golang.testcontainers.org/)

---

**Document généré le 2026-01-19**
**Outil : Claude Code**
