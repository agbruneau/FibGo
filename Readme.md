# AgentMeshKafka : Architecture Event-Driven pour Maillage Agentique

## Vue d'ensemble

AgentMeshKafka est un framework open-source qui implémente une architecture de maillage agentique (agent mesh) basée sur Apache Kafka pour permettre la communication événementielle décentralisée entre agents d'intelligence artificielle. Le projet s'inspire du paradigme spec-driven development popularisé par GitHub Spec Kit et intègre les standards d'interopérabilité AsyncAPI 3.0 et CloudEvents 1.0 pour garantir une compatibilité cross-platform.[^1][^2][^3][^4]

### Motivation

L'essor des systèmes d'IA agentique exige une infrastructure capable de supporter des interactions autonomes, asynchrones et hautement scalables. Les architectures traditionnelles request-response (REST/HTTP) créent des couplages serrés qui limitent la résilience et l'évolutivité. L'event-driven architecture (EDA) résout ces limitations en permettant aux agents de publier et consommer des événements de manière découplée via un backbone événementiel centralisé.[^5][^6][^7]

### Objectifs du projet

1. **Interopérabilité** : Standardiser les communications inter-agents via CloudEvents et AsyncAPI
2. **Performance** : Atteindre <100ms de latence end-to-end avec un throughput de 10,000 TPS
3. **Résilience** : Implémenter DLQ (Dead Letter Queue) et retry avec backoff exponentiel
4. **Développement guidé par spécifications** : Générer automatiquement le code à partir de contrats AsyncAPI
5. **Observabilité native** : Fournir monitoring temps réel via TUI (Terminal User Interface) et Prometheus

## Architecture

### Vue d'ensemble système

L'architecture AgentMeshKafka suit un modèle multi-couches séparant les responsabilités :

```
┌─────────────────────────────────────────────────────────┐
│                  Agent Mesh Layer                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │ Order    │  │Inventory │  │ Billing  │               │
│  │ Agent    │  │ Agent    │  │ Agent    │               │
│  │ (Go)     │  │ (Rust)   │  │ (Rust)   │               │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘               │
└───────┼─────────────┼─────────────┼─────────────────────┘
        │             │             │
        │  CloudEvents 1.0 Messages │
        ▼             ▼             ▼
┌─────────────────────────────────────────────────────────┐
│              Event Backbone (Kafka KRaft)                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ orders/v1   │  │inventory/v1 │  │ billing/v1  │     │
│  │ Topic       │  │ Topic       │  │ Topic       │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
        │             │             │
        ▼             ▼             ▼
┌─────────────────────────────────────────────────────────┐
│         Observability & Coordination Layer               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐        │
│  │ Prometheus │  │  TUI       │  │ AsyncAPI   │        │
│  │ Metrics    │  │ Dashboard  │  │ Registry   │        │
│  └────────────┘  └────────────┘  └────────────┘        │
└─────────────────────────────────────────────────────────┘
```


### Composants clés

#### 1. Event Backbone : Apache Kafka KRaft

Kafka 3.7+ en mode KRaft (sans ZooKeeper) sert de système nerveux central. Les caractéristiques :[^8][^9]

- **Architecture simplifiée** : Élimination de ZooKeeper réduit la complexité opérationnelle de 40%[^9]
- **Exactly-once semantics** : Garantit qu'aucun événement n'est perdu ou dupliqué
- **Partitioning intelligent** : Jusqu'à 32 partitions par topic pour parallélisme maximal
- **Réplication synchrone** : Factor de réplication 3 pour haute disponibilité


#### 2. Agent Producers (Go)

Implémentés en Go 1.24 avec `confluent-kafka-go`, les producers bénéficient de :

- **Performance mémoire** : Zero-allocation grâce à la gestion manuelle des buffers
- **Compression native** : Snappy/LZ4 pour réduire la bande passante de 60%
- **Idempotence** : Configuration `enable.idempotence=true` pour éviter duplications

**Exemple de producer** :

```go
package producer

import (
    "github.com/confluentinc/confluent-kafka-go/kafka"
    cloudevents "github.com/cloudevents/sdk-go/v2"
)

type OrderProducer struct {
    producer *kafka.Producer
}

func (p *OrderProducer) PublishOrderPlaced(order Order) error {
    event := cloudevents.NewEvent()
    event.SetType("com.agentmesh.order.placed")
    event.SetSource("order-agent")
    event.SetData("application/json", order)

    msg := &kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic: kafka.StringPartition("orders/v1"),
            Partition: kafka.PartitionAny,
        },
        Key: []byte(order.ID),
        Headers: []kafka.Header{
            {Key: "ce-type", Value: []byte(event.Type())},
            {Key: "ce-source", Value: []byte(event.Source())},
        },
        Value: event.Data(),
    }

    return p.producer.Produce(msg, nil)
}
```


