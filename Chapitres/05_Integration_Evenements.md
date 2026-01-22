
Je vais rédiger le Chapitre V sur l'Intégration des Événements. Ce chapitre représente le troisième volet du continuum d'intégration (App → Data →  **Event** ) et couvre les patrons de découplage maximal.

# Chapitre V — Intégration des Événements (Le Signal)

*Focus : La réactivité, le découplage temporel maximal et l'autonomie des consommateurs.*

---

## 5.1 Le Signal : Troisième Pilier du Continuum d'Intégration

Les chapitres précédents ont exploré l'intégration des applications (le Verbe) et l'intégration des données (le Nom). Avec l'intégration des événements, nous atteignons le troisième et dernier pilier du continuum : le Signal. Si le Verbe orchestre les actions synchrones et si le Nom structure l'état partagé, le Signal incarne la communication asynchrone qui libère les systèmes de leurs dépendances temporelles et spatiales.

Cette métaphore du Signal n'est pas anodine. Un signal, par nature, est émis sans garantie de réception immédiate. L'émetteur poursuit son activité sans attendre de confirmation synchrone. Le récepteur, de son côté, traite le signal selon sa propre cadence et ses propres priorités. Cette asymétrie fondamentale constitue la force et la complexité de l'architecture événementielle.

### 5.1.1 Du Couplage Fort au Découplage Maximal

Le chapitre II a établi le concept de couplage spatio-temporel comme contrainte fondamentale de l'intégration. L'intégration des applications impose un couplage fort : l'appelant attend une réponse synchrone, créant une dépendance temporelle directe. L'intégration des données réduit ce couplage en partageant l'état plutôt que l'action, mais maintient souvent une dépendance sur la fraîcheur et la localisation des données.

L'intégration des événements rompt ces deux dépendances. Le producteur d'un événement ne connaît pas ses consommateurs. Il publie un fait accompli sans se soucier de qui l'exploitera, quand ni comment. Les consommateurs, inversement, s'abonnent à des flux d'événements sans dépendre de la disponibilité instantanée du producteur. Cette double ignorance mutuelle constitue le découplage maximal.

> **Définition formelle**
> **Intégration des événements** : Patron d'architecture où les systèmes communiquent par l'émission et la consommation de messages représentant des faits métier (événements), sans couplage synchrone entre producteurs et consommateurs.

Ce découplage n'est pas une fin en soi. Il répond à des contraintes concrètes que rencontrent les architectes d'entreprise : la nécessité de maintenir la disponibilité malgré les pannes partielles, le besoin de traiter des volumes massifs de transactions, l'exigence d'ajouter de nouveaux consommateurs sans modifier les producteurs existants, ou encore l'impératif de construire des systèmes résilients face à l'incertitude.

### 5.1.2 Enjeux Spécifiques de l'Architecture Événementielle

L'adoption d'une architecture événementielle introduit des défis que les modèles synchrones n'exposent pas aussi directement. Ces enjeux structurent l'ensemble des patrons présentés dans ce chapitre.

**L'asynchronisme et ses conséquences.** Lorsqu'un producteur émet un événement, il ne reçoit pas de confirmation que cet événement a été traité correctement. Le succès de l'émission ne garantit pas le succès du traitement. Cette incertitude fondamentale impose des mécanismes de compensation, de retry et de réconciliation que les architectures synchrones gèrent implicitement via les codes de retour.

**L'ordre des messages.** Dans un système distribué, deux événements émis dans un certain ordre peuvent être reçus dans l'ordre inverse. Pire, un même événement peut être reçu plusieurs fois. La gestion de l'ordre et de la déduplication devient une responsabilité explicite de l'architecture, là où les appels synchrones séquentiels garantissent naturellement l'ordonnancement.

**L'idempotence comme exigence.** La garantie de livraison « au moins une fois » (at-least-once), privilégiée pour sa fiabilité, implique que les consommateurs doivent pouvoir traiter le même événement plusieurs fois sans effet de bord indésirable. Cette contrainte d'idempotence irrigue la conception de chaque consommateur.

**La traçabilité distribuée.** Suivre le parcours d'une requête à travers une chaîne de microservices synchrones reste relativement simple : chaque appel attend une réponse. Dans une architecture événementielle, un événement initial peut déclencher une cascade de traitements asynchrones, rendant la corrélation et le débogage nettement plus complexes.

**Le volume et la rétention.** Les architectures événementielles génèrent naturellement un volume de messages supérieur aux architectures synchrones. Un système qui émet un événement pour chaque changement d'état produit potentiellement des millions de messages quotidiens. La gestion de ce volume, sa rétention pour audit ou replay, et son coût de stockage deviennent des préoccupations architecturales de premier plan.

### 5.1.3 Anatomie d'un Événement

Avant d'explorer les patrons, il convient de définir précisément ce qu'est un événement dans le contexte de l'intégration d'entreprise. Un événement est un enregistrement immuable d'un fait qui s'est produit dans le passé. Cette définition, apparemment simple, contient trois caractéristiques essentielles.

L'événement est  **immuable** . Une fois émis, il ne peut être modifié. On peut émettre un événement correctif ou annulatif, mais l'événement original demeure dans le flux. Cette immuabilité garantit l'intégrité de l'historique et permet le replay.

L'événement représente un  **fait passé** . Il décrit quelque chose qui s'est déjà produit, non une intention ou une commande. « Commande créée » est un événement ; « Créer une commande » est une commande. Cette distinction, parfois subtile, influence profondément la conception des systèmes.

L'événement possède une  **structure définie** . Au minimum, un événement contient un type (ce qui s'est passé), un horodatage (quand), une source (où), et un payload (les données associées). Le standard CloudEvents, présenté au chapitre VI, formalise cette structure.

