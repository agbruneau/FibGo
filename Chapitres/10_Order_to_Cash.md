# Chapitre X — Étude de Cas Intégrée : Le Processus « Order-to-Cash » Omnicanal

*Focus : Application concrète des trois domaines d'intégration sur un scénario métier complet.*

---

## Introduction

Les chapitres précédents ont établi un cadre conceptuel riche : les fondements théoriques de l'interopérabilité (chapitre II), les patrons d'intégration des applications (chapitre III), des données (chapitre IV) et des événements (chapitre V), les standards d'interface (chapitre VI), les mécanismes de résilience et d'observabilité (chapitre VII), ainsi que les technologies de collaboration et d'automatisation (chapitre VIII). Le chapitre IX a proposé une synthèse sous forme d'architecture de référence convergente. Il est maintenant temps de confronter ces concepts à la réalité d'un processus métier complet.

Le processus « Order-to-Cash » (O2C) — de la prise de commande à l'encaissement — constitue le fil conducteur idéal pour cette démonstration. Ce processus traverse l'ensemble de l'entreprise : il débute par une interaction client (vente), mobilise la logistique (expédition), implique la finance (facturation, paiement) et alimente l'analytique (tableaux de bord, prévisions). Il requiert simultanément des interactions synchrones (validation de stock en temps réel), des propagations de données (mise à jour des inventaires) et des orchestrations événementielles (coordination des étapes logistiques). En somme, il incarne parfaitement le continuum d'intégration App → Data → Event défendu par cette monographie.

Ce chapitre adopte une structure en quatre phases qui reflètent la progression naturelle du processus métier. La première phase,  *Capture* , illustre l'intégration des applications avec ses patrons de couplage fort : API Gateway, Backend for Frontend, Circuit Breaker. La deuxième phase,  *Persistance* , démontre l'intégration des données via le Change Data Capture et le Transactional Outbox. La troisième phase,  *Orchestration* , met en œuvre l'intégration des événements avec le Saga Pattern et la chorégraphie d'événements. La quatrième phase,  *Reporting* , combine les trois domaines pour alimenter des vues matérialisées et assurer la traçabilité de bout en bout.

À travers ce parcours, nous verrons comment les patrons théoriques prennent vie dans un contexte concret, comment les choix architecturaux répondent à des contraintes métier spécifiques, et comment la convergence des trois domaines d'intégration produit un système cohérent, résilient et observable.

---

## 10.1 Contexte et Architecture Cible

### 10.1.1 Le Scénario Métier

Notre étude de cas porte sur  *TechnoCommerce* , un détaillant omnicanal fictif mais représentatif des défis contemporains. L'entreprise opère à travers trois canaux de vente : une application mobile native (iOS et Android), un site web responsive et un réseau de boutiques physiques équipées de points de vente connectés. Le catalogue compte 50 000 références, l'inventaire est réparti sur cinq entrepôts régionaux, et le volume quotidien atteint 25 000 commandes en période normale, avec des pics à 150 000 lors des événements promotionnels.

Le processus Order-to-Cash de TechnoCommerce se décompose en étapes distinctes mais interdépendantes. Un client sélectionne des produits et initie une commande. Le système vérifie la disponibilité en temps réel et calcule les options de livraison. Le client finalise sa commande avec paiement. La commande est transmise à l'entrepôt approprié pour préparation. Le colis est expédié et suivi jusqu'à livraison. La facture est émise et le paiement rapproché. Tout au long du processus, le client peut consulter l'état de sa commande, et les équipes internes disposent de tableaux de bord temps réel.

Ce scénario impose des exigences variées en termes de couplage. La vérification de stock doit être synchrone : le client ne peut commander un article indisponible. Le traitement du paiement requiert une confirmation immédiate pour des raisons de sécurité et d'expérience utilisateur. En revanche, la transmission à l'entrepôt, la génération de la facture et la mise à jour des tableaux de bord peuvent tolérer quelques secondes de latence. Ces exigences différenciées justifient une stratégie hybride mobilisant les trois domaines d'intégration.

### 10.1.2 Vue d'Ensemble de l'Architecture

L'architecture cible de TechnoCommerce s'articule autour des trois couches identifiées au chapitre IX : le System of Engagement (interactions utilisateur), l'Integration Backbone (bus événementiel) et le System of Record (sources de vérité).

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SYSTEM OF ENGAGEMENT                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                      │
│  │  Mobile BFF  │  │   Web BFF    │  │   POS BFF    │                      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                      │
│         └──────────────────┼──────────────────┘                             │
│                            ▼                                                │
│                   ┌────────────────────┐                                    │
│                   │    API Gateway     │                                    │
│                   │  (Auth, Routing,   │                                    │
│                   │   Rate Limiting)   │                                    │
│                   └─────────┬──────────┘                                    │
└─────────────────────────────┼───────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼───────────────────────────────────────────────┐
│                    MICROSERVICES LAYER                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Order     │  │  Inventory  │  │   Payment   │  │  Shipping   │        │
│  │  Service    │  │   Service   │  │   Service   │  │   Service   │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
└─────────┼────────────────┼────────────────┼────────────────┼────────────────┘
          │                │                │                │
┌─────────┼────────────────┼────────────────┼────────────────┼────────────────┐
│         │      INTEGRATION BACKBONE (Kafka)                │                │
│         ▼                ▼                ▼                ▼                │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Topics: orders.created | payments.processed | inventory.reserved   │   │
│  │          shipments.dispatched | orders.completed | orders.cancelled │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
          │                │                │                │
┌─────────┼────────────────┼────────────────┼────────────────┼────────────────┐
│                       SYSTEM OF RECORD                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Orders DB  │  │ Inventory DB│  │ Payments DB │  │ Shipping DB │        │
│  │ (PostgreSQL)│  │ (PostgreSQL)│  │ (PostgreSQL)│  │ (PostgreSQL)│        │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │              Analytics Data Warehouse (ClickHouse)                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

