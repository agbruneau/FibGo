
# Chapitre IV — Intégration des Données (Le Nom)

*Focus : La cohérence de l'état, la gouvernance des structures et l'accessibilité de l'information.*

---

## Introduction

Le chapitre précédent a exploré l'intégration des applications sous l'angle du *Verbe* — l'orchestration des actions, les appels synchrones et la coordination des services. Cette perspective, bien qu'essentielle, ne capture qu'une facette de l'interopérabilité. Car derrière chaque action se trouve un  *état* , derrière chaque requête se cache une  *donnée* . L'intégration des données constitue le *Nom* de notre trilogie architecturale : elle désigne les entités, qualifie leur état et garantit leur accessibilité à travers l'écosystème.

Si l'intégration des applications répond à la question « que faire ? », l'intégration des données répond à « que savons-nous ? ». Cette distinction n'est pas qu'académique. Dans une architecture distribuée, la gestion de l'état représente le défi le plus redoutable. Le théorème CAP, présenté au chapitre II, nous rappelle que la cohérence parfaite et la disponibilité totale demeurent mutuellement exclusives en présence de partitions réseau. L'intégration des données navigue précisément dans cet espace de compromis, cherchant l'équilibre optimal entre fraîcheur, cohérence et performance.

Ce chapitre se situe au cœur du continuum d'intégration. Là où l'intégration des applications impose un couplage temporel fort (l'appelant attend une réponse), l'intégration des données introduit un découplage intermédiaire. Les systèmes ne s'interrogent plus directement; ils partagent, répliquent ou fédèrent leurs états. Ce glissement prépare naturellement le terrain pour l'intégration des événements (chapitre V), où le découplage atteindra son expression maximale.

Les enjeux contemporains amplifient l'importance de ce domaine. La prolifération des sources de données — applications patrimoniales, services infonuagiques, objets connectés, partenaires externes — crée un paysage fragmenté où la « vérité » devient plurielle. Les exigences réglementaires (RGPD, Loi 25 au Québec) imposent une traçabilité et une gouvernance rigoureuses. Parallèlement, les attentes métier en matière d'analytique temps réel et d'intelligence artificielle requièrent un accès fluide à des données fraîches et de qualité.

Ce chapitre présente sept patrons d'architecture essentiels pour relever ces défis. Du Change Data Capture (CDC), qui transforme les mutations de base de données en flux exploitables, au Data Mesh, qui décentralise la propriété des données par domaine métier, chaque patron répond à des problématiques spécifiques tout en s'inscrivant dans une vision cohérente de l'intégration. L'objectif n'est pas de prescrire une solution universelle, mais d'outiller l'architecte pour choisir judicieusement selon le contexte : latence acceptable, cohérence requise, volume de données et couplage toléré.

---

## 4.1 Enjeux de l'Intégration des Données

Avant d'explorer les patrons d'architecture, il convient de cartographier les défis fondamentaux que toute stratégie d'intégration des données doit affronter. Ces enjeux ne sont pas indépendants; ils forment un système de tensions où l'optimisation d'un axe affecte inévitablement les autres.

### 4.1.1 Fraîcheur des Données et Latence

La fraîcheur désigne l'écart temporel entre le moment où une donnée change dans son système source et le moment où cette modification devient visible aux systèmes consommateurs. Cette latence d'intégration peut varier de quelques millisecondes (quasi temps réel) à plusieurs heures, voire jours (traitement par lots).

Le choix du niveau de fraîcheur acceptable dépend du cas d'usage. Un tableau de bord de direction peut tolérer des données actualisées quotidiennement. Un système de détection de fraude exige une latence de l'ordre de la seconde. Un moteur de recommandation en ligne navigue entre ces extrêmes, optimisant le compromis entre pertinence et coût computationnel.

> **Perspective stratégique**
> La fraîcheur des données n'est pas un objectif en soi, mais un levier d'affaires. Chaque milliseconde de latence réduite a un coût (infrastructure, complexité, maintenance). L'architecte doit quantifier la valeur métier de la fraîcheur avant de choisir les mécanismes d'intégration.

Trois modes de synchronisation coexistent dans la pratique. Le mode par lots ( *batch* ) agrège les modifications sur une période définie et les transfère en bloc, minimisant les coûts de transfert mais maximisant la latence. Le mode micro-lots ( *micro-batch* ) réduit cette fenêtre à quelques minutes ou secondes, offrant un compromis intermédiaire. Le mode continu ( *streaming* ) capture et propage chaque modification individuellement, atteignant des latences sub-secondes au prix d'une infrastructure plus sophistiquée.

### 4.1.2 Vérité Unique versus Vérité Distribuée

L'idéal d'une « source unique de vérité » (*Single Source of Truth* — SSOT) imprègne la littérature architecturale depuis des décennies. Ce principe postule qu'une donnée donnée ne devrait exister qu'en un seul endroit, éliminant les risques de désynchronisation et les ambiguïtés sur la version « correcte ».

La réalité des systèmes distribués complexifie cette vision. Dans une architecture de microservices, chaque service possède sa propre base de données pour préserver son autonomie. Le service « Inventaire » détient la vérité sur les stocks, le service « Commandes » sur les transactions, le service « Clients » sur les profils. La vérité devient intrinsèquement distribuée, fragmentée par domaine.

Cette distribution n'est pas un défaut à corriger mais une caractéristique à orchestrer. Le défi consiste à définir clairement les périmètres de responsabilité (quel système fait autorité sur quelle donnée ?) et les mécanismes de propagation (comment les autres systèmes accèdent-ils à cette vérité ?).

> **Définition formelle**
> **System of Record** : Système désigné comme source faisant autorité pour un ensemble de données spécifique. Toute modification doit transiter par ce système, et les autres systèmes synchronisent leur copie locale à partir de cette source.

Le concept de *System of Record* (SoR) formalise cette distribution contrôlée. Chaque catégorie de données possède un propriétaire désigné. Les systèmes consommateurs maintiennent des copies dérivées, explicitement reconnues comme potentiellement périmées. Cette reconnaissance explicite de la réplication, plutôt que sa dissimulation, constitue un progrès architectural majeur.

### 4.1.3 Qualité et Gouvernance Sémantique

L'interopérabilité technique — la capacité à transférer des octets entre systèmes — ne suffit pas. L'interopérabilité sémantique exige que les systèmes partagent une compréhension commune du *sens* des données échangées. Un champ « montant » dans un système peut représenter un montant HT, TTC, ou en devise locale selon le contexte. Sans alignement sémantique, l'intégration produit des résultats erronés malgré une exécution technique irréprochable.

La gouvernance des données englobe les politiques, processus et responsabilités garantissant que les données demeurent exactes, cohérentes, sécurisées et conformes aux réglementations. Dans un contexte d'intégration, cette gouvernance doit s'étendre au-delà des frontières des systèmes individuels pour couvrir les flux de données eux-mêmes.

Plusieurs dimensions composent cette gouvernance. La qualité des données vérifie leur exactitude, complétude et cohérence. La lignée des données ( *data lineage* ) trace leur provenance et leurs transformations. La catalogation documente leur existence et leur signification. Le contrôle d'accès régit qui peut lire ou modifier quoi. La conformité assure le respect des obligations légales et contractuelles.

> **Anti-patron**
> **Intégration sans gouvernance** : Multiplier les pipelines de données sans établir de catalogue centralisé, de standards de qualité ou de responsabilités claires. Résultat : prolifération de copies divergentes, impossibilité de tracer l'origine des erreurs, violations réglementaires potentielles.

### 4.1.4 Tension entre Autonomie et Cohérence

L'architecture moderne valorise l'autonomie des équipes et des services. Chaque équipe devrait pouvoir développer, déployer et opérer son service indépendamment, sans coordination excessive avec les autres équipes. Cette autonomie favorise la vélocité et réduit les goulots d'étranglement organisationnels.

Toutefois, l'intégration des données crée intrinsèquement des dépendances. Si le service A consomme des données du service B, toute modification du schéma par B peut impacter A. L'autonomie parfaite supposerait l'absence de toute dépendance, ce qui est incompatible avec l'objectif même d'intégration.

La gestion de cette tension passe par des contrats explicites entre producteurs et consommateurs de données. Ces contrats définissent les structures, les garanties de qualité et les règles d'évolution. Le chapitre VI approfondira ces mécanismes de contractualisation; les patrons présentés ici s'appuient sur cette notion sans la détailler.

---

## 4.2 Patrons d'Architecture pour l'Intégration des Données

Les sept patrons suivants constituent le catalogue essentiel de l'intégration des données. Ils ne sont pas mutuellement exclusifs; une architecture mature combine généralement plusieurs d'entre eux selon les contextes. L'ordre de présentation suit une logique de découplage croissant, en cohérence avec le continuum App → Data → Event de cette monographie.

### 4.2.1 Change Data Capture (CDC)

#### Définition

Le Change Data Capture (CDC) désigne l'ensemble des techniques permettant d'identifier et de capturer les modifications apportées aux données d'une base de données, puis de les rendre disponibles sous forme de flux exploitable par d'autres systèmes.

#### Problème Résolu

Les bases de données relationnelles constituent souvent le cœur des systèmes d'information d'entreprise. Des décennies de logique métier y sont encodées. Extraire les changements de ces systèmes sans impacter leurs performances ni modifier leur code représente un défi récurrent. Les approches traditionnelles — requêtes périodiques sur des colonnes de date de modification, déclencheurs ( *triggers* ) applicatifs — présentent des limitations significatives en termes de performance, de fiabilité ou d'intrusivité.

#### Mécanisme

Le CDC basé sur les journaux ( *log-based CDC* ) exploite les mécanismes natifs de réplication des bases de données. Chaque système de gestion de base de données maintient un journal des transactions ( *transaction log* ,  *write-ahead log* ,  *redo log* ) pour garantir la durabilité et permettre la récupération après incident. Ce journal contient l'historique détaillé de toutes les modifications : insertions, mises à jour, suppressions.

Le connecteur CDC s'abonne à ce journal et convertit chaque entrée en un événement structuré. Pour une mise à jour, l'événement contient typiquement l'état avant modification ( *before* ), l'état après modification ( *after* ), les métadonnées de la transaction (horodatage, identifiant de transaction) et l'identification de la table et de la clé primaire concernées.

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│   Base de données   │     │   Connecteur CDC    │     │   Bus de messages   │
│                     │     │                     │     │                     │
│  ┌───────────────┐  │     │  Lecture du journal │     │  ┌───────────────┐  │
│  │ Table Orders  │──┼────►│  Conversion en      │────►│  │ Topic orders  │  │
│  └───────────────┘  │     │  événements         │     │  └───────────────┘  │
│         │           │     │  Gestion des offsets│     │                     │
│         ▼           │     │                     │     │                     │
│  ┌───────────────┐  │     └─────────────────────┘     └─────────────────────┘
│  │ Transaction   │  │
│  │ Log (WAL)     │  │
│  └───────────────┘  │
└─────────────────────┘
```

Debezium, projet open source sous l'égide de Red Hat, s'est imposé comme la référence pour le CDC log-based. Il supporte les principales bases de données (PostgreSQL, MySQL, Oracle, SQL Server, MongoDB) et s'intègre nativement avec Apache Kafka via Kafka Connect. Les offres infonuagiques proposent des équivalents gérés : AWS Database Migration Service, Azure Data Factory, Google Cloud Datastream.

> **Note technique**
> Le CDC log-based requiert une configuration spécifique de la base de données source. Pour PostgreSQL, cela implique l'activation de la réplication logique (`wal_level = logical`) et la création d'un slot de réplication. Ces paramètres doivent être planifiés car ils impactent la rétention des journaux et, potentiellement, l'espace disque.

#### Avantages et Inconvénients

Le CDC log-based présente des avantages considérables. Il n'impacte pas les performances de la base source au-delà de la lecture du journal, opération déjà optimisée pour la réplication. Il capture *toutes* les modifications, y compris celles effectuées par des processus ne transitant pas par l'application (scripts de maintenance, corrections manuelles). Il préserve l'ordre des transactions et leur atomicité.

Les inconvénients méritent attention. La configuration initiale peut s'avérer complexe, notamment pour les bases de données propriétaires. Le schéma des événements CDC est étroitement couplé au schéma de la base source; toute modification de ce dernier se répercute sur les consommateurs. La gestion des opérations DDL (création/modification de tables) varie selon les implémentations et peut nécessiter des procédures spécifiques.

> **Bonnes pratiques**
>
> * Établir un registre de schémas (Schema Registry) pour versionner les structures des événements CDC
> * Prévoir une stratégie de « snapshot initial » pour les tables existantes avant d'activer le flux continu
> * Monitorer la latence entre l'écriture en base et la disponibilité de l'événement sur le bus

#### Exemple d'Usage

Une entreprise de commerce électronique souhaite alimenter son entrepôt de données analytique en quasi temps réel. Plutôt que d'exécuter des extractions nocturnes qui surchargent la base de production et introduisent 24 heures de latence, elle déploie Debezium sur sa base de commandes PostgreSQL. Chaque nouvelle commande, modification de statut ou annulation génère un événement sur Kafka. Un consommateur Kafka Streams transforme ces événements en format analytique et les charge dans le data warehouse. La latence passe de 24 heures à moins de 30 secondes, et la charge sur la base de production devient négligeable.

> **Quand utiliser ce patron**
> *Contexte* : Besoin de propager les modifications d'une base de données vers d'autres systèmes sans modifier le code applicatif existant; exigence de faible latence et de capture exhaustive.
> *Alternatives* : Pour des cas simples avec latence tolérante, des requêtes périodiques sur colonnes de timestamp peuvent suffire. Pour des modifications applicatives maîtrisées, le patron Transactional Outbox (chapitre V) offre plus de contrôle sur le format des événements.

---

### 4.2.2 Data Virtualization et Federation

#### Définition

La virtualisation des données ( *Data Virtualization* ) est une approche d'intégration qui fournit une vue unifiée de données provenant de sources hétérogènes sans les déplacer physiquement. Les consommateurs interrogent une couche d'abstraction qui traduit leurs requêtes vers les systèmes sources, agrège les résultats et les présente de manière transparente.

#### Problème Résolu

La réplication des données vers un emplacement central (entrepôt de données, lac de données) introduit de la latence, consomme du stockage et crée des copies qu'il faut maintenir synchronisées. Pour certains cas d'usage — requêtes ad hoc, exploration de données, rapports impliquant des sources diverses — cette duplication s'avère disproportionnée par rapport au besoin. La virtualisation permet d'accéder aux données « là où elles vivent » tout en offrant une interface unifiée.

#### Mécanisme

Une plateforme de virtualisation des données comprend plusieurs composants. La couche de connecteurs établit les connexions vers les sources hétérogènes : bases de données relationnelles, systèmes NoSQL, APIs REST, fichiers (CSV, Parquet), services infonuagiques. Le moteur de requêtes analyse les requêtes entrantes, planifie leur exécution distribuée, pousse les prédicats et projections vers les sources ( *predicate pushdown* ) pour minimiser les transferts, puis agrège les résultats. Le cache optionnel conserve les résultats de requêtes fréquentes pour améliorer les performances. La couche de métadonnées gère le catalogue des sources, les mappings sémantiques et les politiques de sécurité.

```
┌─────────────────────────────────────────────────────────────┐
│                    Couche de Virtualisation                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Interface SQL/REST unifiée             │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          Moteur de requêtes fédérées                │   │
│  │    (optimisation, pushdown, agrégation)             │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │Connecteur│  │Connecteur│  │Connecteur│  │Connecteur│   │
│  │PostgreSQL│  │ MongoDB  │  │Salesforce│  │  S3/CSV  │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
└───────┼─────────────┼─────────────┼─────────────┼─────────┘
        │             │             │             │
        ▼             ▼             ▼             ▼
   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
   │PostgreSQL│  │ MongoDB │   │Salesforce│  │   S3    │
   └─────────┘   └─────────┘   └─────────┘   └─────────┘
```

Les produits de virtualisation incluent Denodo, Dremio, Trino (anciennement PrestoSQL), IBM Data Virtualization et Starburst Enterprise. Chacun présente des forces distinctes : Denodo excelle dans les environnements d'entreprise complexes avec de nombreuses sources; Trino/Starburst cible les requêtes analytiques à grande échelle sur des lacs de données; Dremio combine virtualisation et accélération via son « data lake engine ».

> **Note technique**
> L'efficacité de la virtualisation dépend fortement de la capacité du moteur à « pousser » les opérations vers les sources ( *pushdown* ). Une jointure entre deux tables de sources différentes nécessite de rapatrier les données, ce qui peut devenir prohibitif. Les requêtes impliquant principalement une source avec filtres poussables offrent les meilleures performances.

#### Avantages et Inconvénients

La virtualisation élimine la latence de réplication : les données sont accédées dans leur état actuel. Elle réduit la duplication et les coûts de stockage associés. Elle simplifie l'architecture en évitant la prolifération de pipelines ETL. Elle s'adapte naturellement à l'ajout de nouvelles sources.

Cependant, les performances dépendent des sources les plus lentes impliquées dans la requête. Les requêtes complexes (jointures multi-sources, agrégations volumineuses) peuvent s'avérer coûteuses. La disponibilité du système virtualisé dépend de la disponibilité de toutes les sources interrogées. Certaines transformations complexes restent difficiles à exprimer sans matérialisation intermédiaire.

> **Anti-patron**
> **Virtualisation comme entrepôt de données** : Utiliser la virtualisation pour des requêtes analytiques lourdes et récurrentes sur des volumes massifs. Le coût de requêtage répété sur les sources dépasse rapidement celui d'une matérialisation périodique. La virtualisation convient aux requêtes exploratoires, ad hoc ou sur des données fraîches, non aux rapports quotidiens sur des téraoctets.

#### Exemple d'Usage

Une équipe de science des données doit explorer des données clients réparties entre un CRM Salesforce, une base de transactions PostgreSQL et des logs comportementaux stockés en fichiers Parquet sur S3. Plutôt que de créer un pipeline ETL complexe pour centraliser ces données, elle déploie Trino avec des connecteurs vers chaque source. Les analystes écrivent des requêtes SQL standard joignant ces sources comme si elles formaient une base unique. Pour les requêtes devenues récurrentes, ils matérialisent les résultats dans des tables dérivées, conservant la flexibilité pour l'exploration.

> **Quand utiliser ce patron**
> *Contexte* : Accès ad hoc à des données hétérogènes; besoin de fraîcheur maximale; exploration avant industrialisation; sources nombreuses et évolutives.
> *Alternatives* : Pour des requêtes récurrentes et prévisibles, un entrepôt de données alimenté par CDC ou ETL offre de meilleures performances. Pour des flux temps réel, le streaming (chapitre V) s'avère plus approprié.

---

### 4.2.3 CQRS — Command Query Responsibility Segregation

#### Définition

CQRS (Command Query Responsibility Segregation) est un patron architectural qui sépare explicitement les modèles de données utilisés pour les opérations d'écriture ( *commands* ) de ceux utilisés pour les opérations de lecture ( *queries* ). Plutôt qu'un modèle unique servant les deux usages, deux modèles distincts et optimisés coexistent.

#### Problème Résolu

Un modèle de données unique constitue toujours un compromis. Les contraintes d'intégrité et la normalisation favorisent les écritures cohérentes mais pénalisent les lectures complexes (jointures multiples). Les dénormalisations accélèrent les lectures mais compliquent les mises à jour et risquent les incohérences. Dans les systèmes à charge asymétrique — où les lectures dominent largement les écritures, ou inversement — ce compromis unique devient sous-optimal pour les deux cas d'usage.

#### Mécanisme

Dans une architecture CQRS, le modèle d'écriture (*write model* ou  *command model* ) est optimisé pour garantir l'intégrité des modifications. Il est typiquement normalisé, encapsule la logique métier de validation et constitue la source de vérité. Le modèle de lecture (*read model* ou  *query model* ) est optimisé pour les requêtes anticipées. Il peut être dénormalisé, pré-agrégé, indexé de manière spécifique, voire répliqué en plusieurs variantes selon les besoins des consommateurs.

La synchronisation entre les deux modèles s'effectue de manière asynchrone. Lorsqu'une commande modifie le modèle d'écriture, un mécanisme de propagation (événements, CDC, messages) met à jour le ou les modèles de lecture. Cette asynchronicité introduit une *cohérence à terme* ( *eventual consistency* ) : le modèle de lecture peut temporairement refléter un état périmé.

```
┌───────────────────────────────────────────────────────────────────┐
│                         Application                               │
│   ┌─────────────────┐                    ┌─────────────────┐     │
│   │   Commandes     │                    │    Requêtes     │     │
│   │ (Create, Update,│                    │ (List, Search,  │     │
│   │    Delete)      │                    │    Report)      │     │
│   └────────┬────────┘                    └────────┬────────┘     │
└────────────┼─────────────────────────────────────┼───────────────┘
             │                                     │
             ▼                                     ▼
┌────────────────────┐                 ┌────────────────────┐
│   Modèle d'écriture│                 │   Modèle de lecture│
│   (Write Model)    │                 │   (Read Model)     │
│                    │                 │                    │
│  - Normalisé       │   ─────────►   │  - Dénormalisé     │
│  - Intégrité forte │   Propagation  │  - Pré-agrégé      │
│  - Source de vérité│   asynchrone   │  - Multi-vues      │
│                    │                 │                    │
│  PostgreSQL        │                 │  Elasticsearch,    │
│                    │                 │  Redis, Cassandra  │
└────────────────────┘                 └────────────────────┘
```

> **Note technique**
> La propagation entre modèles peut emprunter plusieurs chemins. Le CDC capture les changements du modèle d'écriture et les propage vers les modèles de lecture. Alternativement, l'application peut émettre explicitement des événements de domaine lors des écritures. L'Event Sourcing (chapitre V) pousse cette logique à l'extrême en faisant des événements la source de vérité primaire.

#### Avantages et Inconvénients

CQRS permet d'optimiser indépendamment les performances de lecture et d'écriture. Les modèles de lecture peuvent être technologiquement distincts : Elasticsearch pour la recherche textuelle, Redis pour les accès clé-valeur rapides, une base colonne pour l'analytique. L'architecture tolère des ratios lecture/écriture extrêmes sans compromis. Elle facilite également la mise à l'échelle indépendante des capacités de lecture et d'écriture.

La complexité constitue le principal inconvénient. Deux modèles à maintenir, un mécanisme de synchronisation à fiabiliser, une cohérence à terme à gérer. Les interfaces utilisateur doivent anticiper le délai de propagation : un utilisateur créant une ressource pourrait ne pas la voir immédiatement dans une liste. Cette complexité n'est justifiée que pour des systèmes où la charge ou les exigences de performance l'imposent.

> **Bonnes pratiques**
>
> * Accepter explicitement la cohérence à terme et concevoir l'expérience utilisateur en conséquence (indicateurs de traitement, confirmations optimistes)
> * Prévoir des mécanismes de réconciliation pour détecter et corriger les divergences entre modèles
> * Commencer simple : CQRS peut émerger progressivement en ajoutant des vues de lecture à un système existant

#### Exemple d'Usage

Un système de gestion de contenu (CMS) pour un média en ligne traite des millions de lectures par minute (affichage des articles) contre quelques centaines d'écritures (publications, modifications). Le modèle d'écriture, une base PostgreSQL normalisée, garantit l'intégrité éditoriale avec workflow de validation, versioning et gestion des métadonnées. Le modèle de lecture combine Elasticsearch pour la recherche et un cache Redis pour les articles populaires. Un pipeline CDC synchronise les publications validées vers ces cibles en moins de 500 ms. Les journalistes travaillent sur un backend à charge maîtrisée tandis que les millions de lecteurs bénéficient de temps de réponse inférieurs à 50 ms.

> **Quand utiliser ce patron**
> *Contexte* : Ratio lecture/écriture fortement asymétrique; exigences de performance distinctes pour les deux opérations; besoins de requêtage complexes (recherche textuelle, agrégations) incompatibles avec le modèle transactionnel.
> *Alternatives* : Pour des systèmes à charge modérée et symétrique, un modèle unique avec indexation appropriée suffit généralement. Les Materialized Views (section 4.2.6) offrent une version allégée de ce patron.

---

### 4.2.4 Data Mesh — Décentralisation par Domaine Métier

#### Définition

Le Data Mesh est une approche sociotechnique de l'architecture de données qui décentralise la propriété et la gestion des données par domaine métier. Plutôt qu'une équipe centrale responsable de toutes les données, chaque domaine (ventes, logistique, finance) possède, produit et expose ses données comme des produits à part entière, consommables par les autres domaines.

#### Problème Résolu

Les architectures de données centralisées — entrepôts de données, lacs de données — créent des goulots d'étranglement organisationnels. Une équipe centrale de données doit comprendre tous les domaines métier, prioriser les demandes concurrentes et maintenir des pipelines toujours plus nombreux. Cette équipe devient un facteur limitant, incapable de suivre le rythme d'évolution des domaines. Par ailleurs, les équipes métier se déresponsabilisent de la qualité des données qu'elles produisent, la considérant comme « le problème de l'équipe data ».

#### Mécanisme

Le Data Mesh repose sur quatre principes fondamentaux formulés par Zhamak Dehghani (Thoughtworks, 2019).

**La propriété par domaine** ( *Domain Ownership* ) attribue la responsabilité des données aux équipes qui les génèrent. L'équipe du domaine « Commandes » possède les données de commandes, en garantit la qualité et les expose. Elle ne délègue pas cette responsabilité à une équipe centrale.

**Les données comme produit** ( *Data as a Product* ) imposent que chaque ensemble de données exposé respecte des standards de qualité produit : documentation, contrats d'interface, SLA de disponibilité, facilité de découverte. Un « produit de données » n'est pas un export brut mais une interface soignée, versionnée et maintenue.

**La plateforme de données en libre-service** ( *Self-Serve Data Platform* ) fournit l'infrastructure et les outils permettant aux domaines de publier leurs produits de données sans expertise infrastructure profonde. Cette plateforme abstrait la complexité technique (stockage, compute, sécurité) derrière des interfaces standardisées.

**La gouvernance fédérée computationnelle** ( *Federated Computational Governance* ) établit des standards globaux (formats, qualité, sécurité) tout en laissant l'autonomie d'implémentation aux domaines. Ces standards sont codifiés et vérifiés automatiquement, non imposés par des processus manuels.

```
┌─────────────────────────────────────────────────────────────────────┐
│                   Gouvernance Fédérée                               │
│           (Standards, politiques, interopérabilité)                 │
└─────────────────────────────────────────────────────────────────────┘
                                  │
       ┌──────────────────────────┼──────────────────────────┐
       │                          │                          │
       ▼                          ▼                          ▼
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│ Domaine Ventes  │      │Domaine Logistique│     │ Domaine Finance │
│                 │      │                 │      │                 │
│ Produits de     │      │ Produits de     │      │ Produits de     │
│ données :       │      │ données :       │      │ données :       │
│ - Commandes     │◄────►│ - Expéditions   │◄────►│ - Transactions  │
│ - Clients       │      │ - Stocks        │      │ - Rapprochements│
│ - Panier moyen  │      │ - Livraisons    │      │ - Provisions    │
│                 │      │                 │      │                 │
│ Équipe : 2 data │      │ Équipe : 1 data │      │ Équipe : 2 data │
│ engineers       │      │ engineer        │      │ engineers       │
└─────────────────┘      └─────────────────┘      └─────────────────┘
       │                          │                          │
       └──────────────────────────┼──────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│              Plateforme de Données en Libre-Service                 │
│  (Infrastructure, outils, catalogage, monitoring, sécurité)         │
└─────────────────────────────────────────────────────────────────────┘
```

> **Perspective stratégique**
> Le Data Mesh est autant un changement organisationnel que technique. Son adoption requiert une restructuration des équipes, une redéfinition des responsabilités et une évolution culturelle vers la responsabilisation des domaines. Sans ce volet organisationnel, la mise en œuvre technique échouera.

#### Avantages et Inconvénients

Le Data Mesh distribue la charge de travail et élimine le goulot central. Il rapproche les données de l'expertise métier, améliorant potentiellement leur qualité et leur pertinence. Il aligne l'architecture de données sur l'architecture organisationnelle (loi de Conway), facilitant l'évolution parallèle des domaines.

Les défis sont substantiels. L'interopérabilité entre domaines exige des standards rigoureux et une gouvernance effective. La duplication des compétences data dans chaque domaine augmente les besoins en recrutement. Sans plateforme mature, chaque domaine réinvente des solutions, créant de la fragmentation. L'investissement initial est conséquent avant de récolter les bénéfices.

> **Anti-patron**
> **Data Mesh cosmétique** : Renommer l'équipe centrale « plateforme », créer quelques APIs et déclarer avoir adopté le Data Mesh. Sans transfert réel de propriété aux domaines, sans produits de données véritablement autonomes, le résultat est une architecture centralisée déguisée avec la complexité additionnelle d'une rhétorique distribuée.

#### Exemple d'Usage

Un groupe de distribution multienseigne comptant 15 000 employés et des dizaines de systèmes opérationnels peinait avec son lac de données centralisé : 18 mois de backlog de demandes, qualité des données médiocre, équipe data épuisée. L'adoption du Data Mesh sur trois ans a transféré la propriété des données aux cinq domaines métier principaux. Chaque domaine a constitué une cellule data de deux à quatre personnes, publie ses produits de données sur une plateforme commune (basée sur Databricks et un catalogue interne) et respecte des contrats de qualité vérifiés automatiquement. Le backlog a été résorbé, le time-to-insight est passé de mois à semaines, et la satisfaction des consommateurs de données a significativement augmenté.

> **Quand utiliser ce patron**
> *Contexte* : Organisation de grande taille avec des domaines métier distincts; équipe data centrale devenue goulot d'étranglement; maturité organisationnelle permettant la responsabilisation des domaines.
> *Alternatives* : Pour les organisations plus petites ou moins matures, une approche centralisée avec des processus de priorisation améliorés peut suffire. Le Data Fabric (section 4.2.7) offre une alternative plus technologique que sociotechnique.

---

### 4.2.5 Schema Registry — Gouvernance des Contrats de Données

#### Définition

Un Schema Registry est un service centralisé qui stocke, versionne et valide les schémas (structures) des données échangées entre systèmes. Il constitue le référentiel faisant autorité sur le format des messages, événements ou enregistrements, permettant aux producteurs et consommateurs de données de s'accorder sur un contrat explicite.

#### Problème Résolu

Dans une architecture distribuée où de multiples services échangent des données, l'absence de contrat explicite sur la structure de ces données conduit à des ruptures silencieuses. Un producteur ajoute un champ, en renomme un autre ou change un type : les consommateurs, ignorants de ce changement, échouent en production. La découverte tardive de ces incompatibilités génère des incidents coûteux et de la méfiance entre équipes.

#### Mécanisme

Le Schema Registry maintient un catalogue de schémas identifiés par un sujet ( *subject* ) — typiquement le nom du topic Kafka ou du type d'événement — et une version. Chaque schéma est exprimé dans un format structuré permettant la sérialisation compacte et la validation : Apache Avro, Protocol Buffers (Protobuf) ou JSON Schema sont les plus courants.

Lors de la production d'un message, le producteur enregistre (ou vérifie) le schéma auprès du registre. Le registre attribue un identifiant unique au schéma. Le message sérialisé inclut cet identifiant en préfixe, permettant au consommateur de récupérer le schéma correspondant et de désérialiser correctement le message.

Le registre applique des règles de compatibilité lors de l'évolution des schémas. La compatibilité ascendante ( *backward* ) garantit qu'un nouveau consommateur peut lire les messages produits avec d'anciennes versions du schéma. La compatibilité descendante ( *forward* ) garantit qu'un ancien consommateur peut lire les messages produits avec de nouvelles versions. La compatibilité complète ( *full* ) combine les deux. Ces règles préviennent les modifications cassantes.

```
┌─────────────┐         ┌─────────────────┐         ┌─────────────┐
│  Producteur │         │ Schema Registry │         │ Consommateur│
│             │         │                 │         │             │
│  1. Enregistre        │  Stocke schéma  │         │             │
│     schéma   ───────► │  Attribue ID=42 │         │             │
│             │         │  Vérifie compat.│         │             │
│             │         │                 │         │             │
│  2. Sérialise         │                 │         │             │
│     message  ───────────────────────────────────► │ 3. Récupère │
│     [ID=42|payload]   │                 │         │    schéma   │
│             │         │ ◄─────────────────────────│    ID=42    │
│             │         │  Retourne schéma│         │             │
│             │         │                 │         │ 4. Désérialise
│             │         │                 │         │    message  │
└─────────────┘         └─────────────────┘         └─────────────┘
```

Confluent Schema Registry, composant de la plateforme Confluent, constitue l'implémentation de référence pour l'écosystème Kafka. AWS Glue Schema Registry, Apicurio Registry (open source, Red Hat) et Azure Schema Registry offrent des alternatives selon l'environnement.

> **Note technique**
> Avro est souvent privilégié pour Kafka en raison de sa sérialisation compacte et de son support natif de l'évolution de schéma. Protobuf excelle pour les communications gRPC et offre une meilleure performance de sérialisation/désérialisation. JSON Schema convient aux contextes où la lisibilité humaine prime sur la compacité.

#### Avantages et Inconvénients

Le Schema Registry élimine l'ambiguïté sur la structure des données. Il prévient les déploiements de schémas incompatibles grâce aux règles de compatibilité. Il centralise la documentation des structures de données, facilitant la découverte. Il permet une évolution contrôlée et traçable des schémas.

Les contraintes incluent une dépendance à la disponibilité du registre (atténuée par le cache côté client). L'obligation de définir explicitement les schémas ajoute du travail initial, perçu comme contraignant par certaines équipes. Le choix du format de schéma (Avro, Protobuf, JSON Schema) influence l'écosystème d'outils disponible.

> **Bonnes pratiques**
>
> * Définir une politique de compatibilité par défaut (backward est souvent un bon point de départ)
> * Intégrer la validation de schéma dans les pipelines CI/CD pour détecter les incompatibilités avant déploiement
> * Documenter les schémas au-delà de leur structure technique : signification des champs, unités, contraintes métier

#### Exemple d'Usage

Une plateforme de paiement traite des millions de transactions quotidiennes, chacune produisant un événement consommé par une douzaine de services (détection de fraude, comptabilité, reporting, notifications). L'adoption d'un Schema Registry avec Avro a stabilisé les échanges : tout nouveau champ doit avoir une valeur par défaut (compatibilité backward), toute suppression de champ est interdite (compatibilité forward). Un champ « montant » mal typé (string au lieu de decimal) a été détecté en CI avant déploiement, évitant une panne potentielle. Le registre sert également de documentation vivante, consultée par les équipes pour comprendre la structure des événements.

> **Quand utiliser ce patron**
> *Contexte* : Échanges de données structurées entre multiples producteurs et consommateurs; besoin de garantir la compatibilité lors des évolutions; volume d'échanges justifiant une sérialisation efficace.
> *Alternatives* : Pour des échanges simples entre deux systèmes, une documentation API et des tests d'intégration peuvent suffire. Le chapitre VI approfondira les contrats d'interface au-delà des schémas de données.

---

### 4.2.6 Materialized View — Vues Pré-Calculées

#### Définition

Une Materialized View (vue matérialisée) est une structure de données persistante contenant le résultat pré-calculé d'une requête. Contrairement à une vue standard qui exécute sa requête à chaque accès, la vue matérialisée stocke physiquement les données, offrant des performances de lecture proches d'une table tout en maintenant (avec un certain délai) la cohérence avec les données sources.

#### Problème Résolu

Les requêtes analytiques complexes — jointures multiples, agrégations, fenêtres temporelles — consomment des ressources significatives et présentent des temps de réponse incompatibles avec les interactions utilisateur. Exécuter ces requêtes à chaque demande surcharge les systèmes sources et dégrade l'expérience. La Materialized View déplace le coût computationnel du moment de la lecture vers celui de l'écriture (ou d'un rafraîchissement périodique), amortissant ce coût sur de multiples lectures.

#### Mécanisme

La création d'une Materialized View implique la définition de la requête source, le choix de la stratégie de rafraîchissement et la configuration du stockage. La requête source spécifie les données à matérialiser : tables impliquées, jointures, filtres, agrégations. Le rafraîchissement peut être complet (recalcul total périodique), incrémental (application des seuls changements depuis le dernier rafraîchissement) ou continu (mise à jour en quasi temps réel via streaming).

Les bases de données relationnelles (PostgreSQL, Oracle, SQL Server) supportent nativement les Materialized Views avec rafraîchissement complet ou incrémental selon les cas. Les systèmes de streaming (Kafka Streams, Apache Flink, ksqlDB) permettent de maintenir des vues matérialisées en temps réel alimentées par des flux d'événements. Les bases de données analytiques (ClickHouse, Apache Druid) optimisent le stockage et le rafraîchissement pour les cas d'usage OLAP.

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Sources de données                              │
│   ┌───────────┐    ┌───────────┐    ┌───────────┐                  │
│   │  Orders   │    │ Customers │    │ Products  │                  │
│   └─────┬─────┘    └─────┬─────┘    └─────┬─────┘                  │
│         │                │                │                         │
│         └────────────────┼────────────────┘                         │
│                          │                                          │
│                          ▼                                          │
│              ┌─────────────────────┐                                │
│              │  Processus de       │                                │
│              │  rafraîchissement   │                                │
│              │  (batch/streaming)  │                                │
│              └──────────┬──────────┘                                │
│                         │                                           │
│                         ▼                                           │
│              ┌─────────────────────┐                                │
│              │  Materialized View  │                                │
│              │  sales_by_category  │                                │
│              │                     │                                │
│              │ category | total    │                                │
│              │ ---------|--------- │                                │
│              │ Électro  | 1.2M$    │                                │
│              │ Mode     | 850K$    │                                │
│              │ Maison   | 620K$    │                                │
│              └─────────────────────┘                                │
└─────────────────────────────────────────────────────────────────────┘
```

> **Note technique**
> Le rafraîchissement incrémental n'est pas toujours possible. Certaines opérations (DISTINCT global, certaines agrégations avec HAVING) nécessitent un recalcul complet. L'analyse de la requête et des contraintes du SGBD détermine la stratégie applicable. PostgreSQL, par exemple, supporte le rafraîchissement concurrent (sans verrouillage exclusif) depuis la version 9.4.

#### Avantages et Inconvénients

Les Materialized Views offrent des performances de lecture prévisibles et rapides, indépendantes de la complexité de la requête source. Elles soulagent les systèmes sources en concentrant la charge de calcul lors du rafraîchissement. Elles permettent de servir des requêtes analytiques depuis des systèmes transactionnels sans les dégrader.

Le compromis principal concerne la fraîcheur des données. Une vue rafraîchie toutes les heures présente jusqu'à une heure de retard. Le stockage supplémentaire consommé peut être significatif pour les vues volumineuses. La gestion des rafraîchissements (ordonnancement, monitoring, reprise sur erreur) ajoute de la complexité opérationnelle.

> **Bonnes pratiques**
>
> * Nommer explicitement les vues matérialisées pour indiquer leur nature dérivée (préfixe `mv_` ou suffixe `_materialized`)
> * Documenter la fenêtre de fraîcheur attendue et la communiquer aux consommateurs
> * Monitorer la durée des rafraîchissements pour anticiper les dépassements de fenêtre

#### Exemple d'Usage

Un site de commerce électronique affiche sur sa page d'accueil les « meilleures ventes par catégorie » et les « tendances de la semaine ». Ces indicateurs impliquent des agrégations sur des millions de lignes de commandes. Plutôt que d'exécuter ces requêtes lourdes à chaque chargement de page, des Materialized Views sont rafraîchies toutes les 15 minutes. La page d'accueil lit ces vues en moins de 10 ms, offrant une expérience fluide. Le délai de 15 minutes est acceptable pour des indicateurs de tendance, et les utilisateurs ne perçoivent pas la différence.

> **Quand utiliser ce patron**
> *Contexte* : Requêtes analytiques récurrentes sur des données volumineuses; tolérance à une fraîcheur légèrement dégradée; besoin de performances de lecture prévisibles.
> *Alternatives* : Pour une fraîcheur temps réel, les systèmes de streaming avec tables matérialisées (ksqlDB, Flink) conviennent mieux. Pour des requêtes ad hoc variables, la virtualisation (section 4.2.2) offre plus de flexibilité.

---

### 4.2.7 Data Fabric — Intégration Automatisée par Métadonnées

#### Définition

Le Data Fabric est une approche architecturale qui utilise les métadonnées et l'automatisation pour créer une couche d'intégration intelligente reliant dynamiquement les sources de données hétérogènes. Plutôt que de construire manuellement chaque pipeline d'intégration, le Data Fabric découvre, catalogue et relie les données de manière semi-automatique, réduisant l'effort d'intégration et accélérant l'accès à l'information.

#### Problème Résolu

La prolifération des sources de données (applications SaaS, bases de données, lacs de données, APIs partenaires) crée un défi de découvrabilité et de connectivité. Les équipes data passent un temps considérable à localiser les données pertinentes, comprendre leur structure, construire les pipelines d'extraction et maintenir ces intégrations. Le Data Fabric vise à automatiser une partie significative de ce travail grâce aux métadonnées.

#### Mécanisme

Le Data Fabric repose sur plusieurs capacités complémentaires. Le catalogue de métadonnées actif découvre automatiquement les sources de données, collecte leurs métadonnées techniques (schémas, volumes, fraîcheur) et enrichit ces métadonnées avec des informations sémantiques (descriptions, classifications, propriétaires). Le graphe de connaissances modélise les relations entre entités de données : un « client » dans le CRM correspond au « customer » dans l'ERP, lui-même lié aux « transactions » du système de paiement. Le moteur de recommandation suggère des sources pertinentes pour un cas d'usage, des transformations applicables ou des données similaires. L'orchestration automatisée génère et exécute des pipelines d'intégration à partir de spécifications de haut niveau.

L'intelligence artificielle et le machine learning jouent un rôle croissant dans le Data Fabric. La classification automatique des données (données personnelles, données sensibles) améliore la gouvernance. La détection d'anomalies identifie les problèmes de qualité. La correspondance sémantique ( *semantic matching* ) relie automatiquement des entités nommées différemment mais désignant le même concept.

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Data Fabric                                 │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              Couche d'Intelligence / ML                      │   │
│  │  (Classification, matching sémantique, recommandations)      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │         Graphe de Connaissances & Catalogue Actif            │   │
│  │  (Entités, relations, lignée, qualité, propriétaires)        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              Orchestration & Pipelines Dynamiques            │   │
│  │  (Génération, exécution, monitoring, self-healing)           │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │Connecteur│  │Connecteur│  │Connecteur│  │Connecteur│           │
│  │   ERP    │  │   CRM    │  │ Data Lake│  │   APIs   │           │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘           │
└───────┼─────────────┼─────────────┼─────────────┼─────────────────┘
        ▼             ▼             ▼             ▼
   Sources de données hétérogènes
```

Les plateformes de Data Fabric incluent Informatica Intelligent Data Management Cloud, Talend Data Fabric, IBM Cloud Pak for Data et les offres émergentes des hyperscalers (Google Dataplex, Microsoft Purview). Le positionnement exact varie : certains mettent l'accent sur le catalogue, d'autres sur l'orchestration ou la gouvernance.

> **Perspective stratégique**
> Le Data Fabric représente une vision ambitieuse où l'intégration devient largement automatisée. La réalité actuelle est plus nuancée : l'automatisation complète reste un horizon, et une expertise humaine significative demeure nécessaire. Les organisations doivent évaluer les capacités réelles des produits au-delà du discours marketing.

#### Avantages et Inconvénients

Le Data Fabric accélère la découverte et l'accès aux données grâce au catalogue centralisé. Il réduit l'effort de construction de pipelines via l'automatisation. Il améliore la gouvernance en centralisant les métadonnées de qualité, de lignée et de conformité. Il s'adapte à l'hétérogénéité croissante des sources.

Les inconvénients incluent la complexité et le coût des plateformes commerciales. L'automatisation a ses limites : les cas complexes requièrent toujours une intervention experte. La dépendance à un fournisseur unique de plateforme crée un risque de verrouillage ( *vendor lock-in* ). L'investissement initial pour cataloguer et connecter les sources existantes est substantiel.

> **Anti-patron**
> **Data Fabric sans gouvernance** : Déployer une plateforme de Data Fabric sans avoir défini les rôles, les politiques de qualité et les processus de gestion des métadonnées. Le catalogue se remplit de données non documentées, non certifiées et non maintenues, devenant un cimetière de métadonnées plutôt qu'un accélérateur d'accès.

#### Exemple d'Usage

Un groupe bancaire comptant des centaines de systèmes sources et des décennies d'historique a déployé une plateforme de Data Fabric pour améliorer l'accès aux données réglementaires. Le catalogue actif a automatiquement scanné les bases de données, identifié les champs contenant des données personnelles (classification ML) et cartographié les relations entre entités clients à travers les systèmes. Les analystes réglementaires, auparavant dépendants d'experts techniques pour localiser les données, peuvent désormais explorer le catalogue, comprendre la lignée et accéder aux données en libre-service. Le temps de préparation d'un rapport réglementaire a diminué de 60 %.

> **Quand utiliser ce patron**
> *Contexte* : Paysage de données très fragmenté (centaines de sources); enjeux de gouvernance et de conformité; besoin d'accélérer l'accès en libre-service; ressources disponibles pour l'investissement initial.
> *Alternatives* : Pour des paysages plus simples, un catalogue de données (Alation, Collibra) sans les capacités d'orchestration avancée peut suffire. Le Data Mesh (section 4.2.4) offre une approche complémentaire, plus organisationnelle que technologique.

---

## 4.3 Synthèse et Critères de Sélection

Les sept patrons présentés ne constituent pas des alternatives exclusives mais des outils complémentaires. Une architecture de données mature combine typiquement plusieurs d'entre eux selon les contextes. Le tableau suivant synthétise leurs caractéristiques et aide à orienter les choix.

| Patron              | Fraîcheur               | Couplage                    | Complexité                  | Cas d'usage principal                                  |
| ------------------- | ------------------------ | --------------------------- | ---------------------------- | ------------------------------------------------------ |
| CDC                 | Temps réel              | Faible (via événements)   | Moyenne                      | Propagation des changements DB vers d'autres systèmes |
| Data Virtualization | Maximale (accès direct) | Fort (requête synchrone)   | Moyenne                      | Exploration ad hoc, requêtes fédérées              |
| CQRS                | Configurable             | Faible (modèles séparés) | Élevée                     | Systèmes à charge asymétrique lecture/écriture     |
| Data Mesh           | Variable                 | Faible (par domaine)        | Élevée (organisationnelle) | Grandes organisations, décentralisation               |
| Schema Registry     | N/A (métadonnées)      | Faible                      | Faible                       | Gouvernance des échanges structurés                  |
| Materialized View   | Périodique              | Faible                      | Faible                       | Requêtes analytiques récurrentes                     |
| Data Fabric         | Variable                 | Variable                    | Élevée                     | Paysages fragmentés, gouvernance, libre-service       |

### Critères de Décision

Plusieurs questions guident le choix des patrons appropriés.

Quelle latence le cas d'usage tolère-t-il ? Pour du temps réel, CDC ou streaming. Pour du quasi temps réel, CDC avec consommation différée. Pour du quotidien, batch ou Materialized Views.

Quel niveau de cohérence est requis ? Pour une cohérence forte, les lectures synchrones (virtualisation, appels API) garantissent l'état actuel. Pour une cohérence à terme, les mécanismes asynchrones (CDC, CQRS) offrent de meilleures performances.

Quelle est la volumétrie ? Les volumes massifs favorisent les approches par lots ou streaming. La virtualisation peine sur les jointures volumineuses. Les Materialized Views absorbent le coût computationnel lors du rafraîchissement.

Quelle maturité organisationnelle ? Le Data Mesh requiert des équipes autonomes et responsabilisées. Le Data Fabric suppose une gouvernance structurée. Le CDC peut être adopté de manière incrémentale avec un impact organisationnel limité.

> **Règle d'or**
> Commencer simple. Un CDC bien opéré couvre une majorité des besoins de propagation de données. Ajouter de la complexité (CQRS, Data Mesh) uniquement lorsque les limites de l'approche simple sont atteintes et documentées.

---

## Conclusion et Transition

Ce chapitre a exploré l'intégration des données sous l'angle du *Nom* — la représentation de l'état, la cohérence de l'information et sa gouvernance à travers l'écosystème. Les sept patrons présentés offrent un spectre de solutions, du CDC qui convertit les mutations de base de données en flux exploitables, au Data Mesh qui réorganise la propriété des données par domaine métier.

Trois enseignements traversent ce catalogue. Premièrement, la fraîcheur parfaite a un coût, et ce coût doit être mis en balance avec la valeur métier délivrée. Deuxièmement, la « source unique de vérité » est un principe de gouvernance plus qu'une réalité physique : dans les architectures distribuées, la vérité est distribuée, et l'enjeu consiste à organiser cette distribution plutôt qu'à la nier. Troisièmement, l'intégration des données ne peut être dissociée de la gouvernance : sans catalogue, sans contrats, sans responsabilités claires, la prolifération des pipelines crée plus de confusion que de valeur.

L'intégration des données occupe une position intermédiaire dans le continuum de couplage. Moins immédiate que l'intégration des applications (pas d'attente synchrone d'une réponse), elle maintient néanmoins une forme de dépendance sur la structure des données partagées. Le chapitre suivant franchira une étape supplémentaire vers le découplage en explorant l'intégration des événements — le  *Signal* . Là où les données représentent l'état, les événements représentent les faits : ce qui s'est passé, dans quel ordre, avec quelles conséquences. Le patron CDC, pont naturel entre ces deux mondes, illustre cette continuité : il transforme précisément les changements d'état (données) en flux de faits (événements).

L'étude de cas du chapitre X illustrera concrètement comment ces patrons s'articulent dans un scénario Order-to-Cash. La phase de persistance y mobilisera CDC et Transactional Outbox pour garantir l'atomicité entre l'écriture de la commande et sa publication vers les systèmes aval. Mais avant d'en arriver là, il convient d'explorer le troisième domaine d'intégration : les événements, terrain du découplage maximal et de la réactivité.

---

## Résumé du Chapitre

**Thème central** : L'intégration des données — « Le Nom » — concerne la cohérence de l'état, la gouvernance des structures et l'accessibilité de l'information à travers l'écosystème d'entreprise.

**Enjeux fondamentaux** :

* La fraîcheur des données impose un compromis entre latence et coût d'infrastructure
* La vérité unique est un principe organisationnel; la vérité distribuée est la réalité technique des architectures modernes
* La qualité et la gouvernance sémantique conditionnent l'utilité des données intégrées
* L'autonomie des équipes doit être équilibrée avec les besoins de cohérence globale

**Catalogue des patrons** :

1. **CDC** : Capture des modifications de base de données en flux d'événements
2. **Data Virtualization** : Accès unifié sans déplacement physique des données
3. **CQRS** : Séparation des modèles d'écriture et de lecture
4. **Data Mesh** : Décentralisation de la propriété par domaine métier
5. **Schema Registry** : Gouvernance des contrats de données
6. **Materialized View** : Pré-calcul de vues complexes pour la performance
7. **Data Fabric** : Intégration automatisée via métadonnées et intelligence

**Position dans le continuum** : L'intégration des données se situe entre le couplage fort de l'intégration applicative et le découplage maximal de l'intégration événementielle, préparant le terrain pour le chapitre V.

---

*Références techniques : Debezium Documentation (Red Hat, 2024), Confluent Schema Registry (Confluent, 2024), Data Mesh Principles (Dehghani/Thoughtworks, 2019), Gartner Magic Quadrant for Data Integration Tools (2024)*
