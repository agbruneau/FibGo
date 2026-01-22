
# Chapitre III — Intégration des Applications (Le Verbe)

*Focus : L'orchestration des processus, l'exposition des fonctionnalités et les interactions synchrones.*

---

## Introduction

Le chapitre précédent a posé les fondements théoriques de l'interopérabilité : le théorème CAP nous a rappelé l'impossible trinité entre cohérence, disponibilité et tolérance au partitionnement; le concept de couplage spatio-temporel a révélé la tension fondamentale entre interactions synchrones et asynchrones; enfin, les modèles de gouvernance nous ont confrontés au choix entre centralisation et décentralisation. Ces contraintes ne sont pas abstraites — elles se manifestent concrètement dès qu'une organisation tente de faire communiquer ses systèmes.

Ce troisième chapitre inaugure notre exploration du continuum d'intégration par son premier domaine : l'intégration des applications. Nous la désignons métaphoriquement comme « le Verbe » car elle concerne l'action, l'invocation directe de fonctionnalités, l'orchestration de processus. Contrairement à l'intégration des données (le Nom) qui se préoccupe de l'état, ou à l'intégration des événements (le Signal) qui privilégie la réactivité asynchrone, l'intégration des applications repose sur un dialogue synchrone : un système appelle, un autre répond, et le premier attend.

Cette modalité d'intégration demeure incontournable malgré ses contraintes. Lorsqu'un utilisateur soumet une commande, il attend une confirmation immédiate. Lorsqu'un système de paiement valide une transaction, le marchand ne peut procéder sans réponse. Le couplage temporel, aussi contraignant soit-il, reflète parfois une réalité métier non négociable. L'enjeu n'est donc pas d'éliminer les interactions synchrones, mais de les architecurer avec rigueur pour en maîtriser les risques.

Ce chapitre présente un catalogue de sept patrons d'architecture qui adressent les défis spécifiques de l'intégration applicative. Chaque patron sera examiné selon une structure uniforme : définition, problème résolu, mécanisme, avantages et inconvénients, puis exemple d'usage. Cette approche systématique vise à outiller les architectes et les développeurs seniors avec des solutions éprouvées, immédiatement actionnables.

---

## 3.1 Enjeux de l'Intégration des Applications

Avant d'aborder les patrons, il convient de caractériser précisément les défis que pose l'intégration des applications dans un écosystème d'entreprise moderne.

### 3.1.1 Le Couplage Temporel Fort

L'intégration des applications implique un couplage temporel : l'appelant doit attendre la réponse de l'appelé pour poursuivre son traitement. Cette dépendance crée une chaîne de disponibilité où la défaillance d'un maillon affecte l'ensemble. Si le service A appelle le service B qui appelle le service C, une panne de C se propage instantanément jusqu'à A. Cette propagation des défaillances constitue le risque principal des architectures fortement couplées.

Le couplage temporel impose également des contraintes de performance. La latence totale d'une opération correspond à la somme des latences individuelles de chaque appel. Dans une architecture de microservices où une requête utilisateur peut traverser une dizaine de services, les latences s'accumulent rapidement. Une requête qui semble simple côté client peut impliquer des dizaines d'appels réseau côté serveur.

### 3.1.2 Les Dépendances Directes

Au-delà du temps, l'intégration applicative crée des dépendances structurelles entre systèmes. Un service consommateur doit connaître l'adresse du service producteur, comprendre son contrat d'interface, gérer ses erreurs spécifiques. Cette connaissance mutuelle génère un couplage qui complique l'évolution indépendante des systèmes.

Lorsqu'un producteur modifie son interface, tous ses consommateurs sont potentiellement affectés. Cette fragilité s'amplifie avec le nombre de dépendances. Dans une architecture où chaque service communique directement avec plusieurs autres, la matrice des dépendances devient rapidement ingérable. Une modification apparemment mineure peut provoquer des régressions en cascade.

### 3.1.3 La Coordination des Services

L'intégration des applications soulève également la question de la coordination. Comment orchestrer une séquence d'appels pour réaliser un processus métier complet ? Comment garantir la cohérence lorsque certains appels réussissent et d'autres échouent ? Comment gérer les transactions distribuées sans les mécanismes ACID d'une base de données unique ?

Ces questions n'ont pas de réponse unique. Certains scénarios tolèrent une cohérence éventuelle, d'autres exigent une atomicité stricte. Les patrons présentés dans ce chapitre offrent des réponses adaptées à différents contextes, mais le choix architectural demeure une décision qui dépend des exigences métier spécifiques.

Prenons l'exemple d'une réservation de voyage impliquant un vol, un hôtel et une voiture de location. Ces trois réservations sont logiquement liées — le client ne veut pas d'un vol sans hôtel, ni d'un hôtel sans vol. Pourtant, elles sont gérées par trois systèmes distincts, potentiellement opérés par trois entreprises différentes. Comment garantir l'atomicité de cette transaction composite ? Les approches traditionnelles de verrouillage pessimiste ne fonctionnent pas à travers les frontières organisationnelles. Des patrons comme le Saga, que nous explorerons au chapitre V, offrent des alternatives basées sur des compensations, mais ils introduisent une complexité significative.

### 3.1.4 La Surface d'Attaque Sécuritaire

Chaque point d'intégration constitue une surface d'attaque potentielle. Les APIs exposées doivent être authentifiées, autorisées, protégées contre les abus. Les données en transit doivent être chiffrées. Les entrées doivent être validées pour prévenir les injections.

Dans une architecture de microservices, la multiplication des services multiplie les points d'entrée à sécuriser. Le trafic Est-Ouest (entre services internes) est souvent négligé alors qu'un attaquant ayant compromis un service peut l'utiliser pour explorer le réseau interne. L'approche Zero Trust, où chaque appel est authentifié même en interne, devient une nécessité.

Les patrons présentés dans ce chapitre contribuent à cette sécurisation. L'API Gateway centralise l'authentification externe. Le Service Mesh (que nous aborderons au chapitre VII) sécurise le trafic interne. Mais la sécurité demeure une préoccupation transversale qui doit être intégrée dès la conception, non ajoutée après coup.

> **Perspective stratégique**
> L'intégration des applications ne doit pas être abordée comme un problème purement technique. Les choix de couplage reflètent des compromis organisationnels : qui peut évoluer indépendamment, qui dépend de qui, quelle équipe porte la responsabilité en cas de défaillance. Ces questions méritent une attention explicite lors de la conception. Les organisations qui négligent cette dimension organisationnelle découvrent souvent que leur architecture technique, aussi élégante soit-elle, se heurte à des frictions humaines qui en limitent les bénéfices.

---

## 3.2 Catalogue des Patrons d'Architecture

Les sept patrons suivants constituent la boîte à outils essentielle de l'architecte d'intégration applicative. Ils ne sont pas mutuellement exclusifs — une architecture réelle les combine souvent pour adresser différents aspects d'un même système.

### 3.2.1 API Gateway

#### Définition

L'API Gateway est un composant d'infrastructure qui sert de point d'entrée unifié pour l'ensemble des requêtes entrantes vers un système. Il centralise les préoccupations transversales telles que l'authentification, l'autorisation, la limitation de débit, le routage et la transformation des requêtes.

#### Problème Résolu