| Composant | Description             | Exemple                                        |
| --------- | ----------------------- | ---------------------------------------------- |
| Type      | Nature de l'événement | `order.created`                              |
| Source    | Système émetteur      | `/services/order-service`                    |
| ID        | Identifiant unique      | `evt-2025-01-15-abc123`                      |
| Time      | Horodatage ISO 8601     | `2025-01-15T14:30:00Z`                       |
| Data      | Payload métier         | `{ "orderId": "ORD-456", "amount": 150.00 }` |

Cette structure standardisée facilite l'interopérabilité entre systèmes hétérogènes et permet aux outils d'observabilité de traiter uniformément les événements de différentes sources.

---

## 5.2 Fondamentaux de l'Architecture Événementielle

### 5.2.1 Topologies de Communication

L'architecture événementielle s'appuie sur des topologies de communication qui diffèrent fondamentalement du modèle requête-réponse. Comprendre ces topologies permet de choisir le patron approprié pour chaque cas d'usage.

**La file d'attente point-à-point.** Dans ce modèle, un message émis par un producteur est consommé par exactement un consommateur. Si plusieurs consommateurs écoutent la même file, ils se partagent les messages selon un mécanisme de compétition. Ce modèle convient aux tâches de traitement distribué où chaque message doit être traité une seule fois.

**Le modèle publication-abonnement.** Ici, un message publié sur un sujet (topic) est diffusé à tous les abonnés de ce sujet. Chaque consommateur reçoit une copie indépendante du message. Ce modèle permet le découplage M:N où M producteurs émettent vers N consommateurs sans se connaître mutuellement.

**Le flux persistant (log).** Apache Kafka a popularisé ce modèle où les messages sont écrits dans un journal (log) ordonné et persistant. Les consommateurs lisent ce journal à leur propre rythme, maintenant un pointeur (offset) sur leur position. Contrairement aux files traditionnelles, les messages ne sont pas supprimés après consommation mais conservés selon une politique de rétention.

> **Note technique**
> La distinction entre file d'attente et log persistant a des implications profondes sur l'architecture. Une file garantit le traitement unique mais perd l'historique. Un log préserve l'historique mais nécessite une gestion explicite des offsets et de l'idempotence côté consommateur.

### 5.2.2 Garanties de Livraison

Les systèmes de messagerie offrent différentes garanties de livraison, chacune impliquant des compromis entre fiabilité, performance et complexité.

**Au plus une fois (at-most-once).** Le message est envoyé sans confirmation. En cas de panne, il peut être perdu. Cette garantie offre les meilleures performances mais convient uniquement aux données non critiques comme la télémétrie ou les métriques.

**Au moins une fois (at-least-once).** Le message est reémis jusqu'à confirmation de réception. En cas de panne pendant le traitement, le message peut être reçu plusieurs fois. Cette garantie, la plus courante, exige des consommateurs idempotents.

**Exactement une fois (exactly-once).** Le message est traité exactement une fois, sans perte ni duplication. Cette garantie, longtemps considérée comme impossible dans un système distribué, est aujourd'hui proposée par certaines plateformes comme Kafka via des transactions. Elle impose cependant une surcharge significative et des contraintes d'architecture.

| Garantie      | Risque de perte | Risque de duplication | Complexité | Cas d'usage                           |
| ------------- | --------------- | --------------------- | ----------- | ------------------------------------- |
| At-most-once  | Oui             | Non                   | Faible      | Métriques, logs non critiques        |
| At-least-once | Non             | Oui                   | Moyenne     | Transactions métier avec idempotence |
| Exactly-once  | Non             | Non                   | Élevée    | Transactions financières critiques   |

### 5.2.3 Partitionnement et Ordre

Dans les systèmes à haut débit, les flux d'événements sont partitionnés pour permettre le traitement parallèle. Chaque partition constitue un sous-flux ordonné traité par un consommateur dédié.

Le partitionnement introduit un compromis fondamental : l'ordre n'est garanti qu'au sein d'une partition. Deux événements portant sur des clés de partition différentes peuvent être traités dans un ordre quelconque. L'architecte doit donc choisir judicieusement la clé de partitionnement pour que les événements devant être ordonnés se retrouvent dans la même partition.

Prenons l'exemple d'un système de gestion de commandes. Si la clé de partition est l'identifiant de la commande, tous les événements d'une même commande (création, paiement, expédition) seront ordonnés. Mais les événements de commandes différentes pourront être traités en parallèle et dans un ordre indéterminé.

> **Bonnes pratiques**
> Choisir comme clé de partitionnement l'identifiant de l'agrégat métier (commande, client, compte) dont l'ordre des événements doit être préservé. Éviter les clés à cardinalité trop faible (risque de partition déséquilibrée) ou trop élevée (ordre perdu entre événements liés).

---

## 5.3 Catalogue des Patrons d'Intégration Événementielle

### 5.3.1 Publish/Subscribe

**Définition.** Le patron Publish/Subscribe (Pub/Sub) établit un découplage fondamental entre les producteurs et les consommateurs de messages. Les producteurs publient des événements sur des sujets (topics) sans connaître les abonnés. Les consommateurs s'abonnent aux sujets qui les intéressent sans connaître les producteurs.

**Problème résolu.** Dans une architecture traditionnelle point-à-point, chaque nouveau consommateur nécessite une modification du producteur pour ajouter une destination. Cette dépendance bidirectionnelle freine l'évolution du système et crée un couplage structurel fort. Le Pub/Sub élimine ce couplage en introduisant un intermédiaire (broker) qui gère les abonnements.