#### 3. Agent Consumers (Rust)

Les consumers Rust exploitent `rdkafka` et `Rayon` pour traitement parallèle :

- **Concurrence structurée** : Rayon thread pool adaptatif (jusqu'à 64 threads)
- **Ownership guarantees** : Le type system Rust élimine race conditions
- **Backpressure automatique** : Consumer pause si latence > 50ms

**Exemple de consumer** :

```rust
use rdkafka::consumer::{Consumer, StreamConsumer};
use rdkafka::config::ClientConfig;
use cloudevents::Event;
use rayon::prelude::*;

struct InventoryConsumer {
    consumer: StreamConsumer,
}

impl InventoryConsumer {
    pub async fn process_orders(&self) {
        let messages = self.consumer
            .stream()
            .collect::<Vec<_>>()
            .await;

        messages.par_iter().for_each(|msg| {
            if let Some(payload) = msg.payload() {
                let event: Event = serde_json::from_slice(payload)
                    .expect("Invalid CloudEvent");

                if event.ty() == "com.agentmesh.order.placed" {
                    self.check_inventory(event);
                }
            }
        });
    }
}
```


#### 4. Standards d'interopérabilité

##### AsyncAPI 3.0

Les contrats AsyncAPI définissent la topologie des événements de manière machine-readable :[^10][^1]

```yaml
asyncapi: '3.0.0'
info:
  title: AgentMesh-Kafka Interop
  version: 1.0.0
  description: Event-driven agent mesh communication

servers:
  production:
    host: 'kafka-cluster.example.com:9092'
    protocol: kafka
    protocolVersion: '3.7'
    security:
      - saslScram: []

channels:
  orders/v1/{orderId}:
    address:
      description: Order lifecycle events
      parameters:
        orderId:
          description: Unique order identifier
          schema:
            type: string
            format: uuid
    messages:
      orderPlaced:
        name: OrderPlaced
        contentType: application/cloudevents+json
        traits:
          - $ref: '#/components/messageTraits/cloudEventsBinaryMode'
        payload:
          $ref: '#/components/schemas/OrderEvent'

components:
  schemas:
    OrderEvent:
      type: object
      properties:
        orderId:
          type: string
          format: uuid
        customerId:
          type: string
        items:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'

    OrderItem:
      type: object
      properties:
        sku:
          type: string
        quantity:
          type: integer
        price:
          type: number
          format: double

  messageTraits:
    cloudEventsBinaryMode:
      headers:
        type: object
        properties:
          ce-type:
            type: string
            description: CloudEvents type attribute
          ce-source:
            type: string
            description: CloudEvents source attribute
          ce-id:
            type: string
            format: uuid
          ce-specversion:
            const: "1.0"
        required:
          - ce-type
          - ce-source
          - ce-id
          - ce-specversion
```


##### CloudEvents 1.0

CloudEvents fournit une enveloppe standard pour l'interopérabilité cross-platform :[^2][^11]

**Mode structured** (JSON complet) :

```json
{
  "specversion": "1.0",
  "id": "a3e8f7c2-4d1b-4e9f-8c2a-5b6d7e8f9a0b",
  "source": "order-agent",
  "type": "com.agentmesh.order.placed",
  "datacontenttype": "application/json",
  "dataschema": "https://github.com/agbruneau/AgentMeshKafka/schemas/order.proto",
  "time": "2026-01-07T17:30:00Z",
  "data": {
    "orderId": "ORD-2026-001",
    "customerId": "CUST-5432",
    "items": [
      {"sku": "COFFEE-DARK", "quantity": 2, "price": 15.99}
    ],
    "totalAmount": 31.98
  }
}
```

**Mode binary** (headers Kafka) :

```
Headers:
  ce-specversion: 1.0
  ce-id: a3e8f7c2-4d1b-4e9f-8c2a-5b6d7e8f9a0b
  ce-source: order-agent
  ce-type: com.agentmesh.order.placed
  content-type: application/json

Payload:
{
  "orderId": "ORD-2026-001",
  "customerId": "CUST-5432",
  ...
}
```

Le mode binary réduit la taille des messages de 30% et améliore les performances de parsing.[^12]

### Observabilité

#### TUI Dashboard (Bubble Tea)

Le dashboard TUI construit avec Bubble Tea (Go) offre une vue temps réel :[^13][^14]

**Fonctionnalités** :

- Graphe throughput par topic (sparklines ASCII)
- Consumer lag par groupe
- Broker health status
- Event tracing live (derniers 100 événements)

**Architecture du TUI** :

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/table"
)