Dans une architecture de microservices, chaque service expose potentiellement sa propre interface. Sans coordination, les clients doivent connaître l'adresse de chaque service, gérer autant de connexions distinctes, et répéter la logique d'authentification pour chaque appel. Cette dispersion complique le développement client, multiplie la surface d'attaque sécuritaire et rend difficile l'application de politiques uniformes.

L'API Gateway résout ce problème en offrant une façade unique. Les clients n'interagissent qu'avec la passerelle, qui se charge de router les requêtes vers les services appropriés après avoir appliqué les contrôles nécessaires.

#### Mécanisme

Le fonctionnement d'une API Gateway s'articule autour de plusieurs étapes. Lorsqu'une requête arrive, la passerelle vérifie d'abord l'identité de l'appelant via un mécanisme d'authentification, typiquement un jeton JWT ou une clé API. Elle contrôle ensuite que l'appelant dispose des permissions nécessaires pour l'opération demandée. Elle peut appliquer une limitation de débit pour protéger les services en aval contre les surcharges. Enfin, elle route la requête vers le service approprié, potentiellement en transformant le format ou en enrichissant les en-têtes.

La passerelle peut également agréger les réponses de plusieurs services pour éviter au client de multiplier les appels. Elle gère le protocole de communication, permettant par exemple d'exposer une interface REST aux clients tout en communiquant en gRPC avec les services internes.

```
┌─────────────────────────────────────────────────────────────┐
│                        CLIENT                               │
└─────────────────────────┬───────────────────────────────────┘
                          │ Requête HTTPS
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                     API GATEWAY                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │  Auth    │ │  Rate    │ │ Routing  │ │ Transform│       │
│  │  Check   │ │  Limit   │ │          │ │          │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
└─────────────────────────┬───────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ Service  │    │ Service  │    │ Service  │
    │    A     │    │    B     │    │    C     │
    └──────────┘    └──────────┘    └──────────┘
```

#### Avantages et Inconvénients

L'API Gateway simplifie considérablement l'expérience des développeurs clients. Au lieu de gérer de multiples points d'entrée avec leurs spécificités respectives, ils interagissent avec une interface unifiée et cohérente. La centralisation des préoccupations transversales évite la duplication de code et garantit l'application uniforme des politiques de sécurité.

La passerelle offre également une observabilité précieuse. Puisque toutes les requêtes la traversent, elle constitue un point idéal pour collecter des métriques, générer des traces et détecter des anomalies. Elle permet aussi de découpler l'interface externe de l'architecture interne, autorisant des restructurations sans impact sur les clients.

En contrepartie, l'API Gateway introduit un point de défaillance unique. Si elle devient indisponible, l'ensemble du système est inaccessible depuis l'extérieur. Cette criticité exige une attention particulière à la haute disponibilité : déploiement multi-instances, répartition de charge, basculement automatique. La passerelle peut également devenir un goulot d'étranglement si elle n'est pas dimensionnée correctement pour le volume de trafic.

> **Bonnes pratiques**
>
> * Déployer la passerelle en configuration haute disponibilité avec au minimum trois instances réparties sur différentes zones.
> * Éviter d'implémenter de la logique métier dans la passerelle; elle doit rester un composant d'infrastructure pur.
> * Surveiller attentivement la latence ajoutée par la passerelle et optimiser si elle dépasse quelques millisecondes.

#### Exemple d'Usage

Une entreprise de commerce électronique expose son catalogue, son système de commandes et son service client via une API Gateway unique. Les applications mobiles et le site web n'ont qu'un seul point d'entrée à configurer. L'authentification par jeton JWT est vérifiée une seule fois à la passerelle, puis les requêtes sont routées vers les microservices appropriés avec un en-tête interne identifiant l'utilisateur. La limitation de débit protège le système lors des pics de trafic comme le Vendredi fou, tandis que les tableaux de bord centralisés permettent de surveiller la santé de l'ensemble de l'écosystème.

La passerelle gère également la transformation de protocole. Les partenaires B2B qui utilisent encore des services SOAP sont accueillis par la gateway qui traduit leurs requêtes en appels REST internes. Cette adaptation évite de maintenir deux interfaces dans chaque microservice tout en préservant la compatibilité avec les systèmes existants.

L'équipe de plateforme a configuré différentes politiques de limitation selon les catégories de clients. Les applications mobiles natives bénéficient de quotas généreux car elles représentent le canal principal. Les intégrations partenaires ont des quotas négociés contractuellement. Les requêtes sans authentification, permises pour la consultation du catalogue public, sont limitées agressivement pour prévenir les abus de scraping.

> **Quand utiliser ce patron**
> *Contexte* : Architecture de microservices exposant des APIs à des clients externes (applications mobiles, partenaires, sites web). Également pertinent lorsque des préoccupations transversales (authentification, journalisation, limitation) doivent être appliquées uniformément.
> *Alternatives* : Pour des architectures plus simples avec peu de services, un répartiteur de charge avec terminaison TLS peut suffire. Pour des besoins d'adaptation par canal, combiner avec le patron BFF. Dans un environnement Kubernetes mature, un Ingress Controller peut assumer certaines fonctions de gateway.

---

### 3.2.2 Backend for Frontend (BFF)

#### Définition

Le Backend for Frontend est un patron où l'on crée un service d'agrégation dédié à chaque type de client. Au lieu d'exposer une API générique que tous les clients consomment, on développe une couche intermédiaire spécifiquement adaptée aux besoins de chaque canal : une pour l'application mobile, une pour le site web, une pour les partenaires B2B.

#### Problème Résolu

Les différents canaux consommateurs ont des besoins distincts. Une application mobile sur un réseau cellulaire nécessite des réponses compactes et un minimum d'appels réseau. Un site web peut se permettre des échanges plus verbeux. Une intégration B2B requiert des formats spécifiques et des niveaux de détail différents.

Face à cette diversité, deux approches inadaptées émergent souvent. La première consiste à créer une API « taille unique » qui tente de satisfaire tous les clients, aboutissant à des compromis sous-optimaux pour chacun. La seconde multiplie les points d'entrée dans les services métier eux-mêmes, polluant leur code avec des préoccupations de présentation.

Le BFF résout ce dilemme en isolant l'adaptation par canal dans une couche dédiée, laissant les services métier se concentrer sur leur domaine.

#### Mécanisme

Chaque BFF est un service à part entière, développé et déployé indépendamment. Il connaît les spécificités de son canal cible : format de réponse optimal, champs requis, transformations nécessaires. Lorsqu'il reçoit une requête, il orchestre les appels vers les services métier appropriés, agrège les résultats et les formate selon les attentes du client.

Le BFF peut implémenter des optimisations spécifiques. Pour un client mobile, il peut fusionner plusieurs appels en un seul pour réduire la latence réseau. Pour un partenaire B2B, il peut transformer les données vers un format standardisé de l'industrie. Pour un site web, il peut inclure des données de préchargement anticipant les prochaines actions de l'utilisateur.

