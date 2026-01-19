# PLAN.MD - Plan d'implémentation EDA-Lab

## Vue d'ensemble

Ce document définit le plan d'implémentation détaillé du projet **EDA-Lab** selon une approche **Test Driven Development (TDD)** avec progression incrémentale.

**Principes directeurs :**
- Chaque étape produit du code fonctionnel et testé
- Tests avec données réelles et vrais appels API (testcontainers-go)
- Aucun code orphelin - intégration continue
- Progression sans sauts de complexité

---

## Structure des phases

| Phase | Nom | Objectif |
|-------|-----|----------|
| 0 | Infrastructure | Docker Compose, Kafka, PostgreSQL, Schema Registry |
| 1 | Fondations Go | Structure monorepo, bibliothèques partagées |
| 2 | Schémas Avro | Définition et enregistrement des schémas |
| 3 | Producteur | Service Simulator - génération d'événements |
| 4 | Consommateur | Service Bancaire - consommation et traitement |
| 5 | Gateway | API REST + WebSocket |
| 6 | Observabilité | Prometheus, Grafana, métriques |
| 7 | Web UI | Interface React + React Flow |
| 8 | Intégration | Tests E2E, scénarios complets |

---

# PHASE 0 : Infrastructure

## Objectif
Mettre en place l'infrastructure Docker Compose avec Kafka (KRaft), Schema Registry et PostgreSQL fonctionnels et testables.

---

### Étape 0.1 : Structure du projet et Docker Compose minimal

#### Sous-étape 0.1.1 : Création de la structure de répertoires

**Contexte :** Créer la structure de base du monorepo selon le PDR.MD.

```text
PROMPT CLAUDE CODE - Étape 0.1.1

Crée la structure de répertoires suivante pour le projet EDA-Lab :

eda-lab/
├── services/
│   ├── bancaire/
│   ├── simulator/
│   └── gateway/
├── schemas/
│   └── bancaire/
├── web-ui/
├── infra/
│   ├── kafka/
│   ├── grafana/
│   └── prometheus/
├── scenarios/
├── scripts/
├── tests/
│   ├── integration/
│   └── e2e/
├── .gitignore
├── Makefile
└── README.md

Crée également un fichier .gitignore approprié pour Go, Node.js et Docker.
Le README.md doit contenir une brève description du projet et les prérequis (Docker Desktop, Go 1.21+, Node.js 20 LTS).

Ne crée pas de code source Go ou React pour l'instant - uniquement la structure et les fichiers de configuration de base.

Crée également un fichier `go.work` à la racine pour le workspace Go monorepo :
```go
go 1.21

use (
    ./pkg/config
    ./pkg/kafka
    ./pkg/database
    ./pkg/events
    ./pkg/observability
    ./services/simulator
    ./services/bancaire
    ./services/gateway
    ./tests/integration
)
```
```

---

#### Sous-étape 0.1.2 : Docker Compose - Kafka KRaft

**Contexte :** Configurer Kafka en mode KRaft (sans ZooKeeper) avec Confluent Platform Community.

```text
PROMPT CLAUDE CODE - Étape 0.1.2

Dans le répertoire infra/, crée un fichier docker-compose.yml avec un broker Kafka en mode KRaft.

Spécifications :
- Image : confluentinc/cp-kafka:7.5.0
- Mode KRaft (sans ZooKeeper) - utilise KAFKA_KRAFT_CLUSTER_ID
- Port externe : 9092 (PLAINTEXT)
- Port interne : 29092 (pour communication inter-conteneurs)
- Configuration des listeners appropriée pour Windows/WSL2
- Volume persistant pour les données Kafka

Ajoute un healthcheck pour vérifier que Kafka est prêt.

Crée également un script scripts/wait-for-kafka.sh qui attend que Kafka soit disponible avant de continuer.

Teste la configuration en démarrant le conteneur et en vérifiant qu'il est healthy.
```

---

#### Sous-étape 0.1.3 : Docker Compose - Schema Registry

**Contexte :** Ajouter Confluent Schema Registry connecté à Kafka.

```text
PROMPT CLAUDE CODE - Étape 0.1.3

Modifie infra/docker-compose.yml pour ajouter Schema Registry.

Spécifications :
- Image : confluentinc/cp-schema-registry:7.5.0
- Port : 8081
- Connecté au broker Kafka de l'étape précédente
- Healthcheck approprié

Crée un script scripts/test-schema-registry.sh qui :
1. Attend que Schema Registry soit prêt
2. Enregistre un schéma Avro de test via curl
3. Récupère le schéma enregistré
4. Vérifie que le schéma est correct

Le schéma de test doit être simple :
{
  "type": "record",
  "name": "TestEvent",
  "namespace": "com.edalab.test",
  "fields": [
    {"name": "id", "type": "string"},
    {"name": "timestamp", "type": "long"}
  ]
}

Exécute le script pour valider que l'infrastructure fonctionne.
```

---

#### Sous-étape 0.1.4 : Docker Compose - PostgreSQL

**Contexte :** Ajouter PostgreSQL pour la persistance des services.

```text
PROMPT CLAUDE CODE - Étape 0.1.4

Modifie infra/docker-compose.yml pour ajouter PostgreSQL.

Spécifications :
- Image : postgres:16-alpine
- Port : 5432
- Base de données : edalab
- Utilisateur : edalab / edalab_password
- Volume persistant pour les données
- Healthcheck approprié

Crée un fichier infra/postgres/init.sql qui :
1. Crée le schéma 'bancaire'
2. Crée une table de test 'bancaire.health_check' avec une colonne 'status'
3. Insère une ligne de test

Crée un script scripts/test-postgres.sh qui :
1. Attend que PostgreSQL soit prêt
2. Exécute une requête SELECT sur la table health_check
3. Affiche le résultat

Exécute le script pour valider que PostgreSQL fonctionne.
```

---

#### Sous-étape 0.1.5 : Makefile et scripts utilitaires

**Contexte :** Créer les commandes make pour simplifier les opérations courantes.

```text
PROMPT CLAUDE CODE - Étape 0.1.5

Crée un Makefile à la racine du projet avec les targets suivantes :

# Infrastructure
make infra-up        : Démarre tous les conteneurs d'infrastructure
make infra-down      : Arrête tous les conteneurs
make infra-logs      : Affiche les logs de tous les conteneurs
make infra-clean     : Supprime les volumes et repart de zéro

# Tests infrastructure
make test-infra      : Exécute tous les scripts de test d'infrastructure

# Kafka
make kafka-topics    : Liste les topics Kafka
make kafka-create-topic TOPIC=<name> : Crée un topic

# Utilitaires
make clean           : Nettoie les artefacts de build
make help            : Affiche l'aide

Crée également scripts/create-topics.sh qui crée les topics Kafka initiaux :
- bancaire.compte.ouvert
- bancaire.compte.ferme
- bancaire.depot.effectue
- bancaire.retrait.effectue
- bancaire.virement.emis
- bancaire.virement.recu
- bancaire.paiement-prime.effectue
- system.dlq

Teste toutes les commandes make pour vérifier qu'elles fonctionnent correctement.
```

---

### Étape 0.2 : Validation de l'infrastructure

#### Sous-étape 0.2.1 : Test d'intégration infrastructure

**Contexte :** Créer un test Go qui valide toute l'infrastructure.

```text
PROMPT CLAUDE CODE - Étape 0.2.1

Crée un module Go dans tests/integration/ pour tester l'infrastructure.

Fichier : tests/integration/go.mod
- Module : github.com/edalab/tests/integration
- Dépendances : testcontainers-go, confluent-kafka-go, pgx/v5, testify

Fichier : tests/integration/infrastructure_test.go
Crée des tests qui utilisent l'infrastructure Docker Compose existante (pas testcontainers pour cette étape) :

1. TestKafkaConnection :
   - Se connecte à Kafka sur localhost:9092
   - Crée un topic de test
   - Produit un message
   - Consomme le message
   - Vérifie le contenu

2. TestSchemaRegistryConnection :
   - Se connecte à Schema Registry sur localhost:8081
   - Enregistre un schéma Avro
   - Récupère le schéma par ID
   - Vérifie la compatibilité

3. TestPostgreSQLConnection :
   - Se connecte à PostgreSQL
   - Exécute une requête
   - Vérifie le résultat

Ajoute une target Makefile :
make test-integration : Exécute les tests d'intégration (nécessite infra-up)

Les tests doivent passer avec l'infrastructure démarrée via make infra-up.
```

---

# PHASE 1 : Fondations Go

## Objectif
Créer les bibliothèques partagées Go pour Kafka, Avro et la configuration.

---

### Étape 1.1 : Module partagé - Configuration

#### Sous-étape 1.1.1 : Structure du package pkg/config

**Contexte :** Créer un package de configuration centralisé.

