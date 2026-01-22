# Chapitre IX — Synthèse : Architecture de Référence Convergente

*Focus : L'assemblage des pièces (App, Data, Event) pour former un écosystème cohérent.*

---

## Introduction

Les six chapitres précédents ont exploré les trois domaines d'intégration — applications, données, événements — ainsi que les standards, la résilience et les mécanismes de collaboration qui les soutiennent. Chaque domaine a révélé ses patrons caractéristiques, ses forces et ses compromis inhérents. Le Verbe orchestre les actions synchrones avec la rigueur des appels API. Le Nom structure l'état partagé avec les mécanismes de propagation et de cohérence. Le Signal libère les systèmes de leurs dépendances temporelles par la communication asynchrone. Ces trois piliers, pris isolément, offrent des réponses partielles aux défis d'interopérabilité. Leur véritable puissance émerge de leur combinaison.

Ce neuvième chapitre accomplit la synthèse promise depuis l'introduction de cette monographie. Il ne s'agit plus d'examiner les patrons individuellement, mais de les assembler en une architecture de référence cohérente — un blueprint applicable aux écosystèmes d'entreprise contemporains. Cette architecture convergente valide la thèse centrale : l'interopérabilité forme un continuum où chaque position implique des compromis explicites, et seule une stratégie hybride permet de naviguer ce continuum avec discernement.

La convergence architecturale que nous proposons s'articule autour de trois axes. Le premier axe établit les principes de l'hybridation : comment les systèmes réactifs et l'approche « Inside-Out » unifient les trois domaines sous une philosophie commune. Le deuxième axe présente le blueprint lui-même : une vue logique en couches distinguant le System of Record, l'Integration Backbone et le System of Engagement, accompagnée de règles d'or non négociables. Le troisième axe fournit les instruments de décision : matrices de sélection des patrons, critères de placement et guides contextuels permettant à l'architecte de choisir la bonne approche pour chaque interaction.

L'objectif n'est pas de prescrire une architecture unique applicable à toutes les organisations — une telle prétention ignorerait la diversité des contextes. Il est plutôt de fournir un cadre de raisonnement, un vocabulaire partagé et des heuristiques éprouvées qui accélèrent les décisions architecturales. L'étude de cas du chapitre X démontrera l'application concrète de ce cadre sur un scénario Order-to-Cash, tandis que le chapitre XI projettera cette architecture vers son évolution agentique.

---

## 9.1 Convergence Architecturale : L'Hybridation

### 9.1.1 Les Principes des Systèmes Réactifs

La convergence des trois domaines d'intégration trouve son expression la plus élégante dans les principes des Reactive Systems, formalisés par le Manifeste Réactif (Reactive Manifesto, 2014). Ce manifeste, signé par des figures influentes de l'architecture distribuée dont Jonas Bonér et Martin Thompson, propose quatre propriétés fondamentales que tout système distribué moderne devrait exhiber : la réactivité, la résilience, l'élasticité et l'orientation message.

> **Définition formelle**
> **Reactive Systems** : Architecture distribuée exhibant quatre propriétés : réactivité (temps de réponse prévisible), résilience (maintien du service malgré les pannes), élasticité (adaptation à la charge) et orientation message (communication asynchrone découplée).

La **réactivité** (responsive) désigne la capacité du système à répondre dans un temps prévisible et acceptable. Cette propriété constitue l'objectif ultime; les trois autres propriétés sont les moyens de l'atteindre. Un système réactif ne se contente pas de fonctionner — il fonctionne de manière fiable et prévisible, permettant aux utilisateurs et aux systèmes clients d'établir des attentes raisonnables.

La **résilience** (resilient) caractérise la capacité du système à maintenir sa réactivité malgré les défaillances. Les pannes sont inévitables dans les systèmes distribués; la question n'est pas de savoir si elles surviendront, mais comment le système y réagira. La résilience s'obtient par la réplication, l'isolation (bulkhead), la délégation et la dégradation gracieuse. Les patrons du chapitre VII — Circuit Breaker, Retry with Exponential Backoff, Fallback — sont les instruments de cette résilience.

L'**élasticité** (elastic) désigne la capacité du système à s'adapter aux variations de charge sans intervention manuelle. Lors d'un pic de demande, le système alloue dynamiquement des ressources supplémentaires; lorsque la charge diminue, il libère ces ressources. Cette propriété est essentielle dans les environnements infonuagiques où les ressources sont facturées à l'usage. L'élasticité repose sur l'absence de goulots d'étranglement centralisés et sur la capacité à partitionner le travail.

L'**orientation message** (message-driven) constitue le fondement architectural des trois propriétés précédentes. Les systèmes réactifs communiquent par messages asynchrones, établissant des frontières claires entre composants et permettant le découplage temporel et spatial. Cette orientation message n'est pas simplement une technique d'implémentation; c'est un choix architectural fondamental qui façonne la structure même du système.

```
PROPRIÉTÉS DES SYSTÈMES RÉACTIFS

                    ┌─────────────────┐
                    │   RÉACTIVITÉ    │
                    │  (Responsive)   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              │              ▼
     ┌─────────────┐        │       ┌─────────────┐
     │ RÉSILIENCE  │        │       │ ÉLASTICITÉ │
     │ (Resilient) │        │       │  (Elastic)  │
     └──────┬──────┘        │       └──────┬──────┘
            │               │              │
            └───────────────┼──────────────┘
                            │
                            ▼
               ┌─────────────────────┐
               │ ORIENTATION MESSAGE │
               │  (Message-Driven)   │
               └─────────────────────┘
```

L'articulation avec les trois domaines d'intégration devient alors évidente. L'intégration des événements (Le Signal) incarne directement l'orientation message. L'intégration des données (Le Nom) contribue à la résilience par la réplication et à l'élasticité par le partitionnement. L'intégration des applications (Le Verbe) assure la réactivité perçue par les utilisateurs tout en s'appuyant sur les deux autres domaines pour sa robustesse.

### 9.1.2 L'Approche Inside-Out : Database Unbundling