Cette architecture respecte les principes établis aux chapitres précédents. Les BFF adaptent l'interface aux spécificités de chaque canal. L'API Gateway centralise les préoccupations transversales. Chaque microservice possède sa propre base de données (principe de l'autonomie des données). Le bus Kafka assure le découplage événementiel entre services. L'entrepôt de données analytiques agrège l'information pour le reporting.

### 10.1.3 Les Acteurs du Système

Avant de plonger dans les phases du processus, identifions les principaux composants et leurs responsabilités.

| Service              | Responsabilité                                    | Base de données       | Événements produits                                              |
| -------------------- | -------------------------------------------------- | ---------------------- | ------------------------------------------------------------------ |
| Order Service        | Création et gestion du cycle de vie des commandes | PostgreSQL (orders)    | `OrderCreated`,`OrderCancelled`,`OrderCompleted`             |
| Inventory Service    | Gestion des stocks, réservations                  | PostgreSQL (inventory) | `InventoryReserved`,`InventoryReleased`                        |
| Payment Service      | Traitement des paiements, remboursements           | PostgreSQL (payments)  | `PaymentProcessed`,`PaymentFailed`,`RefundIssued`            |
| Shipping Service     | Planification et suivi des expéditions            | PostgreSQL (shipments) | `ShipmentScheduled`,`ShipmentDispatched`,`ShipmentDelivered` |
| Notification Service | Envoi de courriels et notifications push           | — (sans état)        | —                                                                 |
| Analytics Service    | Agrégation et tableaux de bord                    | ClickHouse             | —                                                                 |

Chaque service est déployé en plusieurs instances pour la haute disponibilité, orchestré par Kubernetes. La communication synchrone transite par l'API Gateway ; la communication asynchrone emprunte les topics Kafka. Cette séparation claire entre les deux modes de communication reflète la distinction App/Event du continuum d'intégration.

---

## 10.2 Phase 1 — Capture : L'Intégration des Applications

La première phase du processus Order-to-Cash concerne la capture de la commande : le client interagit avec l'interface, sélectionne des produits, vérifie leur disponibilité et soumet sa commande. Cette phase illustre l'intégration des applications — « Le Verbe » — avec ses exigences de réponse synchrone et son couplage temporel fort.

### 10.2.1 Backend for Frontend : Adaptation au Canal

Le client mobile de TechnoCommerce présente des contraintes spécifiques : bande passante limitée, connexions intermittentes, écran réduit. Le BFF mobile optimise l'expérience en agrégeant les données nécessaires en un minimum d'appels et en formatant les réponses pour une consommation efficace.

Lorsqu'un client consulte une fiche produit, le BFF mobile orchestre plusieurs appels internes : récupération des détails du produit (catalogue), vérification de disponibilité (inventaire), récupération du prix (tarification), chargement des avis (recommandations). Plutôt que d'exposer ces quatre endpoints au client mobile — ce qui impliquerait quatre allers-retours réseau —, le BFF les consolide en une réponse unique.

> **Point d'intégration**
> *Domaine* : Applications (Le Verbe)
> *Patron utilisé* : Backend for Frontend (BFF) + Aggregator
> *Justification* : Réduire le nombre d'appels réseau pour les clients mobiles, adapter le format de réponse aux contraintes d'affichage, permettre l'évolution indépendante du canal mobile

```
┌─────────────────┐
│  Client Mobile  │
└────────┬────────┘
         │ GET /products/12345
         ▼
┌─────────────────┐      ┌─────────────────┐
│   Mobile BFF    │─────▶│ Catalog Service │ (détails produit)
│                 │      └─────────────────┘
│  - Agrège       │      ┌─────────────────┐
│  - Formate      │─────▶│Inventory Service│ (disponibilité)
│  - Cache        │      └─────────────────┘
│                 │      ┌─────────────────┐
│                 │─────▶│ Pricing Service │ (prix actuel)
└─────────────────┘      └─────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Réponse agrégée :                       │
│ {                                       │
│   "id": "12345",                        │
│   "name": "Widget Pro",                 │
│   "price": 149.99,                      │
│   "available": true,                    │
│   "stock": 42,                          │
│   "deliveryOptions": [...]              │
│ }                                       │
└─────────────────────────────────────────┘
```

Le BFF web, destiné aux navigateurs, adopte une approche différente. La connexion plus stable et la puissance de calcul supérieure permettent des réponses plus riches. Le BFF web peut inclure des données de personnalisation supplémentaires, des recommandations étendues et des options de filtrage avancées que le BFF mobile omet volontairement.

Cette différenciation par canal illustre la valeur du patron BFF : chaque interface évolue selon les besoins de son audience sans impacter les autres. L'équipe mobile peut ajouter un nouveau champ à son BFF sans coordonner avec l'équipe web, et inversement.

### 10.2.2 API Gateway : Point d'Entrée Unifié

Tous les BFF communiquent avec les microservices via l'API Gateway, qui assure plusieurs fonctions critiques.

L'authentification vérifie l'identité de l'appelant. Pour les clients finaux, un jeton JWT émis lors de la connexion est validé à chaque requête. Pour les partenaires B2B, des clés API avec scopes spécifiques contrôlent les permissions. La gateway extrait l'identité et l'injecte dans les en-têtes internes pour que les services en aval connaissent le contexte de l'appelant.

La limitation de débit protège les services contre les surcharges. TechnoCommerce applique des quotas différenciés : les clients authentifiés bénéficient de limites généreuses (1000 requêtes/minute), les appels anonymes sont plus restreints (100 requêtes/minute), les partenaires B2B ont des quotas négociés contractuellement. Lors des pics promotionnels, ces limites peuvent être ajustées dynamiquement.

Le routage dirige les requêtes vers les services appropriés. La gateway maintient une table de routage qui associe les chemins d'URL aux services cibles. Elle s'intègre au Service Registry (Consul dans notre cas) pour découvrir dynamiquement les instances disponibles et répartir la charge.

> **Point d'intégration**
> *Domaine* : Applications (Le Verbe)
> *Patron utilisé* : API Gateway
> *Justification* : Centraliser l'authentification, appliquer des politiques de sécurité uniformes, abstraire la topologie interne des microservices

```
┌─────────────────────────────────────────────────────────────────────┐
│                          API GATEWAY (Kong)                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐  │
│  │    Auth     │  │    Rate     │  │   Routing   │  │  Logging  │  │
│  │   Plugin    │  │   Limiter   │  │             │  │  Plugin   │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘  │
│         │                │                │                │        │
│         └────────────────┴────────────────┴────────────────┘        │
│                                   │                                  │
│                          Route Table:                                │
│                          /orders/* → Order Service                   │
│                          /inventory/* → Inventory Service            │
│                          /payments/* → Payment Service               │
└──────────────────────────────────┬──────────────────────────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    ▼                             ▼
          ┌─────────────────┐           ┌─────────────────┐
          │  Order Service  │           │Inventory Service│
          │  (3 instances)  │           │  (3 instances)  │
          └─────────────────┘           └─────────────────┘
```

### 10.2.3 Validation Synchrone avec Circuit Breaker

Lors de la soumission d'une commande, le système doit valider plusieurs conditions avant d'accepter : disponibilité des produits, validité de l'adresse de livraison, éligibilité du client. Ces validations sont synchrones car le client attend une confirmation immédiate.

L'Order Service orchestre ces validations en appelant les services concernés. Cependant, si l'Inventory Service est temporairement indisponible, faut-il faire échouer l'ensemble de la commande ? Le patron Circuit Breaker, présenté au chapitre VII, répond à cette question.

TechnoCommerce a implémenté une stratégie nuancée. La vérification d'inventaire est critique : sans stock confirmé, la commande ne peut être acceptée. Le Circuit Breaker protège contre les pannes prolongées, mais un service d'inventaire défaillant bloque effectivement les commandes. En revanche, la vérification de l'adresse (via un service externe de géolocalisation) est non critique : si elle échoue, la commande est acceptée avec un drapeau « adresse non vérifiée » pour traitement manuel ultérieur.

> **Point d'intégration**
> *Domaine* : Applications (Le Verbe)
> *Patron utilisé* : Circuit Breaker + Fallback
> *Justification* : Protéger le système contre les défaillances en cascade, permettre une dégradation gracieuse pour les validations non critiques

```
Order Service                    Inventory Service
     │                                  │
     │──── checkAvailability() ────────▶│
     │                                  │
     │     [Circuit Breaker: CLOSED]    │
     │                                  │
     │◀─── { available: true } ─────────│
     │                                  │

Scénario de défaillance :

Order Service                    Inventory Service
     │                                  │
     │──── checkAvailability() ────────▶│ ✗ Timeout
     │──── checkAvailability() ────────▶│ ✗ Timeout
     │──── checkAvailability() ────────▶│ ✗ Timeout
     │                                  │
     │     [Circuit Breaker: OPEN]      │
     │                                  │
     │──── checkAvailability() ───╳     │ (appel court-circuité)
     │◀─── ServiceUnavailableException ─│
     │                                  │
     │     → Réponse client :           │
     │       "Vérification temporairement│
     │        indisponible, réessayez"  │
```

La configuration du Circuit Breaker pour l'Inventory Service reflète sa criticité :

| Paramètre         | Valeur                 | Justification                                                  |
| ------------------ | ---------------------- | -------------------------------------------------------------- |
| Seuil d'échec     | 5 échecs consécutifs | Tolérer les erreurs transitoires                              |
| Timeout            | 2 secondes             | L'utilisateur attend ; au-delà, l'expérience se dégrade     |
| Durée d'ouverture | 30 secondes            | Laisser le temps au service de récupérer                     |
| Demi-ouverture     | 1 requête test        | Vérifier la récupération avant de restaurer le flux complet |

### 10.2.4 Flux de Création de Commande

Assemblons ces éléments dans le flux complet de création de commande. Le client soumet sa commande via le BFF, qui transmet à l'API Gateway, qui route vers l'Order Service. Celui-ci orchestre les validations, crée la commande si tout est valide, et retourne une confirmation.

```
Client          BFF           Gateway       Order Svc      Inventory      Payment
  │              │               │              │              │              │
  │─ POST /order─▶              │              │              │              │
  │              │─ POST /orders ▶              │              │              │
  │              │               │─ route ─────▶│              │              │
  │              │               │              │              │              │
  │              │               │              │─ check stock ▶              │
  │              │               │              │◀─ available ──│              │
  │              │               │              │              │              │
  │              │               │              │─ reserve ────▶│              │
  │              │               │              │◀─ reserved ───│              │
  │              │               │              │              │              │
  │              │               │              │─ auth payment ──────────────▶
  │              │               │              │◀─ authorized ───────────────│
  │              │               │              │              │              │
  │              │               │              │ [CREATE ORDER]│              │
  │              │               │              │ [PUBLISH EVENT]              │
  │              │               │              │              │              │
  │              │               │◀─ 201 Created│              │              │
  │              │◀─ order confirmation ────────│              │              │
  │◀─ confirmation ─────────────│              │              │              │
```

Ce flux illustre plusieurs patrons en action. Le BFF adapte la requête client au format interne. La Gateway authentifie et route. L'Order Service utilise le patron Aggregator pour collecter les validations. Le Circuit Breaker protège chaque appel sortant. Une fois la commande créée, un événement est publié pour déclencher la suite du processus — transition vers le domaine événementiel que nous explorerons en phase 3.

> **Note technique**
> La réservation d'inventaire et l'autorisation de paiement sont effectuées de manière synchrone lors de la création de la commande. Cette approche garantit que seules les commandes réellement exécutables sont acceptées. Le débit du paiement (capture) et la décrémentation définitive du stock interviennent plus tard, de manière asynchrone, lorsque la commande est confirmée pour expédition.

---

## 10.3 Phase 2 — Persistance : L'Intégration des Données

Une fois la commande créée, elle doit être persistée de manière durable et les systèmes en aval doivent être informés. Cette phase illustre l'intégration des données — « Le Nom » — avec ses mécanismes de capture de changements et de garantie de cohérence.

### 10.3.1 Écriture Transactionnelle

L'Order Service persiste la commande dans sa base de données PostgreSQL. Cette opération doit être atomique : soit la commande est entièrement enregistrée, soit elle ne l'est pas du tout. Les propriétés ACID de PostgreSQL garantissent cette atomicité pour l'écriture locale.

Le modèle de données de la commande comprend plusieurs entités liées :

```sql
-- Table principale des commandes
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'CAD',
    shipping_address JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Lignes de commande
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    reserved_at TIMESTAMP
);

-- Table outbox pour publication garantie
CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    published_at TIMESTAMP
);
```

### 10.3.2 Transactional Outbox : Atomicité Garantie

Le défi critique de cette phase réside dans la publication de l'événement `OrderCreated`. Comment garantir que si la commande est enregistrée en base, l'événement est publié sur Kafka, et inversement ? Sans cette garantie, le système risque des incohérences : commandes orphelines sans événement, ou événements sans commande correspondante.

Le patron Transactional Outbox, présenté au chapitre V, résout ce problème. Au lieu de publier directement sur Kafka (opération qui pourrait échouer après le commit de la transaction), l'Order Service insère l'événement dans une table `outbox` de la même base de données, au sein de la même transaction que la commande.

> **Point d'intégration**
> *Domaine* : Données (Le Nom)
> *Patron utilisé* : Transactional Outbox
> *Justification* : Garantir l'atomicité entre la persistance de la commande et la publication de l'événement, sans recourir aux transactions distribuées (2PC)

```java
@Transactional
public Order createOrder(CreateOrderRequest request) {
    // 1. Valider et créer la commande
    Order order = new Order();
    order.setCustomerId(request.getCustomerId());
    order.setItems(request.getItems());
    order.setStatus(OrderStatus.PENDING);
    order.setTotalAmount(calculateTotal(request.getItems()));
  
    // 2. Persister la commande
    orderRepository.save(order);
  
    // 3. Insérer l'événement dans l'outbox (même transaction)
    OutboxEvent event = OutboxEvent.builder()
        .aggregateType("Order")
        .aggregateId(order.getId())
        .eventType("OrderCreated")
        .payload(buildOrderCreatedPayload(order))
        .build();
  
    outboxRepository.save(event);
  
    // 4. La transaction est commitée atomiquement
    return order;
}
```

Un processus séparé — le *relay* — surveille la table outbox et publie les événements vers Kafka. Chez TechnoCommerce, ce relay est implémenté via Debezium qui capture les insertions dans la table outbox en utilisant le log de réplication de PostgreSQL (WAL).

### 10.3.3 Change Data Capture avec Debezium

Debezium est configuré pour surveiller la table `outbox`. Chaque nouvelle entrée génère un événement de changement (change event) que Debezium transforme et publie sur le topic Kafka approprié.

> **Point d'intégration**
> *Domaine* : Données (Le Nom)
> *Patron utilisé* : Change Data Capture (CDC)
> *Justification* : Capturer les modifications de manière non intrusive via les logs de réplication, sans impact sur les performances de la base source

La configuration Debezium pour TechnoCommerce :

```json
{
  "name": "order-outbox-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "orders-db",
    "database.port": "5432",
    "database.user": "debezium",
    "database.password": "${secrets.db.password}",
    "database.dbname": "orders",
    "database.server.name": "orders",
  
    "table.include.list": "public.outbox",
  
    "transforms": "outbox",
    "transforms.outbox.type": "io.debezium.transforms.outbox.EventRouter",
    "transforms.outbox.table.field.event.type": "event_type",
    "transforms.outbox.table.field.event.payload": "payload",
    "transforms.outbox.route.topic.replacement": "events.${routedByValue}",
  
    "slot.name": "orders_outbox_slot",
    "publication.name": "orders_outbox_pub"
  }
}
```

Cette configuration utilise la transformation `EventRouter` de Debezium qui extrait le type d'événement et le payload de la table outbox pour publier sur des topics distincts. Un événement `OrderCreated` est routé vers `events.OrderCreated`, un événement `OrderCancelled` vers `events.OrderCancelled`, etc.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FLUX CDC                                       │
│                                                                             │
│  ┌───────────────┐     ┌───────────────┐     ┌───────────────────────────┐ │
│  │ Order Service │     │  PostgreSQL   │     │        Debezium          │ │
│  │               │     │               │     │                           │ │
│  │ INSERT order  │────▶│ orders table  │     │                           │ │
│  │ INSERT outbox │────▶│ outbox table  │     │                           │ │
│  │               │     │       │       │     │                           │ │
│  │   COMMIT      │     │       │       │     │                           │ │
│  └───────────────┘     │       ▼       │     │                           │ │
│                        │   WAL Log     │────▶│  Capture changes          │ │
│                        │               │     │  Transform (EventRouter)  │ │
│                        └───────────────┘     │  Publish to Kafka         │ │
│                                              └─────────────┬─────────────┘ │
│                                                            │               │
│                                                            ▼               │
│                                              ┌───────────────────────────┐ │
│                                              │   Kafka Topic             │ │
│                                              │   events.OrderCreated     │ │
│                                              │                           │ │
│                                              │   { "orderId": "123",     │ │
│                                              │     "customerId": "456",  │ │
│                                              │     "items": [...],       │ │
│                                              │     "total": 299.99 }     │ │
│                                              └───────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.3.4 Gestion des Erreurs de Publication

Que se passe-t-il si Debezium échoue à publier sur Kafka ? Le connecteur Debezium est conçu pour la reprise automatique. Il maintient un pointeur (offset) sur sa position dans le WAL. En cas d'échec de publication, il réessaie. En cas de redémarrage, il reprend depuis sa dernière position confirmée.

Cependant, cette robustesse a un coût potentiel : la duplication de messages. Si Debezium publie un événement puis échoue avant de confirmer l'offset, au redémarrage il republiera le même événement. Les consommateurs doivent donc être idempotents — point que nous aborderons en phase 3.

> **Note technique**
> TechnoCommerce a configuré Debezium avec `exactly.once.support` activé (disponible depuis Debezium 2.0 avec Kafka transactions). Cette configuration garantit la publication exactly-once en utilisant les transactions Kafka, mais elle impose une latence légèrement supérieure et une complexité de configuration accrue. Pour les charges de travail où l'idempotence côté consommateur est déjà assurée, le mode at-least-once par défaut suffit souvent.

### 10.3.5 Propagation vers les Autres Services

L'événement `OrderCreated` publié sur Kafka déclenche des actions dans plusieurs services. Cette propagation illustre la transition du domaine des données vers celui des événements : la capture du changement (CDC) alimente un flux événementiel que les consommateurs traitent de manière autonome.

| Consommateur         | Action déclenchée                     | Domaine |
| -------------------- | --------------------------------------- | ------- |
| Inventory Service    | Confirmer la réservation temporaire    | Event   |
| Payment Service      | Initier la capture du paiement          | Event   |
| Notification Service | Envoyer le courriel de confirmation     | Event   |
| Analytics Service    | Mettre à jour les métriques de ventes | Data    |

La distinction entre les domaines reste pertinente. L'Inventory Service et le Payment Service réagissent à l'événement pour orchestrer la suite du processus (domaine Event). L'Analytics Service consomme l'événement pour alimenter des vues matérialisées (domaine Data). Le Notification Service effectue une action « fire-and-forget » sans impact sur l'état du processus.

---

## 10.4 Phase 3 — Orchestration : L'Intégration des Événements

La commande est créée et persistée. Il faut maintenant orchestrer la suite du processus : confirmer le paiement, préparer l'expédition, mettre à jour l'inventaire. Cette phase illustre l'intégration des événements — « Le Signal » — avec le Saga Pattern et la chorégraphie distribuée.

### 10.4.1 Architecture de la Saga Order-to-Cash

Le processus Order-to-Cash constitue une transaction distribuée impliquant plusieurs services. Aucun de ces services ne peut verrouiller les ressources des autres ; une approche transactionnelle traditionnelle (2PC) serait à la fois techniquement complexe et néfaste pour la disponibilité. Le Saga Pattern offre une alternative basée sur des transactions locales et des compensations.

TechnoCommerce a opté pour une saga en chorégraphie plutôt qu'en orchestration. Dans ce modèle, chaque service réagit aux événements et publie ses propres événements sans coordinateur central. Cette approche maximise l'autonomie des services mais requiert une conception soigneuse pour éviter les incohérences.

> **Point d'intégration**
> *Domaine* : Événements (Le Signal)
> *Patron utilisé* : Saga Pattern (Chorégraphie)
> *Justification* : Coordonner une transaction distribuée sans verrouillage global, maintenir l'autonomie des services, permettre des compensations en cas d'échec

```
> **Flux d'événements**
> `OrderCreated` → `PaymentCaptured` → `InventoryConfirmed` → `ShipmentScheduled` → `OrderCompleted`
> Chaque service consomme l'événement précédent, effectue son traitement, et publie le suivant
```

Le diagramme suivant illustre le flux nominal :

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                       SAGA ORDER-TO-CASH (NOMINAL)                           │
│                                                                              │
│  Order Service         Payment Svc        Inventory Svc       Shipping Svc  │
│        │                    │                   │                   │       │
│   [OrderCreated]            │                   │                   │       │
│        │                    │                   │                   │       │
│        └───────────────────▶│                   │                   │       │
│                        [Capture payment]        │                   │       │
│                             │                   │                   │       │
│                        [PaymentCaptured]        │                   │       │
│                             │                   │                   │       │
│                             └──────────────────▶│                   │       │
│                                            [Confirm reservation]    │       │
│                                                 │                   │       │
│                                            [InventoryConfirmed]     │       │
│                                                 │                   │       │
│                                                 └──────────────────▶│       │
│                                                                [Schedule]   │
│                                                                     │       │
│                                                           [ShipmentScheduled]
│                                                                     │       │
│        ◀────────────────────────────────────────────────────────────┘       │
│   [Update status: READY_TO_SHIP]                                            │
│        │                                                                    │
│   [OrderCompleted]                                                          │
│        │                                                                    │
└────────┼────────────────────────────────────────────────────────────────────┘
         ▼
   Notification Service → Courriel "Commande en préparation"
```

### 10.4.2 Gestion des Échecs et Compensations

Le flux nominal est satisfaisant, mais que se passe-t-il lorsqu'une étape échoue ? La saga doit prévoir des compensations pour annuler les effets des étapes déjà accomplies.

Prenons le scénario où le paiement échoue après que la réservation d'inventaire a été confirmée (cas peu fréquent mais possible si l'autorisation initiale expire ou si la carte est ensuite déclarée frauduleuse) :

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                    SAGA ORDER-TO-CASH (ÉCHEC PAIEMENT)                       │
│                                                                              │
│  Order Service         Payment Svc        Inventory Svc       Shipping Svc  │
│        │                    │                   │                   │       │
│   [OrderCreated]            │                   │                   │       │
│        │                    │                   │                   │       │
│        └───────────────────▶│                   │                   │       │
│                        [Attempt capture]        │                   │       │
│                             │                   │                   │       │
│                        ✗ ÉCHEC (carte refusée)  │                   │       │
│                             │                   │                   │       │
│                        [PaymentFailed]          │                   │       │
│                             │                   │                   │       │
│        ◀────────────────────┘                   │                   │       │
│                             │                   │                   │       │
│   [Compensation: Cancel order]                  │                   │       │
│        │                    │                   │                   │       │
│   [OrderCancelled]          │                   │                   │       │
│        │                    │                   │                   │       │
│        └───────────────────────────────────────▶│                   │       │
│                                            [Release reservation]    │       │
│                                                 │                   │       │
│                                            [InventoryReleased]      │       │
│                                                 │                   │       │
└────────┼────────────────────────────────────────────────────────────────────┘
         ▼
   Notification Service → Courriel "Paiement refusé, commande annulée"
```

Les compensations sont définies pour chaque étape :

| Étape                    | Action                 | Compensation                |
| ------------------------- | ---------------------- | --------------------------- |
| Création commande        | Commande créée       | Annuler la commande         |
| Réservation inventaire   | Stock réservé        | Libérer le stock réservé |
| Capture paiement          | Montant capturé       | Émettre un remboursement   |
| Planification expédition | Expédition planifiée | Annuler l'expédition       |

> **Note technique**
> Les compensations ne sont pas toujours des « annulations parfaites ». Une notification envoyée ne peut être « désenvoyée » ; la compensation consiste alors à envoyer une notification d'annulation. Un paiement capturé n'est pas instantanément reversé ; un remboursement est initié, qui apparaîtra sur le relevé du client dans quelques jours. L'architecte doit anticiper ces cas et concevoir des compensations sémantiquement correctes pour chaque contexte.

### 10.4.3 Idempotence des Consommateurs

Dans une architecture événementielle, les messages peuvent être délivrés plusieurs fois (garantie at-least-once). Chaque service consommateur doit être idempotent : traiter le même événement plusieurs fois ne doit pas produire d'effets de bord indésirables.

> **Point d'intégration**
> *Domaine* : Événements (Le Signal)
> *Patron utilisé* : Idempotent Consumer
> *Justification* : Garantir un comportement correct malgré les livraisons multiples inhérentes aux systèmes de messagerie fiables

Le Payment Service implémente l'idempotence via une table de suivi :

```sql
CREATE TABLE payment_operations (
    event_id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    operation_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW()
);
```

```java
public void handleOrderCreated(OrderCreatedEvent event) {
    // Vérifier si déjà traité
    if (paymentOperationRepository.existsByEventId(event.getEventId())) {
        log.info("Event {} already processed, skipping", event.getEventId());
        return;
    }
  
    // Traiter le paiement
    PaymentResult result = paymentGateway.capturePayment(
        event.getPaymentAuthorizationId(),
        event.getTotalAmount()
    );
  
    // Enregistrer l'opération (même transaction que la mise à jour d'état)
    paymentOperationRepository.save(PaymentOperation.builder()
        .eventId(event.getEventId())
        .orderId(event.getOrderId())
        .operationType("CAPTURE")
        .status(result.isSuccess() ? "SUCCESS" : "FAILED")
        .build());
  
    // Publier l'événement résultant
    if (result.isSuccess()) {
        publishEvent(new PaymentCapturedEvent(event.getOrderId(), result.getTransactionId()));
    } else {
        publishEvent(new PaymentFailedEvent(event.getOrderId(), result.getErrorCode()));
    }
}
```

### 10.4.4 Dead Letter Queue et Gestion des Exceptions

Certains événements peuvent échouer de manière persistante : données corrompues, cas d'affaires non anticipés, bogues applicatifs. Plutôt que de bloquer la consommation ou de perdre ces événements, TechnoCommerce utilise des Dead Letter Queues (DLQ).

> **Point d'intégration**
> *Domaine* : Événements (Le Signal)
> *Patron utilisé* : Dead Letter Queue (DLQ)
> *Justification* : Isoler les messages en échec pour analyse sans bloquer le flux principal, préserver les messages pour retraitement ultérieur

La configuration Kafka pour le Payment Service :

```properties
# Consumer configuration
spring.kafka.consumer.group-id=payment-service
spring.kafka.consumer.auto-offset-reset=earliest
spring.kafka.consumer.enable-auto-commit=false

# Retry configuration
spring.kafka.listener.concurrency=3
spring.kafka.listener.ack-mode=MANUAL

# DLQ configuration via Spring Kafka
spring.kafka.listener.common-container-properties.default-error-handler=org.springframework.kafka.listener.DefaultErrorHandler
```

```java
@Bean
public DefaultErrorHandler errorHandler(KafkaTemplate<String, Object> kafkaTemplate) {
    // Après 3 tentatives, envoyer à la DLQ
    DeadLetterPublishingRecoverer recoverer = 
        new DeadLetterPublishingRecoverer(kafkaTemplate,
            (record, exception) -> new TopicPartition(record.topic() + ".dlq", record.partition()));
  
    // Backoff exponentiel : 1s, 2s, 4s
    ExponentialBackOff backOff = new ExponentialBackOff(1000L, 2.0);
    backOff.setMaxElapsedTime(10000L);
  
    return new DefaultErrorHandler(recoverer, backOff);
}
```

L'équipe d'opérations surveille les DLQ via des alertes. Un tableau de bord affiche le nombre de messages en DLQ par service et par type d'erreur. Les messages peuvent être analysés, corrigés si nécessaire, puis réinjectés dans le topic original pour retraitement.

### 10.4.5 Observabilité de la Saga

Tracer le parcours d'une commande à travers la saga nécessite une instrumentation soigneuse. TechnoCommerce utilise OpenTelemetry pour propager le contexte de trace à travers les services et les messages Kafka.

Chaque événement inclut un `traceId` dans ses en-têtes, permettant de corréler tous les traitements relatifs à une même commande. Le collecteur OpenTelemetry agrège les spans et les envoie vers Jaeger pour visualisation.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    TRACE: Order 550e8400-e29b-41d4-a716                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [Order Service]                                                            │
│  ├── createOrder (150ms)                                                    │
│  │   ├── validateInventory (45ms)                                           │
│  │   ├── authorizePayment (80ms)                                            │
│  │   └── persistOrder (25ms)                                                │
│  │                                                                          │
│  [Payment Service]                                                          │
│  ├── handleOrderCreated (200ms)                                             │
│  │   ├── capturePayment (180ms)                                             │
│  │   └── publishPaymentCaptured (20ms)                                      │
│  │                                                                          │
│  [Inventory Service]                                                        │
│  ├── handlePaymentCaptured (50ms)                                           │
│  │   ├── confirmReservation (30ms)                                          │
│  │   └── publishInventoryConfirmed (20ms)                                   │
│  │                                                                          │
│  [Shipping Service]                                                         │
│  ├── handleInventoryConfirmed (100ms)                                       │
│  │   ├── scheduleShipment (70ms)                                            │
│  │   └── publishShipmentScheduled (30ms)                                    │
│  │                                                                          │
│  Total: 500ms                                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

Cette traçabilité permet non seulement le débogage en cas de problème, mais aussi l'analyse de performance. L'équipe peut identifier les goulots d'étranglement, mesurer l'impact des modifications et valider les SLO (Service Level Objectives).

---

## 10.5 Phase 4 — Reporting : Convergence des Domaines

La quatrième et dernière phase combine les trois domaines d'intégration pour alimenter les systèmes de reporting et de prise de décision. Les données transactionnelles, capturées via CDC et propagées via événements, sont agrégées dans des vues matérialisées consultables en temps quasi réel.

### 10.5.1 Architecture du Flux Analytique

TechnoCommerce distingue deux catégories de reporting : les tableaux de bord opérationnels (temps réel) et les rapports analytiques (consolidés). Les premiers exigent une fraîcheur de l'ordre de la minute ; les seconds tolèrent un délai de quelques heures.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         FLUX ANALYTIQUE                                     │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                     KAFKA TOPICS (événements)                         │ │
│  │  orders.created | payments.processed | shipments.dispatched | ...     │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                              │                                              │
│              ┌───────────────┴───────────────┐                             │
│              │                               │                             │
│              ▼                               ▼                             │
│  ┌─────────────────────┐        ┌─────────────────────────────────────┐   │
│  │   Kafka Streams     │        │          Flink Job                  │   │
│  │   (temps réel)      │        │       (agrégations complexes)       │   │
│  │                     │        │                                     │   │
│  │  - Compteurs live   │        │  - Jointures multi-streams          │   │
│  │  - Alertes seuils   │        │  - Fenêtres temporelles             │   │
│  │                     │        │  - Enrichissement                   │   │
│  └──────────┬──────────┘        └───────────────┬─────────────────────┘   │
│             │                                   │                         │
│             ▼                                   ▼                         │
│  ┌─────────────────────┐        ┌─────────────────────────────────────┐   │
│  │       Redis         │        │          ClickHouse                 │   │
│  │  (métriques live)   │        │      (entrepôt analytique)          │   │
│  │                     │        │                                     │   │
│  │  orders_today: 1234 │        │  Fact tables:                       │   │
│  │  revenue_hour: 45K  │        │  - fact_orders                      │   │
│  │                     │        │  - fact_order_items                 │   │
│  └──────────┬──────────┘        │  - fact_shipments                   │   │
│             │                   │                                     │   │
│             │                   │  Materialized views:                │   │
│             │                   │  - mv_daily_sales                   │   │
│             │                   │  - mv_product_performance           │   │
│             │                   │  - mv_customer_ltv                  │   │
│             │                   └───────────────┬─────────────────────┘   │
│             │                                   │                         │
│             └───────────────┬───────────────────┘                         │
│                             │                                             │
│                             ▼                                             │
│             ┌───────────────────────────────────────────────────────────┐ │
│             │                    Grafana / Superset                     │ │
│             │                  (visualisation unifiée)                  │ │
│             └───────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.5.2 Vues Matérialisées en Temps Réel

Pour les tableaux de bord opérationnels, TechnoCommerce utilise Kafka Streams pour maintenir des compteurs en temps réel. Un processeur Kafka Streams consomme les événements `OrderCreated` et met à jour des compteurs dans Redis.

> **Point d'intégration**
> *Domaine* : Données (Le Nom) + Événements (Le Signal)
> *Patron utilisé* : Materialized View (streaming)
> *Justification* : Fournir des métriques temps réel sans charger les bases transactionnelles, exploiter la nature incrémentale du streaming pour une efficacité maximale

```java
public class OrderMetricsProcessor {
  
    @Autowired
    private StreamsBuilder streamsBuilder;
  
    @Autowired
    private RedisTemplate<String, Long> redisTemplate;
  
    @Bean
    public KStream<String, OrderCreatedEvent> processOrderMetrics() {
        KStream<String, OrderCreatedEvent> orders = streamsBuilder
            .stream("events.OrderCreated", Consumed.with(Serdes.String(), orderEventSerde));
      
        // Compteur de commandes par heure
        orders.groupBy((key, value) -> 
                LocalDateTime.now().truncatedTo(ChronoUnit.HOURS).toString())
            .count(Materialized.as("orders-per-hour"))
            .toStream()
            .foreach((hour, count) -> 
                redisTemplate.opsForValue().set("orders:" + hour, count));
      
        // Revenu par catégorie (fenêtre glissante 1h)
        orders.flatMapValues(order -> order.getItems())
            .groupBy((key, item) -> item.getCategoryId())
            .windowedBy(TimeWindows.ofSizeWithNoGrace(Duration.ofHours(1)))
            .aggregate(
                () -> BigDecimal.ZERO,
                (key, item, total) -> total.add(item.getPrice().multiply(BigDecimal.valueOf(item.getQuantity()))),
                Materialized.with(Serdes.String(), bigDecimalSerde))
            .toStream()
            .foreach((windowedKey, revenue) ->
                redisTemplate.opsForValue().set(
                    "revenue:" + windowedKey.key() + ":" + windowedKey.window().start(),
                    revenue.toString()));
      
        return orders;
    }
}
```

Le tableau de bord Grafana interroge Redis toutes les 10 secondes pour afficher les métriques à jour. La latence totale — de l'événement à l'affichage — est inférieure à 30 secondes.

### 10.5.3 Entrepôt Analytique avec ClickHouse

Pour les analyses plus complexes — tendances sur plusieurs mois, segmentation client, performance produit —, TechnoCommerce utilise ClickHouse, une base de données colonne optimisée pour l'analytique.

Un job Apache Flink consomme les événements Kafka, effectue des enrichissements (jointure avec les données produit, géocodage des adresses) et charge les résultats dans ClickHouse.

```sql
-- Table de faits des commandes dans ClickHouse
CREATE TABLE fact_orders (
    order_id UUID,
    customer_id UUID,
    order_date DateTime,
    total_amount Decimal(10,2),
    currency String,
    status String,
    channel String,
    shipping_region String,
    payment_method String,
    items_count UInt32,
    processing_time_ms UInt32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(order_date)
ORDER BY (order_date, customer_id);

-- Vue matérialisée : ventes quotidiennes par catégorie
CREATE MATERIALIZED VIEW mv_daily_sales_by_category
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(sale_date)
ORDER BY (sale_date, category_id)
AS SELECT
    toDate(order_date) AS sale_date,
    category_id,
    count() AS order_count,
    sum(item_amount) AS total_revenue,
    avg(item_amount) AS avg_order_value
FROM fact_order_items
GROUP BY sale_date, category_id;
```

> **Point d'intégration**
> *Domaine* : Données (Le Nom)
> *Patron utilisé* : Materialized View + CQRS
> *Justification* : Séparer les modèles de lecture analytique des modèles d'écriture transactionnelle, pré-calculer les agrégations pour des requêtes rapides

Les analystes et dirigeants accèdent à ces données via Apache Superset, qui génère des rapports interactifs et des tableaux de bord stratégiques. La latence des données dans ClickHouse est de l'ordre de 5 minutes — acceptable pour l'analytique stratégique.

### 10.5.4 Traçabilité de Bout en Bout

La traçabilité complète d'une commande, de sa création à son analyse, constitue un avantage concurrentiel pour TechnoCommerce. En cas de litige ou d'audit, l'entreprise peut reconstituer l'historique complet : qui a passé la commande, quand, via quel canal, quel était l'état du stock, quand le paiement a été capturé, quand l'expédition a eu lieu.

Cette traçabilité repose sur plusieurs mécanismes. Les identifiants de corrélation (trace ID) propagés à travers tous les services permettent de lier les logs et événements relatifs à une même commande. L'Event Sourcing partiel — TechnoCommerce conserve l'historique complet des événements de commande dans un topic Kafka avec rétention illimitée — permet de reconstituer l'état à n'importe quel moment. La lignée des données (data lineage) documentée dans le catalogue de données indique d'où provient chaque métrique analytique.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│              TRAÇABILITÉ COMMANDE #550e8400-e29b-41d4-a716                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Événements (topic orders.events, rétention illimitée) :                    │
│  ──────────────────────────────────────────────────────                     │
│  2025-01-15 10:23:45 │ OrderCreated      │ { items: [...], total: 299.99 }  │
│  2025-01-15 10:23:46 │ PaymentCaptured   │ { txn: "PAY-789", amount: 299.99}│
│  2025-01-15 10:23:47 │ InventoryConfirmed│ { warehouse: "MTL-01" }          │
│  2025-01-15 10:23:48 │ ShipmentScheduled │ { carrier: "Purolator" }         │
│  2025-01-15 14:30:00 │ ShipmentDispatched│ { tracking: "1Z999AA10123456784"}│
│  2025-01-17 09:15:00 │ ShipmentDelivered │ { signature: "J. Smith" }        │
│  2025-01-17 09:15:01 │ OrderCompleted    │ { duration_hours: 46.85 }        │
│                                                                             │
│  Traces (Jaeger) :                                                          │
│  ──────────────────                                                         │
│  Trace ID: abc123... → 47 spans, 7 services, durée totale: 46h 51m 16s     │
│                                                                             │
│  Analytique (ClickHouse) :                                                  │
│  ────────────────────────                                                   │
│  Inclus dans : mv_daily_sales (2025-01-15), mv_customer_ltv (cust-456)     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 10.6 Synthèse : Convergence des Trois Domaines

### 10.6.1 Récapitulatif des Patrons Utilisés

L'étude de cas Order-to-Cash a mobilisé des patrons des trois domaines d'intégration. Le tableau suivant récapitule leur utilisation par phase :

| Phase         | Domaine | Patron               | Composant          | Justification                 |
| ------------- | ------- | -------------------- | ------------------ | ----------------------------- |
| Capture       | App     | BFF                  | Mobile/Web/POS BFF | Adaptation au canal           |
| Capture       | App     | API Gateway          | Kong               | Authentification, routage     |
| Capture       | App     | Aggregator           | Order Service      | Validation multi-services     |
| Capture       | App     | Circuit Breaker      | Tous les appels    | Résilience aux pannes        |
| Persistance   | Data    | Transactional Outbox | Order Service      | Atomicité DB + event         |
| Persistance   | Data    | CDC                  | Debezium           | Capture sans impact perf      |
| Orchestration | Event   | Saga (Chorégraphie) | Inter-services     | Transaction distribuée       |
| Orchestration | Event   | Idempotent Consumer  | Tous les consumers | Livraisons multiples          |
| Orchestration | Event   | DLQ                  | Kafka              | Gestion des échecs           |
| Reporting     | Data    | Materialized View    | Redis, ClickHouse  | Agrégations pré-calculées  |
| Reporting     | Data    | CQRS                 | Analytics          | Séparation lecture/écriture |
| Transversal   | —      | OpenTelemetry        | Tous les services  | Observabilité distribuée    |

### 10.6.2 Points de Décision Architecturale

Plusieurs décisions architecturales clés ont façonné cette solution. Ces décisions, documentées ci-dessous, illustrent comment les critères du chapitre IX guident les choix concrets.

> **Décision architecturale**
> *Contexte* : Choix entre orchestration et chorégraphie pour la saga Order-to-Cash
> *Options* : (A) Orchestration avec un service dédié (Temporal, Camunda) ; (B) Chorégraphie pure via événements Kafka
> *Décision* : Option B — Chorégraphie pure
> *Justification* : Le processus O2C est relativement linéaire avec peu de branchements conditionnels. La chorégraphie préserve l'autonomie des services et évite un point central de coordination. Pour des processus plus complexes avec logique conditionnelle élaborée, l'orchestration serait reconsidérée.

> **Décision architecturale**
> *Contexte* : Choix de la technologie de streaming pour les métriques temps réel
> *Options* : (A) Kafka Streams ; (B) Apache Flink ; (C) ksqlDB
> *Décision* : Option A — Kafka Streams pour les compteurs simples, Option B — Flink pour les enrichissements complexes
> *Justification* : Kafka Streams, embarqué dans les applications, convient aux agrégations simples sans infrastructure supplémentaire. Flink, plus puissant, est réservé aux traitements nécessitant des jointures multi-sources et des fenêtres complexes.

> **Décision architecturale**
> *Contexte* : Stratégie de rétention des événements Kafka
> *Options* : (A) Rétention temporelle (7 jours) ; (B) Rétention basée sur la taille ; (C) Compaction ; (D) Rétention illimitée pour certains topics
> *Décision* : Option D — Rétention illimitée pour `orders.events`, compaction pour les topics d'état, 7 jours pour les topics opérationnels
> *Justification* : L'historique complet des commandes constitue une donnée critique pour l'audit et l'analytique. La compaction convient aux topics représentant l'état courant d'entités. La rétention courte suffit pour les événements opérationnels sans valeur historique.

### 10.6.3 Métriques de Performance

La solution déployée atteint les objectifs de performance suivants :

| Métrique                            | Cible               | Réalisé   | Observation                         |
| ------------------------------------ | ------------------- | ----------- | ----------------------------------- |
| Latence création commande (P99)     | < 500 ms            | 380 ms      | Synchrone, validation incluse       |
| Latence événement → analytics     | < 1 minute          | 45 secondes | Kafka Streams → Redis              |
| Débit pics promotionnels            | 150K commandes/jour | 180K testé | Scaling horizontal Kafka + services |
| Disponibilité processus O2C         | 99.9%               | 99.95%      | Circuit breakers + replicas         |
| Temps moyen de récupération (MTTR) | < 5 minutes         | 3.2 minutes | Alerting + runbooks automatisés    |

### 10.6.4 Leçons Apprises

Le déploiement de cette architecture a généré plusieurs enseignements que TechnoCommerce partage avec la communauté.

**L'idempotence n'est pas négociable.** Lors des premiers déploiements, plusieurs bogues ont causé des duplications de paiement ou d'envoi de courriels. Depuis, chaque consommateur implémente systématiquement un mécanisme d'idempotence, même si cela semble superflu au premier abord.

**Le monitoring de la saga est critique.** Sans observabilité appropriée, le débogage d'une saga distribuée devient cauchemardesque. L'investissement dans OpenTelemetry et les tableaux de bord Grafana a été rentabilisé dès les premières semaines de production.

**Les compensations doivent être testées.** Les chemins d'échec sont moins fréquents que les chemins nominaux, mais leur impact est disproportionné. TechnoCommerce a intégré des tests de chaos qui simulent des pannes de services pour valider les compensations en environnement de staging.

**Le CDC introduit une latence incompressible.** Même avec Debezium et Kafka, la latence entre l'écriture en base et la disponibilité de l'événement est de l'ordre de 100-500 ms. Pour les cas d'usage exigeant une réactivité inférieure, l'émission directe d'événement depuis l'application (en complément du CDC) peut être nécessaire.

**La gouvernance des schémas évite les catastrophes.** L'adoption du Schema Registry avec règles de compatibilité strictes a prévenu plusieurs déploiements qui auraient cassé les consommateurs existants. Le coût initial de rigueur est largement compensé par la stabilité gagnée.

---

## Conclusion et Transition

Ce chapitre a démontré l'application concrète des patrons d'intégration présentés tout au long de cette monographie. Le processus Order-to-Cash de TechnoCommerce illustre comment les trois domaines d'intégration — applications, données, événements — convergent pour former un système cohérent, résilient et observable.

La phase de Capture a mobilisé l'intégration des applications avec ses patrons de couplage synchrone : BFF pour l'adaptation au canal, API Gateway pour les préoccupations transversales, Aggregator pour la composition, Circuit Breaker pour la résilience. Ces patrons répondent à l'exigence d'immédiateté de l'expérience utilisateur.

La phase de Persistance a illustré l'intégration des données avec le Transactional Outbox et le CDC. Ces mécanismes garantissent l'atomicité entre la persistance locale et la propagation vers l'écosystème, sans recourir aux transactions distribuées.

La phase d'Orchestration a déployé l'intégration des événements avec la saga chorégraphiée, les consommateurs idempotents et les Dead Letter Queues. Ce découplage maximal permet l'autonomie des services tout en maintenant la cohérence du processus métier global.

La phase de Reporting a combiné les trois domaines : les événements alimentent des processeurs de streaming (Event), qui maintiennent des vues matérialisées (Data), exposées via des interfaces dédiées (App). Cette convergence produit une visibilité temps réel et historique sur l'ensemble du processus.

Au-delà des patrons techniques, cette étude de cas souligne l'importance de la cohérence architecturale. Les choix ne sont pas isolés ; ils forment un système où chaque décision influence les autres. La chorégraphie choisie pour la saga impose l'idempotence des consommateurs. Le CDC requiert une gestion soigneuse des schémas. Les vues matérialisées dépendent de la fiabilité du flux événementiel.

Le chapitre XI, conclusion de cette monographie, élargira la perspective vers l'Entreprise Agentique. Les patrons d'intégration décrits ici constituent les fondations sur lesquelles les agents IA pourront opérer : des APIs bien définies, des données cohérentes, des événements fiables. L'interopérabilité maîtrisée n'est pas une fin en soi ; elle est le substrat qui rend possible l'automatisation intelligente de demain.

---

## Résumé du Chapitre

**Thème central** : Le processus Order-to-Cash illustre l'application intégrée des patrons d'architecture des chapitres III à IX sur un scénario métier complet, de la prise de commande à l'encaissement et au reporting.

**Phase 1 — Capture (Domaine App)** :

* Backend for Frontend adapte l'interface à chaque canal (mobile, web, POS)
* API Gateway centralise authentification, routage et rate limiting
* Aggregator Pattern compose les validations multi-services
* Circuit Breaker protège contre les défaillances en cascade

**Phase 2 — Persistance (Domaine Data)** :

* Transactional Outbox garantit l'atomicité entre écriture DB et publication d'événement
* Change Data Capture (Debezium) capture les modifications via les logs de réplication
* La propagation vers Kafka alimente les consommateurs en aval

**Phase 3 — Orchestration (Domaine Event)** :

* Saga Pattern en chorégraphie coordonne la transaction distribuée
* Compensations définies pour chaque étape en cas d'échec
* Idempotent Consumer gère les livraisons multiples
* Dead Letter Queue isole les messages en échec

**Phase 4 — Reporting (Convergence)** :

* Kafka Streams maintient des métriques temps réel dans Redis
* Apache Flink alimente ClickHouse pour l'analytique
* Materialized Views pré-calculent les agrégations
* OpenTelemetry assure la traçabilité de bout en bout

**Décisions clés** :

* Chorégraphie plutôt qu'orchestration pour préserver l'autonomie des services
* Rétention illimitée des événements de commande pour l'audit
* Schema Registry avec règles de compatibilité strictes

**Métriques atteintes** :

* Latence P99 création de commande : 380 ms
* Disponibilité O2C : 99.95%
* Débit testé : 180K commandes/jour

**Enseignements** :

* L'idempotence est non négociable dans toute architecture événementielle
* L'observabilité de la saga est critique pour le débogage
* Les compensations doivent être testées comme les chemins nominaux
* La gouvernance des schémas prévient les régressions

---

*Chapitre suivant : XI — Conclusion et Perspectives : Vers l'Entreprise Agentique*