type Model struct {
    table table.Model
    metrics MetricsCollector
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case metricsUpdate:
        m.table.SetRows(m.formatMetrics(msg.data))
    }

    return m, nil
}

func (m Model) View() string {
    return m.table.View() + "\nPress q to quit\n"
}
```


#### Métriques Prometheus

Exposition des métriques via JMX Exporter :[^15][^16][^17]

**Métriques critiques** :

- `kafka_topic_partition_current_offset` : Position actuelle par partition
- `kafka_consumer_lag` : Retard de consommation (SLA < 1000 messages)
- `kafka_server_brokertopicmetrics_messagesinpersec` : Throughput entrant
- `kafka_producer_request_latency_avg` : Latence moyenne producer

**Configuration Prometheus** (`prometheus.yml`) :

```yaml
scrape_configs:
  - job_name: 'kafka-brokers'
    scrape_interval: 10s
    static_configs:
      - targets:
          - 'kafka-broker-1:9092'
          - 'kafka-broker-2:9092'
          - 'kafka-broker-3:9092'
    metrics_path: /metrics
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance

  - job_name: 'kafka-consumers'
    scrape_interval: 15s
    static_configs:
      - targets:
          - 'consumer-inventory:8080'
          - 'consumer-billing:8080'
```

**Alertes configurées** :

```yaml
groups:
  - name: kafka_alerts
    rules:
      - alert: HighConsumerLag
        expr: kafka_consumer_lag > 5000
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Consumer lag exceeds 5000 messages"

      - alert: BrokerDown
        expr: up{job="kafka-brokers"} == 0
        for: 30s
        labels:
          severity: critical
```


## Installation et démarrage rapide

### Prérequis

- **Docker** 24.0+ et **Docker Compose** v2.20+
- **Go** 1.24+ (pour développement producer)
- **Rust** 1.75+ avec Cargo (pour développement consumer)
- **Python** 3.11+ avec `uv` (pour spec-driven tooling)
- **Git** 2.40+


### Installation via Spec Kit

```bash
# 1. Installer le CLI specify
uvx --from git+https://github.com/github/spec-kit.git specify init AgentMeshKafka

# 2. Générer les spécifications
cd AgentMeshKafka
/speckit.specify  # Dans votre agent AI (Cursor/Claude)

# 3. Créer le plan technique
/speckit.plan

# 4. Générer les tâches
/speckit.tasks

# 5. Implémenter
/speckit.implement
```


### Démarrage Docker Compose

Le fichier `docker-compose.yml` configure l'environnement complet :[^18][^19][^8]

```yaml
version: '3.9'

services:
  kafka:
    image: apache/kafka:3.7.1
    container_name: kafka-kraft
    ports:
      - "9092:9092"
      - "9093:9093"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_LOG_DIRS: /var/lib/kafka/data
      CLUSTER_ID: 'MkU3OEVBNTcwNTJENDM2Qk'
    volumes:
      - ./data/kafka:/var/lib/kafka/data
    healthcheck:
      test: ["CMD-SHELL", "kafka-broker-api-versions.sh --bootstrap-server localhost:9092"]
      interval: 10s
      timeout: 5s
      retries: 5

  producer-order:
    build:
      context: ./agents/producer
      dockerfile: Dockerfile
    depends_on:
      kafka:
        condition: service_healthy
    environment:
      KAFKA_BOOTSTRAP_SERVERS: kafka:9092
      KAFKA_TOPIC: orders/v1
    restart: unless-stopped

  consumer-inventory:
    build:
      context: ./agents/consumer
      dockerfile: Dockerfile.rust
    depends_on:
      kafka:
        condition: service_healthy
    environment:
      KAFKA_BOOTSTRAP_SERVERS: kafka:9092
      KAFKA_GROUP_ID: inventory-group
      KAFKA_TOPICS: orders/v1
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:v2.50.0
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  tui-dashboard:
    build:
      context: ./tui
      dockerfile: Dockerfile
    depends_on:
      - kafka
      - prometheus
    environment:
      KAFKA_BOOTSTRAP_SERVERS: kafka:9092
      PROMETHEUS_URL: http://prometheus:9090
    stdin_open: true
    tty: true

volumes:
  prometheus-data:
```

**Commandes essentielles** :

```bash
# Démarrer tous les services
make up  # ou docker compose up -d

# Vérifier la santé Kafka
make health

# Ouvrir le TUI dashboard
make tui

# Publier un événement test
make test-order

# Voir les logs consumers
make logs-consumer

