# SearchPlan.md - Recherche exploratoire comparative (Scénario 7)

Plan d'exécution pour l'évaluation d'alternatives au système de multiplication FFT dans `internal/bigfft/`.

---

## Suivi des tâches

| # | Tâche | Phase | Agent | Dépend de | Statut |
|---|-------|-------|-------|-----------|--------|
| 1 | Explorer le code FFT actuel (NTT perspective) | 1 - Exploration | `ntt-researcher` | — | `[x]` Terminé |
| 2 | Explorer le code de multiplication (Montgomery perspective) | 1 - Exploration | `montgomery-researcher` | — | `[x]` Terminé |
| 3 | Explorer le code Schönhage-Strassen actuel | 1 - Exploration | `ss-researcher` | — | `[x]` Terminé |
| 4 | Explorer le code et benchmarks pour revue adversariale | 1 - Exploration | `devils-advocate` | — | `[x]` Terminé |
| 5 | Produire le rapport de recherche NTT | 2-3 - Recherche & Rapport | `ntt-researcher` | #1 | `[x]` Terminé |
| 6 | Produire le rapport de recherche Montgomery | 2-3 - Recherche & Rapport | `montgomery-researcher` | #2 | `[x]` Terminé |
| 7 | Produire le rapport d'optimisation Schönhage-Strassen | 2-3 - Recherche & Rapport | `ss-researcher` | #3 | `[x]` Terminé |
| 8 | Produire le contre-rapport adversarial | 4 - Revue adversariale | `devils-advocate` | #5, #6, #7 | `[x]` Terminé |
| 9 | Répondre aux objections NTT | 5 - Réponses | `ntt-researcher` | #8 | `[x]` Terminé |
| 10 | Répondre aux objections Montgomery | 5 - Réponses | `montgomery-researcher` | #8 | `[x]` Terminé |
| 11 | Répondre aux objections Schönhage-Strassen | 5 - Réponses | `ss-researcher` | #8 | `[x]` Terminé |
| 12 | Produire la recommandation finale | 6 - Synthèse | `leader` | #9, #10, #11 | `[x]` Terminé |

**Légende** : `[ ]` En attente · `[~]` En cours · `[x]` Terminé

---

## 1. Objectifs et périmètre

### Objectif principal

Évaluer si le système de multiplication FFT actuel (Schönhage-Strassen modulo 2^n+1) dans `internal/bigfft/` est optimal pour FibGo, ou si une alternative offrirait de meilleures performances, une meilleure maintenabilité, ou un meilleur passage à l'échelle.

### Périmètre de recherche

Trois alternatives à évaluer contre l'implémentation actuelle :

| Alternative                                | Principe                                                                            | Complexité théorique             |
| ------------------------------------------ | ----------------------------------------------------------------------------------- | ---------------------------------- |
| **NTT** (Number Theoretic Transform) | FFT sur corps finis (Z/pZ) au lieu de Fermat (2^n+1)                                | O(n log n log log n)               |
| **Montgomery**                       | Multiplication modulaire avec réduction sans division                              | O(n^2) mais constante très faible |
| **Schönhage-Strassen amélioré**   | Optimisations de l'implémentation actuelle (twiddle factors, cache, vectorisation) | O(n log n log log n)               |

### Hors périmètre

- Remplacement de l'algorithme Fast Doubling (couche `internal/fibonacci/`)
- Modifications de l'orchestration ou de la présentation
- Implémentation effective des alternatives (recherche seulement)

---

## 2. Composition de l'équipe

### 2.1 Agent "NTT Researcher" (`ntt-researcher`)

**Rôle** : Évaluer le Number Theoretic Transform comme alternative à la FFT sur nombres de Fermat.

**Fichiers à analyser** :

- `internal/bigfft/fermat.go` - Arithmétique modulo 2^n+1 (type `fermat`, 340 lignes)
- `internal/bigfft/fft_core.go` - Transformée de Fourier (`fourier`, `fourierWithState`)
- `internal/bigfft/fft_recursion.go` - Récursion FFT parallèle (`fourierRecursiveUnified`)
- `internal/bigfft/fft_poly.go` - Opérations polynomiales (Transform, InvTransform, Mul, Sqr)
- `internal/bigfft/fft.go` - API publique (`Mul`, `Sqr`, `fftSizeThreshold`)
- `internal/fibonacci/fft.go` - Intégration FFT (`mulFFT`, `sqrFFT`, `smartMultiply`, `executeDoublingStepFFT`)

**Livrables** : Rapport NTT selon le template (section 5).

### 2.2 Agent "Montgomery Researcher" (`montgomery-researcher`)

**Rôle** : Évaluer la multiplication de Montgomery pour les grands entiers dans le contexte Fibonacci.

**Fichiers à analyser** :

- `internal/bigfft/fft.go` - Seuils FFT (`defaultFFTThresholdWords = 1800`, `fftSizeThreshold`)
- `internal/bigfft/arith_amd64.go` - Arithmétique vectorielle (linkname vers `math/big`)
- `internal/bigfft/arith_generic.go` - Fallback portable
- `internal/bigfft/arith_decl.go` - Déclarations d'interface assembleur
- `internal/fibonacci/strategy.go` - Stratégies de multiplication (`AdaptiveStrategy`, `FFTOnlyStrategy`, `KaratsubaStrategy`)
- `internal/fibonacci/constants.go` - Seuils (`DefaultFFTThreshold = 500_000`, `DefaultParallelThreshold = 4096`)