**Mécanisme.** Le producteur émet un message sur un topic identifié par un nom logique. Le broker reçoit ce message et le distribue à tous les consommateurs abonnés à ce topic. Chaque abonné reçoit une copie indépendante du message et le traite selon sa propre logique.

```
┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│  Producteur  │────▶│  Topic: Orders  │────▶│ Consommateur │
│   Order-Svc  │     │                 │     │  Inventory   │
└──────────────┘     │   [Message 1]   │     └──────────────┘
                     │   [Message 2]   │
                     │       ...       │────▶┌──────────────┐
                     └─────────────────┘     │ Consommateur │
                                             │   Billing    │
                                             └──────────────┘
```

**Avantages.** Le découplage spatial permet d'ajouter de nouveaux consommateurs sans modifier le producteur. Le découplage temporel permet aux consommateurs de traiter les messages à leur propre rythme. La scalabilité horizontale est facilitée par la possibilité de multiplier les instances de consommateurs.

**Inconvénients.** La visibilité de bout en bout est réduite : le producteur ne sait pas si son message a été traité ni par qui. Le débogage des flux complexes devient plus difficile. La prolifération de topics peut créer une complexité de gouvernance.

> **Quand utiliser ce patron**
> *Contexte* : Plusieurs systèmes doivent réagir au même événement métier, ou le nombre de consommateurs est amené à évoluer.
> *Alternatives* : Appel synchrone si un seul consommateur et feedback immédiat requis ; file point-à-point si exactement un consommateur doit traiter chaque message.

**Exemple d'usage.** Un système de commerce électronique publie un événement `order.placed` sur un topic. Le service d'inventaire s'abonne pour réserver les produits. Le service de facturation s'abonne pour générer la facture. Le service de notification s'abonne pour envoyer un courriel de confirmation. Le service d'analytique s'abonne pour mettre à jour les tableaux de bord. Chaque service évolue indépendamment, et de nouveaux services peuvent s'abonner sans modification du service de commande.

### 5.3.2 Event Sourcing

**Définition.** L'Event Sourcing persiste l'état d'un agrégat métier non pas comme un instantané courant, mais comme une séquence ordonnée et immuable de tous les événements qui ont modifié cet état depuis sa création.

**Problème résolu.** Les bases de données traditionnelles stockent l'état courant et écrasent les valeurs précédentes lors de chaque mise à jour. Cette approche perd l'historique des changements, rendant difficile l'audit, le débogage et la compréhension de l'évolution du système. Elle empêche également de reconstruire l'état à un instant passé ou de projeter l'état selon différentes perspectives.

**Mécanisme.** Chaque modification de l'agrégat génère un événement qui est appendé à un journal (event store). L'état courant est obtenu en rejouant (replay) tous les événements depuis l'origine ou depuis un snapshot périodique. Les consommateurs peuvent construire leurs propres projections en interprétant les mêmes événements selon leurs besoins.

```
Event Store pour Commande #123:
┌────────────────────────────────────────────────────────────┐
│ Seq 1: OrderCreated { customerId: "C456", items: [...] }   │
│ Seq 2: ItemAdded { productId: "P789", quantity: 2 }        │
│ Seq 3: PaymentReceived { amount: 150.00, method: "CC" }    │
│ Seq 4: OrderShipped { trackingNumber: "TRK-001" }          │
└────────────────────────────────────────────────────────────┘
État courant = Replay(Seq 1 → Seq 4)
```

**Avantages.** L'auditabilité est totale : chaque changement est tracé avec son horodatage et son contexte. La reconstruction temporelle permet de répondre à des questions comme « quel était l'état de cette commande hier à 14h ? ». Les projections multiples permettent à différents consommateurs de construire des vues optimisées pour leurs cas d'usage. La correction d'erreurs devient possible via des événements compensatoires sans perte d'historique.

**Inconvénients.** La complexité de mise en œuvre est significative : le modèle mental diffère des approches CRUD traditionnelles. Les requêtes sur l'état courant nécessitent soit un replay (coûteux), soit des projections maintenues (complexité supplémentaire). L'évolution des schémas d'événements pose des défis de compatibilité. Le volume de stockage croît indéfiniment avec le nombre d'événements.

> **Anti-patron**
> **Event Sourcing partiel** : Utiliser l'Event Sourcing pour certaines entités et le CRUD pour d'autres dans le même agrégat crée une incohérence architecturale et complique la reconstruction de l'état. Si l'Event Sourcing est adopté, il doit l'être pour l'ensemble de l'agrégat.

**Exemple d'usage.** Un système bancaire utilise l'Event Sourcing pour les comptes clients. Chaque opération (dépôt, retrait, transfert, intérêts) génère un événement. L'état du compte est la somme de tous ces événements. L'auditeur peut retracer chaque centime depuis l'ouverture du compte. Le système de détection de fraude peut analyser les patterns d'événements. Le service fiscal peut projeter les événements selon les règles comptables en vigueur à chaque période.

### 5.3.3 Saga Pattern

**Définition.** Le Saga Pattern coordonne des transactions distribuées longue durée en les décomposant en une séquence de transactions locales, chacune émettant un événement déclenchant la suivante. En cas d'échec, des transactions compensatoires sont exécutées pour annuler les effets des transactions déjà complétées.

**Problème résolu.** Les transactions distribuées traditionnelles (2PC, Two-Phase Commit) imposent un verrouillage des ressources pendant toute la durée de la transaction, créant des goulots d'étranglement et des risques de blocage. Elles nécessitent également que tous les participants soient disponibles simultanément, ce qui contredit le principe de découplage des architectures événementielles.