```text
PROMPT CLAUDE CODE - Étape 1.1.1

Crée un module Go partagé pour la configuration.

Fichier : pkg/config/go.mod
- Module : github.com/edalab/pkg/config

Fichier : pkg/config/config.go
Implémente une structure de configuration avec :

type Config struct {
    Kafka    KafkaConfig
    Schema   SchemaConfig
    Postgres PostgresConfig
    Service  ServiceConfig
}

type KafkaConfig struct {
    BootstrapServers string
    GroupID          string
    AutoOffsetReset  string
}

type SchemaConfig struct {
    URL string
}

type PostgresConfig struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
}

type ServiceConfig struct {
    Name string
    Port int
}

Fonctions à implémenter :
- LoadFromEnv() (*Config, error) : Charge depuis variables d'environnement
- LoadFromFile(path string) (*Config, error) : Charge depuis fichier YAML
- Validate() error : Valide la configuration

Fichier : pkg/config/config_test.go
Tests TDD :
1. TestLoadFromEnv_Success
2. TestLoadFromEnv_MissingRequired
3. TestLoadFromFile_Success
4. TestLoadFromFile_FileNotFound
5. TestValidate_ValidConfig
6. TestValidate_InvalidConfig

Utilise testify pour les assertions. Tous les tests doivent passer.
```

---

#### Sous-étape 1.1.2 : Fichiers de configuration par environnement

**Contexte :** Créer les fichiers de configuration pour le développement local.

```text
PROMPT CLAUDE CODE - Étape 1.1.2

Crée les fichiers de configuration YAML pour l'environnement de développement.

Fichier : config/local.yaml
kafka:
  bootstrap_servers: "localhost:9092"
  auto_offset_reset: "earliest"

schema:
  url: "http://localhost:8081"

postgres:
  host: "localhost"
  port: 5432
  database: "edalab"
  user: "edalab"
  password: "edalab_password"

Fichier : config/docker.yaml (pour exécution dans Docker)
kafka:
  bootstrap_servers: "kafka:29092"
  auto_offset_reset: "earliest"

schema:
  url: "http://schema-registry:8081"

postgres:
  host: "postgres"
  port: 5432
  database: "edalab"
  user: "edalab"
  password: "edalab_password"

Modifie pkg/config/config.go pour supporter le chargement de ces fichiers YAML.
Ajoute la dépendance gopkg.in/yaml.v3.

Crée un test d'intégration qui charge config/local.yaml et vérifie les valeurs.
```

---

### Étape 1.2 : Module partagé - Client Kafka

#### Sous-étape 1.2.1 : Producer Kafka avec Avro

**Contexte :** Créer un client producteur Kafka réutilisable avec sérialisation Avro.

```text
PROMPT CLAUDE CODE - Étape 1.2.1

Crée un module Go pour le client Kafka.

Fichier : pkg/kafka/go.mod
- Module : github.com/edalab/pkg/kafka
- Dépendances : confluent-kafka-go, srclient (Schema Registry client)

Fichier : pkg/kafka/producer.go
Implémente un producteur Kafka avec sérialisation Avro :

type Producer interface {
    Produce(ctx context.Context, topic string, key string, value interface{}) error
    ProduceWithHeaders(ctx context.Context, topic string, key string, value interface{}, headers map[string]string) error
    Close() error
}

type AvroProducer struct {
    producer     *kafka.Producer
    schemaClient *srclient.SchemaRegistryClient
    schemas      map[string]*srclient.Schema
}

func NewAvroProducer(config *config.KafkaConfig, schemaConfig *config.SchemaConfig) (*AvroProducer, error)

La méthode Produce doit :
1. Récupérer le schéma Avro depuis Schema Registry (avec cache)
2. Sérialiser la valeur en Avro
3. Produire le message dans Kafka
4. Attendre la confirmation (synchrone pour simplifier)

Fichier : pkg/kafka/producer_test.go
Tests TDD (nécessitent infrastructure réelle via docker-compose) :
1. TestNewAvroProducer_Success
2. TestNewAvroProducer_InvalidConfig
3. TestProduce_Success (avec vrai Kafka et Schema Registry)
4. TestProduce_SchemaNotFound
5. TestProduce_SerializationError

Utilise un schéma de test simple enregistré préalablement.
```

---

#### Sous-étape 1.2.2 : Consumer Kafka avec Avro

**Contexte :** Créer un client consommateur Kafka réutilisable.

```text
PROMPT CLAUDE CODE - Étape 1.2.2

Fichier : pkg/kafka/consumer.go
Implémente un consommateur Kafka avec désérialisation Avro :

type MessageHandler func(ctx context.Context, msg *Message) error

type Message struct {
    Topic     string
    Partition int32
    Offset    int64
    Key       string
    Value     interface{}
    Headers   map[string]string
    Timestamp time.Time
}

type Consumer interface {
    Subscribe(topics []string) error
    Consume(ctx context.Context, handler MessageHandler) error
    Close() error
}

type AvroConsumer struct {
    consumer     *kafka.Consumer
    schemaClient *srclient.SchemaRegistryClient
}

func NewAvroConsumer(config *config.KafkaConfig, schemaConfig *config.SchemaConfig, groupID string) (*AvroConsumer, error)

La méthode Consume doit :
1. Lire les messages en boucle
2. Extraire le schema ID du message
3. Désérialiser avec le schéma approprié
4. Appeler le handler
5. Commit l'offset si succès
6. Gérer les erreurs et retries

Fichier : pkg/kafka/consumer_test.go
Tests TDD :
1. TestNewAvroConsumer_Success
2. TestSubscribe_Success
3. TestConsume_Success (produire puis consommer)
4. TestConsume_HandlerError
5. TestConsume_DeserializationError

Les tests doivent produire des messages réels puis les consommer.
```

---

#### Sous-étape 1.2.3 : Tests d'intégration Producer-Consumer

**Contexte :** Valider le flux complet production-consommation.

```text
PROMPT CLAUDE CODE - Étape 1.2.3

Fichier : pkg/kafka/integration_test.go

Crée un test d'intégration complet qui :

1. TestProducerConsumerRoundTrip :
   - Crée un producteur
   - Crée un consommateur avec un groupe unique
   - Enregistre un schéma Avro de test
   - Produit 10 messages
   - Consomme les 10 messages
   - Vérifie que tous les messages sont reçus correctement
   - Vérifie les headers, timestamps, etc.

2. TestMultipleConsumersInGroup :
   - Crée un topic avec 3 partitions
   - Crée 2 consommateurs dans le même groupe
   - Produit 100 messages
   - Vérifie que les messages sont distribués entre les consommateurs
   - Vérifie qu'aucun message n'est perdu ou dupliqué

3. TestConsumerRebalance :
   - Démarre un consommateur
   - Produit des messages
   - Ajoute un deuxième consommateur
   - Vérifie le rebalance
   - Arrête le premier consommateur
   - Vérifie que le second reprend toutes les partitions

Ces tests utilisent l'infrastructure Docker Compose réelle.
Ajoute un build tag //go:build integration pour ces tests.
```

---

### Étape 1.3 : Module partagé - Client PostgreSQL

#### Sous-étape 1.3.1 : Repository pattern avec pgx

**Contexte :** Créer une abstraction pour l'accès à PostgreSQL.

```text
PROMPT CLAUDE CODE - Étape 1.3.1

Crée un module Go pour l'accès PostgreSQL.

Fichier : pkg/database/go.mod
- Module : github.com/edalab/pkg/database
- Dépendances : pgx/v5, pgxpool

Fichier : pkg/database/pool.go
type DBPool struct {
    pool *pgxpool.Pool
}

func NewDBPool(config *config.PostgresConfig) (*DBPool, error)
func (p *DBPool) Close()
func (p *DBPool) Pool() *pgxpool.Pool
func (p *DBPool) HealthCheck(ctx context.Context) error

Fichier : pkg/database/transaction.go
type TxFunc func(ctx context.Context, tx pgx.Tx) error

func (p *DBPool) WithTransaction(ctx context.Context, fn TxFunc) error
- Démarre une transaction
- Exécute fn
- Commit si succès, rollback si erreur
- Gère les panics

Fichier : pkg/database/pool_test.go
Tests TDD (avec PostgreSQL réel) :
1. TestNewDBPool_Success
2. TestNewDBPool_InvalidConfig
3. TestHealthCheck_Success
4. TestWithTransaction_Commit
5. TestWithTransaction_Rollback
6. TestWithTransaction_Panic

Les tests doivent créer/supprimer des données réelles dans PostgreSQL.
```

---

### Étape 1.4 : Module partagé - Observabilité

#### Sous-étape 1.4.1 : Métriques Prometheus

**Contexte :** Créer les métriques communes pour tous les services.

```text
PROMPT CLAUDE CODE - Étape 1.4.1

Fichier : pkg/observability/go.mod
- Module : github.com/edalab/pkg/observability
- Dépendances : prometheus/client_golang

Fichier : pkg/observability/metrics.go
Implémente les métriques communes :

var (
    MessagesProduced = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "edalab_messages_produced_total",
            Help: "Total number of messages produced",
        },
        []string{"service", "topic"},
    )

    MessagesConsumed = prometheus.NewCounterVec(...)
    MessageLatency = prometheus.NewHistogramVec(...)
    ProcessingErrors = prometheus.NewCounterVec(...)
)

func RegisterMetrics()
func NewMetricsServer(port int) *http.Server

Fichier : pkg/observability/metrics_test.go
Tests TDD :
1. TestRegisterMetrics_Success
2. TestMessagesProduced_Increment
3. TestMessageLatency_Observe
4. TestMetricsServer_Endpoint

Vérifie que /metrics retourne les métriques au format Prometheus.
```