**Livrables** : Rapport Montgomery selon le template (section 5).

### 2.3 Agent "Schönhage-Strassen Researcher" (`ss-researcher`)

**Rôle** : Évaluer les optimisations possibles de l'implémentation Schönhage-Strassen actuelle.

**Fichiers à analyser** :

- `internal/bigfft/fft_core.go` - Coeur FFT (fourier, twiddle factors, butterfly)
- `internal/bigfft/fft_recursion.go` - Parallélisme récursif (seuils : `ParallelFFTRecursionThreshold = 4`, `MaxParallelFFTDepth = 3`)
- `internal/bigfft/fft_cache.go` - Cache LRU thread-safe pour transforms
- `internal/bigfft/bump.go` - Bump allocator O(1)
- `internal/bigfft/pool.go` - `sync.Pool` pour `big.Int`
- `internal/bigfft/pool_warming.go` - Pré-chauffage des pools
- `internal/bigfft/cpu_amd64.go` - Détection CPU (AVX2/AVX-512)
- `internal/bigfft/memory_est.go` - Estimation mémoire

**Livrables** : Rapport d'amélioration SS selon le template (section 5).

### 2.4 Agent "Devil's Advocate" (`devils-advocate`)

**Rôle** : Challenger systématiquement chaque proposition sur la complexité d'intégration, les risques de régression et la maintenabilité.

**Fichiers à analyser** :

- Tous les fichiers analysés par les 3 autres agents
- `internal/fibonacci/doubling_framework.go` - Framework de calcul (consommateur principal)
- `internal/fibonacci/calculator.go` - Interface `Calculator` publique
- `internal/fibonacci/options.go` - Configuration (`Options` struct)
- `internal/fibonacci/dynamic_threshold.go` - Ajustement dynamique des seuils
- `Docs/algorithms/COMPARISON.md` - Benchmarks de référence
- Tests : `internal/bigfft/*_test.go` (13 fichiers de tests)

**Livrables** : Contre-rapport selon le protocole (section 7).

---

## 3. Exécution en 6 phases

```
Phase 1          Phase 2          Phase 3        Phase 4          Phase 5        Phase 6
Exploration      Recherche        Rapports       Revue            Réponses       Synthèse
(parallèle)      (parallèle)      (parallèle)    adversariale     (parallèle)    (séquentiel)

┌──────────┐    ┌──────────┐    ┌──────────┐   ┌──────────┐    ┌──────────┐   ┌──────────┐
│NTT       │    │NTT       │    │NTT       │   │          │    │NTT       │   │          │
│explore   │───>│research  │───>│rapport   │──>│ Devil's  │───>│réponse   │──>│          │
├──────────┤    ├──────────┤    ├──────────┤   │ Advocate │    ├──────────┤   │ Leader   │
│Montgomery│    │Montgomery│    │Montgomery│   │          │    │Montgomery│   │ synthèse │
│explore   │───>│research  │───>│rapport   │──>│ contre-  │───>│réponse   │──>│ finale   │
├──────────┤    ├──────────┤    ├──────────┤   │ rapport  │    ├──────────┤   │          │
│SS        │    │SS        │    │SS        │   │          │    │SS        │   │          │
│explore   │───>│research  │───>│rapport   │──>│          │───>│réponse   │──>│          │
└──────────┘    └──────────┘    └──────────┘   └──────────┘    └──────────┘   └──────────┘
```

### Phase 1 : Exploration du code actuel (parallèle, ~5 min)

**Objectif** : Chaque agent lit et comprend le code existant.

- Les 3 chercheurs lisent leurs fichiers assignés (section 2)
- Le Devil's Advocate lit l'ensemble des fichiers et les benchmarks existants
- **Aucune communication inter-agents nécessaire**

**Critère de sortie** : Chaque agent a lu et compris les fichiers assignés.

### Phase 2 : Recherche approfondie (parallèle, ~15 min)

**Objectif** : Chaque chercheur investigue son alternative en profondeur.

- Recherche web sur les implémentations existantes (GMP, FLINT, java.math.BigInteger)
- Analyse de la littérature académique
- Comparaison théorique avec l'implémentation actuelle
- Estimation des performances pour les tailles de F(n) du repo

**Critère de sortie** : Chaque chercheur a les données nécessaires pour son rapport.

### Phase 3 : Rédaction des rapports (parallèle, ~10 min)

**Objectif** : Chaque chercheur produit son rapport selon le template standardisé.

- Rédaction selon le template (section 5)
- Inclusion des estimations de performance
- Envoi du rapport au Devil's Advocate via `SendMessage`

**Critère de sortie** : 3 rapports envoyés au Devil's Advocate.

### Phase 4 : Revue adversariale (séquentiel, ~10 min)

**Objectif** : Le Devil's Advocate challenge chaque proposition.

- Lecture des 3 rapports
- Application du protocole adversarial (section 7)
- Rédaction du contre-rapport unifié
- Envoi du contre-rapport aux 3 chercheurs via `SendMessage`

**Critère de sortie** : Contre-rapport envoyé aux 3 chercheurs.

### Phase 5 : Réponses aux objections (parallèle, ~5 min)

