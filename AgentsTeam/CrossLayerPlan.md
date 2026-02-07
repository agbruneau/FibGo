# Plan : Rédaction de `AgentsTeam/CrossLayerPlan.md`

## Tableau de suivi d'implémentation

> **Date d'exécution** : 2026-02-07 | **Plateforme** : Intel Core Ultra 9 275HX (24 cores), Windows, Go 1.25+

### Phases

| Phase | Description | Statut | Résultat |
|-------|-------------|--------|----------|
| Phase 1 | Profiling baseline | ✅ Terminé | Benchmarks F(1M) et F(10M) collectés |
| Phase 2a | Optimisations BigFFT (T3-T7) | ✅ Terminé | Allocs FFT réduites de 93-99% |
| Phase 2b | Optimisations Fibonacci (T8-T12) | ✅ Terminé | Wall time -19 à -33% sur F(10M) |
| Phase 2c | Optimisations Orchestration (T13-T15) | ✅ Terminé | Fast path + lock-free observers |
| Phase 3 | Vérification | ✅ Terminé | 18/18 packages OK, fuzz 14.5M exec OK |

### Tâches détaillées

| Tâche | Couche | Description | Statut | Impact mesuré |
|-------|--------|-------------|--------|---------------|
| T3 | bigfft | Pool `transform()` output buffers | ✅ | -95.8% allocs FFTMul 10K |
| T4 | bigfft | Pool `mul()` output buffers | ✅ | -98.7% allocs FFTMul 100K |
| T5 | bigfft | Pool `invTransform()` + `sqr()` output | ✅ | -99.4% allocs FFTSqr 1M |
| T6 | bigfft | Cache deep copy → buffer contigu | ✅ | -88.2% allocs CacheHit |
| T7 | bigfft | Marge bump allocator 20% → 10% | ✅ | ~10% mémoire épargnée |
| T8 | fibonacci | `MaxPooledBitLen` 4M → 100M bits | ✅ | Pool réutilisation F(10^8) |
| T9 | fibonacci | `ParallelFFTThreshold` 10M → 5M bits | ✅ | Parallélisme plus tôt |
| T10 | fibonacci | Short-circuit progress quiet mode | ✅ | No-op sans observers |
| T11 | fibonacci | DynamicThresholdManager sans mutex | ✅ | Hot path lock-free |
| T12 | fibonacci | Pré-dimensionner T1-T4 selon N | ✅ | Réduit réallocations |
| T13 | orchestration | `ProgressBufferMultiplier` 5 → 50 | ✅ | Moins de contention |
| T14 | orchestration | Fast path mono-calculateur | ✅ | Élimine errgroup overhead |
| T15 | orchestration | `Freeze()` lock-free observers | ✅ | Snapshot sans lock |

### Résultats benchmarks comparatifs

**Fibonacci (ns/op) :**

| Benchmark | Baseline | Après | Gain |
|-----------|----------|-------|------|
| FastDoubling/1M | 3,926,435 | 2,919,423 | **-25.7%** |
| MatrixExp/1M | 4,882,297 | 4,808,411 | -1.5% |
| FFTBased/1M | 6,940,974 | 6,362,186 | **-8.3%** |
| FastDoubling/10M | 66,673,188 | 53,799,700 | **-19.3%** |
| MatrixExp/10M | 55,205,833 | 37,221,345 | **-32.6%** |
| FFTBased/10M | 64,237,084 | 57,427,850 | **-10.6%** |

**BigFFT allocs/op :**

| Benchmark | Baseline | Après | Gain |
|-----------|----------|-------|------|
| FFTMul 10K words | 2,127 | 89 | **-95.8%** |
| FFTMul 100K words | 8,276 | 106 | **-98.7%** |
| FFTMul 1M words | 32,858 | 104 | **-99.7%** |
| FFTSqr 10K words | 1,084 | 66 | **-93.9%** |
| FFTSqr 100K words | 4,163 | 76 | **-98.2%** |
| FFTSqr 1M words | 16,468 | 100 | **-99.4%** |

### Critères d'acceptation

| Critère | Cible | Résultat | Statut |
|---------|-------|----------|--------|
| Wall time FastDoubling/10M | ≥15% | -19.3% | ✅ |
| Wall time MatrixExp/10M | ≥15% | -32.6% | ✅ |
| Allocs FFT operations | ≥30% | -93 à -99% | ✅ |
| Tests pass | 100% | 18/18 packages | ✅ |
| Fuzz tests 30s | 0 failures | 14.5M execs, 0 fail | ✅ |

---

## Contexte