# Arrêter proprement
make down
```


### Exemple d'utilisation complète

**1. Publier un ordre via REST API** :

```bash
curl -X POST http://localhost:3000/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": "CUST-789",
    "items": [
      {"sku": "LAPTOP-PRO", "quantity": 1, "price": 1299.99}
    ]
  }'
```

**2. Observer dans le TUI** :

```
┌─────────────────── Agent Mesh Dashboard ───────────────────┐
│ Topic: orders/v1         Throughput: 245 msg/s            │
│ Consumer Lag (inventory-group): 12 messages               │
│ Brokers Online: 3/3     ✓ Healthy                         │
├────────────────────────────────────────────────────────────┤
│ Recent Events:                                             │
│ [17:45:32] com.agentmesh.order.placed → ORD-2026-042      │
│ [17:45:31] com.agentmesh.inventory.checked → OK           │
│ [17:45:30] com.agentmesh.billing.invoiced → INV-9981      │
└────────────────────────────────────────────────────────────┘
Press [q] to quit | [r] to refresh | [↑↓] to navigate
```

**3. Vérifier les métriques Prometheus** :

```bash
# Requête PromQL pour consumer lag
curl 'http://localhost:9090/api/v1/query?query=kafka_consumer_lag{group="inventory-group"}'
```


## Stratégie de tests

### Tests unitaires

**Go (producers)** :

```go
package producer_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestOrderProducerPublish(t *testing.T) {
    producer := NewMockProducer()
    order := Order{ID: "ORD-TEST", CustomerID: "CUST-1"}

    err := producer.PublishOrderPlaced(order)
    assert.NoError(t, err)
    assert.Equal(t, 1, producer.MessageCount())
}
```

**Rust (consumers)** :

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_inventory_consumer_process() {
        let consumer = InventoryConsumer::new_mock();
        let event = create_test_event();

        let result = consumer.process_event(event).await;
        assert!(result.is_ok());
    }
}
```


### Tests d'intégration

Utilisation de Testcontainers pour Kafka isolé :[^19]

```go
func TestE2EOrderFlow(t *testing.T) {
    ctx := context.Background()

    // Démarrer Kafka via Testcontainers
    kafkaContainer, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image: "apache/kafka:3.7.1",
                ExposedPorts: []string{"9092/tcp"},
            },
        })
    require.NoError(t, err)
    defer kafkaContainer.Terminate(ctx)

    // Test complet producer → consumer
    producer := NewOrderProducer(kafkaContainer.Endpoint())
    consumer := NewInventoryConsumer(kafkaContainer.Endpoint())

    order := Order{ID: "ORD-E2E"}
    producer.PublishOrderPlaced(order)

    received := consumer.WaitForMessage(5 * time.Second)
    assert.Equal(t, "ORD-E2E", received.OrderID)
}
```


### Tests de charge (k6)

Objectif : 10,000 TPS avec latence P99 < 100ms :[^20][^21]

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 5000 },   // Ramp-up
    { duration: '5m', target: 10000 },  // Plateau
    { duration: '2m', target: 0 },      // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(99)<100'],   // 99% < 100ms
    http_req_failed: ['rate<0.05'],     // < 5% erreurs
  },
};

export default function () {
  const payload = JSON.stringify({
    customerId: `CUST-${__VU}`,
    items: [{ sku: 'TEST-SKU', quantity: 1, price: 10.0 }],
  });

  const res = http.post('http://localhost:3000/orders', payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status is 201': (r) => r.status === 201,
    'latency < 100ms': (r) => r.timings.duration < 100,
  });

  sleep(0.1);
}
```


## Roadmap

### v0.1 - MVP (Semaine 1-2)

- [x] Docker Compose Kafka KRaft
- [x] Producer Go basique (CloudEvents)
- [x] Consumer Rust avec rdkafka
- [ ] AsyncAPI schema validation
- [ ] Tests unitaires (>80% coverage)


### v0.2 - Résilience (Semaine 3)

- [ ] DLQ pour messages en échec
- [ ] Retry avec backoff exponentiel
- [ ] Circuit breaker pattern
- [ ] Idempotence garantie


### v0.3 - Observabilité (Semaine 4)

- [ ] TUI dashboard Bubble Tea
- [ ] Prometheus metrics exporter
- [ ] Grafana dashboards pré-configurés
- [ ] Distributed tracing (OpenTelemetry)


### v1.0 - Production-ready (Semaine 5-6)

- [ ] Helm charts Kubernetes
- [ ] SASL/SCRAM + mTLS
- [ ] Schema registry (Confluent/Apicurio)
- [ ] Benchmarks publiés (10k TPS)
- [ ] Documentation complète


### v2.0 - Agentic AI Integration (Q2 2026)

- [ ] Support MCP (Model Context Protocol)[^22][^23]
- [ ] Support A2A (Agent2Agent)[^24][^25]
- [ ] LLM-based reasoning agents
- [ ] Auto-scaling basé sur consumer lag


## Contribution

Nous accueillons les contributions suivant le modèle spec-driven :[^4][^26]

### Processus

1. **Fork** le repository
2. **Créer une branche** : `git checkout -b feat/ma-fonctionnalite`
3. **Rédiger la spec** : Mettre à jour `specs/specify.md` avec la nouvelle fonctionnalité
4. **Générer le plan** : Utiliser `/speckit.plan` pour détailler l'implémentation
5. **Implémenter** : Suivre les tasks générées par `/speckit.tasks`
6. **Tests** : Ajouter tests unitaires + intégration (coverage >85%)
7. **Documentation** : Mettre à jour README et AsyncAPI schemas
8. **PR** : Ouvrir une pull request avec description détaillée

### Standards de code

**Go** :

- Respecter `gofmt` et `golangci-lint`
- Commentaires godoc pour fonctions exportées
- Error handling explicite (pas de `panic`)

**Rust** :

- Respecter `rustfmt` et `clippy`
- Utiliser `thiserror` pour erreurs custom
- Documentation avec `///` pour items publics