```
┌────────────┐    ┌────────────┐    ┌────────────┐
│   Mobile   │    │    Web     │    │    B2B     │
│   Client   │    │   Client   │    │  Partner   │
└─────┬──────┘    └─────┬──────┘    └─────┬──────┘
      │                 │                 │
      ▼                 ▼                 ▼
┌────────────┐    ┌────────────┐    ┌────────────┐
│  Mobile    │    │   Web      │    │   B2B      │
│   BFF      │    │   BFF      │    │   BFF      │
└─────┬──────┘    └─────┬──────┘    └─────┬──────┘
      │                 │                 │
      └────────────┬────┴────────────┬────┘
                   ▼                 ▼
            ┌────────────┐    ┌────────────┐
            │  Service   │    │  Service   │
            │  Catalogue │    │  Commande  │
            └────────────┘    └────────────┘
```

#### Avantages et Inconvénients

Le BFF permet une évolution indépendante de chaque canal. L'équipe mobile peut modifier son BFF sans affecter le site web, et vice versa. Cette autonomie accélère les cycles de livraison et réduit la coordination inter-équipes. Le patron aligne également l'architecture sur l'organisation : l'équipe responsable d'un canal possède son BFF de bout en bout.

Les performances s'améliorent car chaque BFF peut optimiser les échanges pour son contexte spécifique. Un BFF mobile peut agréger agressivement les données pour minimiser les allers-retours réseau, tandis qu'un BFF web peut paralléliser les appels en exploitant une connexion plus stable.

L'inconvénient principal réside dans la multiplication du code. Certaines logiques d'agrégation peuvent se retrouver dupliquées entre BFF. Une discipline rigoureuse est nécessaire pour factoriser les éléments communs dans des bibliothèques partagées sans recréer un couplage excessif. La maintenance de plusieurs BFF requiert également plus de ressources que celle d'une API unique.

Un risque subtil est la dérive des BFF vers des « mini-monolithes ». Sous la pression des délais, les équipes peuvent être tentées d'ajouter de la logique métier dans le BFF plutôt que dans les services appropriés. Cette dérive dégrade progressivement l'architecture et doit être surveillée activement.

> **Anti-patron**
> Créer un BFF « générique » qui tente de servir tous les canaux défait l'objectif du patron. Si les besoins des canaux sont réellement similaires, une API unique avec paramétrage suffit. Le BFF n'a de sens que lorsque les divergences sont significatives.

> **Quand utiliser ce patron**
> *Contexte* : Organisation avec des équipes distinctes par canal (équipe mobile, équipe web) ayant des cycles de livraison indépendants et des besoins d'interface significativement différents.
> *Alternatives* : Si les besoins sont homogènes, une API unique avec versionnement peut suffire. GraphQL offre une alternative intéressante en permettant au client de spécifier précisément les données souhaitées, réduisant le besoin d'adaptation côté serveur.

#### Exemple d'Usage

Une banque offre ses services via une application mobile, un portail web et des APIs pour les agrégateurs financiers. Le BFF mobile implémente une authentification biométrique, des réponses JSON compactes et un regroupement des opérations pour fonctionner efficacement sur des connexions intermittentes. Le BFF web gère les sessions traditionnelles et fournit des réponses plus riches pour alimenter une interface complexe. Le BFF partenaire expose les données au format standardisé de l'industrie financière et implémente les contrôles réglementaires exigés pour les échanges inter-institutions.

---

### 3.2.3 Anti-Corruption Layer (ACL)

#### Définition

L'Anti-Corruption Layer, ou couche anti-corruption, est un patron de conception issu du Domain-Driven Design. Il consiste à insérer une couche de traduction entre deux systèmes dont les modèles de domaine diffèrent significativement, empêchant ainsi la « pollution » conceptuelle de l'un par l'autre.

#### Problème Résolu

Lorsqu'un système moderne doit interagir avec un système patrimonial (legacy), un défi fondamental émerge : les modèles de données et les concepts métier ont souvent évolué significativement. Le système patrimonial peut utiliser des termes obsolètes, des structures de données archaïques, des codes cryptiques dont le sens s'est perdu.

Sans protection, le nouveau système risque d'hériter de ces impuretés. Les développeurs commencent à utiliser les termes du système legacy dans leur code, à reproduire ses structures de données, à propager ses incohérences. Cette contamination conceptuelle dégrade progressivement la qualité du nouveau système et rend son évolution de plus en plus difficile.

L'ACL agit comme une barrière sanitaire. Il traduit les concepts du système externe vers le langage du domaine interne, isolant ainsi le cœur métier des impuretés extérieures.

#### Mécanisme

L'ACL se positionne à la frontière entre deux contextes délimités (bounded contexts). Il expose une interface conforme au modèle du contexte interne et implémente la traduction vers le modèle externe. Cette traduction peut être simple, comme un renommage de champs, ou complexe, impliquant des restructurations de données, des calculs de conversion, voire des appels à des services de référence.

La couche anti-corruption comprend typiquement trois éléments. Un service façade expose les opérations dans les termes du domaine interne. Un traducteur convertit les objets entre les deux modèles. Un adaptateur gère les spécificités techniques de communication avec le système externe.