---

#### Sous-étape 1.4.2 : Tracing OpenTelemetry

**Contexte :** Ajouter le tracing distribué.

```text
PROMPT CLAUDE CODE - Étape 1.4.2

Fichier : pkg/observability/tracing.go
Implémente le tracing OpenTelemetry :

func InitTracer(serviceName string, jaegerEndpoint string) (*trace.TracerProvider, error)
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span)
func SpanFromContext(ctx context.Context) trace.Span
func InjectTraceContext(ctx context.Context, headers map[string]string)
func ExtractTraceContext(ctx context.Context, headers map[string]string) context.Context

Dépendances : go.opentelemetry.io/otel, otel/exporters/jaeger

Fichier : pkg/observability/tracing_test.go
Tests TDD :
1. TestInitTracer_Success
2. TestStartSpan_CreatesSpan
3. TestInjectExtractTraceContext_RoundTrip

Note : Pour les tests, utiliser un NoopTracerProvider pour éviter de dépendre de Jaeger.
Ajouter des tests d'intégration séparés qui utilisent Jaeger réel.
```

---

#### Sous-étape 1.4.3 : Logging structuré

**Contexte :** Configurer le logging structuré JSON.

```text
PROMPT CLAUDE CODE - Étape 1.4.3

Fichier : pkg/observability/logging.go
Implémente le logging structuré avec slog :

func InitLogger(serviceName string, level string) *slog.Logger
func WithTraceID(ctx context.Context, logger *slog.Logger) *slog.Logger

Le logger doit :
- Produire du JSON
- Inclure le nom du service
- Inclure le trace ID si présent dans le contexte
- Supporter les niveaux DEBUG, INFO, WARN, ERROR

Fichier : pkg/observability/logging_test.go
Tests TDD :
1. TestInitLogger_DefaultLevel
2. TestInitLogger_DebugLevel
3. TestWithTraceID_IncludesTraceID
4. TestLogOutput_JSONFormat
```

---

# PHASE 2 : Schémas Avro

## Objectif
Définir et enregistrer les schémas Avro pour le domaine Bancaire.

---

### Étape 2.1 : Schémas du domaine Bancaire

#### Sous-étape 2.1.1 : Schéma CompteOuvert

**Contexte :** Créer le premier schéma Avro avec tous les champs métier.

```text
PROMPT CLAUDE CODE - Étape 2.1.1

Crée le schéma Avro pour l'événement CompteOuvert.

Fichier : schemas/bancaire/compte-ouvert.avsc
{
  "type": "record",
  "name": "CompteOuvert",
  "namespace": "com.edalab.bancaire.events",
  "doc": "Événement émis lors de l'ouverture d'un compte bancaire",
  "fields": [
    {
      "name": "event_id",
      "type": "string",
      "doc": "Identifiant unique de l'événement (UUID)"
    },
    {
      "name": "timestamp",
      "type": {
        "type": "long",
        "logicalType": "timestamp-millis"
      },
      "doc": "Horodatage de l'événement"
    },
    {
      "name": "compte_id",
      "type": "string",
      "doc": "Identifiant unique du compte"
    },
    {
      "name": "client_id",
      "type": "string",
      "doc": "Identifiant du client titulaire"
    },
    {
      "name": "type_compte",
      "type": {
        "type": "enum",
        "name": "TypeCompte",
        "symbols": ["COURANT", "EPARGNE", "JOINT"]
      }
    },
    {
      "name": "devise",
      "type": "string",
      "default": "EUR"
    },
    {
      "name": "solde_initial",
      "type": {
        "type": "bytes",
        "logicalType": "decimal",
        "precision": 18,
        "scale": 2
      }
    },
    {
      "name": "metadata",
      "type": ["null", {
        "type": "map",
        "values": "string"
      }],
      "default": null
    }
  ]
}

Crée un script scripts/register-schemas.sh qui enregistre ce schéma dans Schema Registry.
Vérifie l'enregistrement en récupérant le schéma via l'API.
```

---

#### Sous-étape 2.1.2 : Schémas DepotEffectue et VirementEmis

**Contexte :** Compléter les schémas pour le MVP.

```text
PROMPT CLAUDE CODE - Étape 2.1.2

Crée les schémas Avro supplémentaires.

Fichier : schemas/bancaire/depot-effectue.avsc
Champs :
- event_id (string, UUID)
- timestamp (timestamp-millis)
- compte_id (string)
- montant (decimal 18,2)
- devise (string, default EUR)
- reference (string, référence de l'opération)
- canal (enum: GUICHET, VIREMENT, CHEQUE, CARTE)
- metadata (optional map)

Fichier : schemas/bancaire/virement-emis.avsc
Champs :
- event_id (string, UUID)
- timestamp (timestamp-millis)
- compte_source_id (string)
- compte_destination_id (string)
- montant (decimal 18,2)
- devise (string)
- motif (string)
- reference (string)
- statut (enum: INITIE, EN_COURS, COMPLETE, REJETE)
- metadata (optional map)

Modifie scripts/register-schemas.sh pour enregistrer tous les schémas.
Ajoute une validation qui vérifie que les schémas sont valides avant enregistrement.
```

---

#### Sous-étape 2.1.3 : Génération de code Go depuis Avro

**Contexte :** Générer les structs Go depuis les schémas Avro.

```text
PROMPT CLAUDE CODE - Étape 2.1.3

Crée un module pour les types générés depuis Avro.

Fichier : pkg/events/go.mod
- Module : github.com/edalab/pkg/events

Utilise github.com/hamba/avro/v2 pour la génération.

Crée un script scripts/generate-avro.sh qui :
1. Parcourt tous les fichiers .avsc dans schemas/
2. Génère les structs Go correspondants dans pkg/events/
3. Ajoute les tags JSON et Avro

Fichier attendu : pkg/events/bancaire.go
package events

type CompteOuvert struct {
    EventID      string            `avro:"event_id" json:"event_id"`
    Timestamp    time.Time         `avro:"timestamp" json:"timestamp"`
    CompteID     string            `avro:"compte_id" json:"compte_id"`
    ClientID     string            `avro:"client_id" json:"client_id"`
    TypeCompte   TypeCompte        `avro:"type_compte" json:"type_compte"`
    Devise       string            `avro:"devise" json:"devise"`
    SoldeInitial decimal.Decimal   `avro:"solde_initial" json:"solde_initial"`
    Metadata     map[string]string `avro:"metadata" json:"metadata,omitempty"`
}

// ... autres types

Fichier : pkg/events/bancaire_test.go
Tests TDD :
1. TestCompteOuvert_Serialization_Avro
2. TestCompteOuvert_Deserialization_Avro
3. TestDepotEffectue_RoundTrip
4. TestVirementEmis_RoundTrip

Les tests doivent sérialiser/désérialiser avec de vraies valeurs.
```

---

# PHASE 3 : Service Simulator

## Objectif
Créer le service qui génère des événements fictifs automatiquement.

---

### Étape 3.1 : Structure du service Simulator

#### Sous-étape 3.1.1 : Scaffolding du service

**Contexte :** Créer la structure de base du service Simulator.

```text
PROMPT CLAUDE CODE - Étape 3.1.1

Crée la structure du service Simulator.

Fichier : services/simulator/go.mod
- Module : github.com/edalab/services/simulator
- Dépendances : imports des packages pkg/*

Structure :
services/simulator/
├── cmd/
│   └── simulator/
│       └── main.go
├── internal/
│   ├── generator/
│   │   └── generator.go
│   ├── scenario/
│   │   └── loader.go
│   └── api/
│       └── handler.go
├── Dockerfile
└── go.mod

Fichier : services/simulator/cmd/simulator/main.go
func main() qui :
1. Charge la configuration
2. Initialise le logger
3. Initialise le tracer
4. Crée le producteur Kafka
5. Démarre le serveur HTTP pour l'API de contrôle
6. Démarre le serveur de métriques
7. Gère le graceful shutdown

Le service ne génère pas encore d'événements - juste le scaffolding.

Fichier : services/simulator/Dockerfile
Multi-stage build :
- Stage 1 : golang:1.21-alpine pour build
- Stage 2 : alpine:3.19 pour runtime
- Expose ports 8080 (API) et 9090 (métriques)

Teste que le service démarre sans erreur.
```

---

#### Sous-étape 3.1.2 : Générateur de données fictives

**Contexte :** Créer le générateur de données réalistes pour les événements.