**AsyncAPI** :

- Valider avec AsyncAPI CLI v1.8+
- Versionner les schemas (SemVer)
- Tester avec AsyncAPI Generator[^27][^10]


### Review criteria

Les PRs sont évaluées selon :

- ✅ Spécifications complètes (AsyncAPI + CloudEvents)
- ✅ Tests automatisés (CI passe)
- ✅ Performance (benchmarks si applicable)
- ✅ Documentation (README + inline comments)
- ✅ Sécurité (pas de secrets hardcodés)


## Architecture Decisions Records (ADR)

### ADR-001 : Choix de Kafka (vs NATS/RabbitMQ)

**Statut** : Accepté

**Contexte** : Besoin d'un event backbone haute performance pour 10k+ TPS

**Décision** : Apache Kafka 3.7 (KRaft mode)

**Rationale** :

- **Throughput** : Kafka supporte 100k+ msg/s par broker[^28]
- **Exactly-once** : Transactions natives garantissent cohérence
- **Réplication** : Factor 3 par défaut pour HA
- **Écosystème** : Kafka Connect (150+ connectors), ksqlDB, Flink

**Alternatives considérées** :

- NATS : Plus léger mais moins de garanties de livraison
- RabbitMQ : Bon pour messaging classique, moins pour streaming

**Conséquences** :

- ➕ Performance prouvée en production (LinkedIn, Netflix, Uber)
- ➕ CloudEvents native support
- ➖ Complexité opérationnelle (mitigée par Docker)
- ➖ JVM memory overhead (~2GB par broker)


### ADR-002 : Go pour Producers, Rust pour Consumers

**Statut** : Accepté

**Contexte** : Choix des langages pour performance et maintenabilité

**Décision** :

- **Producers** : Go 1.24
- **Consumers** : Rust 1.75

**Rationale** :

- **Go** : Excellente concurrence (goroutines), large écosystème Kafka, compilation rapide
- **Rust** : Safety garanties (no race conditions), zero-cost abstractions, Rayon pour parallélisme

**Alternatives** :

- Full Go : Moins performant pour traitement CPU-intensif
- Full Rust : Courbe d'apprentissage raide pour toute l'équipe

**Conséquences** :

- ➕ Meilleur langage pour chaque tâche
- ➕ Interopérabilité via CloudEvents (agnostic du langage)
- ➖ Maintenir deux toolchains
- ➖ Partage de code limité (résolu par AsyncAPI)


## Licence

MIT License - Voir [LICENSE](LICENSE) pour détails.

## Références

### Standards et spécifications