```
┌─────────────────────────────────────────────────────────────┐
│                    SYSTÈME MODERNE                          │
│                                                             │
│  Utilise : Client, Commande, Article                        │
│                                                             │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│               ANTI-CORRUPTION LAYER                         │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐              │
│  │ Façade   │ -> │Traducteur│ -> │Adaptateur│              │
│  │          │    │          │    │          │              │
│  │ Client   │    │CUST->Cli │    │ COBOL    │              │
│  │ Commande │    │ORD->Cmd  │    │ Calls    │              │
│  └──────────┘    └──────────┘    └──────────┘              │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   SYSTÈME LEGACY                            │
│                                                             │
│  Utilise : CUST_MSTR, ORD_HDR, ORD_LINE                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Avantages et Inconvénients

L'ACL préserve l'intégrité conceptuelle du système moderne. Les développeurs travaillent avec un modèle propre, cohérent avec le langage métier actuel. Les changements dans le système legacy ne se propagent pas automatiquement — ils sont absorbés par la couche de traduction. Cette isolation facilite également les tests : le système moderne peut être testé indépendamment avec des simulacres de l'ACL.

À plus long terme, l'ACL facilite le remplacement du système legacy. Puisque les dépendances sont encapsulées dans une couche unique, il suffit de réimplémenter cette couche pour migrer vers un nouveau système, sans modifier le code métier.

Le coût principal est la maintenance de la couche de traduction. Chaque changement dans l'interface du système legacy ou dans le modèle interne peut nécessiter des ajustements. La complexité de traduction peut également introduire de la latence et des risques d'erreur. Il est essentiel de tester exhaustivement les règles de conversion.

Une attention particulière doit être portée aux cas limites. Les systèmes patrimoniaux ont souvent accumulé des incohérences de données au fil des décennies. Un champ « date de naissance » peut contenir des valeurs impossibles (31 février, années à deux chiffres ambiguës). L'ACL doit décider comment traiter ces anomalies : les rejeter, les normaliser, ou les signaler. Ces décisions doivent être documentées et validées avec les experts métier.

> **Note technique**
> L'ACL peut être implémenté comme un service séparé ou comme une bibliothèque intégrée au système moderne. Le choix dépend de la complexité de traduction et du nombre de systèmes consommateurs. Un service séparé permet de centraliser la logique et de la faire évoluer indépendamment. Une bibliothèque réduit la latence mais disperse la logique si plusieurs consommateurs l'utilisent.

> **Quand utiliser ce patron**
> *Contexte* : Intégration avec un système patrimonial dont le modèle de données ou les concepts métier diffèrent significativement du système moderne.
> *Alternatives* : Si les modèles sont compatibles, un simple adaptateur technique suffit. Si le système legacy doit être remplacé rapidement, investir dans l'ACL peut être un gaspillage — le Strangler Fig offre alors une meilleure approche.

#### Exemple d'Usage

Une compagnie d'assurance modernise son système de gestion des polices tout en conservant le système de sinistres des années 1990. Ce dernier utilise des codes cryptiques (« CLM_TYP = 'A2' » pour un accident automobile avec tiers), des dates au format AAAAMMJJ et des montants en centimes. L'ACL traduit ces artefacts vers le modèle moderne : un objet TypeSinistre avec des valeurs explicites, des dates ISO 8601, des montants décimaux en dollars. Le système moderne ne voit jamais les codes legacy — il manipule des concepts métier clairs.

---

### 3.2.4 Strangler Fig

#### Définition

Le Strangler Fig, nommé d'après le figuier étrangleur qui enveloppe progressivement son arbre hôte, est une stratégie de migration incrémentale. Au lieu de remplacer un système monolithique d'un coup (« big bang »), on le remplace fonctionnalité par fonctionnalité, le nouveau système « étranglant » progressivement l'ancien jusqu'à son abandon complet.

#### Problème Résolu

La modernisation des systèmes patrimoniaux est une nécessité récurrente des organisations. Ces systèmes, souvent critiques pour les opérations, ne peuvent pas être arrêtés pour une migration prolongée. Une approche « big bang » présente des risques considérables : si le nouveau système échoue au lancement, les opérations sont paralysées.

De plus, les équipes ne possèdent souvent pas une connaissance complète du comportement du système legacy. Des règles métier implicites, des cas limites non documentés, des comportements émergents se révèlent uniquement lorsque le nouveau système diverge. Une migration incrémentale permet de découvrir ces subtilités progressivement.

#### Mécanisme

Le Strangler Fig s'appuie sur une façade qui intercepte toutes les requêtes. Initialement, cette façade route tout le trafic vers le système legacy. Ensuite, fonctionnalité par fonctionnalité, on implémente des équivalents dans le nouveau système. La façade est configurée pour router les requêtes concernant les fonctionnalités migrées vers le nouveau système, et les autres vers l'ancien.

Ce routage peut se faire sur plusieurs critères : par type d'opération (les créations vers le nouveau, les lectures vers l'ancien), par utilisateur (un groupe pilote vers le nouveau), par donnée (les nouveaux clients vers le nouveau système). La flexibilité du routage permet une migration contrôlée avec possibilité de retour arrière rapide.

```
Phase 1 : Début                    Phase 2 : Migration partielle
                                 
┌─────────┐                        ┌─────────┐
│ Façade  │──100%──▶ LEGACY        │ Façade  │──70%──▶ LEGACY
└─────────┘                        │         │──30%──▶ NOUVEAU
                                   └─────────┘
                                 
Phase 3 : Migration avancée        Phase 4 : Fin

┌─────────┐                        ┌─────────┐
│ Façade  │──20%──▶ LEGACY         │ Façade  │──100%──▶ NOUVEAU
│         │──80%──▶ NOUVEAU        └─────────┘
└─────────┘                        LEGACY décommissionné
```

#### Avantages et Inconvénients

Le Strangler Fig réduit drastiquement le risque de migration. Chaque incrément est suffisamment petit pour être testé, validé et, si nécessaire, annulé. Les utilisateurs sont exposés progressivement au nouveau système, permettant de recueillir des retours et d'ajuster avant de migrer la totalité du trafic.

Cette approche permet également de livrer de la valeur plus tôt. Au lieu d'attendre des mois ou des années pour un remplacement complet, les améliorations sont disponibles dès que chaque fonctionnalité est migrée. Les équipes restent motivées par des livraisons régulières plutôt que par un objectif lointain.

L'inconvénient majeur est la complexité de gérer deux systèmes en parallèle pendant la transition. La synchronisation des données entre l'ancien et le nouveau système peut s'avérer délicate. La façade de routage ajoute également une couche de complexité et un point de défaillance potentiel.

La durée de la coexistence mérite une attention particulière. Sans discipline, la migration peut s'éterniser, laissant l'organisation avec le pire des deux mondes : la complexité du nouveau système ET les contraintes de l'ancien. Il est recommandé de définir une échéance ferme pour la décommission du legacy et de la traiter comme un engagement organisationnel, pas une aspiration.

La gestion des données pendant la transition représente un défi technique significatif. Si les deux systèmes maintiennent des copies des données, comment garantir leur synchronisation ? Plusieurs stratégies existent : le système moderne lit toujours depuis le legacy (couplage maintenu), le legacy est mis à jour via CDC lorsque le moderne écrit (complexité accrue), ou les données sont partitionnées entre les systèmes (risque d'incohérence aux frontières).

> **Bonnes pratiques**
>
> * Commencer par les fonctionnalités les plus simples et les moins critiques pour apprendre et ajuster l'approche.
> * Instrumenter abondamment la façade pour détecter rapidement toute divergence de comportement entre les deux systèmes.
> * Définir des critères clairs de succès pour chaque incrément avant de migrer le trafic.
> * Établir un calendrier de décommission et le respecter pour éviter l'enlisement.

> **Quand utiliser ce patron**
> *Contexte* : Modernisation d'un système critique qui ne peut pas être arrêté pour une migration complète. Particulièrement approprié lorsque la connaissance du système legacy est incomplète.
> *Alternatives* : Pour des systèmes non critiques ou des équipes très confiantes dans leur compréhension, une migration directe peut être plus rapide. Pour des systèmes très volumineux, une approche par « branches » (plusieurs Stranglers en parallèle) peut accélérer la migration.

#### Exemple d'Usage

Un détaillant remplace son système de gestion des inventaires monolithique par une architecture de microservices. La migration commence par le module de consultation des stocks, considéré comme simple et non critique. Une façade intercepte les appels : les requêtes de consultation sont routées vers le nouveau service, les autres vers le monolithe. Après validation, le module de réservation est migré, puis celui de réapprovisionnement. Après dix-huit mois, le monolithe ne traite plus que quelques fonctions marginales, puis est finalement décommissionné.

---

### 3.2.5 Aggregator Pattern

#### Définition

L'Aggregator Pattern est un patron de composition où un service collecte les données de plusieurs services en aval et les assemble en une réponse unique pour le client. Il réduit le bavardage réseau en consolidant plusieurs appels en un seul du point de vue du consommateur.

#### Problème Résolu

Dans une architecture de microservices, les données requises pour une opération métier sont souvent distribuées entre plusieurs services. Afficher la page d'un produit peut nécessiter d'interroger le service catalogue pour les détails, le service inventaire pour la disponibilité, le service prix pour les tarifs, et le service avis pour les évaluations clients.

Si le client doit effectuer ces quatre appels lui-même, plusieurs problèmes émergent. La latence totale augmente car les appels sont séquentiels (ou complexes à paralléliser côté client). Le trafic réseau se multiplie, particulièrement problématique pour les clients mobiles. Le client doit comprendre la structure interne du système et gérer les incohérences entre services. Enfin, chaque nouveau client doit réimplémenter cette logique d'agrégation.

#### Mécanisme

L'Aggregator expose un point d'entrée unique qui encapsule la complexité de collecte. Lorsqu'il reçoit une requête, il détermine quels services en aval doivent être interrogés. Il peut effectuer ces appels en parallèle pour minimiser la latence totale. Il collecte les réponses, gère les erreurs partielles et assemble un résultat cohérent.

L'agrégateur peut implémenter différentes stratégies face aux erreurs. En mode strict, l'échec d'un service en aval fait échouer l'ensemble. En mode tolérant, l'agrégateur retourne les données disponibles et signale les parties manquantes. Le choix dépend des exigences métier et de la criticité de chaque source.

```
CLIENT                              AGGREGATOR                    SERVICES
  │                                     │                            │
  │─────── GET /product/123 ───────────▶│                            │
  │                                     │──── GET /catalog/123 ─────▶│
  │                                     │◀─── {name, desc} ──────────│
  │                                     │                            │
  │                                     │──── GET /inventory/123 ───▶│
  │                                     │◀─── {qty: 42} ─────────────│
  │                                     │                            │
  │                                     │──── GET /price/123 ───────▶│
  │                                     │◀─── {price: 29.99} ────────│
  │                                     │                            │
  │◀───── {name, desc, qty, price} ────│                            │
  │                                     │                            │