**Mécanisme.** Une saga se compose de transactions locales T1, T2, ... Tn, chacune associée à une compensation C1, C2, ... Cn. Chaque transaction émet un événement de succès déclenchant la suivante, ou un événement d'échec déclenchant les compensations dans l'ordre inverse.

Deux styles de coordination existent :

**Chorégraphie** : Chaque service écoute les événements et décide de sa propre action. Aucun coordinateur central. Simple mais difficile à suivre pour les flux complexes.

**Orchestration** : Un orchestrateur central dirige le flux, appelant chaque service et gérant les compensations. Plus visible mais crée un point de couplage.

```
Saga "Réservation Voyage" (Chorégraphie):

┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Flight    │───▶│    Hotel    │───▶│     Car     │
│  Reserved   │    │  Reserved   │    │  Reserved   │
└─────────────┘    └─────────────┘    └─────────────┘
      │                  │                  │
      ▼                  ▼                  ▼
  Si échec:         Si échec:         Si échec:
  CancelFlight      CancelHotel       CancelCar
                         │
                         ▼
                    CancelFlight
```

**Avantages.** La cohérence est maintenue sans verrouillage distribué. Chaque service reste autonome et disponible. Les transactions longue durée (heures, jours) sont possibles. L'architecture supporte naturellement les pannes partielles.

**Inconvénients.** La cohérence est éventuelle, non immédiate. La logique de compensation doit être conçue explicitement pour chaque étape. Le suivi du flux global nécessite une corrélation d'événements. Les compensations peuvent elles-mêmes échouer, nécessitant une gestion d'exception.

> **Note technique**
> La compensation n'est pas une annulation parfaite. Si un courriel de confirmation a été envoyé, on ne peut pas le « désavoyer ». La compensation consiste alors à envoyer un second courriel d'annulation. L'architecte doit anticiper ces cas et concevoir des compensations sémantiquement correctes.

> **Quand utiliser ce patron**
> *Contexte* : Transaction impliquant plusieurs services avec exigence de cohérence éventuelle ; durée de transaction pouvant excéder quelques secondes.
> *Alternatives* : Transaction locale si un seul service ; 2PC si cohérence forte requise et tous les participants supportent le protocole.

