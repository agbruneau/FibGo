
# Chapitre VII — Patrons Transversaux de Résilience et Observabilité

*Focus : La robustesse face aux pannes et la visibilité sur les systèmes distribués.*

---

## Introduction

Les chapitres précédents ont exploré les trois domaines fondamentaux de l'intégration : les applications (le Verbe), les données (le Nom) et les événements (le Signal). Chacun de ces domaines propose des patrons architecturaux qui répondent à des besoins spécifiques d'orchestration, de cohérence ou de réactivité. Toutefois, ces patrons partagent une vulnérabilité commune : ils opèrent dans un environnement distribué où les pannes ne sont pas l'exception mais la norme.

Le présent chapitre aborde deux préoccupations transversales qui traversent l'ensemble du continuum d'intégration : la **résilience** et l' **observabilité** . Ces préoccupations ne constituent pas un quatrième domaine d'intégration ; elles forment plutôt le tissu conjonctif qui permet aux trois domaines de fonctionner de manière fiable en production. Sans mécanismes de résilience, une panne locale se propage en cascade jusqu'à paralyser l'ensemble du système. Sans observabilité, les équipes opérationnelles naviguent à l'aveugle, incapables de diagnostiquer les problèmes ou d'anticiper les dégradations.

La distinction entre résilience et observabilité mérite clarification dès l'entrée en matière. La résilience désigne la capacité d'un système à continuer de fonctionner — éventuellement en mode dégradé — malgré les défaillances de certains de ses composants. Elle répond à la question : « Comment survivre aux pannes ? » L'observabilité désigne la capacité de comprendre l'état interne d'un système à partir de ses sorties externes. Elle répond à la question : « Que se passe-t-il réellement dans le système ? »

> **Définition formelle**
> **Résilience** : Capacité d'un système distribué à absorber les perturbations, à maintenir un niveau de service acceptable pendant les défaillances et à récupérer rapidement vers un état nominal.

> **Définition formelle**
> **Observabilité** : Propriété d'un système qui permet de déduire son état interne à partir de la collecte et de l'analyse de ses sorties externes (traces, métriques, journaux).

Ces deux préoccupations entretiennent une relation symbiotique. La résilience sans observabilité produit des systèmes qui survivent aux pannes mais dont personne ne comprend le comportement — les problèmes s'accumulent silencieusement jusqu'à une défaillance catastrophique. L'observabilité sans résilience produit des systèmes parfaitement transparents mais fragiles — les équipes voient les problèmes survenir sans pouvoir les contenir.

Ce chapitre s'organise en trois sections principales. La première présente les patrons de résilience applicative : Circuit Breaker, Retry, Bulkhead, Timeout et Fallback. Ces patrons, implémentés au niveau du code applicatif, constituent la première ligne de défense contre les défaillances. La deuxième section examine l'infrastructure de résilience : le patron Sidecar, le Service Mesh pour le trafic synchrone et l'Event Mesh pour le trafic asynchrone. Ces composants d'infrastructure déchargent les applications de certaines préoccupations transversales. La troisième section explore les trois piliers de l'observabilité — traces, métriques et journaux — ainsi que le standard OpenTelemetry qui les unifie.

L'objectif n'est pas de fournir des tutoriels de configuration pour des outils spécifiques, mais d'établir les principes architecturaux qui guident le choix et l'assemblage de ces mécanismes. Les technologies évoluent rapidement ; les principes demeurent.

---

## 7.1 Patrons de Résilience Applicative

Les systèmes distribués échouent de manières que les systèmes monolithiques ignorent. Un appel réseau peut échouer, expirer ou retourner une réponse après un délai imprévisible. Un service distant peut être surchargé, partiellement défaillant ou totalement indisponible. Ces modes de défaillance, absents d'une application monolithique où les appels de fonction sont instantanés et fiables, exigent des mécanismes de protection explicites.

Les cinq patrons présentés dans cette section forment un arsenal cohérent. Ils ne sont pas mutuellement exclusifs ; une implémentation robuste les combine typiquement selon une stratégie de défense en profondeur.

### 7.1.1 Circuit Breaker

#### Définition

Le Circuit Breaker (disjoncteur) est un patron qui protège un système contre les défaillances en cascade en interrompant temporairement les appels vers un service défaillant. Inspiré des disjoncteurs électriques qui coupent le courant pour prévenir les surcharges, ce patron évite qu'un service en difficulté ne soit submergé de requêtes vouées à l'échec.

#### Problème Résolu

Considérons un service A qui appelle un service B. Si B devient lent ou indisponible, A accumule des requêtes en attente. Ses threads ou connexions se saturent. Bientôt, A devient lui-même incapable de répondre à ses propres appelants. La défaillance de B s'est propagée à A. Si d'autres services dépendent de A, la cascade se poursuit.

Sans protection, un service défaillant dans une chaîne de dépendances peut paralyser l'ensemble du système. Le Circuit Breaker interrompt cette propagation en « ouvrant le circuit » après un nombre défini d'échecs, empêchant les appels futurs d'atteindre le service défaillant.

#### Mécanisme

Le Circuit Breaker maintient trois états possibles. L'état **fermé** (Closed) représente le fonctionnement normal : les requêtes passent vers le service distant, et le disjoncteur comptabilise les succès et les échecs. Lorsque le taux d'échec dépasse un seuil configurable (par exemple, 50 % des requêtes échouent sur une fenêtre de 10 secondes), le disjoncteur passe en état **ouvert** (Open).

En état ouvert, toutes les requêtes échouent immédiatement sans atteindre le service distant. Cette « défaillance rapide » (fail-fast) libère les ressources de l'appelant et évite d'aggraver la situation du service défaillant. Après un délai configurable (par exemple, 30 secondes), le disjoncteur passe en état **semi-ouvert** (Half-Open).

En état semi-ouvert, le disjoncteur laisse passer un nombre limité de requêtes « tests » vers le service distant. Si ces requêtes réussissent, le disjoncteur revient en état fermé. Si elles échouent, il retourne en état ouvert pour une nouvelle période d'attente.

```
           Échecs < seuil          Échecs ≥ seuil
    ┌─────────────────────┐    ┌─────────────────────┐
    │                     │    │                     │
    ▼         Succès      │    ▼      Délai expiré  │
┌────────┐ ◄──────────┐ ┌────────┐ ────────────► ┌────────────┐
│ FERMÉ  │            │ │ OUVERT │               │ SEMI-OUVERT│
│        │ ───────────┼─┤        │ ◄──────────── │            │
└────────┘  Échecs ≥  │ └────────┘    Échec      └────────────┘
            seuil     │      │           │              │
                      │      │           │              │
                      │      └───────────┘              │
                      │      Requêtes échouent         │
                      │      immédiatement              │
                      │                                 │
                      └─────────────────────────────────┘
                              Succès sur requêtes test
```

#### Paramétrage

Le paramétrage du Circuit Breaker requiert une compréhension du contexte applicatif. Plusieurs variables entrent en jeu.

Le **seuil d'ouverture** définit le taux d'échec qui déclenche l'ouverture du circuit. Un seuil trop bas provoque des ouvertures intempestives pour des erreurs transitoires. Un seuil trop élevé laisse le système se dégrader avant d'intervenir. Un seuil de 50 % sur une fenêtre glissante de 10 requêtes constitue un point de départ raisonnable.