```

#### Avantages et Inconvénients

L'agrégateur simplifie considérablement l'expérience client. Un seul appel remplace plusieurs, réduisant la latence perçue et la complexité de développement côté consommateur. L'interface publique reste stable même si l'architecture interne évolue — un nouveau service peut être ajouté sans que les clients ne le remarquent.

La parallélisation des appels internes peut offrir de meilleures performances qu'une exécution séquentielle côté client. L'agrégateur peut également implémenter une mise en cache intelligente, réduisant la charge sur les services en aval pour les données fréquemment demandées.

En contrepartie, l'agrégateur devient un point de couplage central. Il doit connaître tous les services qu'il orchestre et évoluer avec eux. Une modification de l'interface d'un service en aval peut nécessiter une mise à jour de l'agrégateur. La gestion des erreurs partielles ajoute de la complexité et nécessite des décisions de conception explicites.

L'agrégateur peut également devenir un goulot d'étranglement si mal dimensionné. Contrairement aux services en aval qui peuvent être mis à l'échelle indépendamment, l'agrégateur doit absorber l'ensemble du trafic destiné aux données composites. Une attention particulière à la mise en cache, à l'exécution parallèle et au dimensionnement est nécessaire.

> **Anti-patron**
> Implémenter de la logique métier dans l'agrégateur. Son rôle est la composition et la transformation de données, pas le traitement métier. Si l'agrégateur commence à valider des règles ou à modifier l'état, il devient un point de couplage difficile à maintenir. Un autre anti-patron consiste à créer un « super-agrégateur » qui tente de servir trop de cas d'usage différents — mieux vaut plusieurs agrégateurs spécialisés.

> **Quand utiliser ce patron**
> *Contexte* : Interface utilisateur nécessitant des données de plusieurs services pour une seule vue. Clients mobiles ou à faible bande passante où la réduction des appels est critique.
> *Alternatives* : GraphQL permet au client de spécifier exactement les données souhaitées, déplaçant la logique d'agrégation côté serveur de manière déclarative. Pour des cas simples, les appels parallèles côté client peuvent suffire.

#### Exemple d'Usage

Une application de voyage affiche les détails d'une réservation. L'agrégateur interroge simultanément le service vols pour les horaires, le service hôtels pour l'hébergement, le service voitures pour la location, et le service profil pour les préférences du voyageur. Il assemble ces informations en une réponse cohérente. Si le service voitures est temporairement indisponible, l'agrégateur retourne les informations disponibles avec une indication que les détails de location sont temporairement inaccessibles.

---

### 3.2.6 Consumer-Driven Contracts

#### Définition

Consumer-Driven Contracts est une approche de test et de documentation où les consommateurs d'une API définissent leurs attentes sous forme de contrats exécutables. Ces contrats servent ensuite à valider que le producteur respecte les besoins de ses consommateurs, inversant ainsi le contrôle de la validation.

#### Problème Résolu

Dans une architecture de microservices, un service producteur peut avoir de nombreux consommateurs. Lorsque le producteur modifie son interface, comment garantir qu'aucun consommateur n'est affecté ? Les tests d'intégration traditionnels nécessitent de déployer l'ensemble des services, ce qui est lent et fragile. La documentation peut dériver par rapport à l'implémentation réelle.

Le problème s'aggrave avec l'échelle. Chaque consommateur utilise potentiellement un sous-ensemble différent de l'API, avec des hypothèses spécifiques sur les types, les valeurs possibles, les comportements d'erreur. Le producteur ne peut pas tester toutes ces combinaisons manuellement.

#### Mécanisme

Chaque consommateur définit un contrat qui spécifie les requêtes qu'il effectue et les réponses qu'il attend. Ce contrat est écrit dans un format exécutable (Pact, Spring Cloud Contract, etc.) et versionné avec le code du consommateur.

Les contrats sont ensuite partagés avec le producteur, typiquement via un « broker » centralisé. Le pipeline d'intégration continue du producteur récupère tous les contrats de ses consommateurs et les exécute contre son implémentation. Si un contrat échoue, le producteur sait qu'une modification proposée casserait un consommateur existant.

```
CONSOMMATEUR A                     BROKER                      PRODUCTEUR
      │                               │                             │
      │──── Publie contrat A ────────▶│                             │
      │                               │                             │
      │                               │                             │
CONSOMMATEUR B                        │                             │
      │                               │                             │
      │──── Publie contrat B ────────▶│                             │
      │                               │                             │
      │                               │◀──── Récupère contrats ────│
      │                               │                             │
      │                               │────── contrats A, B ───────▶│
      │                               │                             │
      │                               │                     ┌───────┴───────┐
      │                               │                     │ Exécute tests │
      │                               │                     │ contre impl.  │
      │                               │                     └───────────────┘