**Exemple d'usage.** Un processus Order-to-Cash utilise une saga orchestrée : (1) Créer la commande, (2) Réserver l'inventaire, (3) Débiter le paiement, (4) Planifier l'expédition. Si l'étape 3 échoue, l'orchestrateur déclenche la libération de l'inventaire (compensation de l'étape 2). La commande passe en état « échec paiement » plutôt qu'être supprimée, préservant la trace pour analyse.

### 5.3.4 Transactional Outbox

**Définition.** Le patron Transactional Outbox garantit l'atomicité entre la mise à jour d'une base de données et la publication d'un événement en écrivant l'événement dans une table « outbox » de la même base de données, au sein de la même transaction locale.

**Problème résolu.** Considérons un service qui doit persister une commande et publier un événement `OrderCreated`. Deux approches naïves échouent : (1) persister puis publier risque de publier sans avoir persisté en cas de panne entre les deux ; (2) publier puis persister risque de persister sans avoir publié. Sans transaction distribuée entre la base de données et le broker de messages, l'atomicité semble impossible.

**Mécanisme.** Le service écrit l'entité métier et l'événement dans la même base de données, au sein d'une transaction ACID locale. Un processus séparé (poller ou log-based via CDC) lit périodiquement la table outbox et publie les événements vers le broker. Une fois la publication confirmée, l'entrée outbox est marquée comme traitée ou supprimée.

```
Transaction locale:
┌─────────────────────────────────────────────────────────────┐
│ BEGIN TRANSACTION                                           │
│   INSERT INTO orders (id, customer, items) VALUES (...)     │
│   INSERT INTO outbox (event_type, payload) VALUES           │
│          ('OrderCreated', '{"orderId": "123", ...}')        │
│ COMMIT                                                      │
└─────────────────────────────────────────────────────────────┘

Processus de relay (asynchrone):
┌─────────────────────────────────────────────────────────────┐
│ SELECT * FROM outbox WHERE status = 'PENDING'               │
│ PUBLISH message TO broker                                   │
│ UPDATE outbox SET status = 'SENT' WHERE id = ...            │
└─────────────────────────────────────────────────────────────┘
```

**Avantages.** L'atomicité est garantie par la transaction locale de la base de données. Aucune dépendance sur des transactions distribuées (XA) ni sur la disponibilité du broker au moment de l'écriture. Le patron fonctionne avec n'importe quelle base de données relationnelle.

**Inconvénients.** Un processus supplémentaire (relay) doit être déployé et surveillé. La latence de publication dépend de la fréquence de polling (ou de la latence CDC). L'ordre des messages entre différentes entités n'est pas garanti si le relay traite par lots.

> **Bonnes pratiques**
> Utiliser le Change Data Capture (CDC) plutôt que le polling pour le relay. Des outils comme Debezium capturent les insertions dans la table outbox en temps quasi réel via les logs de réplication de la base de données, réduisant la latence et la charge de polling.

**Exemple d'usage.** Le service de commande utilise PostgreSQL. Chaque création de commande insère une ligne dans la table `orders` et une ligne dans la table `outbox` contenant l'événement sérialisé. Debezium surveille la table outbox via le WAL de PostgreSQL et publie chaque nouvelle entrée vers un topic Kafka. Le service de commande n'a aucune dépendance directe sur Kafka ; il ne connaît que sa base de données locale.

### 5.3.5 Event-Carried State Transfer (ECST)

**Définition.** L'Event-Carried State Transfer enrichit l'événement avec l'état complet (ou suffisant) pour que le consommateur puisse traiter l'événement sans rappeler le producteur. L'événement « transporte » l'état nécessaire.

**Problème résolu.** Un événement minimaliste comme `{ "event": "OrderCreated", "orderId": "123" }` oblige le consommateur à rappeler le service de commande pour obtenir les détails de la commande. Ce rappel crée un couplage synchrone, annulant partiellement les bénéfices du découplage événementiel. Si le service de commande est indisponible, le consommateur est bloqué.

**Mécanisme.** L'événement inclut toutes les données dont les consommateurs ont besoin pour leur traitement. Au lieu de publier uniquement l'identifiant, le producteur publie l'entité complète ou un sous-ensemble pertinent.

```
Événement minimaliste (anti-pattern pour ECST):
{ "event": "OrderCreated", "orderId": "123" }

Événement enrichi (ECST):
{
  "event": "OrderCreated",
  "orderId": "123",
  "customer": { "id": "C456", "name": "Alice Martin", "tier": "Gold" },
  "items": [
    { "productId": "P789", "name": "Widget", "quantity": 2, "price": 50.00 }
  ],
  "total": 100.00,
  "currency": "CAD"
}
```

**Avantages.** L'autonomie du consommateur est maximale : il peut traiter l'événement sans aucune dépendance externe. La résilience est améliorée puisque les pannes du producteur n'affectent pas le traitement des événements déjà émis. Les performances sont meilleures car aucun appel synchrone supplémentaire n'est nécessaire.

**Inconvénients.** La taille des messages augmente, impactant le débit et le coût de stockage. La duplication de données entre l'état source et les événements crée un risque d'incohérence si l'état change après l'émission. Le couplage sémantique augmente : les consommateurs dépendent de la structure de l'événement.

> **Anti-patron**
> **Événement obèse** : Inclure systématiquement toutes les données de l'entité, y compris celles qu'aucun consommateur n'utilise, alourdit inutilement les messages. Analyser les besoins réels des consommateurs et n'inclure que les données nécessaires.

> **Quand utiliser ce patron**
> *Contexte* : Les consommateurs nécessitent fréquemment des données supplémentaires ; la latence des rappels synchrones est inacceptable ; les consommateurs doivent fonctionner en mode dégradé si le producteur est indisponible.
> *Alternatives* : Événement minimaliste + appel synchrone si la fraîcheur de l'état est critique et la latence acceptable.

**Exemple d'usage.** Le service d'analytique consomme les événements `OrderCreated` pour alimenter un entrepôt de données. Grâce à l'ECST, chaque événement contient les informations client, les détails produits et les montants. L'analytique peut traiter ces événements en lot, même si le service de commande est en maintenance. L'entrepôt reflète l'état au moment de la commande, ce qui est sémantiquement correct pour l'analyse historique.

### 5.3.6 Claim Check

**Définition.** Le patron Claim Check sépare le contenu volumineux d'un message de son enveloppe événementielle. Le payload lourd est stocké dans un système de stockage externe, et l'événement ne contient qu'une référence (le « claim check ») permettant de récupérer ce payload.

**Problème résolu.** Les brokers de messages ont des limites de taille par message (typiquement 1 Mo pour Kafka par défaut, configurable mais avec impact sur les performances). Certains événements métier incluent des pièces jointes volumineuses : documents PDF, images, fichiers de données. Transmettre ces fichiers directement via le broker sature le réseau et dégrade les performances pour tous les messages.

**Mécanisme.** Le producteur stocke le contenu volumineux dans un système de stockage (S3, Azure Blob, système de fichiers distribué) et obtient une référence unique. L'événement publié contient cette référence plutôt que le contenu. Le consommateur extrait la référence, récupère le contenu depuis le stockage, et procède au traitement.

```
Producteur:
┌─────────────────────────────────────────────────────────────┐
│ 1. Upload document vers S3 → obtient URL signée            │
│ 2. Publie événement avec référence:                        │
│    { "event": "DocumentUploaded",                          │
│      "claimCheck": "s3://bucket/docs/invoice-123.pdf",     │
│      "metadata": { "size": 2500000, "type": "PDF" } }      │
└─────────────────────────────────────────────────────────────┘

Consommateur:
┌─────────────────────────────────────────────────────────────┐
│ 1. Reçoit événement                                         │
│ 2. Télécharge document depuis S3 via claimCheck            │
│ 3. Traite le document                                       │
└─────────────────────────────────────────────────────────────┘
```

**Avantages.** Les messages restent légers, préservant les performances du broker. Les payloads volumineux peuvent être stockés dans des systèmes optimisés pour les fichiers. Le même contenu peut être référencé par plusieurs événements sans duplication.

**Inconvénients.** Le consommateur doit accéder à deux systèmes (broker et stockage), augmentant les points de défaillance. La gestion du cycle de vie du contenu stocké (rétention, suppression) doit être coordonnée avec les événements. La latence de traitement augmente du temps de téléchargement.

> **Note technique**
> Utiliser des URLs signées avec expiration pour sécuriser l'accès au stockage. La durée de validité doit excéder le temps de rétention des messages dans le broker pour garantir que les consommateurs en retard puissent encore accéder au contenu.

**Exemple d'usage.** Un système de gestion documentaire reçoit des factures numérisées. Le service d'ingestion stocke le PDF dans Azure Blob Storage et publie un événement `InvoiceReceived` contenant l'URL du blob. Le service de reconnaissance optique de caractères (OCR) consomme cet événement, télécharge le PDF, extrait le texte, et publie un événement `InvoiceProcessed` avec les données structurées extraites.

### 5.3.7 Idempotent Consumer

**Définition.** Un consommateur idempotent peut traiter le même message plusieurs fois sans produire d'effets de bord indésirables au-delà du premier traitement. Le résultat est identique que le message soit reçu une fois ou N fois.

**Problème résolu.** La garantie « au moins une fois » (at-least-once), standard dans les systèmes de messagerie fiables, implique que les messages peuvent être livrés plusieurs fois. Un consommateur non idempotent exécuterait la logique métier à chaque réception, potentiellement débitant un compte plusieurs fois ou envoyant plusieurs courriels de confirmation pour la même commande.

**Mécanisme.** Plusieurs stratégies permettent d'implémenter l'idempotence :

**Détection par identifiant.** Chaque message porte un identifiant unique. Le consommateur maintient un registre des identifiants déjà traités. À la réception, il vérifie si l'identifiant est connu : si oui, il ignore le message ; sinon, il le traite et enregistre l'identifiant.

**Opérations naturellement idempotentes.** Certaines opérations sont idempotentes par nature. « Définir le statut à SHIPPED » est idempotent : l'exécuter plusieurs fois produit le même résultat. « Incrémenter le compteur de 1 » ne l'est pas.

**Versionnement optimiste.** L'entité porte un numéro de version. Le message spécifie la version attendue. Si la version courante diffère, le message est un doublon ou obsolète et est ignoré.

```
Registre d'idempotence:
┌─────────────────────────────────────────────────────────────┐
│ message_id          │ processed_at         │ status        │
├─────────────────────────────────────────────────────────────┤
│ evt-2025-01-001     │ 2025-01-15 14:30:00  │ SUCCESS       │
│ evt-2025-01-002     │ 2025-01-15 14:30:05  │ SUCCESS       │
│ evt-2025-01-003     │ 2025-01-15 14:30:10  │ FAILED        │
└─────────────────────────────────────────────────────────────┘

Logique du consommateur:
if exists(message.id in idempotency_registry):
    log("Duplicate message, skipping")
    acknowledge(message)
else:
    process(message)
    insert into idempotency_registry(message.id)
    acknowledge(message)
```

**Avantages.** La fiabilité du système est garantie malgré les livraisons multiples. Le consommateur peut être redémarré à tout moment sans risque d'effets de bord. La récupération après panne est simplifiée.

**Inconvénients.** Le registre d'idempotence consomme du stockage et nécessite une gestion de rétention. La vérification ajoute une latence (requête supplémentaire). Les opérations intrinsèquement non idempotentes nécessitent une refonte.

> **Bonnes pratiques**
> Stocker le registre d'idempotence dans la même base de données que l'état métier et dans la même transaction. Cela garantit que si le traitement réussit, l'identifiant est enregistré, et inversement. Utiliser un TTL (Time To Live) pour purger les entrées anciennes et limiter la croissance du registre.

**Exemple d'usage.** Le service de facturation consomme les événements `PaymentReceived`. Chaque événement contient un `paymentId` unique. Avant de créditer le compte du marchand, le service vérifie si ce `paymentId` a déjà été traité. Si oui, il acquitte le message sans action. Cette protection évite les doubles crédits en cas de retry du broker ou de retraitement après panne.

### 5.3.8 Dead Letter Queue (DLQ)

**Définition.** La Dead Letter Queue est une file de destination pour les messages qui n'ont pas pu être traités avec succès après un nombre défini de tentatives. Elle isole les messages en échec pour analyse et retraitement manuel sans bloquer le flux principal.

**Problème résolu.** Certains messages échouent de manière permanente : données invalides, dépendance externe défaillante, bogue applicatif. Sans DLQ, ces messages bloquent la file (si le consommateur refuse de les acquitter) ou sont perdus (si le consommateur les acquitte malgré l'échec). La DLQ offre une troisième voie : les isoler tout en préservant leur contenu.

**Mécanisme.** Le consommateur tente de traiter le message. En cas d'échec, il réessaie selon une politique de retry (typiquement avec backoff exponentiel). Après N tentatives infructueuses, le message est routé vers la DLQ plutôt que d'être acquitté ou de bloquer la file. Une équipe d'opérations surveille la DLQ, analyse les messages en échec, et les réinjecte après correction.

```
Flux normal:
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Topic     │────▶│ Consommateur│────▶│  Traitement │
│   Orders    │     │             │     │   Réussi    │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                    Si échec après N retries
                           │
                           ▼
                    ┌─────────────┐
                    │     DLQ     │
                    │   Orders    │
                    └─────────────┘
                           │
                    Analyse et retraitement manuel
```

**Avantages.** Le flux principal n'est pas bloqué par les messages problématiques. Les messages en échec sont préservés pour analyse. Les métriques sur la DLQ fournissent un indicateur de santé du système. Le retraitement est possible après correction du problème.

**Inconvénients.** La surveillance de la DLQ doit être mise en place (alertes, dashboards). Le retraitement manuel peut être fastidieux si le volume est élevé. L'ordre de traitement n'est plus garanti pour les messages réinjectés.

> **Note technique**
> Enrichir les messages envoyés en DLQ avec des métadonnées de diagnostic : nombre de tentatives, timestamps des échecs, messages d'erreur. Ces informations accélèrent l'analyse et la résolution.

> **Quand utiliser ce patron**
> *Contexte* : Systèmes en production avec des flux de messages critiques où les échecs occasionnels sont inévitables.
> *Alternatives* : Pour les systèmes de développement ou les flux non critiques, un simple log des erreurs peut suffire.

**Exemple d'usage.** Le service de notification consomme les événements `OrderShipped` pour envoyer des courriels aux clients. Certains courriels échouent : adresse invalide, quota SMTP dépassé, timeout. Après trois tentatives, le message est routé vers la DLQ. L'équipe d'opérations constate une accumulation et identifie que le serveur SMTP était temporairement indisponible. Après rétablissement, ils réinjectent les messages de la DLQ vers le topic original.

### 5.3.9 Competing Consumers

**Définition.** Le patron Competing Consumers distribue le traitement des messages d'une file entre plusieurs instances concurrentes du même consommateur. Chaque message est traité par exactement une instance, permettant la scalabilité horizontale.

**Problème résolu.** Un consommateur unique devient un goulot d'étranglement lorsque le volume de messages dépasse sa capacité de traitement. Ajouter des instances identiques permet de paralléliser le traitement, mais nécessite un mécanisme pour que chaque message ne soit traité qu'une fois.

**Mécanisme.** Plusieurs instances du consommateur s'abonnent à la même file ou au même groupe de consommateurs (consumer group). Le broker distribue les messages entre les instances selon un algorithme de partitionnement. Chaque instance traite un sous-ensemble des messages. Si une instance tombe, ses messages sont redistribués aux instances restantes.

```
              ┌─────────────────┐
              │  Topic: Orders  │
              │   Partition 0   │────▶ Consommateur Instance A
              │   Partition 1   │────▶ Consommateur Instance B
              │   Partition 2   │────▶ Consommateur Instance C
              └─────────────────┘
                      │
               Consumer Group "OrderProcessors"
```

**Avantages.** La scalabilité horizontale est linéaire jusqu'au nombre de partitions. La résilience est améliorée par la redistribution automatique en cas de panne d'une instance. L'élasticité permet d'ajuster le nombre d'instances selon la charge.

**Inconvénients.** L'ordre de traitement entre partitions n'est pas garanti. Le nombre maximum de consommateurs parallèles est limité par le nombre de partitions. Le rééquilibrage lors de l'ajout ou du retrait d'instances peut créer une latence temporaire.

> **Bonnes pratiques**
> Dimensionner le nombre de partitions en fonction du parallélisme maximal anticipé. Ajouter des partitions est possible mais complexe sur un topic existant. Prévoir une marge pour la croissance future.

**Exemple d'usage.** Le service de traitement des paiements consomme un topic avec 12 partitions. En période normale, 4 instances se partagent le traitement (3 partitions chacune). Lors du Black Friday, l'équipe ajoute 8 instances supplémentaires pour absorber le pic. Chaque instance traite alors une partition. Après le pic, le nombre d'instances est réduit automatiquement par l'orchestrateur Kubernetes.

---

## 5.4 Synthèse et Matrice de Décision

### 5.4.1 Relations entre les Patrons

Les patrons présentés ne sont pas mutuellement exclusifs. Au contraire, ils se combinent pour former des architectures événementielles robustes. Comprendre leurs relations permet de les assembler judicieusement.

**Pub/Sub comme fondation.** Le patron Publish/Subscribe constitue la base sur laquelle les autres patrons s'appuient. L'Event Sourcing publie ses événements via Pub/Sub. Les Sagas coordonnent leurs étapes par événements publiés. Le Transactional Outbox alimente un topic Pub/Sub.

**Transactional Outbox et Event Sourcing.** Ces deux patrons partagent l'idée de persister les événements dans la base de données avant publication. L'Event Sourcing va plus loin en faisant des événements la source de vérité, alors que le Transactional Outbox maintient un état traditionnel en parallèle.

**Idempotent Consumer comme prérequis.** Tout consommateur dans une architecture événementielle devrait être idempotent. Ce patron est transversal et s'applique aux consommateurs de Sagas, aux processeurs d'Event Sourcing, et à tout service recevant des événements.

**DLQ et Competing Consumers.** Ces patrons de résilience se combinent naturellement. Les Competing Consumers traitent en parallèle, et ceux qui échouent routent vers une DLQ partagée. La DLQ centralise la gestion des erreurs pour toutes les instances.

**ECST et Claim Check.** Ces patrons représentent deux approches opposées de la taille du payload. L'ECST enrichit le message pour l'autonomie du consommateur. Le Claim Check allège le message en externalisant le contenu. Le choix dépend de la taille des données et des contraintes de latence.

### 5.4.2 Matrice de Sélection des Patrons

Le tableau suivant guide le choix du patron selon le contexte et les contraintes.

| Patron                         | Problème Principal                        | Contraintes Clés                         | Quand Éviter                           |
| ------------------------------ | ------------------------------------------ | ----------------------------------------- | --------------------------------------- |
| **Pub/Sub**              | Découplage producteur-consommateur        | Tolérance à la latence de livraison     | Besoin de feedback synchrone immédiat  |
| **Event Sourcing**       | Auditabilité et reconstruction temporelle | Complexité de mise en œuvre acceptable  | CRUD simple sans exigence d'audit       |
| **Saga**                 | Transaction distribuée longue durée      | Cohérence éventuelle acceptable         | Transaction ACID locale suffisante      |
| **Transactional Outbox** | Atomicité DB + publication                | Base de données relationnelle disponible | Event Sourcing déjà en place          |
| **ECST**                 | Autonomie du consommateur                  | Taille du message acceptable              | Payload volumineux, fraîcheur critique |
| **Claim Check**          | Payload volumineux                         | Stockage externe disponible               | Petits messages, latence critique       |
| **Idempotent Consumer**  | Livraisons multiples                       | Toujours applicable                       | Jamais (patron systématique)           |
| **DLQ**                  | Messages en échec                         | Flux de production critique               | Environnement de développement simple  |
| **Competing Consumers**  | Scalabilité horizontale                   | Partitionnement adéquat                  | Ordre global strict requis              |

### 5.4.3 Critères de Décision

Au-delà du tableau, plusieurs critères transversaux influencent le choix des patrons.

**Latence acceptable.** Les architectures événementielles introduisent une latence inhérente. Si le cas d'usage exige une réponse en millisecondes, l'approche synchrone du chapitre III peut être plus appropriée. Les patrons événementiels conviennent aux processus tolérant des latences de secondes à minutes.

**Exigence de cohérence.** La cohérence éventuelle est le modèle par défaut des architectures événementielles. Si une cohérence forte et immédiate est requise (transactions financières critiques), des mécanismes supplémentaires (exactly-once, 2PC) ou une architecture hybride sont nécessaires.

**Volume de messages.** Un volume élevé (millions de messages par jour) justifie l'investissement dans des patrons comme Competing Consumers et Claim Check. Un volume modeste peut se contenter d'une architecture plus simple.

**Complexité organisationnelle.** L'Event Sourcing et le Saga Pattern nécessitent une montée en compétence des équipes. Leur adoption doit être accompagnée de formation et de changement de culture. Un déploiement progressif, commençant par des cas d'usage non critiques, réduit les risques.

> **Perspective stratégique**
> L'adoption de l'architecture événementielle n'est pas binaire. Une entreprise peut commencer par le Transactional Outbox pour fiabiliser la publication, ajouter le Pub/Sub pour découpler les consommateurs, puis introduire l'Event Sourcing sur les agrégats à forte exigence d'audit. Cette progression incrémentale réduit les risques et permet l'apprentissage organisationnel.

---

## 5.5 Résumé et Transition

### Récapitulatif du Chapitre

Ce chapitre a exploré l'intégration des événements, troisième pilier du continuum d'intégration après les applications (le Verbe) et les données (le Nom). Le Signal, métaphore de la communication asynchrone, permet le découplage maximal entre les systèmes.

Nous avons établi les fondamentaux de l'architecture événementielle : la distinction entre les topologies de communication (file, Pub/Sub, log), les garanties de livraison (at-most-once, at-least-once, exactly-once), et les mécanismes de partitionnement qui permettent le parallélisme tout en préservant l'ordre au sein des partitions.

Le catalogue de neuf patrons fournit une boîte à outils complète pour concevoir des systèmes événementiels robustes :

Le **Publish/Subscribe** établit le découplage fondamental M:N entre producteurs et consommateurs. L'**Event Sourcing** transforme les événements en source de vérité, offrant une auditabilité totale et la capacité de reconstruction temporelle. Le **Saga Pattern** coordonne les transactions distribuées sans verrouillage, avec des mécanismes de compensation en cas d'échec.

Le **Transactional Outbox** résout le problème d'atomicité entre la base de données et le broker de messages. L'**Event-Carried State Transfer** enrichit les événements pour maximiser l'autonomie des consommateurs. Le **Claim Check** gère les payloads volumineux en les externalisant.

L' **Idempotent Consumer** , patron transversal, garantit la robustesse face aux livraisons multiples. La **Dead Letter Queue** isole les messages en échec pour analyse sans bloquer le flux principal. Les **Competing Consumers** permettent la scalabilité horizontale du traitement.

### Points Clés à Retenir

L'architecture événementielle n'est pas une panacée. Elle introduit une complexité inhérente (asynchronisme, cohérence éventuelle, débogage distribué) qui doit être justifiée par les bénéfices attendus (découplage, scalabilité, résilience). L'adoption doit être progressive et accompagnée d'une montée en compétence des équipes.

Les patrons se combinent pour former des architectures cohérentes. Le Transactional Outbox alimente le Pub/Sub. Les Competing Consumers traitent en parallèle et routent vers une DLQ partagée. L'Idempotent Consumer est systématique pour tous les consommateurs.

Le choix du patron dépend du contexte : latence acceptable, exigence de cohérence, volume de messages, maturité organisationnelle. La matrice de décision présentée guide ces choix, mais le jugement de l'architecte reste essentiel pour adapter les patrons au contexte spécifique de l'entreprise.

### Transition vers le Chapitre VI

Avec les trois domaines d'intégration désormais couverts (Applications, Données, Événements), le chapitre VI aborde la question des  **standards et contrats d'interface** . Comment documenter ces APIs synchrones et ces flux événementiels ? Comment garantir l'interopérabilité entre systèmes hétérogènes ? Les spécifications OpenAPI, AsyncAPI et CloudEvents fournissent des réponses que nous explorerons en profondeur.

Le chapitre VI établira également les bases de l'interopérabilité sémantique avec JSON-LD et les ontologies, complétant ainsi le cadre technique nécessaire pour construire des architectures d'intégration matures et pérennes.

---

## Références du Chapitre

**Documentation technique**

Apache Kafka.  *Kafka Documentation* . https://kafka.apache.org/documentation/ — Documentation de référence pour les concepts de log distribué, partitionnement et consumer groups.

Confluent.  *Event Sourcing with Kafka* . 2024. — Guide pratique pour l'implémentation de l'Event Sourcing sur Kafka.

Microsoft.  *Cloud Design Patterns: Saga* . Azure Architecture Center, 2024. — Description détaillée du Saga Pattern avec exemples Azure.

**Ouvrages de référence**

Kleppmann, Martin.  *Designing Data-Intensive Applications* . O'Reilly, 2017. — Chapitres 11-12 sur les systèmes de streaming et le traitement événementiel.

Richardson, Chris.  *Microservices Patterns* . Manning, 2018. — Chapitres sur le Transactional Outbox, Saga et Event Sourcing dans le contexte des microservices.

Stopford, Ben.  *Designing Event-Driven Systems* . O'Reilly/Confluent, 2018. — Approche pratique de l'architecture événementielle avec Kafka.

---

*Chapitre suivant : VI — Standards et Contrats d'Interface*