**Objectif** : Chaque chercheur répond aux objections du Devil's Advocate.

- Réponse point par point aux objections
- Ajustement des estimations si nécessaire
- Envoi des réponses au leader

**Critère de sortie** : 3 réponses envoyées au leader.

### Phase 6 : Synthèse finale (séquentiel, ~5 min)

**Objectif** : Le leader produit la recommandation finale.

- Compilation des rapports, contre-rapport et réponses
- Évaluation selon la matrice de décision (section 6)
- Rédaction de la recommandation finale
- Détermination du scénario de sortie (section 9)

**Critère de sortie** : Document de synthèse rédigé.

---

## 4. Méthodologie de recherche par agent

### 4.1 Agent NTT - Questions de recherche

1. **Faisabilité** : La NTT sur Z/pZ peut-elle remplacer la FFT sur 2^n+1 pour la multiplication d'entiers arbitrairement grands ?
2. **Précision** : Quels primes p utiliser pour éviter les erreurs de modular reduction ? Comment gérer le CRT (Chinese Remainder Theorem) pour recombiner ?
3. **Performance** : Quel est le coût du CRT vs le coût de la réduction modulo Fermat ? Pour quelles tailles la NTT devient-elle avantageuse ?
4. **Intégration** : Comment remplacer le type `fermat` (`internal/bigfft/fermat.go`) par une arithmétique Z/pZ ? Quels changements dans `fft_core.go` ?
5. **Références** : Comparer avec l'implémentation NTT de GMP (`mpn_mul_n` avec FFT), FLINT (`fmpz_mul`).

**Démarche** :

1. Lire `fermat.go` pour comprendre l'arithmétique modulo 2^n+1 actuelle
2. Rechercher les primes NTT-friendly (primes de forme k*2^m+1)
3. Estimer le nombre de primes nécessaires par taille d'opérande
4. Comparer le coût CRT vs réduction Fermat
5. Évaluer l'impact sur le cache et la vectorisation

### 4.2 Agent Montgomery - Questions de recherche

1. **Applicabilité** : Montgomery est conçu pour la multiplication modulaire. Comment l'adapter à la multiplication d'entiers libres (non modulaire) utilisée dans Fibonacci ?
2. **Performance** : Pour F(10^6) à F(10^10), la multiplication de Montgomery peut-elle battre Karatsuba (math/big) ou FFT ? À quelles tailles ?
3. **Complémentarité** : Montgomery pourrait-il servir de "tier intermédiaire" entre Karatsuba et FFT dans `smartMultiply` (`internal/fibonacci/fft.go:45`) ?
4. **Assembleur** : Peut-on exploiter les instructions vectorielles AVX2/AVX-512 détectées dans `cpu_amd64.go` pour accélérer Montgomery ?
5. **Références** : Comparer avec OpenSSL BN_mod_mul_montgomery, GMP mpn_redc.

**Démarche** :

1. Lire `strategy.go` pour comprendre l'interface `Multiplier` et les stratégies existantes
2. Analyser `smartMultiply` et `smartSquare` dans `fft.go` pour identifier le crossover Karatsuba→FFT
3. Rechercher les adaptations de Montgomery pour la multiplication non-modulaire
4. Estimer les tailles d'opérandes dans le contexte Fibonacci (F(n) ~ n * log(phi)/log(2) bits)
5. Évaluer si Montgomery peut s'insérer dans la hiérarchie `AdaptiveStrategy`

### 4.3 Agent Schönhage-Strassen - Questions de recherche

1. **Twiddle factors** : L'implémentation actuelle dans `fft_core.go` utilise-t-elle des twiddle factors pré-calculés ? Peut-on les cacher ?
2. **Cache locality** : Le bump allocator (`bump.go`) est-il optimal ? Le pattern d'accès mémoire dans `fourierRecursiveUnified` est-il cache-friendly ?
3. **Vectorisation** : Les opérations `addVV`, `subVV`, `addMulVVW` dans `arith_amd64.go` exploitent-elles pleinement AVX2/AVX-512 ?
4. **Parallélisme** : Les seuils de parallélisme récursif (`ParallelFFTRecursionThreshold = 4`, `MaxParallelFFTDepth = 3`) sont-ils optimaux ?
5. **Transform cache** : Le cache LRU dans `fft_cache.go` est-il dimensionné de manière optimale ? Le taux de hit est-il mesuré ?
6. **Pool warming** : Le pré-chauffage dans `pool_warming.go` est-il calibré pour les tailles typiques de F(n) ?
7. **Algorithme de Harvey-van der Hoeven** : L'algorithme O(n log n) de 2019 est-il implémentable en pratique pour les tailles de FibGo ?

**Démarche** :

1. Profiler mentalement le chemin critique : `fftmul` → `fourier` → `fourierRecursiveUnified` → opérations `fermat`
2. Identifier les allocations évitables et les copies inutiles
3. Rechercher les optimisations dans les implémentations de référence (GMP, FLINT, Python `_pystackless`)
4. Estimer le gain potentiel de chaque optimisation
5. Prioriser par rapport effort/impact

---

## 5. Modèle de rapport standardisé

Chaque chercheur doit produire son rapport en suivant ce template :