La **durée d'ouverture** définit combien de temps le circuit reste ouvert avant de tenter une récupération. Une durée trop courte sollicite prématurément un service qui n'a pas eu le temps de se rétablir. Une durée trop longue prive les utilisateurs d'un service qui a peut-être récupéré. Des valeurs de 30 à 60 secondes sont courantes.

Le **nombre de requêtes en semi-ouvert** définit combien de requêtes tests sont autorisées. Trop peu de requêtes rend la décision statistiquement fragile. Trop de requêtes risque de surcharger un service en cours de récupération.

> **Configuration recommandée**
> Pour un service d'API typique :
>
> * Seuil d'ouverture : 50 % d'échecs sur 20 requêtes minimum
> * Durée d'ouverture : 30 secondes
> * Requêtes en semi-ouvert : 5 requêtes
> * Types d'erreur comptabilisés : timeouts, erreurs 5xx, exceptions de connexion
> * Exclure : erreurs 4xx (erreurs client, pas défaillance du service)

#### Avantages et Inconvénients

Le Circuit Breaker offre une protection immédiate contre les cascades de défaillances. Il libère les ressources de l'appelant qui seraient autrement bloquées en attente. Il donne au service défaillant une opportunité de récupérer en réduisant sa charge. Il fournit un signal clair de l'état de santé d'une dépendance.

En contrepartie, le Circuit Breaker introduit une complexité supplémentaire dans le code ou l'infrastructure. Un mauvais paramétrage peut causer des interruptions de service injustifiées. L'état du circuit doit être partagé entre les instances d'un service (ou chaque instance maintient son propre état, créant des incohérences). Le comportement en mode ouvert doit être géré explicitement (voir le patron Fallback).

> **Incident de terrain**
> *Symptôme* : Un site de commerce électronique devient inaccessible pendant 45 minutes lors d'un pic de trafic.
> *Cause racine* : Le service de recommandations, surchargé, répondait lentement. Les services appelants accumulaient des connexions en attente jusqu'à épuisement de leurs pools de connexions. Sans Circuit Breaker, la défaillance s'est propagée à l'ensemble du système.
> *Solution* : Déploiement d'un Circuit Breaker devant le service de recommandations. En cas d'ouverture, le site affiche des recommandations génériques plutôt que personnalisées.
> *Leçon* : Les dépendances non critiques doivent être isolées par des disjoncteurs pour préserver les fonctionnalités essentielles.

#### Exemple d'Usage

Un service de paiement invoque un service de vérification de fraude avant d'autoriser une transaction. Le service de fraude, hébergé chez un partenaire externe, connaît des problèmes de performance intermittents. Un Circuit Breaker protège le service de paiement : après cinq timeouts consécutifs, le circuit s'ouvre. Les transactions suivantes sont traitées selon une politique de repli : les petits montants (< 100 $) sont autorisés sans vérification de fraude, les montants élevés sont mis en attente pour traitement manuel. Après 60 secondes, le circuit tente une récupération.

> **Quand utiliser ce patron**
> *Contexte* : Appels vers des services distants susceptibles de défaillir ou de devenir lents ; protection contre les cascades de défaillances.
> *Alternatives* : Pour des appels internes très rapides (microsecondes), le Circuit Breaker peut être excessif. Pour des dépendances asynchrones, les mécanismes de backpressure des files de messages sont plus appropriés.

### 7.1.2 Retry with Exponential Backoff

#### Définition

Le patron Retry (réessai) consiste à répéter automatiquement une opération qui a échoué, dans l'espoir que l'échec était transitoire et que la prochaine tentative réussira. L'Exponential Backoff (recul exponentiel) ajoute un délai croissant entre chaque tentative pour éviter de surcharger le système cible.

#### Problème Résolu

De nombreux échecs dans les systèmes distribués sont transitoires : congestion réseau momentanée, redémarrage d'un conteneur, garbage collection prolongée, pic de charge temporaire. Une requête qui échoue à l'instant T peut parfaitement réussir à T+1 seconde. Abandonner immédiatement prive le système de cette opportunité de récupération automatique.

Toutefois, un réessai immédiat et répété peut aggraver le problème. Si un service est surchargé, bombarder ce service de requêtes supplémentaires ne fait qu'empirer la situation. Le backoff exponentiel espace progressivement les tentatives, donnant au service le temps de récupérer.

#### Mécanisme

L'algorithme de base fonctionne ainsi : après un premier échec, attendre un délai initial (par exemple, 100 ms) avant de réessayer. Si la deuxième tentative échoue, doubler le délai (200 ms). Continuer à doubler jusqu'à atteindre un délai maximum ou un nombre maximum de tentatives.

L'ajout de **jitter** (variation aléatoire) au délai est crucial. Si mille clients tentent simultanément d'accéder à un service qui redémarre, et que tous utilisent le même algorithme de backoff, ils réessaieront tous au même instant, créant des « tempêtes de tentatives » (retry storms). Le jitter désynchronise les clients en ajoutant une composante aléatoire au délai.

```
Tentative 1 : échec
  └─► Attente : 100ms × (1 + random(0, 0.5)) = 100-150ms
Tentative 2 : échec
  └─► Attente : 200ms × (1 + random(0, 0.5)) = 200-300ms
Tentative 3 : échec
  └─► Attente : 400ms × (1 + random(0, 0.5)) = 400-600ms
Tentative 4 : échec
  └─► Attente : 800ms × (1 + random(0, 0.5)) = 800-1200ms
Tentative 5 : succès ou abandon
```

#### Paramétrage

Plusieurs paramètres gouvernent le comportement du retry.

Le **nombre maximum de tentatives** limite les ressources consommées. Trois à cinq tentatives suffisent généralement pour les erreurs transitoires. Au-delà, l'échec est probablement persistant.

Le **délai initial** définit le point de départ du backoff. Pour des services rapides, 100 ms est approprié. Pour des services plus lents ou des opérations lourdes, un délai initial plus long (1 seconde) peut être préférable.

Le **facteur multiplicatif** détermine la croissance du délai. Un facteur de 2 (doublement) est standard. Un facteur plus faible (1,5) produit une croissance plus douce.

Le **délai maximum** plafonne le backoff pour éviter des attentes excessives. Des valeurs de 30 à 60 secondes sont courantes.

Les **erreurs réessayables** doivent être définies explicitement. Les erreurs transitoires (timeouts, erreurs 503) méritent un réessai. Les erreurs permanentes (erreurs 400, 404, 401) ne devraient pas être réessayées — elles échoueront toujours.

> **Configuration recommandée**
> Pour un appel d'API REST :
>
> * Tentatives maximum : 3
> * Délai initial : 100 ms
> * Facteur multiplicatif : 2
> * Délai maximum : 10 secondes
> * Jitter : ±50 %
> * Erreurs réessayables : timeouts, 429 (Too Many Requests), 503, 502, 504
> * Erreurs non réessayables : 400, 401, 403, 404, 422

#### Avantages et Inconvénients