```text
PROMPT CLAUDE CODE - Étape 3.1.2

Fichier : services/simulator/internal/generator/fake_data.go

Crée un générateur de données fictives réalistes :

type FakeDataGenerator struct {
    rng *rand.Rand
}

func NewFakeDataGenerator(seed int64) *FakeDataGenerator

func (g *FakeDataGenerator) GenerateClientID() string
func (g *FakeDataGenerator) GenerateCompteID() string
func (g *FakeDataGenerator) GenerateNom() string
func (g *FakeDataGenerator) GeneratePrenom() string
func (g *FakeDataGenerator) GenerateMontant(min, max float64) decimal.Decimal
func (g *FakeDataGenerator) GenerateIBAN() string
func (g *FakeDataGenerator) GenerateReference() string

Les données doivent être réalistes :
- IBAN français valide (format FR + clé de contrôle)
- Noms/prénoms français courants
- Montants avec distribution réaliste

Fichier : services/simulator/internal/generator/fake_data_test.go
Tests TDD :
1. TestGenerateIBAN_ValidFormat
2. TestGenerateIBAN_ValidChecksum
3. TestGenerateMontant_InRange
4. TestGenerateClientID_UniqueFormat
5. TestDeterministicWithSameSeed

Les tests vérifient que les données générées sont valides.
```

---

#### Sous-étape 3.1.3 : Générateur d'événements CompteOuvert

**Contexte :** Créer le générateur spécifique pour l'événement CompteOuvert.

```text
PROMPT CLAUDE CODE - Étape 3.1.3

Fichier : services/simulator/internal/generator/compte_ouvert.go

type CompteOuvertGenerator struct {
    fakeData *FakeDataGenerator
    producer kafka.Producer
    topic    string
}

func NewCompteOuvertGenerator(fakeData *FakeDataGenerator, producer kafka.Producer) *CompteOuvertGenerator

func (g *CompteOuvertGenerator) Generate(ctx context.Context) (*events.CompteOuvert, error)
- Génère un événement CompteOuvert avec données fictives
- Produit l'événement dans Kafka
- Retourne l'événement produit

func (g *CompteOuvertGenerator) GenerateBatch(ctx context.Context, count int, interval time.Duration) error
- Génère count événements avec interval entre chaque
- Respecte le contexte pour annulation
- Log chaque événement produit

Fichier : services/simulator/internal/generator/compte_ouvert_test.go
Tests TDD avec Kafka réel :
1. TestGenerate_ProducesValidEvent
2. TestGenerate_EventInKafka (vérifie que l'événement est dans le topic)
3. TestGenerateBatch_ProducesAllEvents
4. TestGenerateBatch_RespectsInterval
5. TestGenerateBatch_CancellableViaContext
```

---

### Étape 3.2 : API de contrôle du Simulator

#### Sous-étape 3.2.1 : Endpoints REST

**Contexte :** Créer l'API REST pour contrôler la simulation.

```text
PROMPT CLAUDE CODE - Étape 3.2.1

Fichier : services/simulator/internal/api/handler.go

Implémente les endpoints REST :

POST /api/v1/simulation/start
Body: {
  "scenario": "default",
  "rate": 10,  // événements par seconde
  "duration": 300  // secondes, 0 = infini
}
Response: { "simulation_id": "uuid", "status": "running" }

POST /api/v1/simulation/stop
Response: { "status": "stopped", "events_produced": 1234 }

GET /api/v1/simulation/status
Response: {
  "status": "running|stopped",
  "simulation_id": "uuid",
  "events_produced": 1234,
  "started_at": "2024-01-01T00:00:00Z",
  "rate_actual": 9.8
}

POST /api/v1/events/produce
Body: {
  "event_type": "CompteOuvert",
  "count": 1
}
Response: { "events_produced": 1, "event_ids": ["uuid"] }

Utilise chi ou gorilla/mux comme router.

Fichier : services/simulator/internal/api/handler_test.go
Tests TDD :
1. TestStartSimulation_Success
2. TestStartSimulation_AlreadyRunning
3. TestStopSimulation_Success
4. TestStopSimulation_NotRunning
5. TestGetStatus_Running
6. TestGetStatus_Stopped
7. TestProduceEvents_Success
```

---

#### Sous-étape 3.2.2 : Gestionnaire de simulation

**Contexte :** Créer le composant qui orchestre la génération d'événements.

```text
PROMPT CLAUDE CODE - Étape 3.2.2

Fichier : services/simulator/internal/simulation/manager.go

type SimulationManager struct {
    generators map[string]Generator
    producer   kafka.Producer
    status     atomic.Value  // *SimulationStatus
    cancel     context.CancelFunc
    wg         sync.WaitGroup
}

type SimulationStatus struct {
    ID             string
    Status         string  // "running", "stopped"
    EventsProduced int64
    StartedAt      time.Time
    Rate           float64
}

type SimulationConfig struct {
    Scenario string
    Rate     int
    Duration time.Duration
}

func NewSimulationManager(producer kafka.Producer) *SimulationManager
func (m *SimulationManager) Start(ctx context.Context, config SimulationConfig) error
func (m *SimulationManager) Stop() (*SimulationStatus, error)
func (m *SimulationManager) Status() *SimulationStatus

La méthode Start doit :
1. Vérifier qu'aucune simulation n'est en cours
2. Créer un contexte avec cancel
3. Démarrer une goroutine qui génère des événements au rate spécifié
4. Arrêter automatiquement après duration (si > 0)

Fichier : services/simulator/internal/simulation/manager_test.go
Tests TDD :
1. TestStart_Success
2. TestStart_AlreadyRunning
3. TestStop_Success
4. TestStop_NotRunning
5. TestAutoStopAfterDuration
6. TestRateControl
```

---

#### Sous-étape 3.2.3 : Test d'intégration Simulator complet

**Contexte :** Valider le service Simulator de bout en bout.

```text
PROMPT CLAUDE CODE - Étape 3.2.3

Fichier : services/simulator/integration_test.go
Build tag : //go:build integration

Test d'intégration complet :

1. TestSimulatorE2E :
   - Démarre le service Simulator (ou utilise l'instance Docker)
   - Appelle POST /simulation/start avec rate=5, duration=10
   - Attend 10 secondes
   - Vérifie GET /simulation/status
   - Vérifie que ~50 événements ont été produits
   - Consomme les événements depuis Kafka
   - Vérifie que les événements sont valides (schéma Avro)

2. TestSimulatorManualStop :
   - Démarre une simulation avec duration=0
   - Attend 5 secondes
   - Appelle POST /simulation/stop
   - Vérifie que la simulation s'arrête
   - Vérifie le compte d'événements

3. TestSimulatorProduceSingle :
   - Appelle POST /events/produce avec count=1
   - Consomme l'événement depuis Kafka
   - Vérifie le contenu

Ajoute le service Simulator au docker-compose.yml dans un profil "services".
```

---

# PHASE 4 : Service Bancaire

## Objectif
Créer le service qui consomme et traite les événements bancaires.

---

### Étape 4.1 : Structure du service Bancaire

#### Sous-étape 4.1.1 : Scaffolding et modèle de données

**Contexte :** Créer la structure du service Bancaire avec son modèle de données.

```text
PROMPT CLAUDE CODE - Étape 4.1.1

Crée la structure du service Bancaire.

Structure :
services/bancaire/
├── cmd/
│   └── bancaire/
│       └── main.go
├── internal/
│   ├── domain/
│   │   └── compte.go
│   ├── repository/
│   │   └── compte_repository.go
│   ├── handler/
│   │   └── event_handler.go
│   └── api/
│       └── handler.go
├── migrations/
│   └── 001_create_comptes.sql
├── Dockerfile
└── go.mod

Fichier : services/bancaire/internal/domain/compte.go
type Compte struct {
    ID           string
    ClientID     string
    TypeCompte   string
    Devise       string
    Solde        decimal.Decimal
    Statut       string  // ACTIF, FERME, BLOQUE
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Transaction struct {
    ID          string
    CompteID    string
    Type        string  // DEPOT, RETRAIT, VIREMENT_ENTRANT, VIREMENT_SORTANT
    Montant     decimal.Decimal
    Reference   string
    CreatedAt   time.Time
}

Fichier : services/bancaire/migrations/001_create_comptes.sql
CREATE TABLE bancaire.comptes (
    id VARCHAR(36) PRIMARY KEY,
    client_id VARCHAR(36) NOT NULL,
    type_compte VARCHAR(20) NOT NULL,
    devise VARCHAR(3) NOT NULL DEFAULT 'EUR',
    solde DECIMAL(18,2) NOT NULL DEFAULT 0,
    statut VARCHAR(20) NOT NULL DEFAULT 'ACTIF',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bancaire.transactions (
    id VARCHAR(36) PRIMARY KEY,
    compte_id VARCHAR(36) NOT NULL REFERENCES bancaire.comptes(id),
    type VARCHAR(30) NOT NULL,
    montant DECIMAL(18,2) NOT NULL,
    reference VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comptes_client ON bancaire.comptes(client_id);
CREATE INDEX idx_transactions_compte ON bancaire.transactions(compte_id);
```

---

#### Sous-étape 4.1.2 : Repository Pattern

**Contexte :** Implémenter le repository pour la persistance des comptes.

