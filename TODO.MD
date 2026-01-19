# TODO.MD - Liste de tâches détaillée EDA-Lab

## Légende

- `[ ]` : À faire
- `[~]` : En cours
- `[x]` : Terminé
- `[!]` : Bloqué / Problème

---

## Progression globale

| Phase | Nom | Progression |
|-------|-----|-------------|
| 0 | Infrastructure | 0/6 |
| 1 | Fondations Go | 0/9 |
| 2 | Schémas Avro | 0/3 |
| 3 | Service Simulator | 0/6 |
| 4 | Service Bancaire | 0/6 |
| 5 | Service Gateway | 0/3 |
| 6 | Observabilité | 0/5 |
| 7 | Web UI | 0/7 |
| 8 | Intégration finale | 0/6 |
| **TOTAL** | **MVP** | **0/51** |

---

# PHASE 0 : Infrastructure

## Étape 0.1 : Structure du projet et Docker Compose minimal

### 0.1.1 : Création de la structure de répertoires
- [ ] Créer le répertoire racine `eda-lab/`
- [ ] Créer `services/bancaire/`
- [ ] Créer `services/simulator/`
- [ ] Créer `services/gateway/`
- [ ] Créer `schemas/bancaire/`
- [ ] Créer `web-ui/`
- [ ] Créer `infra/kafka/`
- [ ] Créer `infra/grafana/`
- [ ] Créer `infra/prometheus/`
- [ ] Créer `scenarios/`
- [ ] Créer `scripts/`
- [ ] Créer `tests/integration/`
- [ ] Créer `tests/e2e/`
- [ ] Créer `.gitignore` (Go, Node.js, Docker)
- [ ] Créer `Makefile` (placeholder)
- [ ] Créer `README.md` (description + prérequis)
- [ ] Créer `go.work` pour le workspace Go monorepo
- [ ] Valider la structure avec `tree` ou `ls -R`

### 0.1.2 : Docker Compose - Kafka KRaft
- [ ] Créer `infra/docker-compose.yml`
- [ ] Configurer le broker Kafka (image confluentinc/cp-kafka:7.5.0)
- [ ] Configurer le mode KRaft (KAFKA_KRAFT_CLUSTER_ID)
- [ ] Configurer le port externe 9092 (PLAINTEXT)
- [ ] Configurer le port interne 29092 (inter-conteneurs)
- [ ] Configurer les listeners pour Windows/WSL2
- [ ] Ajouter un volume persistant pour Kafka
- [ ] Ajouter un healthcheck pour Kafka
- [ ] Créer `scripts/wait-for-kafka.sh`
- [ ] Tester : `docker-compose up kafka`
- [ ] Vérifier que Kafka est healthy

### 0.1.3 : Docker Compose - Schema Registry
- [ ] Ajouter Schema Registry à `docker-compose.yml`
- [ ] Image : confluentinc/cp-schema-registry:7.5.0
- [ ] Port : 8081
- [ ] Connecter au broker Kafka
- [ ] Ajouter healthcheck pour Schema Registry
- [ ] Créer `scripts/test-schema-registry.sh`
- [ ] Implémenter l'attente Schema Registry prêt
- [ ] Implémenter l'enregistrement d'un schéma de test
- [ ] Implémenter la récupération du schéma
- [ ] Implémenter la vérification du schéma
- [ ] Exécuter et valider le script

### 0.1.4 : Docker Compose - PostgreSQL
- [ ] Ajouter PostgreSQL à `docker-compose.yml`
- [ ] Image : postgres:16-alpine
- [ ] Port : 5432
- [ ] Base de données : edalab
- [ ] Utilisateur : edalab / edalab_password
- [ ] Ajouter volume persistant pour PostgreSQL
- [ ] Ajouter healthcheck pour PostgreSQL
- [ ] Créer `infra/postgres/` directory
- [ ] Créer `infra/postgres/init.sql`
- [ ] Dans init.sql : créer schéma 'bancaire'
- [ ] Dans init.sql : créer table health_check
- [ ] Dans init.sql : insérer ligne de test
- [ ] Créer `scripts/test-postgres.sh`
- [ ] Implémenter l'attente PostgreSQL prêt
- [ ] Implémenter la requête SELECT
- [ ] Exécuter et valider le script

### 0.1.5 : Makefile et scripts utilitaires
- [ ] Créer `Makefile` complet
- [ ] Target : `infra-up`
- [ ] Target : `infra-down`
- [ ] Target : `infra-logs`
- [ ] Target : `infra-clean`
- [ ] Target : `test-infra`
- [ ] Target : `kafka-topics`
- [ ] Target : `kafka-create-topic`
- [ ] Target : `clean`
- [ ] Target : `help`
- [ ] Créer `scripts/create-topics.sh`
- [ ] Créer topic `bancaire.compte.ouvert`
- [ ] Créer topic `bancaire.compte.ferme`
- [ ] Créer topic `bancaire.depot.effectue`
- [ ] Créer topic `bancaire.retrait.effectue`
- [ ] Créer topic `bancaire.virement.emis`
- [ ] Créer topic `bancaire.virement.recu`
- [ ] Créer topic `bancaire.paiement-prime.effectue`
- [ ] Créer topic `system.dlq`
- [ ] Tester `make infra-up`
- [ ] Tester `make infra-down`
- [ ] Tester `make kafka-topics`
- [ ] Tester `make test-infra`
- [ ] Tester `make help`

## Étape 0.2 : Validation de l'infrastructure

### 0.2.1 : Test d'intégration infrastructure
- [ ] Créer `tests/integration/go.mod`
- [ ] Module : github.com/edalab/tests/integration
- [ ] Ajouter dépendance testcontainers-go
- [ ] Ajouter dépendance confluent-kafka-go
- [ ] Ajouter dépendance pgx/v5
- [ ] Ajouter dépendance testify
- [ ] Créer `tests/integration/infrastructure_test.go`
- [ ] Implémenter TestKafkaConnection
  - [ ] Connexion à Kafka localhost:9092
  - [ ] Création topic de test
  - [ ] Production d'un message
  - [ ] Consommation du message
  - [ ] Vérification du contenu
- [ ] Implémenter TestSchemaRegistryConnection
  - [ ] Connexion à Schema Registry localhost:8081
  - [ ] Enregistrement schéma Avro
  - [ ] Récupération schéma par ID
  - [ ] Vérification compatibilité
- [ ] Implémenter TestPostgreSQLConnection
  - [ ] Connexion à PostgreSQL
  - [ ] Exécution requête
  - [ ] Vérification résultat