```

#### Avantages et Inconvénients

Consumer-Driven Contracts détecte les ruptures de compatibilité avant le déploiement en production. Le producteur reçoit un signal clair et précoce lorsqu'une modification affecte un consommateur. Cette approche encourage également une API minimaliste : les contrats documentent ce qui est réellement utilisé, permettant d'identifier et de supprimer les fonctionnalités obsolètes.

Les contrats servent de documentation vivante. Contrairement à une spécification statique, ils sont garantis d'être à jour puisqu'ils font partie du pipeline de validation. Les nouveaux développeurs peuvent comprendre les attentes réelles en examinant les contrats plutôt qu'une documentation potentiellement désuète.

L'inconvénient principal est l'effort de mise en place et de maintenance. Chaque consommateur doit rédiger et maintenir ses contrats. Le broker central devient une dépendance de l'infrastructure de développement. L'approche fonctionne mieux dans un contexte où les équipes contrôlent à la fois les producteurs et les consommateurs — avec des consommateurs externes, la collecte des contrats est plus complexe.

Il existe également une tension entre exhaustivité et maintenabilité. Des contrats trop détaillés deviennent fragiles — le moindre changement de format les fait échouer. Des contrats trop lâches ratent des régressions significatives. Trouver le bon niveau de spécificité requiert de l'expérience et des ajustements itératifs.

> **Note technique**
> Pact est l'outil le plus répandu pour les Consumer-Driven Contracts. Il supporte de nombreux langages (Java, JavaScript, Python, Go, etc.) et offre un broker open source ou hébergé. Spring Cloud Contract est une alternative populaire dans l'écosystème Spring. Pour les APIs GraphQL, des outils spécifiques comme Apollo Contract Tests sont disponibles.

> **Quand utiliser ce patron**
> *Contexte* : Équipes multiples développant des services interdépendants avec des cycles de livraison indépendants. Particulièrement utile lorsque les tests d'intégration complets sont lents ou fragiles.
> *Alternatives* : Pour des équipes co-localisées avec une forte communication, des revues de changements d'API peuvent suffire. Les tests d'intégration dans un environnement partagé restent complémentaires pour valider les comportements de bout en bout.

#### Exemple d'Usage

Une plateforme de paiement est consommée par trois services : facturation, remboursements et rapports. Chaque équipe définit un contrat Pact spécifiant les appels qu'elle effectue. Le service facturation attend un champ « transactionId » dans la réponse. Le service rapports attend un champ « timestamp » au format ISO 8601. Lorsque l'équipe paiement envisage de renommer « transactionId » en « txnId », les tests de contrat échouent immédiatement, signalant que le service facturation serait affecté. L'équipe peut alors coordonner la migration ou maintenir les deux champs pendant une période de transition.

---

### 3.2.7 Service Registry & Discovery

#### Définition

Le Service Registry & Discovery est un mécanisme permettant aux services de s'enregistrer dynamiquement et de découvrir les instances disponibles d'autres services au moment de l'exécution. Il élimine le besoin de configurer statiquement les adresses des dépendances.

#### Problème Résolu

Dans un environnement infonuagique moderne, les instances de services sont éphémères. Les conteneurs sont créés et détruits à la demande, les adresses IP changent, de nouvelles instances apparaissent lors des montées en charge et disparaissent lors des réductions. Configurer statiquement les adresses des dépendances devient impraticable.

Sans découverte dynamique, chaque modification de la topologie du système nécessite une mise à jour de configuration et un redéploiement. Cette rigidité contredit les promesses d'élasticité et d'agilité des architectures de microservices. Un mécanisme de découverte dynamique est essentiel pour exploiter pleinement les environnements infonuagiques.

#### Mécanisme

Le registre de services maintient un catalogue des instances disponibles. Lorsqu'une instance de service démarre, elle s'enregistre auprès du registre en fournissant son identité (nom du service) et sa localisation (adresse, port). Elle envoie ensuite régulièrement des signaux de vie (heartbeats) pour confirmer sa disponibilité. Si les signaux cessent, le registre considère l'instance comme défaillante et la retire du catalogue.

Lorsqu'un service client doit appeler une dépendance, il interroge le registre pour obtenir la liste des instances disponibles. Il peut alors appliquer une stratégie de répartition de charge côté client (round-robin, pondération, affinité) pour sélectionner une instance. Cette approche décentralisée évite qu'un répartiteur de charge central ne devienne un goulot d'étranglement.

```
┌─────────────┐      Enregistrement       ┌─────────────────┐
│ Service A   │──────────────────────────▶│                 │
│ Instance 1  │                           │    SERVICE      │
└─────────────┘                           │    REGISTRY     │
                                          │                 │
┌─────────────┐      Enregistrement       │  ┌───────────┐  │
│ Service A   │──────────────────────────▶│  │ A: [1,2]  │  │
│ Instance 2  │                           │  │ B: [1,2,3]│  │
└─────────────┘                           │  └───────────┘  │
                                          │                 │
┌─────────────┐      Découverte A?        │                 │
│ Service B   │──────────────────────────▶│                 │
│ Instance 1  │◀──────────────────────────│                 │
└─────────────┘      [A:1, A:2]           └─────────────────┘
```

#### Avantages et Inconvénients

La découverte dynamique permet une véritable élasticité. Les instances peuvent être ajoutées ou retirées sans reconfiguration des consommateurs. Le système s'adapte automatiquement aux changements de topologie, que ce soit pour absorber une charge accrue ou pour remplacer une instance défaillante.

Cette approche favorise également les déploiements sans interruption. De nouvelles versions d'un service peuvent être déployées progressivement : les nouvelles instances s'enregistrent, commencent à recevoir du trafic, et les anciennes sont retirées graduellement. Les stratégies de déploiement canari et bleu-vert deviennent naturelles.

Le registre de services introduit cependant une dépendance critique. Si le registre devient indisponible, les services ne peuvent plus découvrir leurs dépendances. Une haute disponibilité du registre est essentielle, typiquement via une réplication multi-nœuds avec consensus distribué (Raft, Paxos). La cohérence éventuelle du registre peut également causer des problèmes transitoires lors des changements rapides de topologie.

Les environnements Kubernetes ont popularisé une approche alternative basée sur le DNS interne. Les services sont accessibles via des noms DNS stables (par exemple, `catalogue.default.svc.cluster.local`) résolus automatiquement par le système. Cette approche simplifie la découverte mais offre moins de flexibilité que les registres dédiés pour la répartition de charge côté client ou les métadonnées enrichies.

> **Bonnes pratiques**
>
> * Implémenter une mise en cache côté client des résultats de découverte pour tolérer une indisponibilité temporaire du registre.
> * Configurer des délais de heartbeat et d'expiration appropriés : trop courts génèrent du bruit, trop longs retardent la détection des pannes.
> * Intégrer les vérifications de santé au registre pour ne pas router vers des instances démarrées mais pas encore prêtes.
> * Prévoir une stratégie de repli statique en cas de défaillance complète du registre.

> **Quand utiliser ce patron**
> *Contexte* : Environnements où les instances de services sont créées et détruites dynamiquement (conteneurs, fonctions sans serveur, auto-scaling). Architectures avec de nombreux services interdépendants.
> *Alternatives* : Pour des architectures simples avec peu de services stables, une configuration DNS traditionnelle peut suffire. Dans Kubernetes, le DNS interne couvre la plupart des besoins sans registre explicite.

#### Exemple d'Usage

Une plateforme de commerce utilise Consul comme registre de services. Le service catalogue déploie cinq instances qui s'enregistrent automatiquement au démarrage. Le service panier, lorsqu'il doit vérifier un produit, interroge Consul pour obtenir les instances catalogue disponibles. L'agent Consul local maintient un cache mis à jour en temps réel. Lors d'un pic de trafic, Kubernetes ajoute automatiquement trois instances catalogue qui s'enregistrent et commencent à recevoir du trafic en quelques secondes, sans aucune intervention manuelle.

---

## 3.3 Matrice de Décision des Patrons

Le choix du patron approprié dépend du contexte spécifique. Le tableau suivant synthétise les critères de sélection pour guider les décisions architecturales.

| Patron           | Contexte d'utilisation               | Couplage induit          | Complexité      | Alternatives                      |
| ---------------- | ------------------------------------ | ------------------------ | ---------------- | --------------------------------- |
| API Gateway      | Exposition externe, multiple clients | Modéré (point central) | Moyenne          | Répartiteur de charge simple     |
| BFF              | Canaux aux besoins divergents        | Faible (par canal)       | Moyenne à haute | API générique avec paramétrage |
| ACL              | Intégration avec legacy             | Faible (isolation)       | Haute            | Réécriture complète            |
| Strangler Fig    | Migration progressive                | Variable (transition)    | Haute            | Big bang (risqué)                |
| Aggregator       | Composition de données distribuées | Modéré (orchestration) | Moyenne          | Appels directs par le client      |
| CDC Contracts    | Évolution d'APIs partagées         | Faible (contractuel)     | Moyenne          | Tests d'intégration classiques   |
| Service Registry | Environnement dynamique              | Faible (découverte)     | Moyenne          | Configuration statique            |

> **Décision architecturale**
> *Contexte* : Choix entre API Gateway unique et multiples BFF.
> *Options* : (A) Gateway unique avec paramétrage par client; (B) Gateway + BFF par canal; (C) BFF sans gateway central.
> *Décision* : L'option B est recommandée pour les organisations matures avec des canaux aux besoins significativement différents. L'option A suffit pour des besoins homogènes. L'option C est appropriée uniquement si les préoccupations transversales sont gérées ailleurs (service mesh).

---

## 3.4 Combinaison des Patrons

Les patrons présentés ne s'utilisent pas isolément. Une architecture d'intégration complète les combine pour adresser différentes préoccupations. Cette section illustre quelques combinaisons fréquentes et les synergies qu'elles créent.

### 3.4.1 Gateway + BFF + Aggregator

Une architecture d'entreprise typique positionne l'API Gateway en première ligne pour les préoccupations de sécurité et de gestion du trafic. Derrière, des BFF adaptent l'interface à chaque canal. Ces BFF utilisent eux-mêmes l'Aggregator Pattern pour composer les données de plusieurs microservices.

Cette combinaison offre une séparation claire des responsabilités : la gateway gère l'infrastructure, les BFF gèrent l'adaptation par canal, les agrégateurs gèrent la composition de données. Chaque couche peut évoluer indépendamment selon ses contraintes propres.

```
┌────────────────────────────────────────────────────────────────┐
│                          CLIENTS                                │
│     Mobile              Web                 Partenaires         │
└─────────┬───────────────┬───────────────────────┬──────────────┘
          │               │                       │
          ▼               ▼                       ▼
