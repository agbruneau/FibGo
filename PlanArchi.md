# Plan d'analyse et d'assainissement de la documentation architecturale

**Projet** : FibGo / FibCalc
**Date** : 2026-02-08
**Objectif** : Épurer, consolider et mettre à jour la documentation en s'appuyant sur `Docs/architecture/` comme source de vérité architecturale.

---

## 0. Tableau de suivi des tâches

| ID | Phase | Agent | Tâche | Statut | Notes |
|----|-------|-------|-------|--------|-------|
| A1 | 1 | A1 | Supprimer `AGENTS.md`, transformer `Docs/ARCHITECTURE.md` en redirection | ✅ Terminé | AGENTS.md supprimé (make build/clean migrés vers CLAUDE.md), ARCHITECTURE.md redirigé vers architecture/README.md |
| A2 | 1 | A2 | Fusionner les catalogues de design patterns | ✅ Terminé | 12 patterns fusionnés dans architecture/patterns/design-patterns.md (1023 lignes), DESIGN_PATTERNS.md redirigé |
| A3 | 1 | A3 | Consolider documentation TUI et FFT | ✅ Terminé | R4: pas d'action (tui-flow.md inexistant), R5: FFT.md épuré, cross-refs ajoutées |
| B1 | 2 | B1 | Auditer diagrammes C4 et graphe de dépendances vs code | ✅ Terminé | 37+ corrections: container-diagram (5 arêtes), component-diagram (18+ signatures), dependency-graph (14 arêtes) |
| B2 | 2 | B2 | Auditer interfaces et patterns documentés vs code | ✅ Terminé | 32 interfaces vérifiées, 2 corrections mineures (BumpAllocator sig + marge sécurité), 4 interfaces non documentées identifiées |
| B3 | 2 | B3 | Auditer flux d'exécution documentés vs appels réels | ✅ Terminé | Data Flow README corrigé, 4 flow docs créés (cli-flow, tui-flow, config-flow, algorithm-flows) avec diagrammes Mermaid |
| B4 | 2 | B4 | Auditer guides opérationnels vs état réel | ✅ Terminé | 14 corrections sur 6 fichiers (README, BUILD, PERFORMANCE, TESTING, .env.example), CALIBRATION.md 100% exact |
| C1 | 3 | C1 | Mettre à jour `architecture/README.md` (index, stats, navigation) | ✅ Terminé | Documentation Map ajoutée, stats codebase, 12 patterns référencés, 21 liens vérifiés (416→479 lignes) |
| C2 | 3 | C2 | Régénérer `validation-report.md` à partir des audits | ✅ Terminé | 38 PASS, 0 FAIL, 0 WARNING, 4 INFO — 53+ corrections documentées, 6 anciens WARNINGs résolus |
| C3 | 3 | C3 | Aligner `README.md` et `CLAUDE.md` avec la structure épurée | ✅ Terminé | Fuzz targets corrigés (17→4), formal verification marquée planifiée, linters 22→24, go.uber.org/mock retiré, arborescence projet mise à jour |

---

## 1. Inventaire de la documentation existante

### 1.1 Racine du projet (5 fichiers — ~1 291 lignes)

| Fichier | Lignes | Rôle | Remarques |
|---------|--------|------|-----------|
| `README.md` | 551 | Point d'entrée utilisateur | Complet, sections performance/architecture à aligner |
| `CLAUDE.md` | 228 | Guide IA (Claude Code) | Très détaillé, source de vérité pour conventions |
| `AGENTS.md` | 56 | Guide IA abrégé | Redondant avec CLAUDE.md — candidat à la suppression |
| `CONTRIBUTING.md` | 370 | Guide de contribution | Autonome, peu de chevauchement |
| `CHANGELOG.md` | 86 | Historique des versions | À jour, à maintenir |

### 1.2 Docs/ — Guides thématiques (7 fichiers — ~2 740 lignes)

| Fichier | Lignes | Rôle | Chevauchement identifié |
|---------|--------|------|------------------------|
| `Docs/ARCHITECTURE.md` | 410 | Vue d'ensemble architecturale | **Fort** avec `Docs/architecture/README.md` |
| `Docs/BUILD.md` | 294 | Système de build | Faible — autonome |
| `Docs/CALIBRATION.md` | 351 | Guide de calibration | Faible — complémentaire avec `flows/config-flow.md` |
| `Docs/DESIGN_PATTERNS.md` | 646 | Catalogue de patterns | **Fort** avec `architecture/patterns/design-patterns.md` |
| `Docs/PERFORMANCE.md` | 265 | Optimisations performance | Modéré — benchmarks à valider |
| `Docs/TESTING.md` | 385 | Stratégie de test | Faible — autonome |
| `Docs/TUI_GUIDE.md` | 389 | Guide développeur TUI | Modéré avec `flows/tui-flow.md` |