- [ ] Ajouter target Makefile : `test-integration`
- [ ] Exécuter `make infra-up`
- [ ] Exécuter `make test-integration`
- [ ] Vérifier que tous les tests passent

---

# PHASE 1 : Fondations Go

## Étape 1.1 : Module partagé - Configuration

### 1.1.1 : Structure du package pkg/config
- [ ] Créer `pkg/config/` directory
- [ ] Créer `pkg/config/go.mod`
- [ ] Module : github.com/edalab/pkg/config
- [ ] Créer `pkg/config/config.go`
- [ ] Définir struct Config
- [ ] Définir struct KafkaConfig
- [ ] Définir struct SchemaConfig
- [ ] Définir struct PostgresConfig
- [ ] Définir struct ServiceConfig
- [ ] Implémenter LoadFromEnv()
- [ ] Implémenter LoadFromFile()
- [ ] Implémenter Validate()
- [ ] Créer `pkg/config/config_test.go`
- [ ] Test : TestLoadFromEnv_Success
- [ ] Test : TestLoadFromEnv_MissingRequired
- [ ] Test : TestLoadFromFile_Success
- [ ] Test : TestLoadFromFile_FileNotFound
- [ ] Test : TestValidate_ValidConfig
- [ ] Test : TestValidate_InvalidConfig
- [ ] Exécuter les tests et vérifier qu'ils passent

### 1.1.2 : Fichiers de configuration par environnement
- [ ] Créer `config/` directory à la racine
- [ ] Créer `config/local.yaml`
- [ ] Configurer kafka.bootstrap_servers: localhost:9092
- [ ] Configurer kafka.auto_offset_reset: earliest
- [ ] Configurer schema.url: http://localhost:8081
- [ ] Configurer postgres (host, port, database, user, password)
- [ ] Créer `config/docker.yaml`
- [ ] Configurer kafka.bootstrap_servers: kafka:29092
- [ ] Configurer schema.url: http://schema-registry:8081
- [ ] Configurer postgres.host: postgres
- [ ] Ajouter dépendance gopkg.in/yaml.v3 à pkg/config
- [ ] Modifier LoadFromFile pour parser YAML
- [ ] Créer test d'intégration pour charger local.yaml
- [ ] Vérifier les valeurs chargées

## Étape 1.2 : Module partagé - Client Kafka

### 1.2.1 : Producer Kafka avec Avro
- [ ] Créer `pkg/kafka/` directory
- [ ] Créer `pkg/kafka/go.mod`
- [ ] Module : github.com/edalab/pkg/kafka
- [ ] Ajouter dépendance confluent-kafka-go
- [ ] Ajouter dépendance srclient
- [ ] Créer `pkg/kafka/producer.go`
- [ ] Définir interface Producer
- [ ] Définir méthode Produce()
- [ ] Définir méthode ProduceWithHeaders()
- [ ] Définir méthode Close()
- [ ] Implémenter struct AvroProducer
- [ ] Implémenter NewAvroProducer()
- [ ] Implémenter récupération schéma depuis Schema Registry
- [ ] Implémenter cache des schémas
- [ ] Implémenter sérialisation Avro
- [ ] Implémenter production dans Kafka
- [ ] Implémenter attente confirmation (synchrone)
- [ ] Créer `pkg/kafka/producer_test.go`
- [ ] Test : TestNewAvroProducer_Success
- [ ] Test : TestNewAvroProducer_InvalidConfig
- [ ] Test : TestProduce_Success
- [ ] Test : TestProduce_SchemaNotFound
- [ ] Test : TestProduce_SerializationError
- [ ] Exécuter les tests avec infrastructure réelle

### 1.2.2 : Consumer Kafka avec Avro
- [ ] Créer `pkg/kafka/consumer.go`
- [ ] Définir type MessageHandler
- [ ] Définir struct Message (Topic, Partition, Offset, Key, Value, Headers, Timestamp)
- [ ] Définir interface Consumer
- [ ] Définir méthode Subscribe()
- [ ] Définir méthode Consume()
- [ ] Définir méthode Close()
- [ ] Implémenter struct AvroConsumer
- [ ] Implémenter NewAvroConsumer()
- [ ] Implémenter lecture messages en boucle
- [ ] Implémenter extraction schema ID
- [ ] Implémenter désérialisation Avro
- [ ] Implémenter appel handler
- [ ] Implémenter commit offset
- [ ] Implémenter gestion erreurs et retries
- [ ] Créer `pkg/kafka/consumer_test.go`
- [ ] Test : TestNewAvroConsumer_Success
- [ ] Test : TestSubscribe_Success
- [ ] Test : TestConsume_Success
- [ ] Test : TestConsume_HandlerError
- [ ] Test : TestConsume_DeserializationError
- [ ] Exécuter les tests avec infrastructure réelle

### 1.2.3 : Tests d'intégration Producer-Consumer
- [ ] Créer `pkg/kafka/integration_test.go`
- [ ] Ajouter build tag //go:build integration
- [ ] Implémenter TestProducerConsumerRoundTrip
  - [ ] Créer producteur
  - [ ] Créer consommateur avec groupe unique
  - [ ] Enregistrer schéma Avro de test
  - [ ] Produire 10 messages
  - [ ] Consommer 10 messages
  - [ ] Vérifier réception correcte
  - [ ] Vérifier headers et timestamps
- [ ] Implémenter TestMultipleConsumersInGroup
  - [ ] Créer topic avec 3 partitions
  - [ ] Créer 2 consommateurs même groupe
  - [ ] Produire 100 messages
  - [ ] Vérifier distribution entre consommateurs
  - [ ] Vérifier aucune perte/duplication
- [ ] Implémenter TestConsumerRebalance
  - [ ] Démarrer 1 consommateur
  - [ ] Produire messages
  - [ ] Ajouter 2ème consommateur
  - [ ] Vérifier rebalance
  - [ ] Arrêter 1er consommateur
  - [ ] Vérifier reprise partitions
- [ ] Exécuter tous les tests d'intégration

## Étape 1.3 : Module partagé - Client PostgreSQL

### 1.3.1 : Repository pattern avec pgx
- [ ] Créer `pkg/database/` directory
- [ ] Créer `pkg/database/go.mod`
- [ ] Module : github.com/edalab/pkg/database
- [ ] Ajouter dépendance pgx/v5
- [ ] Ajouter dépendance pgxpool
- [ ] Créer `pkg/database/pool.go`
- [ ] Implémenter struct DBPool
- [ ] Implémenter NewDBPool()
- [ ] Implémenter Close()
- [ ] Implémenter Pool()
- [ ] Implémenter HealthCheck()
- [ ] Créer `pkg/database/transaction.go`
- [ ] Définir type TxFunc
- [ ] Implémenter WithTransaction()
  - [ ] Démarrer transaction
  - [ ] Exécuter fn
  - [ ] Commit si succès
  - [ ] Rollback si erreur
  - [ ] Gérer panics