```markdown
# Rapport : [Nom de l'alternative]

## Résumé exécutif
[2-3 phrases. Recommandation claire : adopter / rejeter / approfondir]

## 1. Description de l'approche
[Principe algorithmique, complexité théorique, variantes]

## 2. Analyse du code actuel
[Points forts et faiblesses de l'implémentation actuelle
qui motivent l'exploration de cette alternative]

### Fichiers analysés
| Fichier | Lignes | Observations clés |
|---------|--------|-------------------|
| ... | ... | ... |

## 3. Comparaison théorique

### Complexité
| Opération | Actuel (SS/Fermat) | Alternative | Avantage |
|-----------|-------------------|-------------|----------|
| Multiplication | O(n log n log log n) | ... | ... |
| Squaring | ... | ... | ... |
| Transform | ... | ... | ... |

### Constantes cachées
[Analyse des constantes multiplicatives, overhead mémoire,
impact cache L1/L2/L3]

## 4. Estimation de performance

### Données de référence FibGo
Tailles d'opérandes pour les N du benchmark COMPARISON.md :
- F(10^4) : ~2,090 chiffres (~6,942 bits)
- F(10^5) : ~20,899 chiffres (~69,424 bits)
- F(10^6) : ~208,988 chiffres (~694,241 bits)
- F(10^7) : ~2,089,877 chiffres (~6,942,419 bits)
- F(10^8) : ~20,898,764 chiffres (~69,424,191 bits)
- F(5*10^8) : ~104,493,821 chiffres (~347,120,959 bits)

### Estimation par taille
| N | Bits F(n) | Actuel (estimé) | Alternative (estimé) | Ratio |
|---|-----------|-----------------|---------------------|-------|
| 10^6 | 694K | ... | ... | ... |
| 10^7 | 6.9M | ... | ... | ... |
| 10^8 | 69.4M | ... | ... | ... |
| 10^9 | 694M | ... | ... | ... |
| 10^10 | 6.9G | ... | ... | ... |

### Seuil de crossover
[À partir de quelle taille l'alternative devient avantageuse]

## 5. Effort d'intégration

### Architecture d'intégration
[Comment l'alternative s'insère dans l'architecture existante]

### Fichiers à modifier
| Fichier | Type de modification | Effort estimé |
|---------|---------------------|---------------|
| ... | ... | Faible/Moyen/Élevé |

### Points d'intégration clés
- Interface `Multiplier` (strategy.go) : [comment s'adapter]
- Interface `DoublingStepExecutor` (strategy.go) : [impact sur ExecuteStep]
- Options/seuils (options.go, constants.go) : [nouveaux paramètres]
- Tests (bigfft/*_test.go) : [effort de test]

### Effort total estimé
[En jours-développeur, avec hypothèses]

## 6. Risques
| Risque | Probabilité | Impact | Mitigation |
|--------|-------------|--------|------------|
| ... | Faible/Moyen/Élevé | ... | ... |

## 7. Recommandation
[Adopter / Rejeter / Approfondir, avec justification]
```

---

## 6. Critères d'évaluation

### 6.1 Critères quantitatifs

| Critère                                   | Poids | Description                                       | Mesure                              |
| ------------------------------------------ | ----- | ------------------------------------------------- | ----------------------------------- |
| **Performance N moyen** (10^6-10^7)  | 25%   | Gain de performance pour les cas d'usage typiques | Ratio temps estimé vs actuel       |
| **Performance N grand** (10^8-10^10) | 20%   | Gain pour les calculs extrêmes                   | Ratio temps estimé vs actuel       |
| **Consommation mémoire**            | 15%   | Impact sur l'empreinte mémoire                   | Ratio mémoire estimée vs actuelle |
| **Passage à l'échelle**            | 10%   | Comportement asymptotique au-delà de F(10^10)    | Complexité théorique              |

### 6.2 Critères qualitatifs

| Critère                        | Poids | Description                                      | Échelle                       |
| ------------------------------- | ----- | ------------------------------------------------ | ------------------------------ |
| **Effort d'intégration** | 15%   | Quantité de code à modifier/ajouter            | 1 (trivial) - 5 (réécriture) |
| **Maintenabilité**       | 10%   | Lisibilité, testabilité, complexité cognitive | 1 (simple) - 5 (obscur)        |
| **Risque de régression** | 5%    | Probabilité d'introduire des bugs               | 1 (faible) - 5 (élevé)       |

### 6.3 Matrice de décision pondérée

```
Score(alternative) = Σ (poids_i × note_i)

Où note_i ∈ [1, 5] pour chaque critère.

Seuils de décision :
- Score > 3.5 : Recommandation forte d'adoption
- Score 2.5-3.5 : Approfondissement recommandé (prototype)
- Score < 2.5  : Rejet
```

---

## 7. Protocole de l'Avocat du Diable

### 7.1 Principes d'engagement

1. **Adversarial, pas hostile** : L'objectif est de renforcer les propositions, pas de les détruire
2. **Fondé sur le code** : Toute objection doit référencer des fichiers et lignes spécifiques du repo
3. **Quantitatif** : Les objections de performance doivent inclure des estimations chiffrées
4. **Constructif** : Chaque objection doit proposer un critère de résolution

### 7.2 Structure du contre-rapport