### 1.3 Docs/algorithms/ — Documentation algorithmique (7 fichiers — ~2 298 lignes)

| Fichier | Lignes | Rôle | Chevauchement identifié |
|---------|--------|------|------------------------|
| `algorithms/BIGFFT.md` | 700 | Sous-système BigFFT | Modéré avec `flows/fft-pipeline.mermaid` |
| `algorithms/COMPARISON.md` | 225 | Comparaison des algorithmes | Faible — unique |
| `algorithms/FAST_DOUBLING.md` | 291 | Fast Doubling | Modéré avec `flows/fastdoubling.mermaid` |
| `algorithms/FFT.md` | 268 | Théorie FFT | Modéré avec `BIGFFT.md` (théorie vs impl.) |
| `algorithms/GMP.md` | 98 | Intégration GMP | Faible — autonome |
| `algorithms/MATRIX.md` | 343 | Exponentiation matricielle | Modéré avec `flows/matrix.mermaid` |
| `algorithms/PROGRESS_BAR_ALGORITHM.md` | 373 | Algorithme de progression | Faible — spécialisé |

### 1.4 Docs/architecture/ — Documentation architecturale de référence (~1 911 lignes)

**Source de vérité architecturale — 16 fichiers :**

| Fichier | Lignes | Type | Contenu |
|---------|--------|------|---------|
| `architecture/README.md` | 221 | Index | Vue d'ensemble, statistiques, index des livrables |
| `architecture/system-context.mermaid` | 15 | C4 L1 | Diagramme de contexte système |
| `architecture/container-diagram.mermaid` | 30 | C4 L2 | Diagramme de conteneurs |
| `architecture/component-diagram.mermaid` | 188 | C4 L3 | Diagramme de composants détaillé |
| `architecture/dependency-graph.mermaid` | 75 | Graphe | DAG des dépendances inter-paquets |
| `architecture/flows/algorithm-flows.md` | 152 | Flow | Flux d'exécution des algorithmes |
| `architecture/flows/cli-flow.md` | 111 | Flow | Flux d'exécution CLI |
| `architecture/flows/cli-flow.mermaid` | 68 | Flow | Diagramme CLI |
| `architecture/flows/config-flow.md` | 153 | Flow | Flux de configuration |
| `architecture/flows/config-flow.mermaid` | 69 | Flow | Diagramme configuration |
| `architecture/flows/tui-flow.md` | 134 | Flow | Flux d'exécution TUI |
| `architecture/flows/tui-flow.mermaid` | 62 | Flow | Diagramme TUI |
| `architecture/flows/fastdoubling.mermaid` | 71 | Flow | Diagramme Fast Doubling |
| `architecture/flows/fft-pipeline.mermaid` | 66 | Flow | Diagramme pipeline FFT |
| `architecture/flows/matrix.mermaid` | 53 | Flow | Diagramme Matrix Exponentiation |
| `architecture/patterns/design-patterns.md` | 220 | Patterns | 13 patterns avec références code |
| `architecture/patterns/interface-hierarchy.mermaid` | 144 | Patterns | Hiérarchie d'interfaces |
| `architecture/validation/validation-report.md` | 157 | Validation | 22 PASS, 0 FAIL, 6 WARNING |

---

## 2. Diagnostic : redondances et problèmes identifiés

### 2.1 Redondances critiques (à résoudre en priorité)

| # | Documents en conflit | Nature | Action recommandée |
|---|---------------------|--------|-------------------|
| R1 | `AGENTS.md` vs `CLAUDE.md` | Duplication quasi-totale | Supprimer `AGENTS.md`, conserver `CLAUDE.md` |
| R2 | `Docs/ARCHITECTURE.md` vs `architecture/README.md` | Vue d'ensemble dupliquée | Transformer `Docs/ARCHITECTURE.md` en redirection vers `architecture/README.md` |
| R3 | `Docs/DESIGN_PATTERNS.md` (646 l.) vs `architecture/patterns/design-patterns.md` (220 l.) | Catalogue de patterns dupliqué | Fusionner dans `architecture/patterns/`, supprimer le fichier Docs/ |
| R4 | `Docs/TUI_GUIDE.md` vs `architecture/flows/tui-flow.md` | Flux TUI dupliqué | Consolider : garder le guide développeur dans Docs/, référencer les flows |
| R5 | `algorithms/FFT.md` vs `algorithms/BIGFFT.md` | Théorie vs implémentation FFT, chevauchement partiel | Clarifier les frontières, éliminer les sections en double |