- [ ] Créer `pkg/database/pool_test.go`
- [ ] Test : TestNewDBPool_Success
- [ ] Test : TestNewDBPool_InvalidConfig
- [ ] Test : TestHealthCheck_Success
- [ ] Test : TestWithTransaction_Commit
- [ ] Test : TestWithTransaction_Rollback
- [ ] Test : TestWithTransaction_Panic
- [ ] Exécuter les tests avec PostgreSQL réel

## Étape 1.4 : Module partagé - Observabilité

### 1.4.1 : Métriques Prometheus
- [ ] Créer `pkg/observability/` directory
- [ ] Créer `pkg/observability/go.mod`
- [ ] Module : github.com/edalab/pkg/observability
- [ ] Ajouter dépendance prometheus/client_golang
- [ ] Créer `pkg/observability/metrics.go`
- [ ] Définir MessagesProduced (CounterVec)
- [ ] Définir MessagesConsumed (CounterVec)
- [ ] Définir MessageLatency (HistogramVec)
- [ ] Définir ProcessingErrors (CounterVec)
- [ ] Implémenter RegisterMetrics()
- [ ] Implémenter NewMetricsServer()
- [ ] Créer `pkg/observability/metrics_test.go`
- [ ] Test : TestRegisterMetrics_Success
- [ ] Test : TestMessagesProduced_Increment
- [ ] Test : TestMessageLatency_Observe
- [ ] Test : TestMetricsServer_Endpoint
- [ ] Vérifier /metrics retourne format Prometheus

### 1.4.2 : Tracing OpenTelemetry
- [ ] Créer `pkg/observability/tracing.go`
- [ ] Ajouter dépendance go.opentelemetry.io/otel
- [ ] Ajouter dépendance otel/exporters/jaeger
- [ ] Implémenter InitTracer()
- [ ] Implémenter StartSpan()
- [ ] Implémenter SpanFromContext()
- [ ] Implémenter InjectTraceContext()
- [ ] Implémenter ExtractTraceContext()
- [ ] Créer `pkg/observability/tracing_test.go`
- [ ] Test : TestInitTracer_Success
- [ ] Test : TestStartSpan_CreatesSpan
- [ ] Test : TestInjectExtractTraceContext_RoundTrip
- [ ] Utiliser NoopTracerProvider pour tests unitaires

### 1.4.3 : Logging structuré
- [ ] Créer `pkg/observability/logging.go`
- [ ] Implémenter InitLogger() avec slog
- [ ] Configurer sortie JSON
- [ ] Inclure nom du service
- [ ] Supporter niveaux DEBUG, INFO, WARN, ERROR
- [ ] Implémenter WithTraceID()
- [ ] Inclure trace ID si présent dans contexte
- [ ] Créer `pkg/observability/logging_test.go`
- [ ] Test : TestInitLogger_DefaultLevel
- [ ] Test : TestInitLogger_DebugLevel
- [ ] Test : TestWithTraceID_IncludesTraceID
- [ ] Test : TestLogOutput_JSONFormat

---

# PHASE 2 : Schémas Avro

## Étape 2.1 : Schémas du domaine Bancaire

### 2.1.1 : Schéma CompteOuvert
- [ ] Créer `schemas/bancaire/compte-ouvert.avsc`
- [ ] Champ : event_id (string, UUID)
- [ ] Champ : timestamp (timestamp-millis)
- [ ] Champ : compte_id (string)
- [ ] Champ : client_id (string)
- [ ] Champ : type_compte (enum: COURANT, EPARGNE, JOINT)
- [ ] Champ : devise (string, default EUR)
- [ ] Champ : solde_initial (decimal 18,2)
- [ ] Champ : metadata (optional map)
- [ ] Ajouter namespace com.edalab.bancaire.events
- [ ] Ajouter documentation
- [ ] Créer `scripts/register-schemas.sh`
- [ ] Implémenter enregistrement via curl
- [ ] Exécuter et vérifier enregistrement
- [ ] Récupérer schéma via API et valider

### 2.1.2 : Schémas DepotEffectue et VirementEmis
- [ ] Créer `schemas/bancaire/depot-effectue.avsc`
- [ ] Champ : event_id
- [ ] Champ : timestamp
- [ ] Champ : compte_id
- [ ] Champ : montant (decimal)
- [ ] Champ : devise
- [ ] Champ : reference
- [ ] Champ : canal (enum: GUICHET, VIREMENT, CHEQUE, CARTE)
- [ ] Champ : metadata
- [ ] Créer `schemas/bancaire/virement-emis.avsc`
- [ ] Champ : event_id
- [ ] Champ : timestamp
- [ ] Champ : compte_source_id
- [ ] Champ : compte_destination_id
- [ ] Champ : montant
- [ ] Champ : devise
- [ ] Champ : motif
- [ ] Champ : reference
- [ ] Champ : statut (enum: INITIE, EN_COURS, COMPLETE, REJETE)
- [ ] Champ : metadata
- [ ] Modifier register-schemas.sh pour tous les schémas
- [ ] Ajouter validation des schémas avant enregistrement
- [ ] Exécuter et valider

### 2.1.3 : Génération de code Go depuis Avro
- [ ] Créer `pkg/events/` directory
- [ ] Créer `pkg/events/go.mod`
- [ ] Module : github.com/edalab/pkg/events
- [ ] Ajouter dépendance github.com/hamba/avro/v2
- [ ] Créer `scripts/generate-avro.sh`
- [ ] Parcourir fichiers .avsc dans schemas/
- [ ] Générer structs Go dans pkg/events/
- [ ] Ajouter tags JSON et Avro
- [ ] Créer `pkg/events/bancaire.go`
- [ ] Struct : CompteOuvert avec tous les champs
- [ ] Struct : DepotEffectue avec tous les champs
- [ ] Struct : VirementEmis avec tous les champs
- [ ] Enum : TypeCompte
- [ ] Enum : Canal
- [ ] Enum : StatutVirement
- [ ] Créer `pkg/events/bancaire_test.go`
- [ ] Test : TestCompteOuvert_Serialization_Avro
- [ ] Test : TestCompteOuvert_Deserialization_Avro
- [ ] Test : TestDepotEffectue_RoundTrip
- [ ] Test : TestVirementEmis_RoundTrip
- [ ] Exécuter les tests

---

# PHASE 3 : Service Simulator

## Étape 3.1 : Structure du service Simulator