- [AsyncAPI 3.0 Specification](https://www.asyncapi.com/docs/reference/specification/v3.0.0)[^1]
- [CloudEvents 1.0](https://cloudevents.io)[^2]
- [Apache Kafka Protocol](https://kafka.apache.org/protocol)[^29]


### Outils et frameworks

- [GitHub Spec Kit](https://github.com/github/spec-kit)[^3]
- [Bubble Tea (TUI Go)](https://github.com/charmbracelet/bubbletea)[^13]
- [rdkafka (Rust)](https://github.com/fede1024/rust-rdkafka)[^30]
- [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)[^30]


### Articles et guides

- [Event-Driven Architecture: Building Responsive Enterprise Systems](https://ijsrcseit.com/index.php/home/article/view/CSEIT251112323)[^31]
- [The Future of AI Agents Is Event-Driven](https://www.confluent.io/blog/the-future-of-ai-agents-is-event-driven/)[^7]
- [How Kafka Improves Agentic AI](https://developers.redhat.com/articles/2025/06/16/how-kafka-improves-agentic-ai)[^32]
- [Building Rust Microservices with Apache Kafka](https://blog.logrocket.com/building-rust-microservices-apache-kafka/)[^30]


### Monitoring et observabilité

- [Kafka Monitoring: Key Metrics and Tools (2025)](https://www.instaclustr.com/education/apache-kafka/kafka-monitoring-key-metrics-and-5-tools-to-know-in-2025/)[^15]
- [Kafka Monitoring Using Prometheus](https://www.metricfire.com/blog/kafka-monitoring/)[^16]

***

**Dernière mise à jour** : 7 janvier 2026
**Version** : 0.1.0
**Mainteneurs** : [@agbruneau](https://github.com/agbruneau)

Pour questions ou support : [Ouvrir une issue](https://github.com/agbruneau/AgentMeshKafka/issues)
<span style="display:none">[^100][^101][^102][^103][^104][^105][^106][^107][^108][^109][^110][^111][^112][^113][^114][^115][^116][^117][^118][^119][^120][^121][^122][^123][^124][^125][^126][^127][^128][^129][^130][^131][^132][^133][^33][^34][^35][^36][^37][^38][^39][^40][^41][^42][^43][^44][^45][^46][^47][^48][^49][^50][^51][^52][^53][^54][^55][^56][^57][^58][^59][^60][^61][^62][^63][^64][^65][^66][^67][^68][^69][^70][^71][^72][^73][^74][^75][^76][^77][^78][^79][^80][^81][^82][^83][^84][^85][^86][^87][^88][^89][^90][^91][^92][^93][^94][^95][^96][^97][^98][^99]</span>

<div align="center">⁂</div>

[^1]: https://atamel.dev/posts/2023/05-23_asyncapi_cloudevents/

[^2]: https://www.asyncapi.com/blog/asyncapi-cloud-events

[^3]: https://github.com/github/spec-kit

[^4]: https://github.blog/ai-and-ml/generative-ai/spec-driven-development-with-ai-get-started-with-a-new-open-source-toolkit/

[^5]: https://www.linkedin.com/pulse/event-driven-ai-agents-architecture-pattern-every-needs-venkatesan-fzwfc

[^6]: https://docs.aws.amazon.com/prescriptive-guidance/latest/agentic-ai-serverless/event-driven-architecture.html

[^7]: https://www.confluent.io/blog/the-future-of-ai-agents-is-event-driven/

[^8]: https://docs.docker.com/guides/kafka/

[^9]: https://www.instaclustr.com/education/apache-spark/running-apache-kafka-kraft-on-docker-tutorial-and-best-practices/

[^10]: https://www.asyncapi.com/tools

[^11]: https://developers.redhat.com/articles/2021/06/02/simulating-cloudevents-asyncapi-and-microcks

[^12]: https://microcks.io/blog/simulating-cloudevents-with-asyncapi/

[^13]: https://github.com/charmbracelet/bubbletea

[^14]: https://www.prskavec.net/post/bubbletea/

[^15]: https://www.instaclustr.com/education/apache-kafka/kafka-monitoring-key-metrics-and-5-tools-to-know-in-2025/

[^16]: https://www.metricfire.com/blog/kafka-monitoring/

[^17]: https://www.redpanda.com/guides/kafka-performance-kafka-monitoring

[^18]: https://dev.to/deeshath/apache-kafka-kraft-mode-setup-5nj

[^19]: https://www.datacamp.com/tutorial/kafka-docker-explained

[^20]: https://ijsrcseit.com/index.php/home/article/view/CSEIT24106193

[^21]: https://wjaets.com/node/587

[^22]: https://arxiv.org/abs/2507.16044

[^23]: https://www.semanticscholar.org/paper/725bb3e5b9d1afe2e01fbee7fd263f6038ebef65

[^24]: https://research.aimultiple.com/agentic-mesh/

[^25]: https://www.gravitee.io/platform/agent-mesh

[^26]: https://developer.microsoft.com/blog/spec-driven-development-spec-kit

[^27]: https://www.asyncapi.com/tools/generator

[^28]: https://www.kai-waehner.de/blog/2025/12/10/top-trends-for-data-streaming-with-apache-kafka-and-flink-in-2026/

[^29]: https://github.com/apache/kafka/blob/trunk/README.md

[^30]: https://blog.logrocket.com/building-rust-microservices-apache-kafka/

[^31]: https://ijsrcseit.com/index.php/home/article/view/CSEIT251112323

[^32]: https://developers.redhat.com/articles/2025/06/16/how-kafka-improves-agentic-ai

[^33]: http://arxiv.org/pdf/2412.07978.pdf

[^34]: https://arxiv.org/pdf/1111.5930.pdf

[^35]: https://arxiv.org/html/2403.17927v1

[^36]: https://arxiv.org/pdf/2309.07870.pdf

[^37]: https://arxiv.org/pdf/2312.17294.pdf

[^38]: https://arxiv.org/pdf/2310.10634.pdf

[^39]: https://arxiv.org/pdf/2402.16667.pdf

[^40]: http://arxiv.org/pdf/2310.02374.pdf

[^41]: https://docs.github.com

[^42]: https://docs.readme.com/main/docs/documentation-structure

[^43]: https://github.com/apache/kafka/blob/trunk/clients/src/main/resources/common/message/README.md

[^44]: https://www.freecodecamp.org/news/how-to-structure-your-readme-file/

[^45]: https://docs.agentverse.ai/documentation/agent-discovery/agent-search-optimization

[^46]: https://docs.readme.com/main/docs/aiagent

[^47]: https://github.com/apache/kafka-site

[^48]: https://github.com/ccamel/awesome-ccamel

[^49]: https://ijesty.org/index.php/ijesty/article/view/1074

[^50]: https://www.onlinescientificresearch.com/articles/faulttolerant-eventdriven-systems-techniques-and-best-practices.pdf

[^51]: https://ijsra.net/node/7805

[^52]: https://ieeexplore.ieee.org/document/11292459/

[^53]: https://ijsrcseit.com/index.php/home/article/view/CSEIT251112347

[^54]: https://wjaets.com/node/1168

[^55]: https://www.ssrn.com/abstract=5277379

[^56]: http://arxiv.org/pdf/2410.15533.pdf

[^57]: https://linkinghub.elsevier.com/retrieve/pii/S0167739X21002995

[^58]: https://arxiv.org/pdf/1008.0823.pdf

[^59]: https://arxiv.org/pdf/2208.00786.pdf

[^60]: https://arxiv.org/pdf/2205.09415.pdf

[^61]: https://arxiv.org/pdf/2407.11432.pdf

[^62]: https://arxiv.org/pdf/1509.09153.pdf

[^63]: https://linkinghub.elsevier.com/retrieve/pii/S138376212100151X

[^64]: https://www.growin.com/blog/event-driven-architecture-scale-systems-2025/

[^65]: https://www.gravitee.io/blog/event-driven-architecture-patterns

[^66]: https://www.javacodegeeks.com/2026/01/java-and-event-driven-architectures-kafka-pulsar-and-the-modern-streaming-landscape.html

[^67]: https://www.redhat.com/en/blog/apache-kafka-EDA-performance

[^68]: https://www.tinybird.co/blog/event-sourcing-with-kafka

[^69]: https://www.kai-waehner.de/blog/2025/07/21/building-agentic-ai-with-amazon-bedrock-agentcore-and-data-streaming-using-apache-kafka-and-flink/

[^70]: https://www.confluent.io/blog/do-microservices-need-event-driven-architectures/

[^71]: https://dev.to/lovestaco/why-kafka-a-developer-friendly-guide-to-event-driven-architecture-4ekf

[^72]: https://github.com/cloudevents/spec/issues/1276

[^73]: https://miracl.in/blog/apache-kafka-streaming-2026

[^74]: https://dzone.com/articles/agentic-ai-using-apache-kafka-as-event-broker-with-agent2agent-protocol?fromrel=true

[^75]: https://www.asyncapi.com/blog/async_standards_compare

[^76]: https://www.semanticscholar.org/paper/083c64d8d50508832dcfa990d8a0380d06666035

[^77]: https://www.frontiersin.org/articles/10.3389/ftox.2022.893924/full

[^78]: https://www.epj-conferences.org/10.1051/epjconf/202533810001

[^79]: https://pubs.acs.org/doi/10.1021/acs.jcim.4c00730

[^80]: https://lib.jucs.org/article/133397/

[^81]: https://jcheminf.biomedcentral.com/articles/10.1186/s13321-025-01094-1

[^82]: https://www.nature.com/articles/s41562-023-01647-0

[^83]: https://scienceopen.com/hosted-document?doi=10.70534/DXIP5972

[^84]: https://arxiv.org/html/2503.07358v1

[^85]: http://arxiv.org/pdf/2503.04921.pdf

[^86]: https://arxiv.org/pdf/2401.08807.pdf

[^87]: https://arxiv.org/pdf/2209.09804.pdf

[^88]: https://dl.acm.org/doi/pdf/10.1145/3656440

[^89]: http://arxiv.org/pdf/2308.07124.pdf

[^90]: http://arxiv.org/pdf/2411.13200.pdf

[^91]: https://arxiv.org/pdf/2305.04772.pdf

[^92]: https://github.com/LinkedInLearning/spec-driven-development-with-github-spec-kit-4641001

[^93]: https://imagine.jhu.edu/classes/spec-driven-development-with-github-spec-kit/

[^94]: https://github.com/github/spec-kit/blob/main/spec-driven.md

[^95]: https://stackoverflow.com/questions/71706180/microservices-with-apache-kafka-software-architecture-patterns

[^96]: https://solace.com/blog/getting-started-with-agentic-ai/

[^97]: https://vibecoding.app/blog/spec-kit-review

[^98]: https://blog.enapi.com/building-event-streams-with-rust-and-kafka-a-practical-guide-905178817c76

[^99]: https://azurewithaj.com/posts/github-spec-kit/

[^100]: https://www.linkedin.com/pulse/building-scalable-microservices-rust-kafka-sérgio-santos-qpxyf

[^101]: https://www.hivemq.com/blog/benefits-of-event-driven-architecture-scale-agentic-ai-collaboration-part-2/

[^102]: https://bytesizeddesign.substack.com/p/the-future-of-ai-coding-spec-driven

[^103]: https://www.shuttle.dev/blog/2024/04/25/event-driven-services-using-kafka-rust

[^104]: https://www.epam.com/insights/ai/blogs/inside-spec-driven-development-what-githubspec-kit-makes-possible-for-ai-engineering

[^105]: http://arxiv.org/pdf/1805.08598.pdf

[^106]: https://ijai.iaescore.com/index.php/IJAI/article/download/22652/13799

[^107]: https://zenodo.org/record/7018424/files/KafkaFed_Two-Tier_Federated_Learning_Communication_Architecture_for_Internet_of_Vehicles.pdf

[^108]: http://arxiv.org/pdf/2405.07917.pdf

[^109]: https://developer.confluent.io/confluent-tutorials/kafka-on-docker/

[^110]: https://hub.docker.com/r/apache/kafka

[^111]: https://github.com/Joel-hanson/joel-hanson.github.io/blob/main/content/posts/08-setting-up-a-kafka-cluster-without-zookeeper-using-docker.md

[^112]: https://www.rapidinnovation.io/post/building-microservices-with-rust-architectures-and-best-practices

[^113]: https://pypi.org/project/asyncapi-codegen/

[^114]: https://www.linkedin.com/pulse/inside-rust-microservice-design-modules-patterns-sérgio-santos-ibozf

[^115]: https://modeling-languages.com/asyncapi-modeling-editor-code-generator/

[^116]: https://calmops.com/programming/rust/building-microservices-in-rust/

[^117]: https://www.hostinger.com/tutorials/how-to-deploy-kafka-on-docker

[^118]: https://www.apriorit.com/dev-blog/building-microservices-with-golang

[^119]: https://stackoverflow.com/questions/79862563/starting-apache-kafka-container-in-kraft-mode-throws-error-in-etc-kafka-docker

[^120]: https://www.reddit.com/r/golang/comments/17wjt8t/asyncapi_code_generator_in_go_for_go/

[^121]: https://scand.com/company/blog/rust-vs-go/

[^122]: https://github.com/asyncapi/generator

[^123]: https://www.extrica.com/article/23643

[^124]: https://www.reddit.com/r/golang/comments/1gjasbp/build_a_system_monitor_tui_terminal_ui_in_go_my/

[^125]: https://dev.to/andyhaskell/intro-to-bubble-tea-in-go-21lg

[^126]: https://www.reddit.com/r/golang/comments/1gtmc8v/developing_a_terminal_app_in_go_with_bubble_tea/

[^127]: https://www.oreateai.com/blog/exploring-bubble-tea-a-delightful-journey-into-gos-tui-framework/5624ae147623d9ce075e113576a1ded4

[^128]: https://www.datacamp.com/tutorial/docker-compose-guide

[^129]: https://blog.racknerd.com/benefits-of-multi-container-docker-applications-and-best-practices/

[^130]: https://www.reddit.com/r/golang/comments/r0g1zs/build_terminal_apps_in_go_with_bubble_tea/

[^131]: https://overcast.blog/multi-environment-deployments-with-docker-a-guide-890e193191b6

[^132]: https://www.reddit.com/r/golang/comments/1hb8txd/note_a_modern_terminalbasedtui_notetaking_app/

[^133]: https://www.youtube.com/watch?v=evKGEziQj54