### 2.2 Problèmes de cohérence potentiels

| # | Problème | Impact | Vérification requise |
|---|----------|--------|---------------------|
| C1 | `README.md` sections architecture/performance vs docs détaillées | Risque de dérive | Comparer les métriques, schémas, exemples |
| C2 | Diagrammes Mermaid vs code source actuel | Interfaces/packages manquants ou renommés | Valider chaque interface et dépendance |
| C3 | Rapport de validation daté 2026-02-08 | Potentiellement déjà obsolète si code modifié depuis | Re-exécuter la validation |
| C4 | Conventions de code (`CLAUDE.md`) vs pratiques réelles | Dérive possible | Audit échantillonné du code |

### 2.3 Dead code documentaire (d'après le rapport de validation)

- `RenderBrailleChart` — documenté mais potentiellement inutilisé
- `MultiplicationStrategy` — alias déprécié, encore documenté comme actif
- Utilitaires de test non utilisés dans certaines docs

---

## 3. Stratégie d'exécution par équipes d'agents

### 3.1 Organisation des équipes

```
                    ┌─────────────────────┐
                    │  Agent Coordinateur  │
                    │  (Orchestration)     │
                    └──────────┬──────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
    ┌─────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐
    │  Équipe A  │     │  Équipe B   │     │  Équipe C   │
    │ Redondances│     │ Cohérence   │     │ Livrables   │
    │ & Épuration│     │ Code↔Docs   │     │ Mise à jour │
    └────────────┘     └─────────────┘     └─────────────┘
```

### 3.2 Équipe A — Élimination des redondances (3 agents parallèles)

**Objectif** : Résoudre les 5 redondances identifiées (R1–R5).

| Agent | Tâche | Entrées | Sortie attendue |
|-------|-------|---------|-----------------|
| A1 | Résoudre R1 + R2 : supprimer `AGENTS.md`, transformer `Docs/ARCHITECTURE.md` en fichier de redirection | `AGENTS.md`, `CLAUDE.md`, `Docs/ARCHITECTURE.md`, `architecture/README.md` | `AGENTS.md` supprimé, `Docs/ARCHITECTURE.md` réduit à un pointeur |
| A2 | Résoudre R3 : fusionner les catalogues de patterns | `Docs/DESIGN_PATTERNS.md`, `architecture/patterns/design-patterns.md` | Un seul fichier consolidé dans `architecture/patterns/`, `Docs/DESIGN_PATTERNS.md` supprimé ou redirigé |
| A3 | Résoudre R4 + R5 : consolider TUI et FFT | `Docs/TUI_GUIDE.md`, `architecture/flows/tui-flow.md`, `algorithms/FFT.md`, `algorithms/BIGFFT.md` | Frontières clarifiées, sections en double supprimées, cross-références ajoutées |

**Dépendances** : Aucune entre A1, A2, A3 → exécution parallèle.

### 3.3 Équipe B — Vérification de cohérence code ↔ documentation (4 agents parallèles)

**Objectif** : Valider que chaque document reflète l'état actuel du code source.

| Agent | Tâche | Fichiers à vérifier | Méthode |
|-------|-------|--------------------|---------|
| B1 | Valider les diagrammes C4 et le graphe de dépendances | `system-context.mermaid`, `container-diagram.mermaid`, `component-diagram.mermaid`, `dependency-graph.mermaid` | Comparer chaque nœud/arête avec les imports réels du code Go |
| B2 | Valider les interfaces et patterns documentés | `architecture/patterns/design-patterns.md`, `interface-hierarchy.mermaid` | Vérifier chaque interface, méthode, et relation d'implémentation dans le code |
| B3 | Valider les flux d'exécution | `flows/*.md`, `flows/*.mermaid` | Tracer les appels réels dans le code et comparer avec les diagrammes |
| B4 | Valider le README et les guides thématiques | `README.md`, `Docs/BUILD.md`, `Docs/CALIBRATION.md`, `Docs/PERFORMANCE.md`, `Docs/TESTING.md` | Vérifier les commandes, exemples, métriques, et options CLI |