```text
PROMPT CLAUDE CODE - Étape 4.1.2

Fichier : services/bancaire/internal/repository/compte_repository.go

type CompteRepository interface {
    Create(ctx context.Context, compte *domain.Compte) error
    GetByID(ctx context.Context, id string) (*domain.Compte, error)
    GetByClientID(ctx context.Context, clientID string) ([]*domain.Compte, error)
    UpdateSolde(ctx context.Context, id string, nouveauSolde decimal.Decimal) error
    AddTransaction(ctx context.Context, tx *domain.Transaction) error
    GetTransactions(ctx context.Context, compteID string, limit int) ([]*domain.Transaction, error)
}

type PostgresCompteRepository struct {
    pool *database.DBPool
}

func NewPostgresCompteRepository(pool *database.DBPool) *PostgresCompteRepository

Implémente toutes les méthodes avec des requêtes SQL paramétrées.
Utilise les transactions pour les opérations qui modifient plusieurs tables.

Fichier : services/bancaire/internal/repository/compte_repository_test.go
Tests TDD avec PostgreSQL réel :
1. TestCreate_Success
2. TestCreate_DuplicateID
3. TestGetByID_Found
4. TestGetByID_NotFound
5. TestUpdateSolde_Success
6. TestAddTransaction_Success
7. TestAddTransaction_CompteNotFound
8. TestGetTransactions_WithLimit

Chaque test doit nettoyer ses données après exécution.
```

---

#### Sous-étape 4.1.3 : Handler d'événements Kafka

**Contexte :** Créer le handler qui traite les événements Kafka.

```text
PROMPT CLAUDE CODE - Étape 4.1.3

Fichier : services/bancaire/internal/handler/event_handler.go

type EventHandler struct {
    repo   repository.CompteRepository
    logger *slog.Logger
}

func NewEventHandler(repo repository.CompteRepository, logger *slog.Logger) *EventHandler

func (h *EventHandler) HandleCompteOuvert(ctx context.Context, event *events.CompteOuvert) error
- Crée un nouveau compte dans la base de données
- Ajoute une transaction initiale si solde > 0
- Log l'opération
- Retourne une erreur si le compte existe déjà (idempotence)

func (h *EventHandler) HandleDepotEffectue(ctx context.Context, event *events.DepotEffectue) error
- Récupère le compte
- Met à jour le solde
- Ajoute une transaction
- Log l'opération
- Retourne une erreur si compte non trouvé

func (h *EventHandler) HandleVirementEmis(ctx context.Context, event *events.VirementEmis) error
- Vérifie que le compte source existe
- Vérifie le solde suffisant
- Débite le compte
- Ajoute une transaction
- Log l'opération

func (h *EventHandler) Route(ctx context.Context, msg *kafka.Message) error
- Détermine le type d'événement depuis le topic ou les headers
- Appelle le handler approprié

Fichier : services/bancaire/internal/handler/event_handler_test.go
Tests TDD :
1. TestHandleCompteOuvert_Success
2. TestHandleCompteOuvert_Idempotent
3. TestHandleDepotEffectue_Success
4. TestHandleDepotEffectue_CompteNotFound
5. TestHandleVirementEmis_Success
6. TestHandleVirementEmis_SoldeInsuffisant
7. TestRoute_CompteOuvert
8. TestRoute_UnknownEvent
```

---

### Étape 4.2 : Intégration du service Bancaire

#### Sous-étape 4.2.1 : Main et bootstrap

**Contexte :** Assembler tous les composants dans le main.

```text
PROMPT CLAUDE CODE - Étape 4.2.1

Fichier : services/bancaire/cmd/bancaire/main.go

func main() {
    // 1. Configuration
    cfg := config.LoadFromEnv()

    // 2. Observabilité
    logger := observability.InitLogger("bancaire", cfg.LogLevel)
    tracer := observability.InitTracer("bancaire", cfg.JaegerEndpoint)
    observability.RegisterMetrics()

    // 3. Database
    pool := database.NewDBPool(cfg.Postgres)
    repo := repository.NewPostgresCompteRepository(pool)

    // 4. Kafka Consumer
    consumer := kafka.NewAvroConsumer(cfg.Kafka, cfg.Schema, "bancaire-group")
    consumer.Subscribe([]string{
        "bancaire.compte.ouvert",
        "bancaire.depot.effectue",
        "bancaire.virement.emis",
    })

    // 5. Event Handler
    handler := handler.NewEventHandler(repo, logger)

    // 6. API REST (pour health check et queries)
    apiHandler := api.NewHandler(repo)
    httpServer := &http.Server{Addr: ":8080", Handler: apiHandler.Router()}

    // 7. Metrics server
    metricsServer := observability.NewMetricsServer(9090)

    // 8. Start consumers
    go consumer.Consume(ctx, handler.Route)

    // 9. Start HTTP servers
    go httpServer.ListenAndServe()
    go metricsServer.ListenAndServe()

    // 10. Graceful shutdown
    // ...
}

Fichier : services/bancaire/Dockerfile
Multi-stage build similaire au Simulator.

Ajoute le service au docker-compose.yml.
```

---

#### Sous-étape 4.2.2 : API REST pour queries

**Contexte :** Créer l'API REST pour interroger les données.

```text
PROMPT CLAUDE CODE - Étape 4.2.2

Fichier : services/bancaire/internal/api/handler.go

Endpoints :

GET /api/v1/health
Response: { "status": "healthy", "kafka": "connected", "postgres": "connected" }

GET /api/v1/comptes/:id
Response: {
  "id": "uuid",
  "client_id": "uuid",
  "type_compte": "COURANT",
  "devise": "EUR",
  "solde": "1234.56",
  "statut": "ACTIF",
  "created_at": "..."
}

GET /api/v1/comptes/:id/transactions
Query: ?limit=10
Response: {
  "transactions": [
    {
      "id": "uuid",
      "type": "DEPOT",
      "montant": "100.00",
      "reference": "REF123",
      "created_at": "..."
    }
  ]
}

GET /api/v1/clients/:client_id/comptes
Response: { "comptes": [...] }

Fichier : services/bancaire/internal/api/handler_test.go
Tests TDD :
1. TestHealthCheck_AllHealthy
2. TestGetCompte_Found
3. TestGetCompte_NotFound
4. TestGetTransactions_Success
5. TestGetTransactions_EmptyList
6. TestGetComptesByClient_Success
```

---

#### Sous-étape 4.2.3 : Test d'intégration Simulator → Bancaire

**Contexte :** Valider le flux complet de bout en bout.

```text
PROMPT CLAUDE CODE - Étape 4.2.3

Fichier : tests/integration/simulator_bancaire_test.go
Build tag : //go:build integration

Test d'intégration E2E :

1. TestFluxCompteOuvert_E2E :
   - Démarre Simulator et Bancaire (via docker-compose)
   - Appelle Simulator POST /events/produce avec event_type=CompteOuvert
   - Attend quelques secondes pour le traitement
   - Appelle Bancaire GET /comptes/:id avec l'ID généré
   - Vérifie que le compte existe avec les bonnes données

2. TestFluxDepot_E2E :
   - Crée un compte via Simulator
   - Attend le traitement
   - Produit un événement DepotEffectue pour ce compte
   - Attend le traitement
   - Vérifie que le solde a été mis à jour
   - Vérifie que la transaction existe

3. TestFluxMultipleEvents_E2E :
   - Lance une simulation de 10 secondes à 5 events/s
   - Attend 15 secondes
   - Vérifie que tous les comptes ont été créés
   - Vérifie la cohérence des données

4. TestIdempotence_E2E :
   - Produit le même événement deux fois (même event_id)
   - Vérifie qu'un seul compte est créé

Ces tests utilisent l'infrastructure Docker Compose complète.
Ajoute une target Makefile : make test-e2e
```

---

# PHASE 5 : Service Gateway

## Objectif
Créer l'API Gateway avec WebSocket pour le temps réel.

---

### Étape 5.1 : Gateway REST

#### Sous-étape 5.1.1 : Structure et routing

**Contexte :** Créer le service Gateway qui unifie les API.

```text
PROMPT CLAUDE CODE - Étape 5.1.1

Structure :
services/gateway/
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   ├── proxy/
│   │   └── service_proxy.go
│   ├── websocket/
│   │   └── hub.go
│   └── api/
│       └── router.go
├── Dockerfile
└── go.mod

Fichier : services/gateway/internal/proxy/service_proxy.go

type ServiceProxy struct {
    simulatorURL string
    bancaireURL  string
    client       *http.Client
}

func NewServiceProxy(simulatorURL, bancaireURL string) *ServiceProxy

func (p *ServiceProxy) ForwardToSimulator(w http.ResponseWriter, r *http.Request)
func (p *ServiceProxy) ForwardToBancaire(w http.ResponseWriter, r *http.Request)

Le proxy doit :
- Transférer les headers (sauf Host)
- Propager le trace context
- Gérer les timeouts
- Logger les requêtes

Fichier : services/gateway/internal/api/router.go

Routes :
- /api/v1/simulation/*  → Simulator
- /api/v1/bancaire/*    → Bancaire
- /api/v1/health        → Health check agrégé
- /ws                   → WebSocket

Middleware CORS à configurer pour autoriser les requêtes depuis web-ui (localhost:5173 en dev, configurable pour prod).

Fichier : services/gateway/internal/proxy/service_proxy_test.go
Tests TDD :
1. TestForwardToSimulator_Success
2. TestForwardToBancaire_Success
3. TestForward_ServiceUnavailable
4. TestForward_PropagatesHeaders
```