Le scénario 4 de `AgentsTeam/Scenarios.md` décrit une optimisation cross-layer coordonnée de F(10^8) via une équipe de 4 agents Claude Code. L'exploration du codebase a identifié des hotspots concrets dans les 3 couches (`bigfft`, `fibonacci`, `orchestration`). Le livrable est un document Markdown détaillé servant de guide d'implémentation pour lancer ce scénario Agent Teams.

## Action

Créer le fichier `AgentsTeam/CrossLayerPlan.md` avec le contenu structuré suivant :

---

### Structure du document CrossLayerPlan.md

**1. Composition de l'équipe** - 4 agents avec rôles, fichiers assignés, permissions

| Agent             | Rôle                                              | Fichiers principaux                                            |
| ----------------- | -------------------------------------------------- | -------------------------------------------------------------- |
| `profiling`     | Benchmarks, CPU/mem profiles, broadcast résultats | Tous (lecture seule)                                           |
| `bigfft`        | Allocations FFT, pooling, bump allocator           | `fft_poly.go`, `pool.go`, `bump.go`, `fft_cache.go`    |
| `fibonacci`     | Seuils, état poolé, fast path                    | `constants.go`, `common.go`, `calculator.go`, `fft.go` |
| `orchestration` | Contention channels, observer, errgroup            | `orchestrator.go`, `observer.go`, `ui_display.go`        |

**2. Phase 1 - Profiling** (agent `profiling`, bloque tous les autres)

- Benchmarks baseline : `go test -bench -benchmem` sur F(10M)
- CPU profile : `go test -cpuprofile` → `go tool pprof -top`
- Mem profile : `go test -memprofile` → alloc_objects + alloc_space
- Métriques collectées : wall time, alloc/op, bytes/op, top 5 CPU, top 5 alloc, cache hit rate
- Format du broadcast aux autres agents

**3. Phase 2 - Optimisations** (3 agents en parallèle, après broadcast)

Optimisations bigfft (5 tâches) :

- T3: Pooler les buffers de sortie de `transform()` (`fft_poly.go:168-169`) — impact: 10-15% alloc
- T4: Pooler les buffers de `mul()` (`fft_poly.go:338-340`) — impact: 3-5% alloc
- T5: Pooler les buffers de `invTransform()` et `sqr()` (`fft_poly.go:376-378`) — impact: 5-8% alloc
- T6: Réduire le deep copy du cache (`fft_cache.go`) — buffer contigu au lieu de K allocations
- T7: Ajuster la marge du bump allocator (`bump.go:238`) selon profiling réel

Optimisations fibonacci (5 tâches) :

- T8: Augmenter `MaxPooledBitLen` de 4M à 100M bits (`common.go:37`) — impact: 2-3% wall time
- T9: Tuner `ParallelFFTThreshold` basé sur les données profiling (`constants.go:50`)
- T10: Court-circuiter le progress reporting en mode quiet (`calculator.go`, `progress.go`)
- T11: Éliminer le mutex du `DynamicThresholdManager` (single-goroutine)
- T12: Pré-dimensionner les `big.Int` du `CalculationState` selon N (`calculator.go`) — impact: 3-5%

Optimisations orchestration (3 tâches) :

- T13: Augmenter `ProgressBufferMultiplier` de 5 à 50 (`orchestrator.go`)
- T14: Fast path mono-calculateur sans errgroup (`orchestrator.go`)
- T15: Snapshot lock-free des observers avec `Freeze()` (`observer.go`)

**4. Phase 3 - Vérification** (agent `profiling`)

- `go test -v -race -cover ./...`
- Golden file validation, fuzz tests 30s
- Benchmarks comparatifs avec `benchstat baseline.txt after.txt`
- Test end-to-end F(10^8) : `time ./fibcalc -n 100000000 -algo fast`
- Critères d'acceptation : ≥15% wall time, ≥30% alloc, 100% tests pass

**5. Graphe de dépendances des tâches**

```
T1→T2→{T3,T4,T5,T6,T7,T8,T9,T10,T11,T12,T13,T14,T15}→T16→T17
```

**6. Tableau de priorité** - 12 optimisations classées par impact/complexité

## Fichiers critiques

- `internal/bigfft/fft_poly.go` — hotspot principal (lignes 168, 338, 376 : `make()` non-poolés)
- `internal/fibonacci/constants.go` — seuils de performance (lignes 18, 28, 50)
- `internal/fibonacci/common.go` — `MaxPooledBitLen` (ligne 37), sémaphore (ligne 29)
- `internal/fibonacci/calculator.go` — pré-sizing et wiring progress
- `internal/orchestration/orchestrator.go` — buffer channel et errgroup

## Vérification

Après création du fichier :

1. Relire `CrossLayerPlan.md` pour cohérence avec `Scenarios.md` scénario 4
2. Vérifier que chaque optimisation référence des lignes de code réelles (validé par Read)
3. Confirmer que le graphe de dépendances est acyclique