**Dépendances** : Aucune entre B1–B4 → exécution parallèle.
**Pré-requis** : L'équipe A doit avoir terminé (pour éviter de valider des fichiers qui seront supprimés/fusionnés).

### 3.4 Équipe C — Mise à jour des livrables principaux (3 agents séquentiels puis parallèles)

**Objectif** : Mettre à jour les documents architecturaux de référence à partir des résultats des équipes A et B.

| Agent | Tâche | Dépendance | Sortie |
|-------|-------|------------|--------|
| C1 | Mettre à jour `architecture/README.md` — index principal | A1, A2, A3 | Index à jour avec navigation corrigée |
| C2 | Mettre à jour le rapport de validation | B1, B2, B3, B4 | `validation-report.md` régénéré avec résultats frais |
| C3 | Mettre à jour `CLAUDE.md` et `README.md` | A1, C1 | Sections architecture alignées sur `architecture/` |

**Dépendances** :
- C1 dépend de A (fichiers renommés/supprimés)
- C2 dépend de B (résultats de validation)
- C3 dépend de A1 et C1 (pour les références correctes)

---

## 4. Plan d'exécution détaillé

### Phase 1 — Épuration des redondances (Équipe A)

```
Durée estimée : 1 session d'agents parallèles
Parallélisme : 3 agents simultanés
```

#### Étape 1.1 — Agent A1 : Nettoyage AGENTS.md + ARCHITECTURE.md
1. Lire `AGENTS.md` et identifier tout contenu unique absent de `CLAUDE.md`
2. Si contenu unique trouvé → le migrer dans `CLAUDE.md`
3. Supprimer `AGENTS.md`
4. Lire `Docs/ARCHITECTURE.md` et `architecture/README.md`
5. Identifier les sections uniques de `Docs/ARCHITECTURE.md`
6. Migrer le contenu unique dans `architecture/README.md`
7. Remplacer `Docs/ARCHITECTURE.md` par un fichier de redirection :
   ```markdown
   # Architecture
   La documentation architecturale a été consolidée.
   Voir [Docs/architecture/README.md](architecture/README.md).
   ```

#### Étape 1.2 — Agent A2 : Fusion des catalogues de patterns
1. Lire les deux fichiers de patterns côte à côte
2. Créer une matrice de couverture (quel pattern est couvert où, avec quel niveau de détail)
3. Fusionner dans `architecture/patterns/design-patterns.md` :
   - Conserver les exemples de code détaillés de `Docs/DESIGN_PATTERNS.md`
   - Conserver les références fichier:ligne de `architecture/patterns/design-patterns.md`
   - Unifier le format de présentation
4. Supprimer `Docs/DESIGN_PATTERNS.md` ou le remplacer par une redirection

#### Étape 1.3 — Agent A3 : Consolidation TUI + FFT
1. **TUI** : Comparer `Docs/TUI_GUIDE.md` et `architecture/flows/tui-flow.md`
   - `TUI_GUIDE.md` → guide développeur (comment étendre la TUI)
   - `tui-flow.md` → flux d'exécution (comment la TUI fonctionne)
   - Supprimer les sections en double, ajouter des cross-références
2. **FFT** : Comparer `algorithms/FFT.md` et `algorithms/BIGFFT.md`
   - `FFT.md` → théorie mathématique pure (convolution, DFT)
   - `BIGFFT.md` → implémentation (pools, cache, Fermat)
   - Supprimer les sections théoriques dupliquées de `BIGFFT.md`
   - Ajouter un lien de `FFT.md` vers `BIGFFT.md` pour l'implémentation

### Phase 2 — Vérification de cohérence (Équipe B)

```
Durée estimée : 1 session d'agents parallèles
Parallélisme : 4 agents simultanés
Pré-requis : Phase 1 terminée
```

#### Étape 2.1 — Agent B1 : Audit des diagrammes C4
Pour chaque diagramme Mermaid :
1. Extraire les nœuds (packages, interfaces, classes)
2. Vérifier l'existence dans le code source via `Grep`/`Glob`
3. Extraire les arêtes (dépendances, imports)
4. Vérifier via `go list -m` ou analyse des imports
5. Signaler : nœuds manquants, nœuds obsolètes, arêtes incorrectes
6. **Livrable** : Liste de corrections à appliquer aux diagrammes