---

#### Sous-étape 5.1.2 : WebSocket Hub

**Contexte :** Créer le hub WebSocket pour les événements temps réel.

```text
PROMPT CLAUDE CODE - Étape 5.1.2

Fichier : services/gateway/internal/websocket/hub.go

type Client struct {
    conn   *websocket.Conn
    send   chan []byte
    topics map[string]bool  // topics souscrits
}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Message struct {
    Type      string      `json:"type"`      // "event", "status", "error"
    Topic     string      `json:"topic"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

func NewHub() *Hub
func (h *Hub) Run()
func (h *Hub) Broadcast(msg *Message)
func (h *Hub) BroadcastToTopic(topic string, msg *Message)

Fichier : services/gateway/internal/websocket/client.go
func (c *Client) ReadPump(hub *Hub)   // Lit les messages du client (subscribe/unsubscribe)
func (c *Client) WritePump()          // Écrit les messages vers le client

Messages client → serveur :
{ "action": "subscribe", "topics": ["bancaire.compte.ouvert"] }
{ "action": "unsubscribe", "topics": ["bancaire.compte.ouvert"] }

Fichier : services/gateway/internal/websocket/hub_test.go
Tests TDD :
1. TestHub_RegisterClient
2. TestHub_UnregisterClient
3. TestHub_Broadcast
4. TestHub_BroadcastToTopic
5. TestClient_Subscribe
6. TestClient_Unsubscribe
```

---

#### Sous-étape 5.1.3 : Consommateur Kafka pour WebSocket

**Contexte :** Connecter Kafka au WebSocket pour le streaming temps réel.

```text
PROMPT CLAUDE CODE - Étape 5.1.3

Fichier : services/gateway/internal/streaming/kafka_streamer.go

type KafkaStreamer struct {
    consumer *kafka.AvroConsumer
    hub      *websocket.Hub
    topics   []string
}

func NewKafkaStreamer(consumer *kafka.AvroConsumer, hub *websocket.Hub, topics []string) *KafkaStreamer

func (s *KafkaStreamer) Start(ctx context.Context) error
- Souscrit aux topics configurés
- Pour chaque message reçu :
  1. Désérialise l'événement
  2. Crée un Message WebSocket
  3. Broadcast au hub vers les clients abonnés

func (s *KafkaStreamer) Stop()

Fichier : services/gateway/cmd/gateway/main.go
Intègre :
- ServiceProxy
- WebSocket Hub
- KafkaStreamer
- Serveur HTTP

Fichier : services/gateway/integration_test.go
Tests d'intégration :
1. TestWebSocket_ReceivesKafkaEvents
   - Connecte un client WebSocket
   - Subscribe au topic bancaire.compte.ouvert
   - Produit un événement via Simulator
   - Vérifie que le client WebSocket reçoit l'événement

2. TestWebSocket_MultipleClients
   - Connecte 3 clients
   - Vérifie que tous reçoivent les broadcasts
```

---

# PHASE 6 : Observabilité

## Objectif
Configurer Prometheus et Grafana avec des dashboards.

---

### Étape 6.1 : Configuration Prometheus

#### Sous-étape 6.1.1 : Configuration et scraping

**Contexte :** Configurer Prometheus pour collecter les métriques.

```text
PROMPT CLAUDE CODE - Étape 6.1.1

Fichier : infra/prometheus/prometheus.yml

global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'kafka'
    static_configs:
      - targets: ['kafka:9092']

  - job_name: 'simulator'
    static_configs:
      - targets: ['simulator:9090']

  - job_name: 'bancaire'
    static_configs:
      - targets: ['bancaire:9090']

  - job_name: 'gateway'
    static_configs:
      - targets: ['gateway:9090']

Modifie docker-compose.yml pour ajouter Prometheus :
- Image : prom/prometheus:v2.47.0
- Port : 9090
- Volume : ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
- Healthcheck

Crée un script scripts/test-prometheus.sh qui :
1. Attend que Prometheus soit prêt
2. Query l'API pour vérifier les targets
3. Vérifie que les services sont UP
```

---

#### Sous-étape 6.1.2 : Métriques Kafka avec JMX Exporter

**Contexte :** Exposer les métriques Kafka pour Prometheus.

```text
PROMPT CLAUDE CODE - Étape 6.1.2

Modifie la configuration Kafka dans docker-compose.yml pour activer JMX :

Environment variables :
KAFKA_JMX_PORT: 9999
KAFKA_JMX_HOSTNAME: kafka

Ajoute un conteneur JMX Exporter :
- Image : bitnami/jmx-exporter:0.19.0
- Port : 5556
- Configuration pour les métriques Kafka

Fichier : infra/kafka/jmx-exporter-config.yml
Métriques à exposer :
- kafka.server:type=BrokerTopicMetrics (messages in/out)
- kafka.server:type=ReplicaManager (partitions)
- kafka.consumer:type=consumer-fetch-manager-metrics (lag)

Modifie prometheus.yml pour scraper le JMX exporter.

Vérifie que les métriques Kafka apparaissent dans Prometheus.
```

---

### Étape 6.2 : Dashboards Grafana

#### Sous-étape 6.2.1 : Configuration Grafana

**Contexte :** Configurer Grafana avec provisioning automatique.

```text
PROMPT CLAUDE CODE - Étape 6.2.1

Ajoute Grafana à docker-compose.yml :
- Image : grafana/grafana:10.2.0
- Port : 3000
- Volumes pour provisioning

Fichier : infra/grafana/provisioning/datasources/prometheus.yml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true

Fichier : infra/grafana/provisioning/dashboards/dashboards.yml
apiVersion: 1
providers:
  - name: 'EDA-Lab'
    folder: 'EDA-Lab'
    type: file
    options:
      path: /var/lib/grafana/dashboards

Vérifie que Grafana démarre et se connecte à Prometheus.
```

---

#### Sous-étape 6.2.2 : Dashboard Kafka Overview

**Contexte :** Créer le dashboard pour les métriques Kafka.

```text
PROMPT CLAUDE CODE - Étape 6.2.2

Fichier : infra/grafana/dashboards/kafka-overview.json

Crée un dashboard Grafana avec les panels suivants :

Row 1 : Vue d'ensemble
- Stat : Messages/sec (total)
- Stat : Topics actifs
- Stat : Partitions totales

Row 2 : Messages
- Time series : Messages produits par topic (5 min)
- Time series : Messages consommés par topic (5 min)

Row 3 : Consumer Lag
- Time series : Lag par consumer group
- Table : Top 10 partitions avec lag

Row 4 : Bytes
- Time series : Bytes in/out par broker

Variables :
- $topic : Sélection du topic
- $consumer_group : Sélection du consumer group

Le dashboard doit utiliser les métriques du JMX exporter.
```

---

#### Sous-étape 6.2.3 : Dashboard Services

**Contexte :** Créer le dashboard pour les métriques applicatives.

```text
PROMPT CLAUDE CODE - Étape 6.2.3

Fichier : infra/grafana/dashboards/services-overview.json

Dashboard avec panels :

Row 1 : Santé des services
- Stat : Status Simulator (UP/DOWN)
- Stat : Status Bancaire (UP/DOWN)
- Stat : Status Gateway (UP/DOWN)

Row 2 : Événements
- Time series : edalab_messages_produced_total par service
- Time series : edalab_messages_consumed_total par service

Row 3 : Latences
- Heatmap : edalab_message_latency_seconds (Simulator)
- Heatmap : edalab_processing_latency_seconds (Bancaire)

Row 4 : Erreurs
- Time series : edalab_processing_errors_total par service
- Time series : Rate d'erreur (%)

Row 5 : Ressources
- Time series : go_memstats_alloc_bytes par service
- Time series : go_goroutines par service

Variables :
- $service : Sélection du service
```

---

# PHASE 7 : Web UI

## Objectif
Créer l'interface React avec visualisation des flux.

---

### Étape 7.1 : Setup React

#### Sous-étape 7.1.1 : Initialisation du projet React

**Contexte :** Créer l'application React avec les dépendances nécessaires.

```text
PROMPT CLAUDE CODE - Étape 7.1.1

Dans le répertoire web-ui/, initialise un projet React avec Vite :

npm create vite@latest . -- --template react-ts

Installe les dépendances :
npm install @xyflow/react           # React Flow pour les graphes
npm install @tanstack/react-query   # Data fetching
npm install axios                   # HTTP client
npm install zustand                 # State management
npm install tailwindcss postcss autoprefixer  # Styling
npm install lucide-react            # Icons

Configure Tailwind CSS.

Fichier : web-ui/src/App.tsx
Structure de base :
- Header avec titre et status de connexion
- Sidebar avec navigation
- Main content area
- Footer avec métriques