```markdown
# Contre-rapport : Revue adversariale

## Synthèse des objections

| Proposition | Objections critiques | Objections mineures | Verdict préliminaire |
|-------------|---------------------|---------------------|---------------------|
| NTT | ... | ... | Viable / Douteux / Rejeté |
| Montgomery | ... | ... | Viable / Douteux / Rejeté |
| SS amélioré | ... | ... | Viable / Douteux / Rejeté |

## Objections détaillées

### [Proposition X]

#### Objection 1 : [Titre]
- **Catégorie** : Performance / Intégration / Maintenabilité / Risque
- **Sévérité** : Bloquante / Majeure / Mineure
- **Argument** : [Description détaillée avec références au code]
- **Fichiers concernés** : [fichier:ligne]
- **Critère de résolution** : [Ce qui rendrait l'objection caduque]

...

## Éléments de comparaison transversaux
[Points où les propositions se contredisent ou se complètent]

## Questions ouvertes pour les chercheurs
[Questions nécessitant une réponse avant la recommandation finale]
```

### 7.3 Objections anticipées par alternative

**NTT** :

- Overhead du CRT (Chinese Remainder Theorem) pour recombiner les résultats de plusieurs primes
- Choix des primes NTT-friendly : impact sur la taille maximale des opérandes
- Perte de la propriété "squaring saves 33%" (la FFT Fermat optimise le squaring nativement via `Sqr` dans `fft_poly.go`)
- Complexité d'implémentation des primitives mod-p vs arithmétique Fermat existante

**Montgomery** :

- Montgomery est conçu pour la multiplication *modulaire* (opérandes < modulus). F(n) croît sans borne : pas de modulus naturel
- Pour la multiplication non-modulaire, Montgomery n'apporte pas d'avantage sur Karatsuba (math/big)
- L'overhead de conversion Montgomery form / standard form annule les gains pour des multiplications isolées
- Ne remplace pas FFT pour n > 500K bits : complexité O(n^2) vs O(n log n)

**Schönhage-Strassen amélioré** :

- Optimisations incrémentales : gains marginaux (5-15%) vs effort de développement
- Risque de régression sur les 13 fichiers de tests existants dans `internal/bigfft/`
- Les optimisations les plus impactantes (vectorisation AVX-512) dépendent du hardware
- L'algorithme Harvey-van der Hoeven O(n log n) n'est avantageux qu'au-delà de 10^12 bits (hors périmètre FibGo)

---

## 8. Plan de communication

### 8.1 Flux de messages

```
Phase 3 (rapports) :
  ntt-researcher      ──SendMessage──> devils-advocate  (rapport NTT)
  montgomery-researcher──SendMessage──> devils-advocate  (rapport Montgomery)
  ss-researcher        ──SendMessage──> devils-advocate  (rapport SS)

Phase 4 (revue) :
  devils-advocate ──SendMessage──> ntt-researcher       (objections NTT)
  devils-advocate ──SendMessage──> montgomery-researcher (objections Montgomery)
  devils-advocate ──SendMessage──> ss-researcher         (objections SS)

Phase 5 (réponses) :
  ntt-researcher       ──SendMessage──> leader  (réponse aux objections)
  montgomery-researcher──SendMessage──> leader  (réponse aux objections)
  ss-researcher         ──SendMessage──> leader  (réponse aux objections)
  devils-advocate       ──SendMessage──> leader  (synthèse adversariale)
```

### 8.2 Format de message

```
Objet : [Phase] [Agent] - [Sujet]
Exemple : "Phase 3 - Rapport NTT complet"
```

### 8.3 Points de synchronisation

| Point       | Condition                               | Action si non remplie                  |
| ----------- | --------------------------------------- | -------------------------------------- |
| Fin Phase 1 | 4 agents ont lu leurs fichiers          | Leader relance les agents bloqués     |
| Fin Phase 3 | 3 rapports reçus par Devil's Advocate  | Leader vérifie via TaskList           |
| Fin Phase 4 | Contre-rapport envoyé aux 3 chercheurs | Leader déclenche Phase 5 manuellement |
| Fin Phase 5 | 4 réponses reçues par le leader       | Leader commence la synthèse           |

---

## 9. Cadre de décision

### 9.1 Scénarios de sortie

**Scénario A - Adoption claire** : Une alternative obtient un score > 3.5 et le Devil's Advocate n'a pas d'objection bloquante.

> Action : Créer un plan d'implémentation détaillé (nouveau scénario Agent Teams).

**Scénario B - Approfondissement** : Une ou plusieurs alternatives obtiennent un score entre 2.5 et 3.5.