#### Étape 2.2 — Agent B2 : Audit des interfaces
Pour chaque interface documentée :
1. Vérifier l'existence et la signature dans le code
2. Vérifier les implémentations listées
3. Identifier les interfaces non documentées
4. Vérifier les relations (implements, uses, creates)
5. **Livrable** : Matrice interface × implémentation validée

#### Étape 2.3 — Agent B3 : Audit des flux d'exécution
Pour chaque flow documenté (CLI, TUI, Config, Algorithmes) :
1. Tracer le flux réel dans le code (appels de fonctions)
2. Comparer avec le diagramme Mermaid correspondant
3. Comparer avec la description textuelle
4. Identifier les étapes manquantes ou obsolètes
5. **Livrable** : Liste de corrections par flux

#### Étape 2.4 — Agent B4 : Audit des guides opérationnels
Pour chaque guide (BUILD, CALIBRATION, PERFORMANCE, TESTING, README) :
1. Tester les commandes documentées (syntaxe, options)
2. Vérifier les métriques de performance citées
3. Vérifier les options CLI et variables d'environnement
4. Vérifier les dépendances listées et leurs versions
5. **Livrable** : Liste de corrections par guide

### Phase 3 — Mise à jour des livrables (Équipe C)

```
Durée estimée : 1 session d'agents (partiellement parallèle)
Pré-requis : Phases 1 et 2 terminées
```

#### Étape 3.1 — Agent C1 : Mise à jour de l'index architectural
1. Recalculer les statistiques (nombre de fichiers, lignes, packages)
2. Mettre à jour la navigation et les liens (fichiers supprimés/renommés)
3. Mettre à jour le tableau des packages si nécessaire
4. Vérifier tous les liens internes

#### Étape 3.2 — Agent C2 : Régénération du rapport de validation
1. Reprendre les résultats des agents B1–B4
2. Re-exécuter les vérifications automatisables
3. Mettre à jour le rapport avec :
   - Nouvelle date
   - Résultats PASS/FAIL/WARNING actualisés
   - Résolution des 6 WARNING précédents si applicable

#### Étape 3.3 — Agent C3 : Alignement README + CLAUDE.md
1. Mettre à jour les sections architecture de `README.md` pour référencer `Docs/architecture/`
2. Mettre à jour `CLAUDE.md` si des interfaces/packages ont changé
3. Supprimer les références à `AGENTS.md`
4. Vérifier la cohérence des commandes de build/test

---

## 5. Critères de succès

| # | Critère | Mesure | Cible |
|---|---------|--------|-------|
| S1 | Zéro redondance critique | Nombre de doublons R1–R5 résolus | 5/5 |
| S2 | Cohérence code ↔ docs | Taux de conformité des vérifications B1–B4 | > 95% |
| S3 | Index à jour | Tous les liens internes fonctionnels | 100% |
| S4 | Rapport de validation vert | FAIL count | 0 |
| S5 | Réduction du volume | Lignes totales après épuration | Réduction ≥ 15% |
| S6 | Source de vérité unique | Chaque concept documenté en un seul endroit | Oui |

---

## 6. Arborescence cible après épuration