### 3.1.1 : Scaffolding du service
- [ ] Créer `services/simulator/go.mod`
- [ ] Module : github.com/edalab/services/simulator
- [ ] Ajouter imports des packages pkg/*
- [ ] Créer `services/simulator/cmd/simulator/`
- [ ] Créer `services/simulator/cmd/simulator/main.go`
- [ ] Créer `services/simulator/internal/generator/`
- [ ] Créer `services/simulator/internal/scenario/`
- [ ] Créer `services/simulator/internal/api/`
- [ ] Dans main.go : charger configuration
- [ ] Dans main.go : initialiser logger
- [ ] Dans main.go : initialiser tracer
- [ ] Dans main.go : créer producteur Kafka
- [ ] Dans main.go : démarrer serveur HTTP API
- [ ] Dans main.go : démarrer serveur métriques
- [ ] Dans main.go : gérer graceful shutdown
- [ ] Créer `services/simulator/Dockerfile`
- [ ] Stage 1 : golang:1.21-alpine (build)
- [ ] Stage 2 : alpine:3.19 (runtime)
- [ ] Exposer ports 8080 et 9090
- [ ] Tester que le service démarre

### 3.1.2 : Générateur de données fictives
- [ ] Créer `services/simulator/internal/generator/fake_data.go`
- [ ] Implémenter struct FakeDataGenerator
- [ ] Implémenter NewFakeDataGenerator(seed)
- [ ] Implémenter GenerateClientID()
- [ ] Implémenter GenerateCompteID()
- [ ] Implémenter GenerateNom()
- [ ] Implémenter GeneratePrenom()
- [ ] Implémenter GenerateMontant(min, max)
- [ ] Implémenter GenerateIBAN() (format FR valide)
- [ ] Implémenter GenerateReference()
- [ ] Créer liste de noms français courants
- [ ] Créer liste de prénoms français courants
- [ ] Implémenter calcul clé IBAN
- [ ] Créer `services/simulator/internal/generator/fake_data_test.go`
- [ ] Test : TestGenerateIBAN_ValidFormat
- [ ] Test : TestGenerateIBAN_ValidChecksum
- [ ] Test : TestGenerateMontant_InRange
- [ ] Test : TestGenerateClientID_UniqueFormat
- [ ] Test : TestDeterministicWithSameSeed
- [ ] Exécuter les tests

### 3.1.3 : Générateur d'événements CompteOuvert
- [ ] Créer `services/simulator/internal/generator/compte_ouvert.go`
- [ ] Implémenter struct CompteOuvertGenerator
- [ ] Implémenter NewCompteOuvertGenerator()
- [ ] Implémenter Generate(ctx)
  - [ ] Générer données fictives
  - [ ] Créer événement CompteOuvert
  - [ ] Produire dans Kafka
  - [ ] Retourner événement
- [ ] Implémenter GenerateBatch(ctx, count, interval)
  - [ ] Boucle count fois
  - [ ] Respecter interval entre chaque
  - [ ] Respecter contexte pour annulation
  - [ ] Logger chaque événement
- [ ] Créer `services/simulator/internal/generator/compte_ouvert_test.go`
- [ ] Test : TestGenerate_ProducesValidEvent
- [ ] Test : TestGenerate_EventInKafka
- [ ] Test : TestGenerateBatch_ProducesAllEvents
- [ ] Test : TestGenerateBatch_RespectsInterval
- [ ] Test : TestGenerateBatch_CancellableViaContext
- [ ] Exécuter les tests avec Kafka réel

## Étape 3.2 : API de contrôle du Simulator

### 3.2.1 : Endpoints REST
- [ ] Créer `services/simulator/internal/api/handler.go`
- [ ] Ajouter dépendance chi ou gorilla/mux
- [ ] Implémenter POST /api/v1/simulation/start
  - [ ] Parser body (scenario, rate, duration)
  - [ ] Démarrer simulation
  - [ ] Retourner simulation_id et status
- [ ] Implémenter POST /api/v1/simulation/stop
  - [ ] Arrêter simulation
  - [ ] Retourner status et events_produced
- [ ] Implémenter GET /api/v1/simulation/status
  - [ ] Retourner status complet
  - [ ] Inclure rate_actual
- [ ] Implémenter POST /api/v1/events/produce
  - [ ] Parser body (event_type, count)
  - [ ] Produire événements
  - [ ] Retourner event_ids
- [ ] Créer `services/simulator/internal/api/handler_test.go`
- [ ] Test : TestStartSimulation_Success
- [ ] Test : TestStartSimulation_AlreadyRunning
- [ ] Test : TestStopSimulation_Success
- [ ] Test : TestStopSimulation_NotRunning
- [ ] Test : TestGetStatus_Running
- [ ] Test : TestGetStatus_Stopped
- [ ] Test : TestProduceEvents_Success

### 3.2.2 : Gestionnaire de simulation
- [ ] Créer `services/simulator/internal/simulation/manager.go`
- [ ] Implémenter struct SimulationManager
- [ ] Implémenter struct SimulationStatus
- [ ] Implémenter struct SimulationConfig
- [ ] Implémenter NewSimulationManager()
- [ ] Implémenter Start(ctx, config)
  - [ ] Vérifier aucune simulation en cours
  - [ ] Créer contexte avec cancel
  - [ ] Démarrer goroutine génération
  - [ ] Respecter rate spécifié
  - [ ] Arrêt auto après duration (si > 0)
- [ ] Implémenter Stop()
- [ ] Implémenter Status()
- [ ] Créer `services/simulator/internal/simulation/manager_test.go`
- [ ] Test : TestStart_Success
- [ ] Test : TestStart_AlreadyRunning
- [ ] Test : TestStop_Success
- [ ] Test : TestStop_NotRunning
- [ ] Test : TestAutoStopAfterDuration
- [ ] Test : TestRateControl

### 3.2.3 : Test d'intégration Simulator complet
- [ ] Créer `services/simulator/integration_test.go`
- [ ] Ajouter build tag //go:build integration
- [ ] Implémenter TestSimulatorE2E
  - [ ] Démarrer service Simulator
  - [ ] POST /simulation/start (rate=5, duration=10)
  - [ ] Attendre 10 secondes
  - [ ] GET /simulation/status
  - [ ] Vérifier ~50 événements
  - [ ] Consommer événements Kafka
  - [ ] Vérifier schéma Avro
- [ ] Implémenter TestSimulatorManualStop
  - [ ] Démarrer simulation (duration=0)
  - [ ] Attendre 5 secondes
  - [ ] POST /simulation/stop
  - [ ] Vérifier arrêt
  - [ ] Vérifier compte événements
- [ ] Implémenter TestSimulatorProduceSingle
  - [ ] POST /events/produce (count=1)
  - [ ] Consommer événement Kafka
  - [ ] Vérifier contenu
- [ ] Ajouter Simulator à docker-compose.yml (profil services)
- [ ] Exécuter les tests

---

# PHASE 4 : Service Bancaire

## Étape 4.1 : Structure du service Bancaire

### 4.1.1 : Scaffolding et modèle de données
- [ ] Créer `services/bancaire/go.mod`
- [ ] Module : github.com/edalab/services/bancaire
- [ ] Créer structure répertoires (cmd, internal, migrations)
- [ ] Créer `services/bancaire/internal/domain/compte.go`
- [ ] Définir struct Compte
- [ ] Définir struct Transaction
- [ ] Créer `services/bancaire/migrations/001_create_comptes.sql`
- [ ] CREATE TABLE bancaire.comptes
- [ ] CREATE TABLE bancaire.transactions
- [ ] CREATE INDEX idx_comptes_client
- [ ] CREATE INDEX idx_transactions_compte
- [ ] Créer `services/bancaire/cmd/bancaire/main.go` (placeholder)
- [ ] Créer `services/bancaire/Dockerfile`

### 4.1.2 : Repository Pattern
- [ ] Créer `services/bancaire/internal/repository/compte_repository.go`
- [ ] Définir interface CompteRepository
- [ ] Méthode : Create()
- [ ] Méthode : GetByID()
- [ ] Méthode : GetByClientID()
- [ ] Méthode : UpdateSolde()
- [ ] Méthode : AddTransaction()
- [ ] Méthode : GetTransactions()
- [ ] Implémenter struct PostgresCompteRepository
- [ ] Implémenter NewPostgresCompteRepository()
- [ ] Implémenter toutes les méthodes avec SQL paramétré
- [ ] Utiliser transactions pour opérations multi-tables
- [ ] Créer `services/bancaire/internal/repository/compte_repository_test.go`
- [ ] Test : TestCreate_Success
- [ ] Test : TestCreate_DuplicateID
- [ ] Test : TestGetByID_Found
- [ ] Test : TestGetByID_NotFound
- [ ] Test : TestUpdateSolde_Success
- [ ] Test : TestAddTransaction_Success
- [ ] Test : TestAddTransaction_CompteNotFound
- [ ] Test : TestGetTransactions_WithLimit
- [ ] Nettoyer données après chaque test

### 4.1.3 : Handler d'événements Kafka
- [ ] Créer `services/bancaire/internal/handler/event_handler.go`
- [ ] Implémenter struct EventHandler
- [ ] Implémenter NewEventHandler()
- [ ] Implémenter HandleCompteOuvert()
  - [ ] Créer compte en base
  - [ ] Ajouter transaction initiale si solde > 0
  - [ ] Logger opération
  - [ ] Gérer idempotence
- [ ] Implémenter HandleDepotEffectue()
  - [ ] Récupérer compte
  - [ ] Mettre à jour solde
  - [ ] Ajouter transaction
  - [ ] Logger
- [ ] Implémenter HandleVirementEmis()
  - [ ] Vérifier compte source
  - [ ] Vérifier solde suffisant
  - [ ] Débiter compte
  - [ ] Ajouter transaction
- [ ] Implémenter Route()
  - [ ] Déterminer type événement
  - [ ] Appeler handler approprié
- [ ] Créer `services/bancaire/internal/handler/event_handler_test.go`
- [ ] Test : TestHandleCompteOuvert_Success
- [ ] Test : TestHandleCompteOuvert_Idempotent
- [ ] Test : TestHandleDepotEffectue_Success
- [ ] Test : TestHandleDepotEffectue_CompteNotFound
- [ ] Test : TestHandleVirementEmis_Success
- [ ] Test : TestHandleVirementEmis_SoldeInsuffisant
- [ ] Test : TestRoute_CompteOuvert
- [ ] Test : TestRoute_UnknownEvent

## Étape 4.2 : Intégration du service Bancaire

### 4.2.1 : Main et bootstrap
- [ ] Compléter `services/bancaire/cmd/bancaire/main.go`
- [ ] Charger configuration
- [ ] Initialiser logger
- [ ] Initialiser tracer
- [ ] Enregistrer métriques
- [ ] Créer pool database
- [ ] Créer repository
- [ ] Créer consumer Kafka
- [ ] Subscribe aux topics
- [ ] Créer event handler
- [ ] Créer API handler
- [ ] Créer HTTP server
- [ ] Créer metrics server
- [ ] Démarrer consumer (goroutine)
- [ ] Démarrer HTTP servers
- [ ] Implémenter graceful shutdown
- [ ] Compléter Dockerfile
- [ ] Ajouter au docker-compose.yml

### 4.2.2 : API REST pour queries
- [ ] Créer `services/bancaire/internal/api/handler.go`
- [ ] Implémenter GET /api/v1/health
  - [ ] Vérifier connexion Kafka
  - [ ] Vérifier connexion PostgreSQL
- [ ] Implémenter GET /api/v1/comptes/:id
- [ ] Implémenter GET /api/v1/comptes/:id/transactions
  - [ ] Supporter query param ?limit
- [ ] Implémenter GET /api/v1/clients/:client_id/comptes
- [ ] Créer `services/bancaire/internal/api/handler_test.go`
- [ ] Test : TestHealthCheck_AllHealthy
- [ ] Test : TestGetCompte_Found
- [ ] Test : TestGetCompte_NotFound
- [ ] Test : TestGetTransactions_Success
- [ ] Test : TestGetTransactions_EmptyList
- [ ] Test : TestGetComptesByClient_Success

### 4.2.3 : Test d'intégration Simulator → Bancaire
- [ ] Créer `tests/integration/simulator_bancaire_test.go`
- [ ] Ajouter build tag //go:build integration
- [ ] Implémenter TestFluxCompteOuvert_E2E
  - [ ] Démarrer services via docker-compose
  - [ ] POST events/produce (CompteOuvert)
  - [ ] Attendre traitement
  - [ ] GET /comptes/:id
  - [ ] Vérifier données
- [ ] Implémenter TestFluxDepot_E2E
  - [ ] Créer compte via Simulator
  - [ ] Attendre traitement
  - [ ] Produire DepotEffectue
  - [ ] Vérifier solde mis à jour
  - [ ] Vérifier transaction existe
- [ ] Implémenter TestFluxMultipleEvents_E2E
  - [ ] Simulation 10s à 5 events/s
  - [ ] Attendre 15s
  - [ ] Vérifier tous comptes créés
  - [ ] Vérifier cohérence données
- [ ] Implémenter TestIdempotence_E2E
  - [ ] Produire même événement 2 fois
  - [ ] Vérifier 1 seul compte créé
- [ ] Ajouter target Makefile : test-e2e
- [ ] Exécuter les tests

---

# PHASE 5 : Service Gateway

## Étape 5.1 : Gateway REST

### 5.1.1 : Structure et routing
- [ ] Créer `services/gateway/go.mod`
- [ ] Module : github.com/edalab/services/gateway
- [ ] Créer structure répertoires
- [ ] Créer `services/gateway/internal/proxy/service_proxy.go`
- [ ] Implémenter struct ServiceProxy
- [ ] Implémenter NewServiceProxy()
- [ ] Implémenter ForwardToSimulator()
  - [ ] Transférer headers (sauf Host)
  - [ ] Propager trace context
  - [ ] Gérer timeouts
  - [ ] Logger requêtes
- [ ] Implémenter ForwardToBancaire()
- [ ] Créer `services/gateway/internal/api/router.go`
- [ ] Route : /api/v1/simulation/* → Simulator
- [ ] Route : /api/v1/bancaire/* → Bancaire
- [ ] Route : /api/v1/health → Health check agrégé
- [ ] Route : /ws → WebSocket
- [ ] Configurer middleware CORS pour web-ui
- [ ] Créer `services/gateway/internal/proxy/service_proxy_test.go`
- [ ] Test : TestForwardToSimulator_Success
- [ ] Test : TestForwardToBancaire_Success
- [ ] Test : TestForward_ServiceUnavailable
- [ ] Test : TestForward_PropagatesHeaders

### 5.1.2 : WebSocket Hub
- [ ] Créer `services/gateway/internal/websocket/hub.go`
- [ ] Implémenter struct Client
- [ ] Implémenter struct Hub
- [ ] Implémenter struct Message
- [ ] Implémenter NewHub()
- [ ] Implémenter Run()
- [ ] Implémenter Broadcast()
- [ ] Implémenter BroadcastToTopic()
- [ ] Créer `services/gateway/internal/websocket/client.go`
- [ ] Implémenter ReadPump()
- [ ] Implémenter WritePump()
- [ ] Gérer messages subscribe/unsubscribe
- [ ] Créer `services/gateway/internal/websocket/hub_test.go`
- [ ] Test : TestHub_RegisterClient
- [ ] Test : TestHub_UnregisterClient
- [ ] Test : TestHub_Broadcast
- [ ] Test : TestHub_BroadcastToTopic
- [ ] Test : TestClient_Subscribe
- [ ] Test : TestClient_Unsubscribe

### 5.1.3 : Consommateur Kafka pour WebSocket
- [ ] Créer `services/gateway/internal/streaming/kafka_streamer.go`
- [ ] Implémenter struct KafkaStreamer
- [ ] Implémenter NewKafkaStreamer()
- [ ] Implémenter Start()
  - [ ] Souscrire aux topics
  - [ ] Pour chaque message : désérialiser
  - [ ] Créer Message WebSocket
  - [ ] Broadcast au hub
- [ ] Implémenter Stop()
- [ ] Créer `services/gateway/cmd/gateway/main.go`
- [ ] Intégrer ServiceProxy
- [ ] Intégrer WebSocket Hub
- [ ] Intégrer KafkaStreamer
- [ ] Intégrer serveur HTTP
- [ ] Créer `services/gateway/integration_test.go`
- [ ] Test : TestWebSocket_ReceivesKafkaEvents
- [ ] Test : TestWebSocket_MultipleClients
- [ ] Créer Dockerfile
- [ ] Ajouter au docker-compose.yml

---

# PHASE 6 : Observabilité

## Étape 6.1 : Configuration Prometheus

### 6.1.1 : Configuration et scraping
- [ ] Créer `infra/prometheus/prometheus.yml`
- [ ] Configurer global scrape_interval: 15s
- [ ] Configurer job prometheus
- [ ] Configurer job kafka
- [ ] Configurer job simulator
- [ ] Configurer job bancaire
- [ ] Configurer job gateway
- [ ] Ajouter Prometheus à docker-compose.yml
- [ ] Image : prom/prometheus:v2.47.0
- [ ] Port : 9090
- [ ] Volume pour prometheus.yml
- [ ] Healthcheck
- [ ] Créer `scripts/test-prometheus.sh`
- [ ] Attendre Prometheus prêt
- [ ] Query API pour vérifier targets
- [ ] Vérifier services UP
- [ ] Exécuter et valider

### 6.1.2 : Métriques Kafka avec JMX Exporter
- [ ] Modifier config Kafka pour JMX
- [ ] KAFKA_JMX_PORT: 9999
- [ ] KAFKA_JMX_HOSTNAME: kafka
- [ ] Ajouter conteneur JMX Exporter
- [ ] Image : bitnami/jmx-exporter:0.19.0
- [ ] Port : 5556
- [ ] Créer `infra/kafka/jmx-exporter-config.yml`
- [ ] Métriques BrokerTopicMetrics
- [ ] Métriques ReplicaManager
- [ ] Métriques consumer-fetch-manager
- [ ] Modifier prometheus.yml pour scraper JMX
- [ ] Vérifier métriques Kafka dans Prometheus

## Étape 6.2 : Dashboards Grafana

### 6.2.1 : Configuration Grafana
- [ ] Ajouter Grafana à docker-compose.yml
- [ ] Image : grafana/grafana:10.2.0
- [ ] Port : 3000
- [ ] Volumes pour provisioning
- [ ] Créer `infra/grafana/provisioning/datasources/prometheus.yml`
- [ ] Configurer datasource Prometheus
- [ ] Créer `infra/grafana/provisioning/dashboards/dashboards.yml`
- [ ] Configurer provider EDA-Lab
- [ ] Vérifier Grafana démarre
- [ ] Vérifier connexion Prometheus

### 6.2.2 : Dashboard Kafka Overview
- [ ] Créer `infra/grafana/dashboards/kafka-overview.json`
- [ ] Row 1 : Vue d'ensemble
  - [ ] Stat Messages/sec
  - [ ] Stat Topics actifs
  - [ ] Stat Partitions totales
- [ ] Row 2 : Messages
  - [ ] Time series Messages produits
  - [ ] Time series Messages consommés
- [ ] Row 3 : Consumer Lag
  - [ ] Time series Lag par group
  - [ ] Table Top 10 partitions
- [ ] Row 4 : Bytes
  - [ ] Time series Bytes in/out
- [ ] Variables $topic et $consumer_group
- [ ] Vérifier dashboard fonctionnel

### 6.2.3 : Dashboard Services
- [ ] Créer `infra/grafana/dashboards/services-overview.json`
- [ ] Row 1 : Santé services
  - [ ] Status Simulator
  - [ ] Status Bancaire
  - [ ] Status Gateway
- [ ] Row 2 : Événements
  - [ ] messages_produced_total
  - [ ] messages_consumed_total
- [ ] Row 3 : Latences
  - [ ] Heatmap message_latency
  - [ ] Heatmap processing_latency
- [ ] Row 4 : Erreurs
  - [ ] processing_errors_total
  - [ ] Rate d'erreur (%)
- [ ] Row 5 : Ressources
  - [ ] go_memstats_alloc_bytes
  - [ ] go_goroutines
- [ ] Variable $service
- [ ] Vérifier dashboard fonctionnel

---

# PHASE 7 : Web UI

## Étape 7.1 : Setup React

### 7.1.1 : Initialisation du projet React
- [ ] Dans web-ui/ : `npm create vite@latest . -- --template react-ts`
- [ ] Installer @xyflow/react
- [ ] Installer @tanstack/react-query
- [ ] Installer axios
- [ ] Installer zustand
- [ ] Installer tailwindcss postcss autoprefixer
- [ ] Installer lucide-react
- [ ] Configurer Tailwind CSS
- [ ] Créer structure App.tsx de base
- [ ] Header avec titre et status
- [ ] Sidebar avec navigation
- [ ] Main content area
- [ ] Footer avec métriques
- [ ] Créer `web-ui/Dockerfile`
- [ ] Stage 1 : node:20-alpine (build)
- [ ] Stage 2 : nginx:alpine (serve)
- [ ] Vérifier `npm run dev` fonctionne

### 7.1.2 : Configuration API et WebSocket
- [ ] Créer `web-ui/src/lib/api.ts`
- [ ] Configurer API_BASE_URL
- [ ] Créer instance axios
- [ ] Créer simulationApi
  - [ ] start()
  - [ ] stop()
  - [ ] status()
  - [ ] produceEvent()
- [ ] Créer bancaireApi
  - [ ] getCompte()
  - [ ] getTransactions()
- [ ] Créer `web-ui/src/lib/websocket.ts`
- [ ] Implémenter class EventSocket
- [ ] Méthode connect()
- [ ] Méthode disconnect()
- [ ] Méthode subscribe()
- [ ] Méthode unsubscribe()
- [ ] Créer `web-ui/src/lib/api.test.ts`
- [ ] Installer msw (Mock Service Worker)
- [ ] Test : TestSimulationApi_Start
- [ ] Test : TestSimulationApi_Status
- [ ] Test : TestBancaireApi_GetCompte

## Étape 7.2 : Composants UI

### 7.2.1 : Contrôles de simulation
- [ ] Créer `web-ui/src/components/SimulationControls.tsx`
- [ ] Bouton Start/Stop (toggle)
- [ ] Slider pour rate
- [ ] Input pour durée
- [ ] Dropdown scénario
- [ ] Affichage status
- [ ] Compteur événements
- [ ] Créer `web-ui/src/components/EventProducer.tsx`
- [ ] Dropdown type événement
- [ ] Input nombre événements
- [ ] Bouton "Produire"
- [ ] Feedback succès/erreur
- [ ] Créer `web-ui/src/components/__tests__/SimulationControls.test.tsx`
- [ ] Test : TestRender_Initial
- [ ] Test : TestStart_CallsOnStart
- [ ] Test : TestStop_CallsOnStop
- [ ] Test : TestSlider_UpdatesRate

### 7.2.2 : Visualisation React Flow
- [ ] Créer `web-ui/src/components/FlowVisualization.tsx`
- [ ] Node Simulator (source)
- [ ] Node Kafka (broker)
- [ ] Node Bancaire (consumer)
- [ ] Node Client 360 (préparation)
- [ ] Edge Simulator → Kafka (animé)
- [ ] Edge Kafka → Bancaire (animé)
- [ ] Animation selon type événement
- [ ] Compteur événements sur edges
- [ ] Créer `web-ui/src/components/FlowNode.tsx`
- [ ] Icône du service
- [ ] Nom du service
- [ ] Status indicator (couleur)
- [ ] Compteur événements
- [ ] Sparkline 60 secondes
- [ ] Créer `web-ui/src/hooks/useFlowData.ts`
- [ ] Connexion WebSocket
- [ ] Écoute événements
- [ ] Mise à jour compteurs
- [ ] Déclenchement animations
- [ ] Tests visuels

### 7.2.3 : Dashboard de métriques
- [ ] Créer `web-ui/src/components/MetricsDashboard.tsx`
- [ ] Section Simulation
  - [ ] Card événements produits
  - [ ] Card rate actuel
  - [ ] Card durée simulation
- [ ] Section Kafka
  - [ ] Card messages topics
  - [ ] Card consumer lag
  - [ ] Mini graphique messages/sec
- [ ] Section Services
  - [ ] Status chaque service
  - [ ] Latence moyenne
  - [ ] Erreurs récentes
- [ ] Créer `web-ui/src/hooks/useMetrics.ts`
- [ ] Poll métriques toutes les 5s
- [ ] Ou WebSocket temps réel
- [ ] Calcul moyennes mobiles
- [ ] Créer `web-ui/src/components/MetricCard.tsx`
- [ ] Titre
- [ ] Valeur principale
- [ ] Tendance (↑ ↓ →)
- [ ] Sparkline optionnel

## Étape 7.3 : Intégration finale UI

### 7.3.1 : Layout et navigation
- [ ] Compléter `web-ui/src/App.tsx`
- [ ] Layout avec Header/Sidebar/Main/Footer
- [ ] Intégrer SimulationControls dans Sidebar
- [ ] Intégrer EventProducer dans Sidebar
- [ ] Intégrer FlowVisualization dans Main
- [ ] Intégrer MetricsDashboard dans Main
- [ ] Footer avec stats globales
- [ ] Créer `web-ui/src/store/simulationStore.ts`
- [ ] State : simulationStatus
- [ ] State : metrics
- [ ] State : events (derniers 100)
- [ ] State : connectionStatus
- [ ] Action : startSimulation
- [ ] Action : stopSimulation
- [ ] Action : updateMetrics
- [ ] Action : addEvent

### 7.3.2 : Tests E2E UI
- [ ] Installer @playwright/test
- [ ] Créer `web-ui/e2e/simulation.spec.ts`
- [ ] Test : démarre et arrête simulation
  - [ ] Naviguer page
  - [ ] Click Start
  - [ ] Vérifier status running
  - [ ] Attendre 5s
  - [ ] Vérifier compteur
  - [ ] Click Stop
  - [ ] Vérifier status stopped
- [ ] Test : produit événement manuel
  - [ ] Sélectionner CompteOuvert
  - [ ] Entrer count=1
  - [ ] Click Produire
  - [ ] Vérifier toast succès
  - [ ] Vérifier compteur
- [ ] Test : visualisation s'anime
  - [ ] Démarrer simulation
  - [ ] Vérifier edges animés
  - [ ] Vérifier compteurs nodes
- [ ] Test : WebSocket reconnecte
  - [ ] Vérifier connexion
  - [ ] Simuler déconnexion
  - [ ] Vérifier reconnexion
- [ ] Créer `web-ui/playwright.config.ts`
- [ ] Exécuter tests E2E

---

# PHASE 8 : Intégration finale

## Étape 8.1 : Tests E2E complets

### 8.1.1 : Scénarios de test E2E
- [ ] Créer `tests/e2e/mvp_test.go`
- [ ] Ajouter build tag //go:build e2e
- [ ] Implémenter TestMVP_FullFlow
  - [ ] make infra-up
  - [ ] make services-up
  - [ ] POST simulation/start (rate=10, duration=60)
  - [ ] Attendre 60s
  - [ ] Vérifier ~600 événements
  - [ ] Vérifier comptes dans Bancaire
  - [ ] Vérifier métriques Prometheus
  - [ ] Arrêter proprement
- [ ] Implémenter TestMVP_ChaosConsumerRestart
  - [ ] Démarrer simulation continue
  - [ ] Arrêter service Bancaire
  - [ ] Attendre 10s
  - [ ] Redémarrer Bancaire
  - [ ] Vérifier aucune perte
- [ ] Implémenter TestMVP_HighThroughput
  - [ ] Simulation rate=100
  - [ ] Vérifier système supporte
  - [ ] Mesurer latences
  - [ ] Vérifier aucune erreur
- [ ] Créer `tests/e2e/docker-compose.e2e.yml`

### 8.1.2 : Script de validation complète
- [ ] Créer `scripts/validate-mvp.sh`
- [ ] Étape 1 : Starting infrastructure
- [ ] Étape 2 : Running infrastructure tests
- [ ] Étape 3 : Starting services
- [ ] Étape 4 : Running unit tests
- [ ] Étape 5 : Running integration tests
- [ ] Étape 6 : Running E2E tests
- [ ] Étape 7 : Checking Prometheus targets
- [ ] Étape 8 : Checking Grafana
- [ ] Étape 9 : Running simulation test
- [ ] Étape 10 : Cleanup
- [ ] Ajouter target Makefile : validate-mvp
- [ ] Exécuter validation complète

### 8.1.3 : Tests de performance
- [ ] Créer `tests/e2e/performance_test.go`
- [ ] Ajouter build tag //go:build performance
- [ ] Implémenter TestPerformance_Throughput
  - [ ] Simulation rate=100 events/s
  - [ ] Mesurer throughput réel
  - [ ] Vérifier throughput >= 90%
  - [ ] Mesurer latence moyenne
- [ ] Implémenter TestPerformance_Latency
  - [ ] Produire 1000 événements
  - [ ] Mesurer temps production → persistance
  - [ ] Calculer P50, P95, P99
  - [ ] Vérifier P99 < 500ms
- [ ] Implémenter TestPerformance_Burst
  - [ ] Produire 1000 événements en 1s
  - [ ] Vérifier traitement dans 30s
  - [ ] Mesurer temps de récupération
- [ ] Implémenter TestPerformance_Sustained
  - [ ] Simulation rate=50 pendant 5 min
  - [ ] Vérifier stabilité mémoire
  - [ ] Vérifier stabilité latence
- [ ] Créer `scripts/run-performance-tests.sh`
- [ ] Créer `infra/grafana/dashboards/performance.json`
- [ ] Ajouter target Makefile : test-performance
- [ ] Ajouter target Makefile : report-performance

## Étape 8.2 : Documentation finale

### 8.2.1 : README et guides
- [ ] Compléter `README.md`
- [ ] Quick Start
  - [ ] Prérequis
  - [ ] Démarrage rapide
- [ ] Architecture (diagramme ASCII)
- [ ] Utilisation
  - [ ] Démarrer simulation
  - [ ] Visualiser flux
  - [ ] Consulter métriques
- [ ] Développement
  - [ ] Structure projet
  - [ ] Lancer tests
- [ ] License

### 8.2.2 : Documentation technique
- [ ] Créer `docs/ARCHITECTURE.md`
- [ ] Diagramme C4 - Context
- [ ] Diagramme C4 - Container
- [ ] Diagramme C4 - Component
- [ ] Flux de données CompteOuvert
- [ ] Référence aux ADR
- [ ] Créer `docs/adr/` directory
- [ ] Créer ADR 001-choix-kafka.md
- [ ] Créer ADR 002-choix-avro.md
- [ ] Créer ADR 003-choix-go.md
- [ ] Créer ADR 004-choix-postgresql.md
- [ ] Créer ADR 005-choix-react-flow.md

### 8.2.3 : Guide du patron Pub/Sub
- [ ] Créer `docs/patterns/01-pub-sub.md`
- [ ] Section Concept
- [ ] Diagramme patron
- [ ] Section Implémentation
  - [ ] Code Producteur
  - [ ] Configuration Kafka
  - [ ] Code Consommateur
- [ ] Section Expérimentation
  - [ ] Exercice 1 : Observer découplage
  - [ ] Exercice 2 : Scaling consommateurs
- [ ] Section Compromis architecturaux
- [ ] Liens pour aller plus loin

---

# Checklist de validation MVP

## Infrastructure
- [ ] Kafka KRaft fonctionne
- [ ] Schema Registry fonctionne
- [ ] PostgreSQL fonctionne
- [ ] Prometheus fonctionne
- [ ] Grafana fonctionne
- [ ] Tous les services démarrent

## Fonctionnalités
- [ ] Simulation démarre/arrête
- [ ] Événements générés automatiquement
- [ ] Événements consommés par Bancaire
- [ ] Comptes créés en base
- [ ] API REST fonctionnelle
- [ ] WebSocket temps réel fonctionne
- [ ] UI affiche les flux
- [ ] Dashboards Grafana affichent métriques

## Tests
- [ ] Tests unitaires passent
- [ ] Tests d'intégration passent
- [ ] Tests E2E passent
- [ ] Tests de performance passent
- [ ] Script validate-mvp.sh passe

## Documentation
- [ ] README complet
- [ ] Documentation architecture
- [ ] ADR documentés
- [ ] Guide patron Pub/Sub

---

**Dernière mise à jour :** 2026-01-19
**Progression totale :** 0/51 sous-étapes (0%)