> Action : Lancer un prototype limité (implémentation de `Multiplier` pour l'alternative) et benchmarker.

**Scénario C - Optimisation incrémentale** : Aucune alternative ne dépasse 2.5, mais des optimisations SS sont identifiées.

> Action : Intégrer les optimisations dans `internal/bigfft/` via un scénario de type #4 (optimisation cross-layer).

**Scénario D - Statu quo** : Le système actuel est jugé optimal par consensus.

> Action : Documenter la décision et les raisons dans `Docs/algorithms/`.

### 9.2 Format de recommandation finale

```markdown
# Recommandation finale : Recherche exploratoire bigfft

## Décision : [Scénario A/B/C/D]

## Matrice de scores

| Critère | Poids | NTT | Montgomery | SS amélioré |
|---------|-------|-----|------------|-------------|
| Perf N moyen | 25% | .../5 | .../5 | .../5 |
| Perf N grand | 20% | .../5 | .../5 | .../5 |
| Mémoire | 15% | .../5 | .../5 | .../5 |
| Scalabilité | 10% | .../5 | .../5 | .../5 |
| Effort intégration | 15% | .../5 | .../5 | .../5 |
| Maintenabilité | 10% | .../5 | .../5 | .../5 |
| Risque régression | 5% | .../5 | .../5 | .../5 |
| **Score pondéré** | 100% | **...** | **...** | **...** |

## Justification
[Synthèse des arguments des chercheurs et du Devil's Advocate]

## Prochaines étapes
[Actions concrètes selon le scénario choisi]
```

---

## 10. Commandes Claude Code

### 10.1 Création de l'équipe

```
Crée une équipe de recherche "bigfft-research" pour évaluer des alternatives
au système de multiplication FFT dans internal/bigfft/.
```

Cela exécute :

```json
TeamCreate {
  "team_name": "bigfft-research",
  "description": "Recherche exploratoire : alternatives au système FFT de bigfft"
}
```

### 10.2 Création des tâches

#### Phase 1 - Exploration

```json
TaskCreate {
  "subject": "Explorer le code FFT actuel (NTT perspective)",
  "description": "Lire et analyser les fichiers internal/bigfft/fermat.go, fft_core.go, fft_recursion.go, fft_poly.go, fft.go et internal/fibonacci/fft.go. Comprendre l'arithmétique modulo 2^n+1, la structure de la transformée, et les optimisations existantes (squaring, transform reuse). Résumer les points clés pertinents pour une comparaison avec NTT sur Z/pZ.",
  "activeForm": "Exploring FFT code for NTT comparison"
}
```

```json
TaskCreate {
  "subject": "Explorer le code de multiplication (Montgomery perspective)",
  "description": "Lire et analyser les fichiers internal/bigfft/fft.go (seuils), arith_amd64.go, arith_generic.go, arith_decl.go, internal/fibonacci/strategy.go et constants.go. Comprendre la hiérarchie des stratégies (Adaptive, FFT, Karatsuba), les seuils de crossover (DefaultFFTThreshold = 500000 bits), et les opérations vectorielles disponibles. Résumer les points clés pertinents pour évaluer Montgomery comme tier intermédiaire.",
  "activeForm": "Exploring multiplication strategies for Montgomery analysis"
}
```

```json
TaskCreate {
  "subject": "Explorer le code Schönhage-Strassen actuel",
  "description": "Lire et analyser les fichiers internal/bigfft/fft_core.go, fft_recursion.go, fft_cache.go, bump.go, pool.go, pool_warming.go, cpu_amd64.go et memory_est.go. Identifier les points d'optimisation potentiels : twiddle factors, cache locality, vectorisation, parallélisme récursif. Résumer les opportunités d'amélioration.",
  "activeForm": "Exploring Schönhage-Strassen implementation details"
}
```

```json
TaskCreate {
  "subject": "Explorer le code et les benchmarks pour la revue adversariale",
  "description": "Lire les fichiers de tous les autres agents plus internal/fibonacci/doubling_framework.go, calculator.go, options.go, dynamic_threshold.go et Docs/algorithms/COMPARISON.md. Comprendre comment le système FFT s'intègre dans l'architecture globale, les benchmarks de référence, et les contraintes d'intégration. Préparer les axes de challenge.",
  "activeForm": "Exploring codebase for adversarial review preparation"
}
```

#### Phase 2 & 3 - Recherche et rapports

```json
TaskCreate {
  "subject": "Produire le rapport de recherche NTT",
  "description": "Rechercher en profondeur le Number Theoretic Transform comme alternative à la FFT Fermat. Couvrir : faisabilité (primes NTT-friendly, CRT), performance estimée pour F(10^6) à F(10^10), effort d'intégration (remplacement du type fermat, modifications de fft_core.go). Comparer avec GMP et FLINT. Rédiger le rapport selon le template standardisé et l'envoyer au devils-advocate.",
  "activeForm": "Researching NTT alternative and writing report"
}
```

```json
TaskCreate {
  "subject": "Produire le rapport de recherche Montgomery",
  "description": "Rechercher en profondeur la multiplication de Montgomery pour grands entiers dans le contexte Fibonacci. Couvrir : applicabilité (multiplication non-modulaire), performance estimée, rôle comme tier intermédiaire dans AdaptiveStrategy, exploitation AVX2/AVX-512. Comparer avec OpenSSL et GMP. Rédiger le rapport selon le template standardisé et l'envoyer au devils-advocate.",
  "activeForm": "Researching Montgomery alternative and writing report"
}
```

```json
TaskCreate {
  "subject": "Produire le rapport d'optimisation Schönhage-Strassen",
  "description": "Rechercher les optimisations applicables à l'implémentation SS actuelle. Couvrir : twiddle factors pré-calculés, cache locality du bump allocator, vectorisation AVX-512 des opérations fermat, calibration des seuils de parallélisme, faisabilité de Harvey-van der Hoeven. Rédiger le rapport selon le template standardisé et l'envoyer au devils-advocate.",
  "activeForm": "Researching SS optimizations and writing report"
}
```

#### Phase 4 - Revue adversariale

```json
TaskCreate {
  "subject": "Produire le contre-rapport adversarial",
  "description": "Après réception des 3 rapports de recherche, challenger systématiquement chaque proposition. Appliquer le protocole adversarial : objections fondées sur le code (avec références fichier:ligne), estimations chiffrées, critères de résolution. Produire le contre-rapport unifié selon le template et l'envoyer aux 3 chercheurs pour réponse.",
  "activeForm": "Writing adversarial counter-report"
}
```

#### Phase 5 - Réponses

```json
TaskCreate {
  "subject": "Répondre aux objections NTT",
  "description": "Répondre point par point aux objections du Devil's Advocate concernant la proposition NTT. Ajuster les estimations si nécessaire. Envoyer la réponse au leader.",
  "activeForm": "Responding to NTT objections"
}
```

```json
TaskCreate {
  "subject": "Répondre aux objections Montgomery",
  "description": "Répondre point par point aux objections du Devil's Advocate concernant la proposition Montgomery. Ajuster les estimations si nécessaire. Envoyer la réponse au leader.",
  "activeForm": "Responding to Montgomery objections"
}
```

```json
TaskCreate {
  "subject": "Répondre aux objections Schönhage-Strassen",
  "description": "Répondre point par point aux objections du Devil's Advocate concernant les optimisations SS. Ajuster les estimations si nécessaire. Envoyer la réponse au leader.",
  "activeForm": "Responding to SS objections"
}
```

#### Phase 6 - Synthèse

```json
TaskCreate {
  "subject": "Produire la recommandation finale",
  "description": "Compiler les 3 rapports, le contre-rapport et les 3 réponses. Évaluer chaque alternative selon la matrice de décision pondérée (7 critères). Déterminer le scénario de sortie (A: adoption, B: prototype, C: optimisation incrémentale, D: statu quo). Rédiger la recommandation finale selon le format défini.",
  "activeForm": "Producing final recommendation"
}
```

### 10.3 Dépendances entre tâches

```
Phase 1 (tâches 1-4)  : aucune dépendance (parallèle)
Phase 2-3 (tâches 5-7): bloquées par les tâches 1-3 respectivement
Phase 4 (tâche 8)     : bloquée par les tâches 5, 6, 7
Phase 5 (tâches 9-11) : bloquées par la tâche 8
Phase 6 (tâche 12)    : bloquée par les tâches 9, 10, 11
```

Après création, établir les dépendances :

```json
TaskUpdate { "taskId": "5", "addBlockedBy": ["1"] }
TaskUpdate { "taskId": "6", "addBlockedBy": ["2"] }
TaskUpdate { "taskId": "7", "addBlockedBy": ["3"] }
TaskUpdate { "taskId": "8", "addBlockedBy": ["5", "6", "7"] }
TaskUpdate { "taskId": "9", "addBlockedBy": ["8"] }
TaskUpdate { "taskId": "10", "addBlockedBy": ["8"] }
TaskUpdate { "taskId": "11", "addBlockedBy": ["8"] }
TaskUpdate { "taskId": "12", "addBlockedBy": ["9", "10", "11"] }
```

### 10.4 Spawn des agents

```json
Task {
  "subagent_type": "general-purpose",
  "name": "ntt-researcher",
  "team_name": "bigfft-research",
  "description": "NTT research agent",
  "prompt": "Tu es l'agent NTT Researcher dans l'équipe bigfft-research. Ta mission est d'évaluer le Number Theoretic Transform comme alternative à la FFT sur nombres de Fermat dans internal/bigfft/. Commence par lire le fichier AgentsTeam/SearchPlan.md pour comprendre le plan complet, puis consulte TaskList pour trouver tes tâches assignées. Suis la méthodologie de la section 4.1 et rédige ton rapport selon le template de la section 5. Envoie ton rapport au devils-advocate quand il est prêt."
}
```

```json
Task {
  "subagent_type": "general-purpose",
  "name": "montgomery-researcher",
  "team_name": "bigfft-research",
  "description": "Montgomery research agent",
  "prompt": "Tu es l'agent Montgomery Researcher dans l'équipe bigfft-research. Ta mission est d'évaluer la multiplication de Montgomery pour les grands entiers dans le contexte Fibonacci. Commence par lire le fichier AgentsTeam/SearchPlan.md pour comprendre le plan complet, puis consulte TaskList pour trouver tes tâches assignées. Suis la méthodologie de la section 4.2 et rédige ton rapport selon le template de la section 5. Envoie ton rapport au devils-advocate quand il est prêt."
}
```

```json
Task {
  "subagent_type": "general-purpose",
  "name": "ss-researcher",
  "team_name": "bigfft-research",
  "description": "SS optimization research agent",
  "prompt": "Tu es l'agent Schönhage-Strassen Researcher dans l'équipe bigfft-research. Ta mission est d'évaluer les optimisations possibles de l'implémentation Schönhage-Strassen actuelle dans internal/bigfft/. Commence par lire le fichier AgentsTeam/SearchPlan.md pour comprendre le plan complet, puis consulte TaskList pour trouver tes tâches assignées. Suis la méthodologie de la section 4.3 et rédige ton rapport selon le template de la section 5. Envoie ton rapport au devils-advocate quand il est prêt."
}
```

```json
Task {
  "subagent_type": "general-purpose",
  "name": "devils-advocate",
  "team_name": "bigfft-research",
  "description": "Devil's advocate agent",
  "prompt": "Tu es l'agent Devil's Advocate dans l'équipe bigfft-research. Ta mission est de challenger systématiquement les propositions des 3 agents chercheurs (NTT, Montgomery, SS). Commence par lire le fichier AgentsTeam/SearchPlan.md (sections 7 et 2.4) pour comprendre ton protocole. Pendant la phase 1, explore le code (fichiers listés en section 2.4) et les benchmarks COMPARISON.md. Puis attends les 3 rapports avant de rédiger ton contre-rapport selon le template de la section 7.2. Sois adversarial mais constructif : chaque objection doit référencer du code et proposer un critère de résolution."
}
```

### 10.5 Assignation des tâches

```json
TaskUpdate { "taskId": "1", "owner": "ntt-researcher" }
TaskUpdate { "taskId": "2", "owner": "montgomery-researcher" }
TaskUpdate { "taskId": "3", "owner": "ss-researcher" }
TaskUpdate { "taskId": "4", "owner": "devils-advocate" }
TaskUpdate { "taskId": "5", "owner": "ntt-researcher" }
TaskUpdate { "taskId": "6", "owner": "montgomery-researcher" }
TaskUpdate { "taskId": "7", "owner": "ss-researcher" }
TaskUpdate { "taskId": "8", "owner": "devils-advocate" }
TaskUpdate { "taskId": "9", "owner": "ntt-researcher" }
TaskUpdate { "taskId": "10", "owner": "montgomery-researcher" }
TaskUpdate { "taskId": "11", "owner": "ss-researcher" }
TaskUpdate { "taskId": "12", "owner": "leader" }
```

### 10.6 Orchestration par le leader

Le leader (vous) surveille la progression via :

```
TaskList    → voir l'état de toutes les tâches
TaskGet     → lire les détails d'une tâche spécifique
```

En cas de blocage, relancer un agent :

```json
SendMessage {
  "type": "message",
  "recipient": "ntt-researcher",
  "content": "Phase 1 terminée pour les autres agents. Merci de finaliser ton exploration et passer à la phase 2.",
  "summary": "Reminder to complete Phase 1"
}
```

---

## 11. Annexes

### A. Données de référence

#### Tailles de F(n) (bits)

| N      | Bits de F(n) | Mots (64-bit) | Taille FFT (k) |
| ------ | ------------ | ------------- | -------------- |
| 10^4   | 6,942        | 109           | < seuil FFT    |
| 10^5   | 69,424       | 1,085         | < seuil FFT    |
| 10^6   | 694,241      | 10,848        | k=5-6          |
| 10^7   | 6,942,419    | 108,476       | k=8-9          |
| 10^8   | 69,424,191   | 1,084,753     | k=11-12        |
| 5*10^8 | 347,120,959  | 5,423,765     | k=13-14        |

#### Seuils FFT actuels (`fft.go:184`)

```go
var fftSizeThreshold = [...]int64{0, 0, 0,
    4<<10, 8<<10, 16<<10,           // k=3,4,5
    32<<10, 64<<10, 1<<18, 1<<20,   // k=6,7,8,9
    3<<20, 8<<20, 30<<20, 100<<20,  // k=10,11,12,13
    300<<20, 600<<20,               // k=14,15
}
```

#### Seuils de stratégie (`constants.go`)

| Constante                             | Valeur    | Unité               |
| ------------------------------------- | --------- | -------------------- |
| `DefaultParallelThreshold`          | 4,096     | bits                 |
| `DefaultFFTThreshold`               | 500,000   | bits                 |
| `DefaultStrassenThreshold`          | 3,072     | bits                 |
| `ParallelFFTThreshold`              | 5,000,000 | bits                 |
| `defaultFFTThresholdWords` (bigfft) | 1,800     | mots = ~115,200 bits |

#### Benchmarks de référence (`COMPARISON.md`)

| N      | Fast Doubling | FFT-Based | Ratio FD/FFT |
| ------ | ------------- | --------- | ------------ |
| 10^5   | 3.2ms         | 5.8ms     | 0.55x        |
| 5*10^5 | 35ms          | 42ms      | 0.83x        |
| 10^6   | 85ms          | 95ms      | 0.89x        |
| 10^7   | 2.1s          | 2.3s      | 0.91x        |
| 10^8   | 45s           | 48s       | 0.94x        |
| 5*10^8 | 8m45s         | 9m15s     | 0.95x        |

**Observation** : Le ratio converge vers 1.0 pour les grands N, indiquant que la multiplication FFT est le facteur dominant. Les alternatives doivent améliorer cette couche pour avoir un impact.

### B. Risques du projet de recherche

| Risque                                                   | Probabilité | Impact                  | Mitigation                                                          |
| -------------------------------------------------------- | ------------ | ----------------------- | ------------------------------------------------------------------- |
| Biais de confirmation des chercheurs                     | Moyen        | Conclusions biaisées   | Devil's Advocate dédié                                            |
| Estimations de performance imprécises                   | Élevé      | Mauvaise recommandation | Croiser avec benchmarks GMP/FLINT publiés                          |
| Scope creep (exploration trop large)                     | Moyen        | Dépassement de temps   | Template de rapport structuré                                      |
| Alternative théoriquement supérieure mais impraticable | Moyen        | Effort gaspillé        | Critère d'effort d'intégration pondéré à 15%                   |
| Insuffisance de la recherche web                         | Faible       | Données manquantes     | Références croisées (publications, implémentations open-source) |