┌────────────────────────────────────────────────────────────────┐
│                       API GATEWAY                               │
│          Authentification, Rate Limiting, TLS                   │
└─────────┬───────────────┬───────────────────────┬──────────────┘
          │               │                       │
          ▼               ▼                       ▼
    ┌──────────┐    ┌──────────┐           ┌──────────┐
    │ BFF      │    │ BFF      │           │ BFF      │
    │ Mobile   │    │ Web      │           │ B2B      │
    └────┬─────┘    └────┬─────┘           └────┬─────┘
         │               │                      │
         └───────────────┼──────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │    AGGREGATOR       │
              │    (si nécessaire)  │
              └──────────┬──────────┘
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
   ┌──────────┐    ┌──────────┐    ┌──────────┐
   │ Service  │    │ Service  │    │ Service  │
   │ Produit  │    │ Stock    │    │ Prix     │
   └──────────┘    └──────────┘    └──────────┘
```

### 3.4.2 ACL + Strangler Fig

Lors de la migration d'un système legacy, l'ACL et le Strangler Fig se complètent naturellement. L'ACL isole le modèle moderne des concepts legacy pendant que le Strangler Fig route progressivement le trafic. À mesure que des fonctionnalités sont migrées, l'ACL se simplifie puisqu'il communique de moins en moins avec le système legacy.

Cette combinaison permet une migration maîtrisée tout en préservant l'intégrité conceptuelle du nouveau système dès le départ. L'équipe peut développer avec un modèle propre sans attendre la fin de la migration.

Un point subtil mérite attention : l'ACL doit être positionné du bon côté de la façade Strangler. Si l'ACL est dans le nouveau système, chaque fonctionnalité migrée bénéficie immédiatement de l'isolation. Si l'ACL est partagé entre les deux, il peut devenir un goulot d'étranglement de la migration.

### 3.4.3 Service Registry + Consumer-Driven Contracts

Dans un écosystème de microservices dynamique, le Service Registry permet la découverte des instances tandis que les Consumer-Driven Contracts garantissent la compatibilité des interfaces. Le registre répond à la question « où est le service ? » tandis que les contrats répondent à « est-ce que le service répond correctement ? ».

Ces deux mécanismes peuvent être intégrés plus profondément. Le broker de contrats peut enrichir le registre avec des métadonnées de compatibilité. Un consommateur peut alors filtrer les instances non seulement par disponibilité mais aussi par compatibilité de contrat, évitant de router vers une version du service dont l'interface a changé de manière incompatible.

### 3.4.4 Gateway + Service Registry + Circuit Breaker

Cette combinaison, que nous approfondirons au chapitre VII, illustre comment les patrons d'intégration et les patrons de résilience s'entrelacent. La gateway utilise le registre pour découvrir les instances disponibles, et le circuit breaker (un patron de résilience) pour éviter les instances défaillantes.

Le circuit breaker peut même informer le registre de l'état de santé observé, créant une boucle de rétroaction. Si un circuit s'ouvre répétitivement vers une instance, le registre peut la marquer comme dégradée et réduire le trafic qui lui est envoyé.

### 3.4.5 Considérations sur la Superposition des Patrons

La combinaison de patrons n'est pas sans risque. Chaque couche ajoutée introduit de la latence, de la complexité opérationnelle et des points de défaillance potentiels. L'architecte doit résister à la tentation de tout mettre en œuvre simultanément.

Une approche pragmatique consiste à commencer avec le minimum nécessaire et à ajouter des couches lorsque la douleur devient tangible. Une gateway simple sans BFF peut suffire initialement. L'ajout de BFF se justifie lorsque les équipes par canal commencent à s'entraver mutuellement. L'ACL devient nécessaire lorsque le modèle legacy commence à contaminer le code moderne.

Cette progression incrémentale permet d'accumuler l'expertise nécessaire pour opérer chaque patron avant d'en ajouter un nouveau. Elle évite également le piège de la sur-ingénierie précoce pour des problèmes hypothétiques qui ne se matérialiseront peut-être jamais.

---

## 3.5 Technologies d'Implémentation

Cette section présente brièvement les technologies couramment utilisées pour implémenter les patrons décrits. Le choix d'un outil spécifique dépend du contexte technique, des compétences de l'équipe et de l'écosystème existant.

### 3.5.1 API Gateways

Plusieurs solutions matures existent pour implémenter une API Gateway. Kong, basé sur NGINX, offre une extensibilité via des plugins Lua et une version entreprise avec fonctionnalités avancées. AWS API Gateway s'intègre naturellement à l'écosystème Amazon et offre des options serverless. Azure API Management et Google Apigee couvrent les besoins similaires dans leurs écosystèmes respectifs. Pour les architectures Kubernetes, l'Ingress Controller avec des annotations peut assumer certaines fonctions de gateway, tandis que des solutions comme Traefik ou Ambassador offrent des capacités plus avancées.

### 3.5.2 Service Discovery

Consul de HashiCorp combine registre de services, configuration distribuée et maillage de service dans une solution unifiée. Eureka, développé par Netflix, reste populaire dans l'écosystème Spring Cloud. Pour les déploiements Kubernetes natifs, le DNS interne (CoreDNS) et les Services Kubernetes fournissent une découverte de base suffisante pour de nombreux cas d'usage.

### 3.5.3 Consumer-Driven Contracts

Pact demeure la référence pour les contrats pilotés par les consommateurs, avec un support multi-langage et un broker pour centraliser les contrats. Spring Cloud Contract offre une alternative idiomatique pour les équipes Java/Spring. Pour les APIs GraphQL, des outils comme Apollo Studio intègrent la validation de schéma dans le flux de travail.

### 3.5.4 Considérations de Mise en Œuvre

L'adoption de ces patrons ne se fait pas en vase clos. L'observabilité (que nous approfondirons au chapitre VII) est un prérequis : sans visibilité sur les appels entre services, le débogage devient rapidement impossible. L'infrastructure as code facilite la reproductibilité et l'évolution des configurations. L'intégration continue automatise la validation des contrats et des tests d'intégration.

> **Note technique**
> La tendance actuelle favorise les solutions « cloud-native » qui s'intègrent aux orchestrateurs de conteneurs. Cependant, la portabilité entre fournisseurs infonuagiques reste un défi. Les standards ouverts comme OpenAPI, CloudEvents et OpenTelemetry réduisent ce risque en découplant la spécification de l'implémentation.

---

## Conclusion et Transition

Ce chapitre a exploré le premier domaine de notre continuum d'intégration : l'intégration des applications, métaphoriquement désignée comme « le Verbe ». Nous avons d'abord caractérisé ses enjeux spécifiques : le couplage temporel fort qui enchaîne les disponibilités, les dépendances directes qui complexifient l'évolution, les défis de coordination dans les processus distribués, et la surface d'attaque sécuritaire qui s'étend avec chaque nouveau point d'intégration.

Les sept patrons présentés constituent une boîte à outils éprouvée pour adresser ces défis. L'API Gateway centralise les préoccupations transversales et simplifie l'accès en offrant un point d'entrée unique. Le Backend for Frontend adapte les interfaces aux besoins spécifiques de chaque canal, permettant aux équipes d'évoluer indépendamment. L'Anti-Corruption Layer préserve l'intégrité conceptuelle face aux systèmes patrimoniaux en traduisant leurs artefacts vers un modèle moderne. Le Strangler Fig permet des migrations progressives et contrôlées sans risquer un « big bang » catastrophique. L'Aggregator Pattern réduit le bavardage réseau en composant les données de multiples sources. Les Consumer-Driven Contracts garantissent la compatibilité lors des évolutions en inversant le contrôle de validation. Enfin, le Service Registry & Discovery permet l'élasticité dynamique indispensable aux environnements infonuagiques modernes.

Ces patrons partagent une caractéristique commune : ils opèrent dans un mode fondamentalement synchrone. L'appelant invoque, attend, reçoit une réponse. Ce couplage temporel, bien que maîtrisé par les patrons présentés, demeure une contrainte intrinsèque de l'intégration applicative. Certains contextes métier exigent cette synchronicité — une transaction de paiement ne peut pas s'accommoder d'une réponse différée. D'autres contextes, cependant, peuvent bénéficier d'approches alternatives.

Le chapitre suivant explore le deuxième domaine de notre continuum : l'intégration des données, que nous désignons comme « le Nom ». Là où ce chapitre s'est concentré sur l'action et l'invocation, le prochain se préoccupera de l'état, de la cohérence et de l'accessibilité de l'information. Nous verrons comment des patrons tels que Change Data Capture (CDC), Data Virtualization et CQRS permettent de synchroniser les données entre systèmes tout en préservant leur autonomie. Cette transition nous fera progresser sur le continuum vers un couplage moins direct, préparant le terrain pour l'intégration des événements qui achèvera notre exploration des trois domaines fondamentaux.

L'architecte averti notera que les frontières entre ces domaines ne sont pas étanches. Un flux d'intégration réel combine souvent des éléments des trois domaines. Le patron CDC, que nous aborderons au chapitre IV, transforme des changements de données en événements — illustrant comment l'intégration des données (le Nom) peut alimenter l'intégration des événements (le Signal). Cette perméabilité justifie l'approche de continuum plutôt qu'une taxonomie rigide.

---

## Résumé du Chapitre III

**Thème central** : L'intégration des applications (« Le Verbe ») adresse l'orchestration des processus et les interactions synchrones entre systèmes. Elle constitue le premier domaine du continuum d'intégration, caractérisé par un couplage temporel inhérent où l'appelant attend la réponse de l'appelé.

**Enjeux identifiés** :

* Le couplage temporel crée des chaînes de dépendance où les défaillances se propagent en cascade
* Les dépendances directes compliquent l'évolution indépendante des systèmes et nécessitent une coordination inter-équipes
* La coordination des appels distribués soulève des défis de cohérence transactionnelle sans les garanties ACID traditionnelles
* La multiplication des points d'entrée étend la surface d'attaque sécuritaire et exige une approche Zero Trust

**Patrons présentés** :

| Patron                       | Rôle principal                                                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| API Gateway                  | Point d'entrée unifié avec préoccupations transversales centralisées (authentification, rate limiting, routage) |
| Backend for Frontend         | Adaptation de l'interface aux besoins spécifiques de chaque canal (mobile, web, B2B)                               |
| Anti-Corruption Layer        | Isolation conceptuelle face aux systèmes patrimoniaux via traduction de modèles                                   |
| Strangler Fig                | Migration progressive du monolithe vers les microservices par étranglement fonctionnel                             |
| Aggregator Pattern           | Composition de données distribuées en réponse unique pour réduire le bavardage réseau                          |
| Consumer-Driven Contracts    | Validation des interfaces par les attentes des consommateurs via contrats exécutables                              |
| Service Registry & Discovery | Localisation dynamique des services en environnement élastique avec heartbeats et mise en cache                    |

**Position dans le continuum** : L'intégration des applications représente l'extrémité « couplage fort » du continuum. Les patrons présentés atténuent ce couplage sans l'éliminer — le caractère synchrone demeure fondamental à ce domaine. Cette contrainte est parfois incontournable (transactions de paiement, validations en temps réel) mais peut souvent être relaxée lorsque les exigences métier le permettent.

**Combinaisons clés** : Les patrons se combinent naturellement — Gateway + BFF + Aggregator pour une architecture multi-canal complète, ACL + Strangler Fig pour une migration maîtrisée, Service Registry + Consumer-Driven Contracts pour un écosystème dynamique et fiable.

**Transition** : Le chapitre IV explorera l'intégration des données (« Le Nom »), progressant vers un couplage intermédiaire où la préoccupation principale devient la cohérence de l'état plutôt que l'orchestration des processus. Les patrons CDC, CQRS et Data Mesh illustreront comment maintenir des vues cohérentes de données distribuées.