La convergence architecturale trouve une expression particulièrement puissante dans l'approche « Inside-Out », également connue sous le nom de Database Unbundling. Cette approche, popularisée par Martin Kleppmann et les équipes de Confluent, propose de repenser le rôle du journal de transactions (log) dans l'architecture d'entreprise.

> **Définition formelle**
> **Database Unbundling (Inside-Out)** : Approche architecturale où le journal de transactions (log) devient la colonne vertébrale du système d'information, permettant à chaque application de construire ses propres vues matérialisées optimisées pour ses besoins spécifiques.

Dans une architecture traditionnelle, la base de données constitue le centre du système. Les applications écrivent dans la base, lisent depuis la base, et dépendent de sa disponibilité pour toute opération. Le schéma de la base représente un compromis entre les besoins d'écriture (intégrité, normalisation) et les besoins de lecture (performance, dénormalisation). Ce compromis satisfait rarement pleinement les deux usages.

L'approche Inside-Out inverse cette perspective. Le journal de transactions — ce mécanisme interne que toute base de données utilise pour garantir la durabilité et la récupération — devient un artefact architectural de premier plan. Les changements sont d'abord capturés dans un log distribué (typiquement Apache Kafka), puis propagés vers différentes « projections » optimisées pour leurs cas d'usage respectifs : une base relationnelle pour les transactions, un moteur de recherche pour les requêtes textuelles, un cache pour les accès rapides, un entrepôt pour l'analytique.

```
APPROCHE TRADITIONNELLE (Outside-In)

┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   App A     │    │    App B    │    │    App C    │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                          ▼
              ┌─────────────────────┐
              │   Base de données   │
              │   (compromis R/W)   │
              └─────────────────────┘


APPROCHE INSIDE-OUT (Database Unbundling)

┌─────────────────────────────────────────────────────────┐
│              LOG DISTRIBUÉ (Kafka)                      │
│         Source de vérité immuable et ordonnée          │
└───────────┬───────────────┬───────────────┬────────────┘
            │               │               │
            ▼               ▼               ▼
    ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
    │  PostgreSQL   │ │ Elasticsearch │ │    Redis      │
    │ (transactions)│ │  (recherche)  │ │   (cache)     │
    └───────────────┘ └───────────────┘ └───────────────┘
            │               │               │
            ▼               ▼               ▼
    ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
    │    App A      │ │     App B     │ │     App C     │
    │ (OLTP)        │ │ (Search)      │ │ (Read-heavy)  │
    └───────────────┘ └───────────────┘ └───────────────┘
```