Fichier : web-ui/Dockerfile
Multi-stage :
- Stage 1 : node:20-alpine pour build
- Stage 2 : nginx:alpine pour servir les fichiers

Vérifie que npm run dev démarre correctement.
```

---

#### Sous-étape 7.1.2 : Configuration API et WebSocket

**Contexte :** Configurer les clients pour communiquer avec le Gateway.

```text
PROMPT CLAUDE CODE - Étape 7.1.2

Fichier : web-ui/src/lib/api.ts

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const api = axios.create({
    baseURL: API_BASE_URL,
});

// Simulation API
export const simulationApi = {
    start: (config: SimulationConfig) => api.post('/api/v1/simulation/start', config),
    stop: () => api.post('/api/v1/simulation/stop'),
    status: () => api.get('/api/v1/simulation/status'),
    produceEvent: (eventType: string, count: number) =>
        api.post('/api/v1/events/produce', { event_type: eventType, count }),
};

// Bancaire API
export const bancaireApi = {
    getCompte: (id: string) => api.get(`/api/v1/bancaire/comptes/${id}`),
    getTransactions: (compteId: string) =>
        api.get(`/api/v1/bancaire/comptes/${compteId}/transactions`),
};

Fichier : web-ui/src/lib/websocket.ts

export class EventSocket {
    private ws: WebSocket | null = null;
    private listeners: Map<string, Set<(event: any) => void>> = new Map();

    connect(url: string): Promise<void>
    disconnect(): void
    subscribe(topic: string, callback: (event: any) => void): () => void
    unsubscribe(topic: string): void
}

export const eventSocket = new EventSocket();

Fichier : web-ui/src/lib/api.test.ts
Tests avec MSW (Mock Service Worker) :
1. TestSimulationApi_Start
2. TestSimulationApi_Status
3. TestBancaireApi_GetCompte
```

---

### Étape 7.2 : Composants UI

#### Sous-étape 7.2.1 : Contrôles de simulation

**Contexte :** Créer les composants pour contrôler la simulation.

```text
PROMPT CLAUDE CODE - Étape 7.2.1

Fichier : web-ui/src/components/SimulationControls.tsx

Composant avec :
- Bouton Start/Stop (toggle selon état)
- Slider pour le rate (événements/seconde)
- Input pour la durée (0 = infini)
- Dropdown pour sélection du scénario
- Affichage du status actuel (running/stopped)
- Compteur d'événements produits

Props :
interface SimulationControlsProps {
    onStart: (config: SimulationConfig) => void;
    onStop: () => void;
    status: SimulationStatus | null;
}

Fichier : web-ui/src/components/EventProducer.tsx

Composant pour produire des événements manuellement :
- Dropdown pour type d'événement (CompteOuvert, DepotEffectue, etc.)
- Input pour le nombre d'événements
- Bouton "Produire"
- Feedback de succès/erreur

Fichier : web-ui/src/components/__tests__/SimulationControls.test.tsx
Tests avec Testing Library :
1. TestRender_Initial
2. TestStart_CallsOnStart
3. TestStop_CallsOnStop
4. TestSlider_UpdatesRate
```

---

#### Sous-étape 7.2.2 : Visualisation React Flow

**Contexte :** Créer la visualisation des flux d'événements.

```text
PROMPT CLAUDE CODE - Étape 7.2.2

Fichier : web-ui/src/components/FlowVisualization.tsx

Utilise React Flow pour visualiser :

Nodes :
- Simulator (source)
- Kafka (broker central)
- Bancaire (consumer)
- Client 360 (consumer) - pour préparation future

Edges :
- Simulator → Kafka (animé quand événements produits)
- Kafka → Bancaire (animé quand événements consommés)

Animation :
- Utilise edgeAnimationDuration pour montrer le flux
- Change la couleur selon le type d'événement
- Affiche le compteur d'événements sur chaque edge

Fichier : web-ui/src/components/FlowNode.tsx

Custom node avec :
- Icône du service
- Nom du service
- Status indicator (vert/orange/rouge)
- Compteur d'événements traités
- Mini graphique sparkline des dernières 60 secondes

Fichier : web-ui/src/hooks/useFlowData.ts

Hook qui :
- Se connecte au WebSocket
- Écoute les événements
- Met à jour les compteurs
- Déclenche les animations

Tests :
1. TestFlowVisualization_RendersNodes
2. TestFlowVisualization_AnimatesOnEvent
3. TestFlowNode_ShowsStatus
```

---

#### Sous-étape 7.2.3 : Dashboard de métriques

**Contexte :** Créer un dashboard léger avec les métriques principales.

```text
PROMPT CLAUDE CODE - Étape 7.2.3

Fichier : web-ui/src/components/MetricsDashboard.tsx

Dashboard avec :

Section 1 : Simulation
- Card : Événements produits (total)
- Card : Rate actuel (events/sec)
- Card : Durée de simulation

Section 2 : Kafka
- Card : Messages dans les topics
- Card : Consumer lag
- Mini graphique : Messages/sec (30 dernières secondes)

Section 3 : Services
- Status de chaque service (badge coloré)
- Latence moyenne de traitement
- Erreurs récentes

Fichier : web-ui/src/hooks/useMetrics.ts

Hook qui :
- Poll les métriques via l'API toutes les 5 secondes
- Ou utilise le WebSocket pour les updates temps réel
- Calcule les moyennes mobiles

Fichier : web-ui/src/components/MetricCard.tsx

Card réutilisable :
- Titre
- Valeur principale (grande)
- Tendance (↑ ↓ →)
- Sparkline optionnel
```

---

### Étape 7.3 : Intégration finale UI

#### Sous-étape 7.3.1 : Layout et navigation

**Contexte :** Assembler tous les composants dans le layout final.

```text
PROMPT CLAUDE CODE - Étape 7.3.1

Fichier : web-ui/src/App.tsx

Layout final :
┌─────────────────────────────────────────────────────────────┐
│  Header : EDA-Lab    |  Status: Connected  |  Settings     │
├─────────────────────────────────────────────────────────────┤
│  Sidebar     │  Main Content                                │
│              │  ┌─────────────────────────────────────────┐ │
│  Controls    │  │                                         │ │
│  - Start     │  │     Flow Visualization                  │ │
│  - Stop      │  │     (React Flow)                        │ │
│  - Rate      │  │                                         │ │
│              │  └─────────────────────────────────────────┘ │
│  Producer    │  ┌─────────────────────────────────────────┐ │
│  - Type      │  │     Metrics Dashboard                   │ │
│  - Count     │  │     (Cards + Charts)                    │ │
│  - Produce   │  │                                         │ │
│              │  └─────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  Footer : Events: 1234  |  Rate: 10/s  |  Uptime: 00:05:30 │
└─────────────────────────────────────────────────────────────┘

Fichier : web-ui/src/store/simulationStore.ts (Zustand)

State global :
- simulationStatus
- metrics
- events (derniers 100 événements)
- connectionStatus

Actions :
- startSimulation
- stopSimulation
- updateMetrics
- addEvent
```

---

#### Sous-étape 7.3.2 : Tests E2E UI

**Contexte :** Créer les tests E2E pour l'interface.

```text
PROMPT CLAUDE CODE - Étape 7.3.2

Installe Playwright :
npm install -D @playwright/test

Fichier : web-ui/e2e/simulation.spec.ts

Tests E2E avec vraie infrastructure :

1. test('démarre et arrête une simulation')
   - Navigue vers la page
   - Clique sur Start
   - Vérifie que le status passe à "running"
   - Attend 5 secondes
   - Vérifie que les événements sont comptés
   - Clique sur Stop
   - Vérifie que le status passe à "stopped"

2. test('produit un événement manuel')
   - Sélectionne "CompteOuvert"
   - Entre 1 comme count
   - Clique sur Produire
   - Vérifie le toast de succès
   - Vérifie que le compteur augmente

3. test('visualisation s\'anime avec les événements')
   - Démarre une simulation
   - Vérifie que les edges s'animent
   - Vérifie que les compteurs sur les nodes changent

4. test('WebSocket se reconnecte après déconnexion')
   - Vérifie la connexion
   - Simule une déconnexion
   - Vérifie la reconnexion automatique

Fichier : web-ui/playwright.config.ts
Configuration pour utiliser l'infrastructure Docker Compose.
```

---

# PHASE 8 : Intégration finale

## Objectif
Valider le système complet et documenter.

---

### Étape 8.1 : Tests E2E complets

#### Sous-étape 8.1.1 : Scénarios de test E2E

**Contexte :** Créer les tests de bout en bout pour le MVP complet.

```text
PROMPT CLAUDE CODE - Étape 8.1.1

Fichier : tests/e2e/mvp_test.go
Build tag : //go:build e2e

Tests E2E complets :