Le Retry récupère automatiquement des erreurs transitoires sans intervention humaine. Il améliore la disponibilité perçue par les utilisateurs finaux. Combiné au backoff exponentiel et au jitter, il évite d'aggraver les problèmes de surcharge.

Cependant, le Retry peut masquer des problèmes systémiques en les faisant paraître comme des ralentissements plutôt que des échecs. Il augmente la latence en cas d'échec (temps de toutes les tentatives). Sans jitter, il peut provoquer des tempêtes de tentatives. Pour les opérations non idempotentes, il peut causer des effets de bord indésirables (double exécution).

> **Anti-patron**
> **Retry sans idempotence** : Réessayer une opération qui n'est pas idempotente (création d'une commande, débit d'un compte) peut provoquer des duplications. Avant d'implémenter le retry, s'assurer que l'opération cible est idempotente ou que le protocole garantit exactement une exécution (exactly-once).

#### Interaction avec le Circuit Breaker

Le Retry et le Circuit Breaker travaillent de concert mais doivent être ordonnés correctement. Le Retry devrait être à l'intérieur du Circuit Breaker : les tentatives comptent pour le calcul du taux d'échec du disjoncteur. Si le retry est à l'extérieur, chaque tentative de retry sur un circuit ouvert compte comme un nouvel échec, perturbant les statistiques.

```
Mauvais : Retry(CircuitBreaker(appel))
  └─► Chaque retry voit un circuit ouvert, compte comme échec

Bon : CircuitBreaker(Retry(appel))
  └─► Retry épuise ses tentatives, puis renvoie un échec unique au circuit
```

### 7.1.3 Bulkhead

#### Définition

Le patron Bulkhead (cloison étanche) isole les ressources d'un système en compartiments indépendants, de sorte que la défaillance d'un compartiment ne se propage pas aux autres. Le nom provient des cloisons étanches des navires qui empêchent l'eau d'envahir l'ensemble du vaisseau en cas de brèche.

#### Problème Résolu

Dans un système qui partage des ressources (pool de threads, pool de connexions, mémoire), une dépendance défaillante peut monopoliser ces ressources au détriment des autres fonctionnalités. Si un service A devient lent et que tous les threads disponibles sont bloqués en attente de A, les requêtes vers les services B, C et D ne peuvent plus être traitées, même si ces services fonctionnent parfaitement.

Le Bulkhead partitionne les ressources de sorte que chaque dépendance dispose de son propre quota. La saturation d'un compartiment n'affecte pas les autres.

#### Mécanisme

Deux formes principales de Bulkhead existent dans la pratique.

Le **Bulkhead par pool de threads** alloue un pool de threads dédié à chaque dépendance. Un service qui appelle trois services externes maintient trois pools de threads distincts. Si le pool associé au service A est saturé, les appels vers B et C continuent de fonctionner normalement.

Le **Bulkhead par sémaphore** limite le nombre d'appels concurrents vers chaque dépendance sans créer de threads dédiés. Un compteur limite le nombre de requêtes en cours ; les requêtes excédentaires échouent immédiatement ou sont mises en file d'attente.

```
Sans Bulkhead :                    Avec Bulkhead :
┌─────────────────────────┐       ┌─────────────────────────┐
│    Pool partagé         │       │  Pool Service A (10)   │
│    (50 threads)         │       │  ████████░░            │
│    ████████████████     │       ├─────────────────────────┤
│    Tous bloqués sur A   │       │  Pool Service B (20)   │
│                         │       │  ██████░░░░░░░░░░░░░░  │
│    B et C inaccessibles │       ├─────────────────────────┤
└─────────────────────────┘       │  Pool Service C (20)   │
                                  │  ████░░░░░░░░░░░░░░░░  │
                                  │                         │
                                  │  B et C fonctionnent   │
                                  └─────────────────────────┘
```

#### Dimensionnement

Le dimensionnement des compartiments exige une analyse du profil de trafic et des caractéristiques de chaque dépendance.

Pour chaque dépendance, estimer le nombre maximum d'appels concurrents en conditions normales, puis ajouter une marge (typiquement 50-100 %) pour absorber les pics. Un compartiment trop petit rejette des requêtes légitimes ; un compartiment trop grand consomme des ressources inutilement et offre moins de protection.

La latence de la dépendance influence le dimensionnement. Une dépendance rapide (5 ms de latence moyenne) nécessite moins de threads concurrents qu'une dépendance lente (500 ms) pour le même débit.

> **Configuration recommandée**
> Pour dimensionner un pool de threads dédié :
>
> * Taille = (Requêtes par seconde) × (Latence moyenne en secondes) × 1,5
> * Exemple : 100 req/s vers un service à 200 ms de latence → 100 × 0,2 × 1,5 = 30 threads
> * Ajouter une file d'attente limitée pour absorber les micro-pics (taille = 50-100 % du pool)

#### Avantages et Inconvénients

Le Bulkhead garantit que les ressources sont disponibles pour les dépendances saines, même lorsqu'une dépendance est défaillante. Il fournit une visibilité claire sur l'utilisation des ressources par dépendance. Il facilite le réglage fin des limites selon la criticité de chaque dépendance.

En contrepartie, le Bulkhead consomme plus de ressources totales qu'un pool partagé (chaque compartiment doit être dimensionné pour son pic). Il ajoute de la complexité de configuration et de surveillance. Le dimensionnement incorrect peut créer des goulots d'étranglement artificiels.

### 7.1.4 Timeout

#### Définition