```
FibGo/
├── README.md                          (point d'entrée, liens vers Docs/)
├── CLAUDE.md                          (guide IA — source de vérité conventions)
├── CONTRIBUTING.md                    (guide de contribution — inchangé)
├── CHANGELOG.md                       (historique — inchangé)
├── [AGENTS.md supprimé]
│
├── Docs/
│   ├── ARCHITECTURE.md                (→ redirection vers architecture/README.md)
│   ├── BUILD.md                       (guide de build — vérifié et à jour)
│   ├── CALIBRATION.md                 (guide de calibration — vérifié et à jour)
│   ├── [DESIGN_PATTERNS.md supprimé → fusionné dans architecture/patterns/]
│   ├── PERFORMANCE.md                 (guide performance — vérifié et à jour)
│   ├── TESTING.md                     (guide de test — vérifié et à jour)
│   ├── TUI_GUIDE.md                   (guide développeur TUI — épuré, cross-réf)
│   │
│   ├── algorithms/                    (documentation algorithmique — épurée)
│   │   ├── BIGFFT.md                  (implémentation FFT — théorie déplacée)
│   │   ├── COMPARISON.md              (comparaison — inchangé)
│   │   ├── FAST_DOUBLING.md           (Fast Doubling — inchangé)
│   │   ├── FFT.md                     (théorie FFT — source unique)
│   │   ├── GMP.md                     (GMP — inchangé)
│   │   ├── MATRIX.md                  (Matrix Expo — inchangé)
│   │   └── PROGRESS_BAR_ALGORITHM.md  (progression — inchangé)
│   │
│   └── architecture/                  (★ SOURCE DE VÉRITÉ ARCHITECTURALE)
│       ├── README.md                  (index — mis à jour)
│       ├── system-context.mermaid     (C4 L1 — validé)
│       ├── container-diagram.mermaid  (C4 L2 — validé)
│       ├── component-diagram.mermaid  (C4 L3 — validé)
│       ├── dependency-graph.mermaid   (graphe dépendances — validé)
│       ├── flows/
│       │   ├── algorithm-flows.md     (flux algorithmes — validé)
│       │   ├── cli-flow.md            (flux CLI — validé)
│       │   ├── cli-flow.mermaid       (diagramme CLI — validé)
│       │   ├── config-flow.md         (flux config — validé)
│       │   ├── config-flow.mermaid    (diagramme config — validé)
│       │   ├── tui-flow.md            (flux TUI — validé, cross-réf TUI_GUIDE)
│       │   ├── tui-flow.mermaid       (diagramme TUI — validé)
│       │   ├── fastdoubling.mermaid   (diagramme Fast Doubling — validé)
│       │   ├── fft-pipeline.mermaid   (diagramme FFT — validé)
│       │   └── matrix.mermaid         (diagramme Matrix — validé)
│       ├── patterns/
│       │   ├── design-patterns.md     (★ catalogue consolidé — 13 patterns)
│       │   └── interface-hierarchy.mermaid  (hiérarchie — validée)
│       └── validation/
│           └── validation-report.md   (rapport régénéré)
```

---

## 7. Commandes d'exécution des agents

### Lancement Phase 1 (3 agents parallèles)

```
Agent A1 : "Résoudre redondances AGENTS.md et Docs/ARCHITECTURE.md"
Agent A2 : "Fusionner les catalogues de design patterns"
Agent A3 : "Consolider documentation TUI et FFT"
```

### Lancement Phase 2 (4 agents parallèles, après Phase 1)

```
Agent B1 : "Auditer diagrammes C4 et graphe de dépendances vs code source"
Agent B2 : "Auditer interfaces et patterns documentés vs code source"
Agent B3 : "Auditer flux d'exécution documentés vs appels réels dans le code"
Agent B4 : "Auditer guides opérationnels (BUILD, CALIBRATION, PERFORMANCE, TESTING, README)"
```

### Lancement Phase 3 (3 agents, partiellement parallèle, après Phase 2)

```
Agent C1 : "Mettre à jour architecture/README.md (index, stats, navigation)"
Agent C2 : "Régénérer validation-report.md à partir des audits B1–B4"
Agent C3 : "Aligner README.md et CLAUDE.md avec la structure épurée"
```

---

## 8. Risques et mitigations

| Risque | Probabilité | Impact | Mitigation |
|--------|-------------|--------|------------|
| Perte de contenu unique lors de la fusion | Moyenne | Élevé | Diff systématique avant suppression, revue par agent coordinateur |
| Liens cassés après réorganisation | Élevée | Moyen | Agent C1 vérifie tous les liens internes en Phase 3 |
| Diagrammes Mermaid incorrects non détectés | Faible | Moyen | Agent B1 effectue une vérification nœud par nœud |
| Régression de la couverture documentaire | Faible | Élevé | Critère S6 : chaque concept doit rester documenté exactement une fois |
| Conflits Git si travail en parallèle sur mêmes fichiers | Moyenne | Moyen | Agents A1–A3 travaillent sur des fichiers distincts |

---

## 9. Résumé exécutif

**32 fichiers** de documentation (~9 120 lignes) ont été inventoriés. L'analyse révèle **5 redondances critiques** et **4 zones de cohérence à vérifier**. Le plan propose une exécution en **3 phases** avec **10 agents** organisés en **3 équipes** :

- **Phase 1** (Équipe A, 3 agents parallèles) : Élimination des redondances → 2 fichiers supprimés, 3 fichiers réduits
- **Phase 2** (Équipe B, 4 agents parallèles) : Vérification de cohérence code ↔ documentation
- **Phase 3** (Équipe C, 3 agents) : Mise à jour des livrables de référence dans `Docs/architecture/`

**Résultat attendu** : Documentation épurée (~15% de réduction), sans redondance, alignée sur le code source, avec `Docs/architecture/` comme source de vérité unique validée.