1. TestMVP_FullFlow
   - Démarre toute l'infrastructure (make infra-up)
   - Démarre tous les services (make services-up)
   - Via Gateway API :
     a. Démarre une simulation (rate=10, duration=60)
     b. Attend 60 secondes
     c. Vérifie que ~600 événements ont été produits
     d. Vérifie que tous les comptes existent dans Bancaire
     e. Vérifie les métriques Prometheus
     f. Vérifie les traces Jaeger (si configuré)
   - Arrête proprement

2. TestMVP_ChaosConsumerRestart
   - Démarre une simulation continue
   - Arrête le service Bancaire
   - Attend 10 secondes (événements s'accumulent)
   - Redémarre Bancaire
   - Vérifie que tous les événements sont traités (pas de perte)

3. TestMVP_HighThroughput
   - Démarre une simulation rate=100
   - Vérifie que le système supporte la charge
   - Mesure les latences
   - Vérifie qu'il n'y a pas d'erreurs

Fichier : tests/e2e/docker-compose.e2e.yml
Configuration spécifique pour les tests E2E.
```

---

#### Sous-étape 8.1.2 : Script de validation complète

**Contexte :** Créer un script de validation qui teste tout.

```text
PROMPT CLAUDE CODE - Étape 8.1.2

Fichier : scripts/validate-mvp.sh

Script qui exécute une validation complète du MVP :

#!/bin/bash
set -e

echo "=== EDA-Lab MVP Validation ==="

echo "1. Starting infrastructure..."
make infra-up
sleep 30  # Attendre que tout soit prêt

echo "2. Running infrastructure tests..."
make test-infra

echo "3. Starting services..."
make services-up
sleep 10

echo "4. Running unit tests..."
make test-unit

echo "5. Running integration tests..."
make test-integration

echo "6. Running E2E tests..."
make test-e2e

echo "7. Checking Prometheus targets..."
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

echo "8. Checking Grafana..."
curl -s http://localhost:3000/api/health | jq .

echo "9. Running simulation test..."
curl -X POST http://localhost:8080/api/v1/simulation/start \
  -H "Content-Type: application/json" \
  -d '{"rate": 10, "duration": 30}'
sleep 35
RESULT=$(curl -s http://localhost:8080/api/v1/simulation/status)
echo "Simulation result: $RESULT"

echo "10. Cleanup..."
make infra-down

echo "=== MVP Validation Complete ==="

Ajoute la target Makefile : make validate-mvp
```

---

#### Sous-étape 8.1.3 : Tests de performance

**Contexte :** Mesurer les performances et les limites du système.

```text
PROMPT CLAUDE CODE - Étape 8.1.3

Crée des tests de performance pour mesurer les limites du système.

Fichier : tests/e2e/performance_test.go
Build tag : //go:build performance

1. TestPerformance_Throughput :
   - Démarre une simulation à rate=100 events/s
   - Mesure le throughput réel atteint
   - Vérifie que throughput >= 90% du rate demandé
   - Mesure la latence moyenne de traitement
   - Durée : 60 secondes

2. TestPerformance_Latency :
   - Produit 1000 événements
   - Mesure le temps entre production et persistance en base
   - Calcule P50, P95, P99 des latences
   - Vérifie P99 < 500ms

3. TestPerformance_Burst :
   - Produit 1000 événements en burst (1s)
   - Vérifie que tous sont traités dans les 30s suivantes
   - Mesure le temps de récupération

4. TestPerformance_Sustained :
   - Simulation à rate=50 pendant 5 minutes
   - Vérifie stabilité mémoire (pas de leak)
   - Vérifie stabilité latence (pas de dégradation)

Fichier : scripts/run-performance-tests.sh
- Configure les limites de ressources
- Exécute les tests de performance
- Génère un rapport avec les métriques

Fichier : infra/grafana/dashboards/performance.json
Dashboard dédié aux tests de performance :
- Throughput en temps réel
- Latences (histogramme)
- Utilisation mémoire/CPU
- Consumer lag

Ajoute les targets Makefile :
- make test-performance : Exécute les tests de performance
- make report-performance : Génère le rapport de performance
```

---

### Étape 8.2 : Documentation finale

#### Sous-étape 8.2.1 : README et guides

**Contexte :** Créer la documentation utilisateur.

```text
PROMPT CLAUDE CODE - Étape 8.2.1

Fichier : README.md

# EDA-Lab

Simulateur d'architecture événementielle pour l'apprentissage des patrons EDA.

## Quick Start

### Prérequis
- Docker Desktop 4.x avec WSL2
- Go 1.21+
- Node.js 20 LTS

### Démarrage rapide

```bash
# Clone
git clone https://github.com/edalab/eda-lab.git
cd eda-lab

# Démarrer l'infrastructure
make infra-up

# Démarrer les services
make services-up

# Ouvrir l'interface
open http://localhost:3000
```

## Architecture

[Diagramme ASCII de l'architecture]

## Utilisation

### Démarrer une simulation
[Instructions avec captures d'écran]

### Visualiser les flux
[Instructions]

### Consulter les métriques
[Instructions Grafana]

## Développement

### Structure du projet
[Description des répertoires]

### Lancer les tests
```bash
make test-unit
make test-integration
make test-e2e
```

## License

MIT
```

---

#### Sous-étape 8.2.2 : Documentation technique

**Contexte :** Créer la documentation pour les développeurs.

```text
PROMPT CLAUDE CODE - Étape 8.2.2

Fichier : docs/ARCHITECTURE.md

# Architecture EDA-Lab

## Vue d'ensemble

[Diagramme C4 - Context]
[Diagramme C4 - Container]
[Diagramme C4 - Component]

## Flux de données

### Flux CompteOuvert
1. Simulator génère un événement CompteOuvert
2. Sérialisation Avro avec Schema Registry
3. Production dans topic bancaire.compte.ouvert
4. Gateway reçoit via consumer et broadcast WebSocket
5. Bancaire consomme et persiste dans PostgreSQL
6. UI met à jour la visualisation

## Décisions techniques

Voir dossier adr/ pour les Architecture Decision Records.

Fichier : docs/adr/001-choix-kafka.md

# ADR 001: Choix de Confluent Platform pour Kafka

## Contexte
[...]

## Décision
[...]

## Conséquences
[...]

Créer les ADR pour :
- 001-choix-kafka.md
- 002-choix-avro.md
- 003-choix-go.md
- 004-choix-postgresql.md
- 005-choix-react-flow.md
```

---

#### Sous-étape 8.2.3 : Guide du patron Pub/Sub

**Contexte :** Documenter le premier patron implémenté.

```text
PROMPT CLAUDE CODE - Étape 8.2.3

Fichier : docs/patterns/01-pub-sub.md

# Patron Producteur/Consommateur (Pub/Sub)

## Concept

Le patron Pub/Sub découple les producteurs des consommateurs via un broker de messages.

[Diagramme]

## Implémentation dans EDA-Lab

### Producteur (Simulator)

```go
// Exemple de code du Simulator
producer.Produce(ctx, "bancaire.compte.ouvert", event)
```

### Broker (Kafka)

- Topic: bancaire.compte.ouvert
- Partitions: 3
- Replication factor: 1 (dev)

### Consommateur (Bancaire)

```go
// Exemple de code du Bancaire
consumer.Subscribe([]string{"bancaire.compte.ouvert"})
consumer.Consume(ctx, handler.Route)
```

## Expérimentation

### Exercice 1: Observer le découplage
1. Arrêter le service Bancaire
2. Produire des événements
3. Redémarrer Bancaire
4. Observer que les événements sont traités

### Exercice 2: Scaling des consommateurs
1. Démarrer 2 instances de Bancaire
2. Observer la répartition des partitions
3. Mesurer le throughput

## Compromis architecturaux

| Avantage | Inconvénient |
|----------|--------------|
| Découplage | Complexité ajoutée |
| Scalabilité | Latence (async) |
| Résilience | Ordering (par partition) |

## Pour aller plus loin

- [Kafka Documentation](...)
- [Patron suivant: Event Sourcing](./02-event-sourcing.md)
```

---

## Résumé du plan

| Phase | Étapes | Sous-étapes | Estimation |
|-------|--------|-------------|------------|
| 0 - Infrastructure | 2 | 6 | Fondation |
| 1 - Fondations Go | 4 | 9 | Core libraries |
| 2 - Schémas Avro | 1 | 3 | Data contracts |
| 3 - Simulator | 2 | 6 | Event generation |
| 4 - Bancaire | 2 | 6 | Event processing |
| 5 - Gateway | 1 | 3 | API + WebSocket |
| 6 - Observabilité | 2 | 5 | Monitoring |
| 7 - Web UI | 3 | 7 | Interface |
| 8 - Intégration | 2 | 6 | Validation + Performance |
| **TOTAL** | **19** | **51** | **MVP complet** |

---

## Notes d'implémentation

1. **Ordre strict** : Chaque sous-étape dépend des précédentes
2. **TDD** : Écrire les tests avant l'implémentation
3. **Données réelles** : Utiliser testcontainers-go et infrastructure Docker
4. **Pas de mocks** : Sauf pour les tests unitaires isolés
5. **Intégration continue** : Chaque sous-étape produit du code fonctionnel
6. **Validation** : Exécuter les tests après chaque sous-étape
