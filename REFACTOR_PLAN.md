# Plan de Refactorisation et Optimisation - FibCalc

## Résumé

Ce plan couvre la refactorisation et l'optimisation du dépôt FibCalc, un calculateur Fibonacci haute performance en Go. Les changements sont organisés en 6 phases par ordre de priorité et de risque.

---

## Phase 1: Couverture de Tests Critique (Risque: FAIBLE)

**Objectif**: Établir un filet de sécurité avant les modifications structurelles.

### 1.1 Tests pour `internal/logging/logger.go` (CRITIQUE - 20 fonctions non testées)
- Créer `internal/logging/logger_test.go`
- Tester: `NewDefaultLogger()`, `NewLogger()`, adaptateurs Zerolog/StdLog
- Tester: helpers Field (`String()`, `Int()`, `Uint64()`, `Float64()`, `Err()`)

### 1.2 Tests pour `internal/server/security.go` et `metrics.go`
- Créer `internal/server/security_test.go` - headers CORS, SecurityMiddleware
- Créer `internal/server/metrics_test.go` - Prometheus, IncrementActiveRequests

### 1.3 Tests pour `cmd/generate-golden/main.go` (CRITIQUE)
- Créer `cmd/generate-golden/main_test.go`
- Valider `fibBig()` pour n=0,1,2,92,93,94

### 1.4 Tests pour `internal/ui/colors.go` et `internal/calibration/runner.go`
- Créer fichiers de tests correspondants

---

## Phase 2: Élimination du Code Dupliqué (Risque: MOYEN)

### 2.1 Supprimer duplication des fonctions couleur
**Fichiers**:
- `internal/cli/ui.go` (lignes 59-84) - SUPPRIMER les 9 wrappers
- Garder `internal/ui/colors.go` comme source canonique
- Mettre à jour imports dans `cli/repl.go`, `cli/output.go`

### 2.2 Extraire la logique de progression dans `cli/repl.go`
**Fichier**: `internal/cli/repl.go`
- Créer fonction `runWithProgress(ctx, calc, n, opts)`
- Remplacer code dupliqué dans `calculate()` (203-262) et `cmdCompare()` (284-352)

### 2.3 Supprimer code mort dans `fibonacci/fastdoubling.go`
**Fichier**: `internal/fibonacci/fastdoubling.go`
- Supprimer `acquireState()`/`releaseState()` (lignes 264-280)
- Utiliser `AcquireState()`/`ReleaseState()` directement

---

## Phase 3: Réduction de la Complexité Cyclomatique (Risque: MOYEN) ✅ COMPLÉTÉE

### 3.1 Refactoriser `DisplayResult` dans `cli/ui.go` (CC≈6-7) ✅
**Fichier**: `internal/cli/ui.go`
- ✅ Extrait `displayResultHeader()` - affiche la taille binaire
- ✅ Extrait `displayDetailedAnalysis()` - affiche les métriques détaillées
- ✅ Extrait `displayCalculatedValue()` - affiche la valeur calculée
- ✅ Refactorisé `DisplayResult()` pour utiliser les nouvelles fonctions
- ✅ Corrigé `emptyStringTest` dans `formatNumberString()`

### 3.2 Considérer registre de commandes pour `processCommand` (CC≈10-12) - DIFFÉRÉ
**Fichier**: `internal/cli/repl.go` (lignes 144-184)
- Option: Convertir switch en `map[string]commandHandler`
- **Décision**: Structure actuelle conservée - le switch est clair et lisible
- Une refactorisation ajouterait de la complexité pour un gain minimal

---

## Phase 4: Optimisations de Performance (Risque: ÉLEVÉ)

### 4.1 Optimiser zeroing manuel avec `clear()` (Go 1.21+)
**Fichiers**:
- `internal/bigfft/pool.go` (lignes 62-66, 141-143, 209-211, 279-281)
- `internal/bigfft/bump.go` (lignes 113-115)
```go
// Avant: for i := range slice { slice[i] = 0 }
// Après: clear(slice)
```

### 4.2 Optimiser cache FFT avec pooling
**Fichier**: `internal/bigfft/fft_cache.go` (lignes 149-160, 191-196)
- Utiliser `acquireFermatSlice()` au lieu de `make()` pour copies

### 4.3 Limiter goroutines Strassen avec sémaphore
**Fichier**: `internal/fibonacci/common.go` (lignes 77-100)
- Ajouter `taskSemaphore` similaire à FFT recursion
- Limiter à `runtime.NumCPU()*2` goroutines concurrentes

### 4.4 Ajouter support context dans FFT (Optionnel)
**Fichier**: `internal/bigfft/fft_recursion.go` (lignes 78-114)
- Vérifier `ctx.Err()` avant acquisition sémaphore
- Risque élevé - changement d'API significatif

---

## Phase 5: Cohérence de Nommage (Risque: FAIBLE)

### 5.1 Documenter conventions de nommage
**Fichier**: `internal/cli/output.go`
- Ajouter documentation package expliquant:
  - `Display*` = écrit vers io.Writer
  - `Format*` = retourne string
  - `Write*` = écrit vers fichier

---

## Phase 6: Améliorations Architecture (Risque: ÉLEVÉ - Différé)

### 6.1 Réduire couplage CLI (À considérer plus tard)
- Introduction d'interfaces à la frontière CLI
- Priorité basse - couplage actuel acceptable

---

## Fichiers Critiques à Modifier

| Phase | Fichier | Changement | Statut |
|-------|---------|------------|--------|
| 1 | `internal/logging/logger_test.go` | CRÉER | ✅ |
| 1 | `internal/server/security_test.go` | CRÉER | ✅ |
| 1 | `internal/server/metrics_test.go` | CRÉER | ✅ |
| 2 | `internal/cli/ui.go` | Supprimer lignes 59-84 | ✅ |
| 2 | `internal/cli/repl.go` | Extraire runWithProgress | ✅ |
| 2 | `internal/fibonacci/fastdoubling.go` | Supprimer lignes 264-280 | ✅ |
| 3 | `internal/cli/ui.go` | Refactoriser DisplayResult | ✅ |
| 4 | `internal/bigfft/pool.go` | Utiliser clear() | ⏳ |
| 4 | `internal/fibonacci/common.go` | Ajouter taskSemaphore | ⏳ |

---

## Vérification

Après chaque phase:
```bash
make test              # Tests passent
make lint              # Pas d'erreurs linting
make coverage          # Couverture maintenue/améliorée
go test -race ./...    # Pas de race conditions
make benchmark         # Performance non dégradée (Phase 4)
```

### Tests End-to-End
```bash
# Calcul CLI basique
go run ./cmd/fibcalc -n 1000 -algo fast

# Mode serveur
go run ./cmd/fibcalc --server --port 8080 &
curl "http://localhost:8080/calculate?n=100"
curl "http://localhost:8080/health"

# REPL interactif
go run ./cmd/fibcalc --interactive
# > calc 100
# > exit
```

---

## Ordre d'Implémentation

```
Phase 1 (Tests) ✅
    → Phase 2 (Déduplication) ✅
        → Phase 3 (Complexité) ✅
            → Phase 4 (Performance)
                → Phase 5 (Nommage)
```

Chaque phase est indépendamment testable et déployable.