Le patron Timeout (délai d'expiration) fixe une durée maximale pour une opération. Si l'opération n'est pas terminée dans ce délai, elle est annulée et une erreur est retournée à l'appelant.

#### Problème Résolu

Sans timeout explicite, un appel vers un service qui ne répond pas peut bloquer indéfiniment. Les ressources de l'appelant (threads, connexions, mémoire) restent allouées. Si suffisamment d'appels s'accumulent, l'appelant devient lui-même incapable de répondre.

Les timeouts par défaut des bibliothèques réseau sont souvent inadaptés : trop longs (plusieurs minutes) ou absents. Un timeout explicite, calibré selon les attentes métier, garantit que les ressources sont libérées dans un délai prévisible.

#### Types de Timeouts

Plusieurs niveaux de timeout coexistent dans une chaîne d'appel.

Le **timeout de connexion** limite le temps pour établir une connexion réseau. Des valeurs de 1 à 5 secondes sont courantes. Un échec de connexion indique généralement un problème réseau ou un service totalement indisponible.

Le **timeout de lecture** limite le temps d'attente pour recevoir une réponse après l'envoi de la requête. Ce timeout doit être calibré selon la latence attendue du service. Un service rapide (P99 < 100 ms) peut avoir un timeout de lecture de 500 ms. Un service de reporting peut nécessiter plusieurs secondes.

Le **timeout de bout en bout** couvre l'ensemble de l'opération, incluant les éventuels retries. Ce timeout est particulièrement important pour les requêtes utilisateur qui ont une tolérance limitée à l'attente.

#### Calibration

La calibration des timeouts requiert des données empiriques sur la latence des dépendances.

Une règle pratique consiste à fixer le timeout légèrement au-dessus du 99e percentile de latence observé en conditions normales. Si un service répond en moins de 200 ms dans 99 % des cas, un timeout de 500 ms capture les cas légitimement lents tout en coupant les blocages anormaux.

Le timeout doit tenir compte de la chaîne d'appel. Si un service A appelle B qui appelle C, le timeout de A doit être supérieur à la somme des timeouts de B et C, plus une marge pour le traitement local.

> **Note technique**
> Les timeouts interagissent avec les retries. Si le timeout est de 1 seconde et le retry de 3 tentatives, le timeout effectif est de 3 secondes (sans compter les délais de backoff). Le timeout de bout en bout doit couvrir l'ensemble de cette durée.

#### Propagation des Deadlines

Dans les architectures de microservices, la propagation des deadlines (échéances) évite les situations où des services en aval travaillent sur des requêtes déjà expirées en amont.

Lorsqu'une requête arrive avec un timeout de 5 secondes, le service A consomme 1 seconde pour son traitement initial. Il devrait propager une deadline de 4 secondes (moins une marge) au service B. Si B consomme 2 secondes, il propage 2 secondes à C. Ainsi, si la deadline globale est dépassée, les services en aval cessent de travailler inutilement.

gRPC implémente nativement cette propagation de deadlines. Pour HTTP, des en-têtes personnalisés (comme `X-Request-Deadline`) peuvent être utilisés.

### 7.1.5 Fallback

#### Définition

Le patron Fallback (repli) définit un comportement alternatif lorsqu'une opération principale échoue. Plutôt que de simplement propager l'erreur, le système fournit une réponse dégradée mais utile.

#### Problème Résolu

Dans de nombreux cas, une erreur brute (« Service indisponible ») est une réponse médiocre pour l'utilisateur final. Un Fallback transforme cet échec en une expérience acceptable, même si sous-optimale.

Le Fallback est particulièrement pertinent pour les dépendances non critiques. Si le service de recommandations est indisponible, le site peut afficher des produits populaires génériques plutôt qu'une page d'erreur. Si le service de personnalisation échoue, l'application peut utiliser des valeurs par défaut.

#### Stratégies de Fallback

Plusieurs stratégies de repli sont courantes dans la pratique.

Le **cache stale** retourne la dernière valeur connue, même si elle est potentiellement périmée. Pour des données qui changent peu (configuration, catalogue), cette stratégie offre une excellente continuité de service.

La **valeur par défaut** retourne une constante prédéfinie. Pour les recommandations, cela peut être une liste de produits populaires. Pour les configurations, les valeurs par défaut codées en dur.

Le **service de secours** route vers une implémentation alternative, potentiellement moins performante ou moins complète. Un service de géolocalisation principal peut avoir un service de secours basé sur une base de données locale moins précise.

La **réponse vide ou neutre** retourne une réponse vide qui permet au reste du système de fonctionner. Une liste de recommandations vide est préférable à une erreur qui bloque le rendu de la page.

La **réponse en cache local** diffère du cache stale en ce qu'elle utilise un cache embarqué dans l'application plutôt qu'un cache distribué, offrant une meilleure résilience au prix de la fraîcheur.

```
Stratégie de Fallback en cascade :
1. Appel au service principal
   └─► Échec
2. Tentative de cache distribué (Redis)
   └─► Échec ou miss
3. Cache local (en mémoire)
   └─► Échec ou miss
4. Valeur par défaut codée en dur
   └─► Toujours disponible
```

#### Signalement du Mode Dégradé

Lorsqu'un Fallback est activé, il est important de le signaler, tant aux utilisateurs qu'aux systèmes de surveillance.

Côté utilisateur, un indicateur discret peut signaler que certaines fonctionnalités sont temporairement limitées. Côté monitoring, des métriques doivent comptabiliser les activations de Fallback pour détecter les problèmes persistants.

> **Bonnes pratiques**
>
> * Tester régulièrement les chemins de Fallback ; ils sont rarement exercés et peuvent être défectueux lorsqu'on en a besoin
> * Documenter le comportement attendu en mode dégradé pour chaque fonctionnalité
> * Émettre des alertes lorsque les Fallbacks sont activés fréquemment

### 7.1.6 Combinaison des Patrons

Les cinq patrons présentés se combinent en une stratégie de défense en profondeur. Une implémentation typique les assemble comme suit :

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Appelant                                   │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                     Circuit Breaker                          │   │
│  │  ┌───────────────────────────────────────────────────────┐  │   │
│  │  │                     Bulkhead                           │  │   │
│  │  │  ┌─────────────────────────────────────────────────┐  │  │   │
│  │  │  │                   Timeout                        │  │  │   │
│  │  │  │  ┌───────────────────────────────────────────┐  │  │  │   │
│  │  │  │  │            Retry + Backoff                 │  │  │  │   │
│  │  │  │  │  ┌─────────────────────────────────────┐  │  │  │  │   │
│  │  │  │  │  │         Appel au service            │  │  │  │  │   │
│  │  │  │  │  └─────────────────────────────────────┘  │  │  │  │   │
│  │  │  │  └───────────────────────────────────────────┘  │  │  │   │
│  │  │  └─────────────────────────────────────────────────┘  │  │   │
│  │  └───────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                    En cas d'échec                                  │
│                              │                                      │
│                              ▼                                      │
│                        ┌──────────┐                                 │
│                        │ Fallback │                                 │
│                        └──────────┘                                 │
└─────────────────────────────────────────────────────────────────────┘
```

L'ordre des couches importe. Le Circuit Breaker enveloppe l'ensemble pour court-circuiter rapidement lorsque le service est connu comme défaillant. Le Bulkhead limite les ressources consommées par cette dépendance. Le Timeout garantit une durée maximale pour chaque tentative. Le Retry gère les erreurs transitoires au niveau de chaque appel individuel. Le Fallback fournit une réponse alternative si tout échoue.

---

## 7.2 Infrastructure de Résilience

Les patrons de la section précédente s'implémentent au niveau du code applicatif. Ils offrent un contrôle fin mais imposent une charge de développement et de maintenance sur chaque équipe. L'infrastructure de résilience décharge les applications de certaines préoccupations transversales en les implémentant dans des composants d'infrastructure partagés.

### 7.2.1 Sidecar Pattern

#### Définition

Le patron Sidecar consiste à déployer un processus auxiliaire aux côtés de chaque instance d'application. Ce processus « accompagnateur » gère des préoccupations transversales (sécurité, observabilité, communication réseau) sans que l'application ne doive les implémenter elle-même.

#### Mécanisme

Dans un déploiement Kubernetes typique, le Sidecar est un conteneur additionnel dans le même pod que le conteneur applicatif. Les deux conteneurs partagent le même espace réseau (localhost) et peuvent partager des volumes de stockage.

Le trafic réseau entrant et sortant de l'application est intercepté par le Sidecar via des règles iptables. L'application communique avec localhost ; le Sidecar gère la communication avec l'extérieur, appliquant au passage les politiques de sécurité, de retry, de Circuit Breaker et de télémétrie.

```
┌─────────────────────────────────────────────────────────────────────┐
│                              Pod                                    │
│                                                                     │
│  ┌─────────────────────┐         ┌─────────────────────────────┐   │
│  │                     │         │          Sidecar Proxy       │   │
│  │    Application      │◄───────►│                              │   │
│  │                     │ localhost│  • mTLS                     │   │
│  │  (code métier      │         │  • Circuit Breaker           │   │
│  │   uniquement)       │         │  • Retry                     │   │
│  │                     │         │  • Métriques                │   │
│  └─────────────────────┘         │  • Traces                    │   │
│                                  │                              │   │
│                                  └──────────────┬───────────────┘   │
│                                                 │                   │
└─────────────────────────────────────────────────┼───────────────────┘
                                                  │
                                                  ▼
                                          Réseau externe
```

#### Avantages et Inconvénients

Le Sidecar permet d'ajouter des fonctionnalités transversales sans modifier le code applicatif. Les équipes de développement se concentrent sur la logique métier ; l'équipe de plateforme gère l'infrastructure de résilience. Les mises à jour des politiques de sécurité ou de résilience se propagent en mettant à jour le Sidecar, sans redéployer les applications.

En contrepartie, le Sidecar ajoute de la latence (traversée d'un processus supplémentaire) et consomme des ressources (CPU, mémoire). La complexité de débogage augmente car le trafic traverse un composant intermédiaire. L'injection automatique de Sidecars peut surprendre les développeurs non familiers avec le mécanisme.

### 7.2.2 Service Mesh

#### Définition

Un Service Mesh (maillage de services) est une couche d'infrastructure dédiée à la gestion du trafic réseau entre les services. Il implémente le patron Sidecar à l'échelle de l'ensemble du cluster, offrant une visibilité et un contrôle uniformes sur toutes les communications inter-services.

#### Composants

Un Service Mesh typique comprend deux composants principaux.

Le **data plane** (plan de données) est constitué de l'ensemble des proxies Sidecar déployés aux côtés de chaque service. Ces proxies interceptent tout le trafic entrant et sortant. Envoy, développé par Lyft, est le proxy le plus répandu, utilisé par Istio, Linkerd et d'autres.

Le **control plane** (plan de contrôle) centralise la configuration et la gestion des proxies. Il distribue les politiques de routage, de sécurité et d'observabilité à l'ensemble des Sidecars. Istiod (pour Istio) ou linkerd-destination (pour Linkerd) sont des exemples de control planes.

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Control Plane                               │
│                                                                     │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │                  Configuration centralisée                  │    │
│  │  • Politiques de routage                                    │    │
│  │  • Règles de sécurité (mTLS)                               │    │
│  │  • Paramètres de résilience                                │    │
│  └────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │ Configuration
                    ┌─────────────┼─────────────┐
                    │             │             │
                    ▼             ▼             ▼
┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────┐
│ Pod Service A       │ │ Pod Service B       │ │ Pod Service C       │
│ ┌─────┐ ┌─────────┐ │ │ ┌─────┐ ┌─────────┐ │ │ ┌─────┐ ┌─────────┐ │
│ │ App │◄┤ Sidecar │◄┼─┼►│ App │◄┤ Sidecar │◄┼─┼►│ App │◄┤ Sidecar │ │
│ └─────┘ └─────────┘ │ │ └─────┘ └─────────┘ │ │ └─────┘ └─────────┘ │
└─────────────────────┘ └─────────────────────┘ └─────────────────────┘
                                  │
                           Data Plane
```

#### Fonctionnalités

Un Service Mesh offre plusieurs catégories de fonctionnalités.

**Sécurité** : Le mTLS (mutual TLS) chiffre automatiquement tout le trafic entre services et authentifie les deux parties. L'application n'a pas à gérer les certificats ; le mesh s'en charge. Les politiques d'autorisation définissent quels services peuvent communiquer entre eux.

**Gestion du trafic** : Le routage intelligent permet les déploiements canari (envoi d'un pourcentage du trafic vers une nouvelle version), les déploiements bleu-vert, et le routage basé sur des en-têtes. Le load balancing avec détection de santé répartit le trafic entre les instances saines.

**Résilience** : Le mesh implémente les patrons de résilience (Circuit Breaker, Retry, Timeout) au niveau de l'infrastructure, sans code applicatif. Les politiques sont définies de manière déclarative et appliquées uniformément.

**Observabilité** : Le mesh collecte automatiquement les métriques de latence, de débit et d'erreur pour chaque communication. Il génère des traces distribuées sans instrumentation applicative. Il produit une carte des dépendances entre services.

#### Istio vs Linkerd

Istio et Linkerd sont les deux Service Mesh les plus répandus. Ils partagent des fonctionnalités similaires mais diffèrent par leur philosophie.

 **Istio** , soutenu par Google, IBM et Lyft, offre un ensemble de fonctionnalités très riche mais avec une complexité de configuration significative. Il convient aux organisations avec des besoins avancés de gestion du trafic et une équipe plateforme dédiée.

 **Linkerd** , développé par Buoyant, privilégie la simplicité et la légèreté. Son empreinte ressources est plus faible, sa courbe d'apprentissage plus douce. Il convient aux organisations qui recherchent les bénéfices essentiels du mesh sans la complexité d'Istio.

| Critère                   | Istio         | Linkerd                 |
| -------------------------- | ------------- | ----------------------- |
| Complexité                | Élevée      | Modérée               |
| Consommation ressources    | Plus élevée | Plus légère           |
| Fonctionnalités avancées | Très riches  | Essentielles            |
| Courbe d'apprentissage     | Abrupte       | Progressive             |
| Communauté                | Très large   | Active mais plus petite |

> **Quand utiliser un Service Mesh**
> *Contexte* : Architectures de microservices avec de nombreux services inter-communicants ; besoin de sécurité mTLS uniforme ; exigence d'observabilité sans modification du code applicatif.
> *Alternatives* : Pour des architectures simples (< 10 services), les bibliothèques de résilience applicatives peuvent suffire avec moins de surcharge opérationnelle.

### 7.2.3 Event Mesh

#### Définition

L'Event Mesh étend le concept de Service Mesh au trafic asynchrone basé sur les événements. Il fournit une infrastructure de routage, de transformation et de gouvernance pour les flux d'événements, indépendamment des brokers de messages sous-jacents.

#### Problème Résolu

Dans une architecture événementielle hybride, les événements peuvent transiter par différentes plateformes : Kafka pour les flux à haut débit, RabbitMQ pour les files de travail, services de messagerie infonuagique (AWS EventBridge, Azure Event Grid). Orchestrer le routage entre ces plateformes, assurer la cohérence des politiques de sécurité et maintenir l'observabilité devient complexe.

L'Event Mesh abstrait cette hétérogénéité en offrant une couche unifiée de routage et de gouvernance.

#### Fonctionnalités

**Routage dynamique** : L'Event Mesh route les événements vers les consommateurs appropriés selon leur type, leur contenu ou des règles personnalisées. Il peut router un même événement vers plusieurs destinations hétérogènes.

**Transformation** : L'Event Mesh peut transformer les événements à la volée : conversion de formats (Avro vers JSON), enrichissement avec des données de référence, filtrage de champs sensibles.

**Gouvernance** : Les politiques de qualité (schéma obligatoire, validation de contenu) et de sécurité (chiffrement, contrôle d'accès) s'appliquent uniformément à tous les flux.

**Observabilité** : Les métriques de latence, de débit et d'erreur sont collectées pour chaque flux d'événements, permettant le monitoring et le débogage.

#### Solutions du Marché

Solace PubSub+ et Confluent (avec son écosystème Kafka) sont les principaux fournisseurs d'Event Mesh. Les solutions infonuagiques natives (AWS EventBridge, Azure Event Grid) offrent des fonctionnalités similaires dans leurs écosystèmes respectifs.

L'Event Mesh complète le Service Mesh : le premier gère le trafic asynchrone (événements), le second le trafic synchrone (requêtes-réponses). Ensemble, ils couvrent l'ensemble du spectre de communication d'une architecture hybride, tel que présenté dans le continuum d'intégration de cette monographie.

---

## 7.3 Observabilité Distribuée

La résilience permet aux systèmes de survivre aux pannes ; l'observabilité permet de comprendre ce qui se passe, de diagnostiquer les problèmes et d'anticiper les dégradations. Dans un système distribué où les requêtes traversent des dizaines de services, l'observabilité traditionnelle (regarder les logs d'un serveur) ne suffit plus.

L'observabilité moderne repose sur trois piliers complémentaires : les traces, les métriques et les journaux. Chaque pilier répond à un type de question différent ; ensemble, ils offrent une vision complète du système.

### 7.3.1 Les Trois Piliers

#### Traces Distribuées

Les traces suivent le parcours d'une requête à travers les différents services du système. Une trace est composée de  *spans* , chaque span représentant une opération (appel HTTP, requête de base de données, traitement local). Les spans sont liés par des relations parent-enfant, formant un arbre qui visualise le flux de la requête.

> **Définition formelle**
> **Trace** : Représentation du parcours complet d'une requête à travers un système distribué, composée de spans interconnectés qui capturent le timing, les métadonnées et les éventuelles erreurs de chaque opération.

Les traces répondent aux questions de type : « Pourquoi cette requête était-elle lente ? », « Quels services ont été impliqués dans le traitement de cette transaction ? », « Où l'erreur s'est-elle produite dans la chaîne d'appels ? »

```
Trace ID: abc123
│
├── Span: API Gateway (45ms)
│   └── Span: Service Commande (120ms)
│       ├── Span: Validation (5ms)
│       ├── Span: Service Inventaire (80ms)    ◄── Lent !
│       │   └── Span: Query PostgreSQL (75ms)  ◄── Base de données
│       └── Span: Service Paiement (25ms)
│           └── Span: Appel Stripe (20ms)
│
Total: 190ms
```

#### Métriques

Les métriques sont des mesures numériques agrégées dans le temps. Elles capturent l'état du système à un niveau statistique : nombre de requêtes par seconde, latence au 99e percentile, taux d'erreur, utilisation CPU, files d'attente.

> **Définition formelle**
> **Métrique** : Mesure numérique, généralement agrégée sur une fenêtre temporelle, qui quantifie un aspect du comportement ou de l'état d'un système (compteur, jauge, histogramme).

Les métriques répondent aux questions de type : « Quel est le débit actuel du système ? », « La latence se dégrade-t-elle au fil du temps ? », « Combien d'erreurs avons-nous eu cette heure ? »

Trois types de métriques sont fondamentaux :

* **Compteurs** (counters) : Valeurs qui ne font qu'augmenter (nombre total de requêtes, octets transférés). On analyse généralement leur taux de changement.
* **Jauges** (gauges) : Valeurs qui peuvent monter ou descendre (utilisation mémoire, nombre de connexions actives, taille de file d'attente).
* **Histogrammes** : Distributions de valeurs (latences des requêtes, tailles des réponses). Permettent de calculer des percentiles.

#### Journaux (Logs)

Les journaux sont des enregistrements textuels horodatés d'événements discrets. Ils capturent des détails contextuels que les métriques agrégées ne peuvent pas représenter.

> **Définition formelle**
> **Journal** : Enregistrement horodaté et structuré d'un événement discret survenu dans un système, incluant contexte, sévérité et données associées.

Les journaux répondent aux questions de type : « Que s'est-il passé exactement lors de cette erreur ? », « Quels paramètres ont été utilisés pour cette requête ? », « Quelle était la séquence exacte d'événements ? »

Les journaux modernes sont structurés (JSON plutôt que texte libre) pour faciliter l'analyse automatisée. Ils incluent des identifiants de corrélation (trace ID, request ID) qui permettent de les relier aux traces.

```json
{
  "timestamp": "2025-01-15T14:30:00.123Z",
  "level": "ERROR",
  "service": "order-service",
  "trace_id": "abc123",
  "span_id": "def456",
  "message": "Payment validation failed",
  "error": {
    "type": "PaymentDeclinedException",
    "code": "INSUFFICIENT_FUNDS"
  },
  "context": {
    "order_id": "ORD-789",
    "amount": 150.00,
    "customer_id": "CUST-456"
  }
}
```

### 7.3.2 Complémentarité des Piliers

Les trois piliers ne sont pas redondants ; ils se complètent pour offrir une observabilité complète.

**Les métriques détectent** qu'un problème existe. Une alerte sur le taux d'erreur ou la latence signale une anomalie.

**Les traces localisent** où le problème se situe. En suivant une requête lente, on identifie le service ou l'opération responsable.

**Les journaux expliquent** pourquoi le problème s'est produit. Les détails contextuels révèlent la cause racine.

| Pilier     | Question   | Force                   | Faiblesse                |
| ---------- | ---------- | ----------------------- | ------------------------ |
| Métriques | Quoi ?     | Vue d'ensemble, alertes | Peu de détails          |
| Traces     | Où ?      | Parcours des requêtes  | Volume, échantillonnage |
| Journaux   | Pourquoi ? | Détails contextuels    | Volume, recherche        |

> **Perspective stratégique**
> Investir dans les trois piliers dès le début du projet. Ajouter l'observabilité après coup est nettement plus coûteux car cela nécessite de ré-instrumenter le code existant. L'observabilité n'est pas un luxe réservé aux systèmes matures ; c'est un prérequis pour opérer des systèmes distribués.

### 7.3.3 Outils et Technologies

L'écosystème d'outils d'observabilité est riche. Voici les principales catégories et les solutions de référence.

#### Traçage Distribué

 **Jaeger** , développé par Uber et maintenant projet CNCF, est une solution de traçage distribué open source. Il s'intègre nativement avec OpenTelemetry et offre une interface de visualisation des traces.

 **Zipkin** , développé par Twitter, est une alternative plus ancienne mais toujours pertinente, particulièrement dans les écosystèmes Java/Spring.

**Tempo** de Grafana Labs offre un backend de traces intégré à l'écosystème Grafana, avec un modèle de stockage économique basé sur le stockage objet.

Les solutions commerciales (Datadog APM, New Relic, Dynatrace) offrent des expériences intégrées avec corrélation automatique entre traces, métriques et journaux.

#### Métriques

**Prometheus** est le standard de facto pour la collecte de métriques dans les environnements Kubernetes. Son modèle de scraping (les services exposent des endpoints `/metrics` que Prometheus interroge périodiquement) s'adapte bien aux architectures dynamiques.

**Grafana** est l'outil de visualisation et de tableaux de bord le plus répandu. Il s'intègre avec Prometheus et de nombreuses autres sources de données.

**InfluxDB** et **TimescaleDB** sont des bases de données temporelles alternatives pour des cas d'usage spécifiques (IoT, haute cardinalité).

#### Journaux

**Elasticsearch, Logstash, Kibana (ELK Stack)** constituent la pile de logging la plus répandue. Elasticsearch indexe les journaux pour une recherche rapide ; Logstash les collecte et les transforme ; Kibana les visualise.

**Loki** de Grafana Labs offre une alternative plus légère, indexant uniquement les labels plutôt que le contenu complet des journaux. Il s'intègre nativement avec Grafana.

**Fluent Bit** et **Fluentd** sont des collecteurs de journaux légers utilisés pour agréger et router les journaux vers diverses destinations.

### 7.3.4 OpenTelemetry

#### Définition et Objectif

OpenTelemetry (OTel) est un projet de la Cloud Native Computing Foundation (CNCF) qui vise à standardiser la collecte et l'export des données d'observabilité. Il fournit des API, des SDK et des outils pour instrumenter les applications et collecter traces, métriques et journaux de manière uniforme.

> **Définition formelle**
> **OpenTelemetry** : Framework open source de télémétrie qui fournit des API, SDK et outils standardisés pour la collecte, le traitement et l'export des traces, métriques et journaux, indépendamment des backends de stockage et d'analyse.

Avant OpenTelemetry, chaque solution d'observabilité utilisait son propre format et son propre SDK d'instrumentation. Instrumenter une application pour Jaeger était différent de l'instrumenter pour Zipkin ou Datadog. OpenTelemetry unifie ces approches en une abstraction commune.

#### Architecture

OpenTelemetry comprend plusieurs composants.

Les **API** définissent les interfaces pour créer des spans, enregistrer des métriques et émettre des journaux. Le code applicatif utilise ces API sans dépendre d'une implémentation spécifique.

Les **SDK** implémentent les API et gèrent l'échantillonnage, le batching et l'export des données. Ils sont disponibles pour de nombreux langages (Java, Python, Go, JavaScript, .NET, etc.).

Les **exporters** transmettent les données vers les backends de stockage. Des exporters existent pour Jaeger, Zipkin, Prometheus, les solutions commerciales, et le protocole natif OTLP (OpenTelemetry Protocol).

Le **Collector** est un composant optionnel qui reçoit les données de télémétrie, les transforme et les exporte vers une ou plusieurs destinations. Il permet de découpler l'instrumentation applicative du choix de backend.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Application                                  │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                  OpenTelemetry SDK                           │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │   │
│  │  │  Traces  │  │ Métriques│  │ Journaux │                   │   │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘                   │   │
│  │       └─────────────┼─────────────┘                         │   │
│  │                     │ OTLP                                   │   │
│  └─────────────────────┼───────────────────────────────────────┘   │
└─────────────────────────┼───────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    OpenTelemetry Collector                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐ │
│  │  Receivers  │→ │ Processors  │→ │        Exporters            │ │
│  │  (OTLP)     │  │ (batch,     │  │ Jaeger, Prometheus, Loki,   │ │
│  │             │  │  filter)    │  │ Datadog, etc.               │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

#### Instrumentation Automatique

OpenTelemetry offre une instrumentation automatique pour de nombreux frameworks et bibliothèques. En Java, un agent JVM injecte automatiquement le code de traçage pour Spring, JDBC, HTTP clients, etc. En Python, des hooks similaires existent. Cette auto-instrumentation permet d'obtenir une observabilité de base sans modifier le code applicatif.

L'instrumentation manuelle permet d'ajouter des spans personnalisés pour les opérations métier significatives, d'enrichir les spans avec des attributs contextuels (ID client, ID commande) et de capturer des métriques business spécifiques.

> **Bonnes pratiques**
>
> * Commencer par l'auto-instrumentation pour obtenir rapidement une visibilité de base
> * Ajouter une instrumentation manuelle pour les opérations métier clés
> * Inclure des attributs de contexte métier (identifiants, montants) pour faciliter le débogage
> * Utiliser le Collector pour découpler l'instrumentation du backend

### 7.3.5 Stratégies d'Échantillonnage

Dans un système à haut débit, collecter 100 % des traces génère un volume de données considérable et un coût de stockage prohibitif. L'échantillonnage sélectionne un sous-ensemble représentatif des traces à conserver.

#### Types d'Échantillonnage

**L'échantillonnage de tête** (head-based sampling) décide au début de la trace si elle sera collectée. Une décision aléatoire (par exemple, 10 % des traces) est prise lors de la création du span racine et propagée à tous les spans enfants. Simple mais aveugle : des traces intéressantes (erreurs, latence élevée) peuvent être ignorées.

**L'échantillonnage de queue** (tail-based sampling) attend la fin de la trace pour décider de la conserver. Toutes les traces sont collectées temporairement, puis un filtre sélectionne celles à conserver selon des critères (présence d'erreur, latence supérieure à un seuil, attributs spécifiques). Plus coûteux mais plus intelligent.

**L'échantillonnage adaptatif** ajuste dynamiquement le taux d'échantillonnage selon la charge. En période de faible trafic, 100 % des traces sont conservées. En période de pic, le taux diminue pour limiter les coûts.

> **Configuration recommandée**
> Pour un système de production à trafic modéré :
>
> * Échantillonnage de base : 10 % des traces normales
> * Conservation systématique : 100 % des traces avec erreur
> * Conservation conditionnelle : traces avec latence > P99
> * Révision périodique du taux selon les coûts et les besoins d'analyse

### 7.3.6 Corrélation des Données

La valeur de l'observabilité se multiplie lorsque les trois piliers sont corrélés. Un identifiant de trace commun permet de naviguer d'une métrique anormale vers les traces correspondantes, puis vers les journaux détaillés.

Cette corrélation requiert une propagation cohérente du contexte. Le **trace context** (W3C Trace Context est le standard) doit être propagé dans tous les appels inter-services, inclus dans tous les journaux et utilisé comme dimension dans les métriques pertinentes.

Les plateformes d'observabilité modernes (Datadog, Grafana Cloud, Elastic Observability) offrent cette corrélation nativement. Pour les solutions assemblées manuellement, le Collector OpenTelemetry et une discipline d'instrumentation rigoureuse permettent d'atteindre un résultat similaire.

---

## 7.4 Synthèse et Application

### 7.4.1 Résilience et Observabilité par Domaine d'Intégration

Les patrons de résilience et d'observabilité s'appliquent différemment selon le domaine d'intégration considéré.

Pour l'**intégration des applications** (chapitre III), le trafic synchrone requête-réponse bénéficie directement des Circuit Breakers, des Retries et des Timeouts. Le Service Mesh s'applique naturellement à ce trafic. Les traces distribuées capturent le parcours des requêtes.

Pour l'**intégration des données** (chapitre IV), les mécanismes de résilience s'appliquent aux pipelines CDC et aux requêtes de virtualisation. Les métriques de latence de réplication et de fraîcheur des données sont essentielles. Les traces peuvent suivre la propagation d'un changement de données à travers le système.

Pour l'**intégration des événements** (chapitre V), les patrons asynchrones ont leur propre forme de résilience intrinsèque (découplage temporel, persistance). La DLQ (chapitre V) est un mécanisme de résilience spécifique aux événements. L'Event Mesh étend les concepts du Service Mesh au trafic événementiel. Les métriques de lag de consommation et de taille de file sont critiques.

| Domaine      | Résilience applicative                  | Infrastructure         | Métriques clés                    |
| ------------ | ---------------------------------------- | ---------------------- | ----------------------------------- |
| Applications | Circuit Breaker, Retry, Timeout          | Service Mesh           | Latence P99, taux d'erreur          |
| Données     | Retry sur pipelines, circuit sur sources | Réplication, failover | Latence de réplication, fraîcheur |
| Événements | DLQ, Idempotent Consumer                 | Event Mesh             | Consumer lag, throughput            |

### 7.4.2 Matrice de Décision

Le choix et le paramétrage des patrons de résilience dépendent de plusieurs critères.

| Critère                         | Indication                 | Patron privilégié                       |
| -------------------------------- | -------------------------- | ----------------------------------------- |
| Dépendance critique             | Échec inacceptable        | Fallback robuste, cache                   |
| Dépendance non critique         | Dégradation acceptable    | Circuit Breaker agressif, Fallback simple |
| Erreurs transitoires fréquentes | Réseau instable           | Retry avec backoff                        |
| Dépendance lente                | Latence variable           | Timeout strict, Bulkhead                  |
| Charge variable                  | Pics imprévisibles        | Bulkhead, load shedding                   |
| Nombreux services                | Architecture microservices | Service Mesh                              |

### 7.4.3 Anti-patrons à Éviter

Plusieurs anti-patrons méritent d'être soulignés pour conclure cette section.

**Résilience sans observabilité** : Implémenter des Circuit Breakers sans monitorer leur état conduit à des défaillances silencieuses. Un circuit perpétuellement ouvert indique un problème persistant qui passe inaperçu.

**Retries non bornés** : Des retries sans limite de tentatives ou de durée totale peuvent amplifier les problèmes de surcharge plutôt que les atténuer.

**Timeouts hérités** : Utiliser les timeouts par défaut des bibliothèques (souvent plusieurs minutes) plutôt que des valeurs calibrées pour le contexte applicatif.

**Observabilité après coup** : Ajouter l'instrumentation après le déploiement en production, lorsque les problèmes sont déjà survenus et qu'il est trop tard pour collecter les données nécessaires au diagnostic.

**Alertes sur tout** : Configurer des alertes pour chaque métrique sans priorisation, conduisant à la fatigue d'alerte et à l'ignorance des signaux importants.

---

## Conclusion et Transition

Ce chapitre a exploré les préoccupations transversales de résilience et d'observabilité qui traversent les trois domaines d'intégration présentés dans les chapitres précédents. Ces préoccupations ne constituent pas un domaine d'intégration supplémentaire mais plutôt le substrat technique qui permet aux patrons d'applications, de données et d'événements de fonctionner de manière fiable en production.

La résilience, première préoccupation examinée, vise à maintenir le fonctionnement du système malgré les défaillances inévitables des environnements distribués. Les cinq patrons applicatifs — Circuit Breaker, Retry with Exponential Backoff, Bulkhead, Timeout et Fallback — forment un arsenal cohérent de défense en profondeur. Combinés judicieusement, ils isolent les défaillances, récupèrent des erreurs transitoires et offrent des réponses dégradées mais utiles lorsque les chemins principaux échouent.

L'infrastructure de résilience — Sidecar, Service Mesh et Event Mesh — décharge les applications de ces préoccupations transversales en les implémentant au niveau de la plateforme. Cette approche uniformise les pratiques, réduit la charge sur les équipes de développement et permet des mises à jour centralisées des politiques.

L'observabilité, seconde préoccupation, permet de comprendre l'état interne des systèmes distribués. Les trois piliers — traces, métriques et journaux — répondent à des questions complémentaires : les métriques détectent les anomalies, les traces localisent les problèmes, les journaux expliquent les causes. OpenTelemetry unifie l'instrumentation et découple le code applicatif des backends d'analyse.

Ces fondations de résilience et d'observabilité sont indispensables pour aborder le chapitre suivant. Le **chapitre VIII** explorera les technologies de collaboration temps réel (CRDTs, Operational Transformation) et d'automatisation (workflows, agents IA). Ces technologies, par nature distribuées et complexes, présupposent une infrastructure capable de gérer les pannes et de fournir la visibilité nécessaire à leur opération. Les workflows de longue durée, en particulier, requièrent une observabilité fine pour suivre leur progression et diagnostiquer leurs blocages.

Plus généralement, la maîtrise de la résilience et de l'observabilité prépare le terrain pour l'architecture de référence convergente du chapitre IX et pour l'Entreprise Agentique du chapitre XI. Des agents IA autonomes opérant dans un environnement distribué sans mécanismes de résilience seraient fragiles et imprévisibles. Sans observabilité, leur comportement resterait opaque, rendant impossible la gouvernance et l'audit de leurs décisions.

---

## Résumé du Chapitre

**Thème central** : La résilience et l'observabilité sont des préoccupations transversales qui traversent les trois domaines d'intégration (Applications, Données, Événements) et constituent le substrat technique indispensable au fonctionnement fiable des architectures distribuées.

**Patrons de résilience applicative** :

| Patron          | Fonction                                                 | Paramètre clé                 |
| --------------- | -------------------------------------------------------- | ------------------------------- |
| Circuit Breaker | Interrompt les appels vers un service défaillant        | Seuil d'ouverture, durée       |
| Retry + Backoff | Réessaie les erreurs transitoires avec délai croissant | Tentatives max, facteur, jitter |
| Bulkhead        | Isole les ressources par dépendance                     | Taille des compartiments        |
| Timeout         | Limite la durée d'attente                               | Calibré sur latence P99        |
| Fallback        | Fournit une réponse alternative en cas d'échec         | Stratégie de repli             |

**Infrastructure de résilience** :

* **Sidecar Pattern** : Processus auxiliaire gérant les préoccupations transversales
* **Service Mesh** : Couche d'infrastructure pour le trafic synchrone (Istio, Linkerd)
* **Event Mesh** : Extension au trafic asynchrone événementiel

**Les trois piliers de l'observabilité** :

* **Traces** : Parcours des requêtes à travers les services (Jaeger, Zipkin, Tempo)
* **Métriques** : Mesures agrégées dans le temps (Prometheus, Grafana)
* **Journaux** : Enregistrements détaillés d'événements (ELK, Loki)

**OpenTelemetry** : Standard unifié pour l'instrumentation et l'export des données d'observabilité, découplant le code applicatif des backends d'analyse.

**Position dans le continuum** : Ces préoccupations transversales s'appliquent à l'ensemble du continuum d'intégration et constituent un prérequis pour les architectures avancées des chapitres suivants (workflows, agents IA, architecture de référence).

---

*Chapitre suivant : VIII — Collaboration Temps Réel et Automatisation*