Cette approche présente plusieurs avantages fondamentaux. Elle élimine le compromis lecture/écriture en permettant à chaque projection d'être optimisée pour son usage. Elle découple les applications les unes des autres : l'ajout d'une nouvelle projection ne nécessite pas de modifier les applications existantes. Elle préserve l'historique complet des changements, permettant le replay, l'audit et la reconstruction des états passés. Elle aligne naturellement l'architecture sur les principes des systèmes réactifs en plaçant le message (l'événement dans le log) au centre.

Le lien avec les patrons des chapitres précédents est direct. Le Change Data Capture (CDC) du chapitre IV alimente le log à partir des bases existantes. L'Event Sourcing du chapitre V pousse cette logique à son terme en faisant du log la source de vérité primaire. Les Materialized Views se multiplient naturellement comme projections optimisées. Le CQRS trouve son expression infrastructurelle dans cette séparation physique des modèles de lecture et d'écriture.

> **Perspective stratégique**
> L'approche Inside-Out ne convient pas à toutes les organisations. Elle requiert une maturité significative en termes d'infrastructure (gestion d'un cluster Kafka), de compétences (streaming, event sourcing) et de culture (cohérence éventuelle). Les organisations au début de leur parcours d'intégration peuvent commencer par des approches plus traditionnelles et évoluer progressivement vers ce modèle.

### 9.1.3 La Convergence des Trois Domaines

La convergence architecturale que nous proposons n'abolit pas la distinction entre les trois domaines d'intégration; elle les articule dans une vision unifiée où chacun joue un rôle spécifique et complémentaire.

L'**intégration des applications (Le Verbe)** demeure le point d'entrée pour les interactions synchrones exigeant une réponse immédiate. L'API Gateway, le Backend for Frontend et les patrons de résilience (Circuit Breaker, Timeout) garantissent que les utilisateurs et systèmes clients obtiennent des réponses prévisibles. Cependant, dans une architecture convergente, ces interactions synchrones ne constituent que la surface visible d'un iceberg largement asynchrone.

L'**intégration des événements (Le Signal)** forme la colonne vertébrale de l'architecture. Chaque changement d'état significatif émet un événement qui se propage à travers le système. Les Sagas coordonnent les processus distribués sans verrouillage global. Le Transactional Outbox garantit l'atomicité entre les écritures locales et la publication des événements. Cette orientation événementielle n'est pas un ajout cosmétique; elle restructure fondamentalement la façon dont les systèmes interagissent.

L'**intégration des données (Le Nom)** assure la cohérence et l'accessibilité de l'état. Le CDC capture les changements des systèmes qui ne peuvent pas émettre directement des événements. Les Schema Registries garantissent la compatibilité des structures de données à travers l'écosystème. Les Materialized Views offrent des projections optimisées pour chaque cas d'usage de lecture. Le Data Mesh organise la propriété et la gouvernance de ces données par domaine métier.

> **Règle d'or**
> Dans une architecture convergente, les trois domaines ne sont pas des alternatives mais des couches complémentaires. Toute transaction métier significative implique typiquement les trois : une entrée synchrone (App), une persistance avec capture de changement (Data), et une propagation asynchrone (Event).

La convergence se manifeste également dans le flux d'une transaction typique. Considérons une modification de commande :

1. **App** : Le client soumet la modification via l'API; le service valide la requête et répond immédiatement avec un accusé de réception.
2. **Data** : Le service persiste la modification dans sa base locale; le CDC ou le Transactional Outbox capture ce changement.
3. **Event** : Un événement `OrderModified` est publié sur le bus; les services intéressés (inventaire, facturation, notification) réagissent de manière autonome.

Cette séquence illustre la complémentarité : l'utilisateur obtient une réponse rapide (App), l'état est persisté de manière fiable (Data), et les conséquences se propagent de manière découplée (Event). Aucun domaine ne suffit seul; leur combinaison crée un système à la fois réactif et robuste.

### 9.1.4 Cohérence et Compromis

La convergence architecturale ne résout pas magiquement le défi fondamental de la cohérence dans les systèmes distribués. Le théorème CAP (chapitre II) demeure une contrainte incontournable. L'architecture convergente ne prétend pas l'abolir; elle propose plutôt un cadre pour naviguer explicitement ses compromis.

La **cohérence forte** reste disponible là où elle est indispensable. Une transaction bancaire qui débite un compte et crédite un autre peut utiliser une transaction distribuée (2PC) ou un mécanisme de compensation fiable. Le coût de cette cohérence — latence accrue, disponibilité réduite lors des partitions — est acceptable pour ces cas critiques.

La **cohérence éventuelle** devient le modèle par défaut pour la majorité des interactions. Un tableau de bord analytique peut afficher des données vieilles de quelques secondes sans impact métier. Une notification peut être légèrement retardée sans conséquence grave. Accepter explicitement cette cohérence éventuelle libère l'architecture des contraintes du couplage fort.

Entre ces extrêmes, des **modèles de cohérence intermédiaires** offrent des compromis nuancés. La cohérence causale garantit que les effets suivent leurs causes, sans imposer un ordre total global. La cohérence de session assure qu'un utilisateur voit ses propres écritures, même si d'autres utilisateurs ont une vue légèrement différée. Ces modèles, supportés par des technologies comme Apache Kafka (garantie d'ordre par partition) ou CRDTs (chapitre VIII), permettent d'ajuster finement le compromis cohérence/disponibilité.

> **Décision architecturale**
> *Contexte* : Choix du niveau de cohérence pour une nouvelle fonctionnalité.
> *Options* : (A) Cohérence forte via transaction distribuée; (B) Cohérence éventuelle via événements; (C) Cohérence causale via partitionnement intelligent.
> *Décision* : Analyser l'impact métier d'une incohérence temporaire. Si l'incohérence peut causer une perte financière ou une violation réglementaire, choisir (A). Si elle cause simplement un affichage légèrement périmé, choisir (B). Si l'ordre des opérations d'un même agrégat doit être préservé mais que l'ordre global importe peu, choisir (C).

---

## 9.2 Blueprint d'Architecture : Le Modèle Cible

### 9.2.1 Vue Logique en Trois Couches

L'architecture de référence convergente s'organise en trois couches logiques, chacune correspondant à un rôle architectural distinct et à une position caractéristique sur le continuum couplage/découplage.

> **Définition formelle**
> **System of Record** (Couche Data) : Ensemble des systèmes responsables de la persistance durable et de l'intégrité des données métier. Chaque donnée possède un System of Record désigné qui fait autorité.

> **Définition formelle**
> **Integration Backbone** (Couche Event) : Infrastructure de communication asynchrone qui relie les composants de l'écosystème. Typiquement implémenté par un bus d'événements distribué (Kafka, Pulsar) complété par un Event Mesh.

> **Définition formelle**
> **System of Engagement** (Couche App) : Ensemble des points d'interaction avec les utilisateurs, partenaires et systèmes externes. Inclut les API Gateways, les BFF, les portails web et les applications mobiles.

```
VUE LOGIQUE EN TROIS COUCHES

┌─────────────────────────────────────────────────────────────────────┐
│                    SYSTEM OF ENGAGEMENT                             │
│                        (Couche App)                                 │
│                                                                     │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐       │
│  │    API    │  │    BFF    │  │    BFF    │  │  Partner  │       │
│  │  Gateway  │  │  Mobile   │  │   Web     │  │    API    │       │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘       │
│        │              │              │              │              │
└────────┼──────────────┼──────────────┼──────────────┼──────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   INTEGRATION BACKBONE                              │
│                      (Couche Event)                                 │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    EVENT BUS (Kafka)                         │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │   │
│  │  │ Orders  │  │Inventory│  │Payments │  │ Users   │        │   │
│  │  │ Topic   │  │ Topic   │  │ Topic   │  │ Topic   │        │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└────────────────────────────┬────────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     SYSTEM OF RECORD                                │
│                       (Couche Data)                                 │
│                                                                     │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │
│  │   Order DB    │  │  Inventory DB │  │   User DB     │           │
│  │  (PostgreSQL) │  │   (MongoDB)   │  │  (PostgreSQL) │           │
│  │               │  │               │  │               │           │
│  │  + CDC        │  │  + CDC        │  │  + CDC        │           │
│  │  (Debezium)   │  │  (Debezium)   │  │  (Debezium)   │           │
│  └───────────────┘  └───────────────┘  └───────────────┘           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

La **couche System of Record** constitue le fondement de l'architecture. Chaque domaine métier possède sa propre base de données, optimisée pour ses besoins spécifiques. Le service Commandes peut utiliser PostgreSQL pour ses garanties ACID; le service Inventaire peut préférer MongoDB pour sa flexibilité de schéma; le service Catalogue peut s'appuyer sur Elasticsearch pour ses capacités de recherche. Cette polyglot persistence (persistence polyglotte) est rendue possible par le découplage qu'offre l'Integration Backbone.

Chaque System of Record est équipé d'un mécanisme de capture de changements — typiquement Debezium pour le CDC log-based. Chaque modification persistée génère un événement qui alimente l'Integration Backbone. Cette émission systématique des changements constitue le pont entre la couche Data et la couche Event.

La **couche Integration Backbone** matérialise la colonne vertébrale événementielle de l'architecture. Le bus d'événements (Kafka dans notre exemple) reçoit les événements de tous les Systems of Record et les rend disponibles à tous les consommateurs intéressés. Les topics sont organisés par domaine métier, reflétant la structure organisationnelle selon les principes du Data Mesh.

L'Integration Backbone ne se limite pas au transport de messages. Il inclut également le Schema Registry pour la gouvernance des contrats de données, les mécanismes de routage intelligent (Event Mesh) pour la distribution multi-région, et les outils de monitoring pour l'observabilité des flux.

La **couche System of Engagement** gère toutes les interactions synchrones avec le monde extérieur. L'API Gateway assure l'authentification, l'autorisation et la limitation de débit pour les appels entrants. Les Backend for Frontend adaptent les réponses aux besoins spécifiques de chaque canal (mobile, web, partenaires). Les services exposés ici orchestrent les appels vers les domaines métier et peuvent également publier des événements vers l'Integration Backbone.

### 9.2.2 Flux de Données Caractéristiques

L'architecture en trois couches se manifeste concrètement dans les flux de données qui traversent le système. Quatre types de flux caractéristiques illustrent les interactions entre couches.

**Flux de commande (Command Flow)** : Une requête utilisateur arrive via le System of Engagement, traverse les validations synchrones nécessaires, persiste son effet dans le System of Record, puis émet un événement vers l'Integration Backbone. Ce flux combine le couplage fort initial (réponse à l'utilisateur) avec le découplage subséquent (propagation asynchrone).

```
FLUX DE COMMANDE

Utilisateur
    │
    ▼
┌─────────────────┐
│  API Gateway    │ ← Authentification, Rate Limiting
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Order Service  │ ← Validation métier
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌───────┐  ┌───────────┐
│  DB   │  │  Outbox   │
└───────┘  └─────┬─────┘
                 │
                 ▼
         ┌─────────────┐
         │   Kafka     │
         │  (Event)    │
         └─────────────┘
```

**Flux de requête (Query Flow)** : Une demande de lecture peut être servie directement depuis une projection optimisée (Materialized View, cache Redis, index Elasticsearch) sans solliciter le System of Record. Ce découplage des lectures permet d'optimiser indépendamment les performances de consultation.

**Flux de réaction (Reaction Flow)** : Un événement publié sur l'Integration Backbone déclenche des traitements dans plusieurs services consommateurs. Chaque consommateur réagit de manière autonome, à son propre rythme, sans coordination explicite avec les autres. Ce flux incarne le découplage maximal de l'intégration événementielle.

**Flux de synchronisation (Sync Flow)** : Les projections de lecture sont maintenues à jour par des consommateurs dédiés qui lisent les événements de l'Integration Backbone et mettent à jour leurs structures optimisées. Ce flux assure la cohérence éventuelle entre le System of Record et les caches/vues matérialisées.

### 9.2.3 Les Règles d'Or

L'architecture de référence s'accompagne de principes directeurs non négociables — des règles d'or qui garantissent la cohérence et l'intégrité de l'ensemble.

> **Règle d'or — Jamais d'écriture directe inter-service**
> Un service ne doit jamais écrire directement dans la base de données d'un autre service. Toute modification de l'état d'un domaine transite par le service propriétaire de ce domaine, via une API ou un événement. Cette règle préserve l'encapsulation des domaines et évite les corruptions de données.

Cette première règle est fondamentale pour maintenir l'intégrité des domaines. Lorsqu'un service écrit directement dans la base d'un autre, il contourne les validations métier, ignore les règles de cohérence et crée des couplages cachés. La base de données devient une « API publique accidentelle » dont le schéma ne peut plus évoluer sans risquer de casser les clients intrusifs.

> **Règle d'or — Tout changement d'état émet un événement**
> Chaque modification d'état significative dans un System of Record doit générer un événement correspondant sur l'Integration Backbone. Cette émission peut être explicite (publication applicative) ou implicite (capture via CDC), mais elle doit être garantie atomiquement avec l'écriture.

Cette deuxième règle assure que l'Integration Backbone reflète fidèlement l'état de l'écosystème. Les services consommateurs peuvent s'abonner aux événements qui les concernent et construire leurs propres projections sans dépendre de la disponibilité des services producteurs. Le Transactional Outbox (chapitre V) et le CDC (chapitre IV) sont les mécanismes techniques qui garantissent cette atomicité.

> **Règle d'or — Requêtes synchrones pour les lectures, événements pour les écritures**
> Les opérations de lecture qui nécessitent une réponse immédiate utilisent des appels synchrones vers des projections optimisées. Les opérations d'écriture dont les conséquences doivent se propager utilisent des événements asynchrones. Cette séparation aligne le style d'intégration sur la nature de l'opération.

Cette troisième règle explicite le principe CQRS à l'échelle de l'architecture. Les lectures tolèrent une cohérence éventuelle mais exigent une latence faible; elles sont servies depuis des caches, vues matérialisées ou réplicas optimisés. Les écritures exigent une fiabilité absolue mais tolèrent une propagation asynchrone; elles transitent par les événements qui garantissent la livraison.

> **Règle d'or — Idempotence par défaut**
> Tout consommateur d'événements doit être idempotent. Le traitement multiple d'un même événement ne doit pas produire d'effets de bord indésirables. Cette propriété est non négociable dans un système où la garantie « au moins une fois » est le modèle standard.

Cette quatrième règle reconnaît la réalité des systèmes distribués : les messages peuvent être dupliqués (retry après timeout, rebalancing des partitions Kafka, reprise après panne). Le consommateur idempotent transforme ce risque en non-événement. Les stratégies d'implémentation — registre d'identifiants traités, opérations naturellement idempotentes, versionnement optimiste — ont été détaillées au chapitre V.

> **Règle d'or — Contrat explicite à chaque frontière**
> Chaque point d'intégration — API synchrone, événement asynchrone, structure de données partagée — doit être documenté par un contrat explicite et versionné. Les contrats utilisent des formats standards (OpenAPI, AsyncAPI, Avro/Protobuf) et sont gérés dans un registre central.

Cette cinquième règle combat l'entropie naturelle des systèmes distribués. Sans contrats explicites, les interfaces dérivent imperceptiblement, les incompatibilités s'accumulent et les intégrations deviennent fragiles. Le Schema Registry (chapitre IV) et les Consumer-Driven Contracts (chapitre III) fournissent les mécanismes de gouvernance. Le chapitre VI a détaillé les standards qui formalisent ces contrats.

### 9.2.4 Patterns d'Implémentation

Les règles d'or se traduisent en patterns d'implémentation concrets pour les situations récurrentes.

**Pattern : Service avec Outbox**

Chaque service applicatif qui modifie son état suit le même pattern : écrire la modification et l'événement correspondant dans la même transaction locale, puis laisser un processus de relay publier l'événement vers Kafka.

```
┌─────────────────────────────────────────────────────┐
│                   Order Service                     │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │            Business Logic                    │   │
│  └─────────────────────┬───────────────────────┘   │
│                        │                           │
│                        ▼                           │
│  ┌─────────────────────────────────────────────┐   │
│  │         Transaction (ACID locale)           │   │
│  │  ┌─────────────┐    ┌─────────────────┐    │   │
│  │  │   orders    │    │     outbox      │    │   │
│  │  │   table     │    │     table       │    │   │
│  │  └─────────────┘    └─────────────────┘    │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │         Debezium CDC Connector              │   │
│  │         (lit la table outbox)               │   │
│  └─────────────────────┬───────────────────────┘   │
└────────────────────────┼────────────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │   Kafka Topic       │
              │   orders.events     │
              └─────────────────────┘
```

**Pattern : Projection de Lecture**

Les services ayant des besoins de lecture complexes maintiennent des projections dédiées alimentées par les événements.

```
┌─────────────────────────────────────────────────────┐
│               Search Projection                     │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │         Kafka Consumer                       │   │
│  │   (orders.events, products.events, ...)     │   │
│  └─────────────────────┬───────────────────────┘   │
│                        │                           │
│                        ▼                           │
│  ┌─────────────────────────────────────────────┐   │
│  │         Projection Logic                     │   │
│  │   (dénormalisation, enrichissement)         │   │
│  └─────────────────────┬───────────────────────┘   │
│                        │                           │
│                        ▼                           │
│  ┌─────────────────────────────────────────────┐   │
│  │         Elasticsearch Index                  │   │
│  │   (optimisé pour la recherche)              │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

**Pattern : Saga Orchestrée**

Les processus métier impliquant plusieurs domaines utilisent une Saga avec orchestrateur dédié.

```
┌─────────────────────────────────────────────────────────────────┐
│                     Order Saga Orchestrator                     │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    État de la Saga                       │   │
│  │   orderId: 123, step: PAYMENT_PENDING, retries: 0       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  1. OrderCreated        →  Publish: ReserveInventory           │
│  2. InventoryReserved   →  Publish: ProcessPayment             │
│  3. PaymentProcessed    →  Publish: ScheduleShipment           │
│  4. ShipmentScheduled   →  Publish: OrderCompleted             │
│                                                                 │
│  Compensation si échec:                                        │
│  - PaymentFailed        →  Publish: ReleaseInventory           │
│  - InventoryFailed      →  Publish: CancelOrder                │
└─────────────────────────────────────────────────────────────────┘
```

### 9.2.5 Considérations de Déploiement

L'architecture de référence se décline différemment selon les contextes de déploiement.

**Déploiement infonuagique natif** : Les trois couches s'appuient sur les services gérés du fournisseur. L'Integration Backbone utilise Amazon MSK, Confluent Cloud ou Azure Event Hubs. Les Systems of Record exploitent les bases de données gérées (RDS, Cloud SQL, Cosmos DB). Le System of Engagement s'appuie sur les API Gateways natifs (API Gateway AWS, Apigee, Azure API Management). Ce déploiement minimise la charge opérationnelle mais crée une dépendance au fournisseur.

**Déploiement Kubernetes** : L'ensemble de l'architecture est conteneurisé et orchestré par Kubernetes. Kafka s'exécute via Strimzi ou l'opérateur Confluent. Les bases de données utilisent des opérateurs dédiés (CloudNativePG pour PostgreSQL, MongoDB Operator). L'API Gateway s'appuie sur un Ingress Controller enrichi (Kong, Ambassador) ou un Service Mesh (Istio). Ce déploiement offre la portabilité mais exige une expertise opérationnelle significative.

**Déploiement hybride** : Les Systems of Record demeurent sur l'infrastructure existante (on-premise ou hébergée), tandis que l'Integration Backbone et le System of Engagement migrent vers l'infonuagique. Le CDC assure le pont entre les deux mondes. Ce déploiement correspond souvent à une stratégie de migration progressive (Strangler Fig à l'échelle de l'infrastructure).

> **Note technique**
> Le choix du déploiement influence significativement les patterns applicables. Un déploiement multi-région avec Kafka nécessite une réflexion sur la réplication (MirrorMaker 2, Cluster Linking) et le routage des événements. Un déploiement hybride doit gérer la latence réseau entre les composants on-premise et infonuagiques.

---

## 9.3 Matrice de Décision des Patrons

### 9.3.1 Guide de Placement

Disposer d'un catalogue de patrons ne suffit pas; l'architecte doit savoir quand et où appliquer chaque patron. Cette section propose un guide de placement systématique.

**Où placer l'API Gateway ?**

L'API Gateway se positionne à la frontière entre le réseau externe (clients, partenaires) et le réseau interne (microservices). Il constitue le point d'entrée unique pour tout trafic Nord-Sud (entrant depuis l'extérieur).

| Situation                                  | Recommandation                           |
| ------------------------------------------ | ---------------------------------------- |
| Exposition d'APIs à des clients externes  | API Gateway obligatoire                  |
| Trafic entre microservices internes        | Pas de Gateway, utiliser Service Mesh    |
| APIs partenaires B2B avec SLA spécifiques | Gateway dédié ou virtual host séparé |
| Environnement multi-tenant                 | Gateway avec isolation par tenant        |

**Où placer le Cache ?**

Le cache s'insère entre le consommateur et la source de données pour absorber les lectures répétitives et réduire la latence.

| Situation                                       | Recommandation                             |
| ----------------------------------------------- | ------------------------------------------ |
| Données fréquemment lues, rarement modifiées | Cache applicatif (Redis) devant le service |
| Réponses API coûteuses à calculer            | Cache de réponse dans l'API Gateway       |
| Données de session utilisateur                 | Cache distribué avec TTL court            |
| Projections de lecture événementielles        | Materialized View plutôt que cache        |

**Quand choisir CDC vs API ?**

Le choix entre CDC et API pour propager les données dépend de plusieurs facteurs.

| Critère                    | Favorise CDC          | Favorise API            |
| --------------------------- | --------------------- | ----------------------- |
| Latence acceptable          | Secondes à minutes   | Millisecondes           |
| Couplage toléré           | Faible (découplé)   | Acceptable (synchrone)  |
| Volume de changements       | Élevé               | Faible                  |
| Contrôle sur le producteur | Limité (legacy)      | Total (nouveau service) |
| Format des événements     | Schéma DB acceptable | Format métier requis   |

**Où placer l'Anti-Corruption Layer (ACL) ?**

L'ACL s'insère à la frontière entre un système moderne et un système legacy dont le modèle conceptuel diffère.

| Situation                             | Recommandation                 |
| ------------------------------------- | ------------------------------ |
| Intégration avec un ERP legacy       | ACL côté service moderne     |
| Migration progressive (Strangler Fig) | ACL dans la façade de routage |
| Multiples systèmes legacy similaires | ACL partagé avec adaptateurs  |
| Legacy exposant une API propre        | ACL potentiellement évitable  |

**Quand utiliser le Saga Pattern ?**

Le Saga s'applique aux transactions distribuées impliquant plusieurs services avec compensation possible.

| Situation                                  | Recommandation                            |
| ------------------------------------------ | ----------------------------------------- |
| Transaction impliquant 2+ services         | Saga (chorégraphie ou orchestration)     |
| Rollback immédiat requis                  | Éviter Saga, utiliser 2PC si possible    |
| Processus longue durée (heures/jours)     | Saga orchestrée avec persistance d'état |
| Services hétérogènes (legacy + moderne) | Saga orchestrée avec adaptateurs         |

### 9.3.2 Matrice de Sélection par Domaine

Le tableau suivant synthétise les patrons recommandés selon le domaine d'intégration et le type de problème.

| Problème                           | Domaine App               | Domaine Data        | Domaine Event        |
| ----------------------------------- | ------------------------- | ------------------- | -------------------- |
| Point d'entrée unifié             | API Gateway               | —                  | —                   |
| Adaptation par canal                | BFF                       | —                  | —                   |
| Isolation legacy                    | ACL                       | —                  | —                   |
| Migration progressive               | Strangler Fig             | —                  | —                   |
| Agrégation de données             | Aggregator                | Data Virtualization | —                   |
| Validation de contrat               | Consumer-Driven Contracts | Schema Registry     | AsyncAPI             |
| Découverte de services             | Service Registry          | Data Catalog        | —                   |
| Capture de changements              | —                        | CDC                 | —                   |
| Séparation lecture/écriture       | —                        | CQRS                | —                   |
| Optimisation de requêtes           | —                        | Materialized View   | —                   |
| Gouvernance distribuée             | —                        | Data Mesh           | —                   |
| Découplage producteur/consommateur | —                        | —                  | Pub/Sub              |
| Audit et reconstruction             | —                        | —                  | Event Sourcing       |
| Transaction distribuée             | —                        | —                  | Saga                 |
| Atomicité DB + publication         | —                        | —                  | Transactional Outbox |
| Payload volumineux                  | —                        | —                  | Claim Check          |
| Gestion des erreurs                 | Circuit Breaker           | —                  | DLQ                  |
| Scalabilité du traitement          | —                        | —                  | Competing Consumers  |

### 9.3.3 Critères de Sélection Détaillés

Au-delà des heuristiques précédentes, quatre critères fondamentaux guident la sélection des patrons pour chaque point d'intégration.

**Critère 1 : Latence acceptable**

La latence acceptable définit la fenêtre temporelle dans laquelle une interaction doit se compléter pour répondre aux exigences métier.

| Catégorie             | Latence      | Patrons appropriés                  |
| ---------------------- | ------------ | ------------------------------------ |
| Temps réel interactif | < 100 ms     | API synchrone, Cache, BFF            |
| Temps réel système   | 100 ms - 1 s | API avec Circuit Breaker, Aggregator |
| Quasi temps réel      | 1 s - 1 min  | CDC streaming, Événements Kafka    |
| Batch court            | 1 min - 1 h  | Micro-batch, Materialized View       |
| Batch long             | > 1 h        | ETL traditionnel, Data Warehouse     |

**Critère 2 : Cohérence requise**

Le niveau de cohérence requis détermine les compromis acceptables en termes de fraîcheur des données.

| Niveau          | Description                                 | Patrons appropriés               |
| --------------- | ------------------------------------------- | --------------------------------- |
| Forte (ACID)    | Lecture reflète immédiatement l'écriture | Transaction locale, 2PC           |
| Causale         | Effets suivent leurs causes                 | Partitionnement Kafka par clé    |
| Session         | Utilisateur voit ses propres écritures     | Sticky sessions, Read-your-writes |
| Éventuelle     | Convergence à terme garantie               | Pub/Sub, CDC, CQRS                |
| Aucune garantie | Best effort                                 | Cache sans invalidation           |

**Critère 3 : Volume de données**

Le volume de données influence le choix entre streaming et batch, entre stockage centralisé et distribué.

| Volume                  | Caractéristiques          | Patrons appropriés            |
| ----------------------- | -------------------------- | ------------------------------ |
| Faible (< 1 Go/jour)    | Traitement simple          | API directe, batch simple      |
| Moyen (1-100 Go/jour)   | Besoin d'optimisation      | CDC, Kafka, Materialized Views |
| Élevé (> 100 Go/jour) | Infrastructure dédiée    | Kafka partitionné, Data Lake  |
| Très élevé (To/jour) | Architecture spécialisée | Pulsar/Kafka tiered storage    |

**Critère 4 : Couplage toléré**

Le niveau de couplage toléré reflète les contraintes organisationnelles et techniques sur l'évolution indépendante des systèmes.

| Niveau   | Implications                               | Patrons appropriés                        |
| -------- | ------------------------------------------ | ------------------------------------------ |
| Fort     | Équipes coordonnées, releases synchrones | API directe, RPC, schéma partagé         |
| Modéré | Contrats explicites, versions compatibles  | Consumer-Driven Contracts, Schema Registry |
| Faible   | Évolution indépendante, découverte      | Pub/Sub, CDC, Event Sourcing               |
| Minimal  | Aucune connaissance mutuelle               | Event-Carried State Transfer               |

### 9.3.4 Arbres de Décision

Pour faciliter la navigation entre les nombreux patrons, deux arbres de décision synthétisent les choix les plus fréquents.

**Arbre de décision : Style d'intégration**

```
L'appelant a-t-il besoin d'une réponse immédiate ?
│
├─ OUI → Intégration synchrone (App)
│        │
│        └─ Quelle est la nature de la requête ?
│           │
│           ├─ Lecture simple → API directe ou Cache
│           ├─ Lecture complexe → Aggregator ou BFF
│           ├─ Écriture → API avec Transactional Outbox
│           └─ Action sur legacy → API via ACL
│
└─ NON → Intégration asynchrone (Event)
         │
         └─ Quelle est la source du changement ?
            │
            ├─ Base de données existante → CDC (Debezium)
            ├─ Service applicatif → Transactional Outbox
            ├─ Événement métier explicite → Pub/Sub direct
            └─ Transaction multi-services → Saga
```

**Arbre de décision : Gestion de la cohérence**

```
Quel niveau de cohérence est requis ?
│
├─ Cohérence forte immédiate
│  │
│  └─ Transaction limitée à un service ?
│     │
│     ├─ OUI → Transaction locale ACID
│     └─ NON → 2PC ou Saga synchrone (avec précaution)
│
├─ Cohérence éventuelle acceptable
│  │
│  └─ Délai acceptable ?
│     │
│     ├─ < 1 seconde → CDC streaming ou Kafka
│     ├─ < 1 minute → CDC avec micro-batch
│     └─ > 1 minute → Batch traditionnel acceptable
│
└─ Cohérence causale ou session
   │
   └─ Utiliser partitionnement intelligent
      (même clé = même partition = ordre garanti)
```

### 9.3.5 Anti-Patrons à Éviter

La connaissance des patrons appropriés doit s'accompagner de la conscience des anti-patrons fréquents.

> **Anti-patron : Distributed Monolith**
> *Symptôme* : Des microservices qui ne peuvent pas être déployés indépendamment car ils partagent une base de données, des bibliothèques versionnées ensemble, ou des dépendances synchrones circulaires.
> *Cause* : Migration superficielle d'un monolithe vers des services sans découplage réel.
> *Remède* : Appliquer rigoureusement la règle « jamais d'écriture directe inter-service »; utiliser l'Integration Backbone pour le découplage.

> **Anti-patron : Event Soup**
> *Symptôme* : Des dizaines de types d'événements différents sans taxonomie claire, avec des noms incohérents et des schémas qui évoluent de manière incontrôlée.
> *Cause* : Adoption des événements sans gouvernance; chaque équipe définit ses propres conventions.
> *Remède* : Établir un catalogue d'événements avec nomenclature standardisée; utiliser un Schema Registry avec règles de compatibilité.

> **Anti-patron : Synchronous Everything**
> *Symptôme* : Toutes les interactions entre services sont synchrones; une panne d'un service cascade immédiatement à tous ses appelants.
> *Cause* : Habitude des architectures monolithiques; confort avec le modèle requête-réponse.
> *Remède* : Identifier les interactions qui tolèrent l'asynchronie; introduire progressivement les événements pour le découplage.

> **Anti-patron : Async Everything**
> *Symptôme* : Même les interactions nécessitant une réponse immédiate passent par des événements, forçant des patterns de corrélation complexes et des timeouts artificiels.
> *Cause* : Application dogmatique du découplage sans considération du contexte.
> *Remède* : Accepter que certaines interactions sont intrinsèquement synchrones; utiliser l'arbre de décision pour choisir le style approprié.

> **Anti-patron : Shared Database Integration**
> *Symptôme* : Plusieurs services lisent et écrivent dans les mêmes tables d'une base de données partagée.
> *Cause* : Facilité apparente; évitement de la complexité des APIs.
> *Remède* : Établir un propriétaire unique pour chaque ensemble de données; exposer les données via APIs ou événements.

---

## 9.4 Gouvernance de l'Architecture

### 9.4.1 Rôles et Responsabilités

L'architecture de référence ne s'auto-implémente pas; elle requiert une gouvernance active pour maintenir sa cohérence à travers le temps et les équipes.

**L'équipe de plateforme** possède et opère l'Integration Backbone (Kafka, Schema Registry), les composants partagés du System of Engagement (API Gateway) et les outils d'observabilité. Elle définit les standards techniques, fournit les templates de services et accompagne les équipes produit dans l'adoption des patrons.

**Les équipes de domaine** possèdent leurs Systems of Record respectifs et les services applicatifs associés. Elles sont responsables de la qualité des événements qu'elles émettent, du respect des contrats définis et de l'implémentation correcte des patrons dans leur périmètre.

**L'architecte d'entreprise** ou le comité d'architecture arbitre les décisions transverses, valide les dérogations aux standards et assure la cohérence globale de l'écosystème. Il maintient le catalogue des patrons approuvés et les critères de sélection.

> **Perspective stratégique**
> La gouvernance architecturale efficace est habilitante, non bloquante. Elle fournit des « golden paths » — des chemins balisés où les bonnes pratiques sont pré-implémentées — plutôt que des processus d'approbation lourds. L'objectif est que suivre l'architecture de référence soit plus facile que la contourner.

### 9.4.2 Évolution et Dérogations

L'architecture de référence n'est pas figée; elle évolue avec les besoins et les technologies. Un processus explicite gère cette évolution.

**Les ADR (Architecture Decision Records)** documentent chaque décision architecturale significative : contexte, options considérées, décision retenue et justification. Ces documents forment la mémoire institutionnelle de l'architecture et facilitent l'intégration des nouveaux membres.

**Les dérogations** aux standards sont parfois nécessaires pour des cas exceptionnels. Elles doivent être explicitement demandées, évaluées par l'architecte responsable et documentées avec leur justification et leur date d'expiration prévue. Une dérogation sans date d'expiration devient un standard de fait.

**Les revues d'architecture** périodiques évaluent l'adéquation de l'architecture de référence aux besoins actuels. Les patrons obsolètes sont dépréciés; les nouveaux patrons validés par l'expérience sont intégrés au catalogue.

### 9.4.3 Métriques de Conformité

La conformité à l'architecture de référence se mesure par des indicateurs concrets.

| Métrique                    | Description                                      | Cible                    |
| ---------------------------- | ------------------------------------------------ | ------------------------ |
| Couverture CDC               | % des Systems of Record avec CDC actif           | > 90%                    |
| Taux de schémas versionnés | % des événements avec schéma dans le Registry | 100%                     |
| Ratio sync/async             | Proportion d'appels synchrones vs événements   | À définir par contexte |
| Couplage inter-services      | Nombre de dépendances directes par service      | < 5                      |
| Temps moyen d'intégration   | Durée pour intégrer un nouveau service         | < 2 semaines             |

Ces métriques alimentent les tableaux de bord de gouvernance et identifient les zones nécessitant une attention particulière.

---

## Conclusion et Transition

Ce chapitre a accompli la synthèse des trois domaines d'intégration en une architecture de référence convergente. Les principes des Reactive Systems et l'approche Inside-Out fournissent le cadre conceptuel. La vue logique en trois couches — System of Record, Integration Backbone, System of Engagement — structure l'implémentation. Les règles d'or garantissent la cohérence. Les matrices de décision guident les choix contextuels.

**Points clés à retenir** :

L'architecture convergente ne choisit pas entre synchrone et asynchrone, entre couplage et découplage; elle les combine selon les besoins de chaque interaction. Les règles d'or — jamais d'écriture directe inter-service, tout changement émet un événement, requêtes synchrones pour les lectures et événements pour les écritures, idempotence par défaut, contrats explicites à chaque frontière — constituent le socle non négociable. Les critères de décision — latence acceptable, cohérence requise, volume de données, couplage toléré — permettent de naviguer le continuum d'intégration avec discernement.

Cette architecture de référence n'est pas une destination finale mais un point de départ. Chaque organisation l'adaptera à son contexte : sa maturité technique, ses contraintes legacy, ses exigences réglementaires, sa culture organisationnelle. L'essentiel est de disposer d'un cadre cohérent pour guider les décisions plutôt que de les prendre au cas par cas sans vision d'ensemble.

Le chapitre suivant démontrera l'application concrète de cette architecture sur un scénario métier complet : le processus Order-to-Cash omnicanal. De la prise de commande à l'encaissement, en passant par la gestion des stocks et l'expédition, nous verrons comment les patrons des chapitres III à V s'assemblent pour former un système cohérent, réactif et résilient. Cette étude de cas validera empiriquement les principes énoncés ici et fournira un modèle réutilisable pour d'autres processus métier.

Au-delà de l'étude de cas, cette architecture de référence prépare le terrain pour l'évolution vers l'Entreprise Agentique (chapitre XI). Les agents IA autonomes qui orchestreront demain les flux d'intégration ne pourront le faire que sur une infrastructure mature : APIs bien documentées pour l'invocation, événements structurés pour la réactivité, données gouvernées pour le contexte décisionnel. L'architecture convergente n'est pas seulement une réponse aux défis actuels; elle constitue le substrat sur lequel l'intelligence artificielle pourra opérer.

---

## Résumé du Chapitre

**Thème central** : La convergence des trois domaines d'intégration (App, Data, Event) en une architecture de référence unifiée, guidée par les principes des Reactive Systems et l'approche Inside-Out.

**Principes de convergence** :

* Les Reactive Systems établissent quatre propriétés fondamentales : réactivité, résilience, élasticité et orientation message
* L'approche Inside-Out (Database Unbundling) place le log distribué au centre de l'architecture, permettant des projections optimisées par cas d'usage
* Les trois domaines ne sont pas des alternatives mais des couches complémentaires d'une même architecture

**Blueprint d'architecture** :

| Couche               | Rôle                         | Composants typiques                 |
| -------------------- | ----------------------------- | ----------------------------------- |
| System of Engagement | Points d'interaction externes | API Gateway, BFF, Portails          |
| Integration Backbone | Communication asynchrone      | Kafka, Schema Registry, Event Mesh  |
| System of Record     | Persistance et intégrité    | Bases de données par domaine + CDC |

**Règles d'or** :

1. Jamais d'écriture directe inter-service
2. Tout changement d'état émet un événement
3. Requêtes synchrones pour les lectures, événements pour les écritures
4. Idempotence par défaut
5. Contrat explicite à chaque frontière

**Critères de décision** :

* Latence acceptable : détermine le style synchrone vs asynchrone
* Cohérence requise : détermine les compromis CAP
* Volume de données : influence le choix streaming vs batch
* Couplage toléré : guide le niveau d'abstraction des interfaces

**Anti-patrons à éviter** : Distributed Monolith, Event Soup, Synchronous Everything, Async Everything, Shared Database Integration.

**Position dans le continuum** : Ce chapitre synthétise les chapitres III-VIII en une vision unifiée, préparant l'application pratique (chapitre X) et l'évolution agentique (chapitre XI).

---

*Références du chapitre : Reactive Manifesto (Bonér et al., 2014), Designing Data-Intensive Applications (Kleppmann, 2017), Building Event-Driven Microservices (Stopford, 2020), Confluent Documentation (2024-2025)*
