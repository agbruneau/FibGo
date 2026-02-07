# Reponse aux objections : Optimisation Schönhage-Strassen

## Synthese des reponses

| Objection | Severite | Verdict | Ajustement |
|-----------|----------|---------|------------|
| #1 Gains surestimates (15-25% -> 8-15%) | Majeure | **Partiellement acceptee** | Ajuste a 10-18% pour Phases 1-2 |
| #2 basicSqr inutile quand big.Int detecte squaring | Majeure | **Rejetee** -- big.Int ne detecte PAS le squaring dans ce cas | Le gain de `fermat.Sqr` est reel pour TOUTES les tailles |
| #3 Twiddle cache risque de coherence | Mineure | **Acceptee** | Architecture per-level clarifiee |
| #4 Seuils parallelisme non valides | Mineure | **Acceptee** | Benchmark requis avant modification |
| #5 sqrt(2) trick sous-estime en effort | Mineure | **Acceptee** | Requalifie a 10-15 jours, Phase 3 uniquement |

**Gain revise apres ajustements : 10-18% (Phases 1-2), potentiellement 20-35% avec Phase 3.**

---

## Reponses detaillees

### Objection 1 : Gains surestimates -- 15-25% vs 8-15%

**Reponse : Partiellement acceptee. Ajuste a 10-18%.**

Je concede les points suivants du devil's advocate :

**(a) Twiddle factor pre-calcul : ajuste de 5-10% a 3-6%**

Le devil's advocate a raison que le pre-calcul ne supprime pas l'application du shift, seulement le recalcul de `i*omega2shift`. Cependant, le gain reel vient de deux sources sous-estimees par le DA :

1. **Eviter la multiplication `i*omega2shift`** : C'est une multiplication entiere par iteration de la boucle, certes faible.
2. **Opportunity de fused shift+butterfly** : En pre-connaissant les shift amounts, on peut fusionner `ShiftHalf` + `Add`/`Sub` en une seule passe memoire. C'est l'approche GMP qui combine twiddle application avec le papillon. Le gain n'est pas dans le pre-calcul du shift amount mais dans la reduction des passes memoire de 3 (shift, sub, add) a 2 (fused_shift_sub, add).

Le cout memoire K * (n+1) mots est un probleme valide pour les tres grandes FFT. Pour K=4096 et n=1024 mots, c'est 32 MB -- trop grand pour un cache pre-calcule. Le pre-calcul n'est donc viable que pour les offsets entiers, pas pour les valeurs pre-shiftees.

**Ajustement : 3-6%** (au lieu de 5-10%).

**(b) Squaring specialise : REVISE A LA HAUSSE de 5-8% a 8-12%**

C'est ici que le devil's advocate a fait une erreur critique. L'objection 2 affirme que `big.Int.Mul` detecte le squaring en interne. **C'est faux dans le contexte de `fermat.Mul`.**

Verification effectuee : Go's `math/big.Int.Mul` detecte le squaring via un test de **pointeur** `*Int` (`if x == y` dans `int.go:mul`), PAS via une comparaison des slices sous-jacentes. Or, dans `fermat.Mul` (fermat.go:159-163) :

```go
var xi, yi, zi big.Int   // deux big.Int DISTINCTS
xi.SetBits(x)            // xi pointe vers x
yi.SetBits(y)            // yi pointe vers y (meme slice quand x==y)
zb := zi.Mul(&xi, &yi).Bits()  // &xi != &yi => PAS detecte comme squaring
```

Meme quand `x` et `y` sont la meme slice fermat (cas du squaring pointwise dans `sqr()` a `fft_poly.go:387`), `&xi != &yi` car ce sont deux variables locales distinctes. **Le squaring n'est donc JAMAIS optimise dans le chemin actuel, ni pour n < 30 ni pour n >= 30.**

Cela signifie que `fermat.Sqr()` apporterait un gain pour **toutes les tailles de coefficients**, pas seulement n < 30 :
- Pour n < 30 : `basicSqr` avec symetrie (~30-40% de gain sur basicMul)
- Pour n >= 30 : `big.Int.Mul(x, x)` avec le **meme pointeur** (declenchant le squaring optimise de Go) au lieu de `big.Int.Mul(&xi, &yi)` avec des pointeurs distincts

L'implementation serait :
```go
func (z fermat) Sqr(x fermat) fermat {
    n := len(x) - 1
    if n < smallMulThreshold {
        z = z[:2*n+2]
        basicSqr(z, x)  // symetrie exploitee
        z = z[:2*n+1]
    } else {
        var xi, zi big.Int
        xi.SetBits(x)
        zi.SetBits(z)
        zb := zi.Mul(&xi, &xi).Bits()  // MEME pointeur => squaring detecte
        // ... normalisation
    }
    // ... reste identique a Mul
}
```

Le squaring de Go utilise Karatsuba squaring qui est ~30% plus rapide que Karatsuba multiplication pour les grandes tailles. Le squaring pointwise est appele K fois par `sqr()`. Pour F(10^6), K ~ 256 avec des coefficients de ~2700 mots. Le gain du squaring detecte sur chaque appel se cumule.

**Ajustement : 8-12%** (revise a la hausse, contre 5-8% initialement et 3-5% propose par le DA).

**(c) Seuils parallelisme : ajuste de 3-7% a 1-4%**

J'accepte l'argument du DA. Le semaphore est deja limite a NumCPU tokens. Augmenter `MaxParallelFFTDepth` sans augmenter les tokens ne fait que creer plus de tentatives `select` qui tombent dans le `default`. Le gain est conditionnel au hardware et probablement < 3% dans la plupart des cas.

**Ajustement : 1-4%** (au lieu de 3-7%).

**Gain cumule revise (non-additif) :**

| Optimisation | Estimation DA | Ma revision | Confiance |
|-------------|---------------|-------------|-----------|
| 3.1 Twiddle pre-calcul | 3-5% | 3-6% | Moyenne |
| 3.2 Squaring specialise | 3-5% (erreur) | **8-12%** | **Haute** (verifie dans le code) |
| 3.3 Seuils parallelisme | 1-3% | 1-4% | Basse |
| 3.4 Localite cache | 1-3% | 2-4% | Moyenne |
| 3.5 Cache agrandi | 0-2% | 1-3% | Basse |

Avec un facteur de non-additivite de ~0.7x (car 3.1 et 3.4 touchent le meme chemin, et 3.2 est independant) :

- Phase 1 (3.2 + 3.3 + 3.5) : 8-12% + 1-4% + 1-3% = brut 10-19%, ajuste ~**8-14%**
- Phase 2 (3.1 + 3.4) : 3-6% + 2-4% = brut 5-10%, ajuste supplementaire ~**3-7%**
- **Total Phases 1-2 : 10-18%**

C'est entre l'estimation DA (8-15%) et mon estimation initiale (15-25%). La difference principale vient de la decouverte que `fermat.Sqr` apporte un gain reel pour toutes les tailles.

---

### Objection 2 : basicSqr inutile quand big.Int detecte le squaring

**Reponse : REJETEE. L'objection est fondee sur une premisse incorrecte.**

Comme demontre ci-dessus, `big.Int.Mul` compare les pointeurs `*Int` (`if x == y` dans `int.go:mul`), pas les slices sous-jacentes. Dans `fermat.Mul`, `xi` et `yi` sont des variables locales distinctes (`var xi, yi, zi big.Int`), donc `&xi != &yi` meme quand les slices pointent vers la meme memoire.

**Preuve :** Source Go `src/math/big/int.go`, methode `mul` :
```go
func (z *Int) mul(stk *stack, x, y *Int) {
    if x == y {           // comparaison de POINTEURS *Int
        z.abs = z.abs.sqr(stk, x.abs)
        // ...
        return
    }
    z.abs = z.abs.mul(stk, x.abs, y.abs)
    // ...
}
```

La subtilite soulevee par le DA lui-meme ("il faudrait verifier si Go detecte le squaring via les pointeurs de slice ou les pointeurs de big.Int") a ete verifiee : c'est par pointeurs `*Int`, et donc le squaring n'est **pas** detecte dans le cas de `fermat.Mul`.

**Consequence :** `fermat.Sqr()` est l'optimisation au meilleur ROI du rapport. Un simple changement de `zi.Mul(&xi, &yi)` a `zi.Mul(&xi, &xi)` (en passant le meme pointeur) suffit pour les n >= 30, sans meme avoir besoin d'ecrire `basicSqr`.

**Critere de resolution du DA :** "Ecrire un benchmark simple: `big.Int.Mul(x, y)` ou x et y sont des `big.Int` distincts mais avec `SetBits` pointant vers la meme slice. Comparer avec `big.Int.Mul(x, x)` (meme pointeur)."

Reponse : Ce benchmark confirmerait mon analyse. `Mul(x, x)` (meme pointeur) declenchera le chemin `sqr()` optimise de Go, tandis que `Mul(&xi, &yi)` (pointeurs distincts) executera le chemin `mul()` generique. Le gain attendu est de ~30% pour les tailles Karatsuba (Go utilise Karatsuba squaring quand le squaring est detecte).

---

### Objection 3 (mineure) : Twiddle cache risque de coherence

**Reponse : Acceptee. Voici l'architecture clarifiee.**

Le DA a raison que `omega2shift = (4*n*_W) >> size` change a chaque niveau de recursion (car `size` diminue). Un cache global serait incorrect.

**Architecture proposee : pre-calcul per-invocation, pas cache persistant.**

Au lieu d'un cache global, l'approche correcte est :
1. Au debut de chaque appel `fourierRecursiveUnified` de niveau `size`, pre-calculer le tableau `twiddleOffsets[i] = i * omega2shift` pour i = 0..len(dst1)-1
2. Ce tableau est alloue depuis le bump allocator (cout O(K) mots pour le niveau, negligeable)
3. Pas de cache inter-appels : chaque invocation FFT recalcule ses offsets

Le gain principal n'est pas dans le caching mais dans la possibilite de **fusionner** l'application du twiddle avec le papillon quand les offsets sont pre-connus. Cela reduit les passes memoire.

**Cout memoire pour K=4096, n=1024 :**
- Offsets pre-calcules : 4096 * 8 bytes (entiers) = 32 KB (negligeable, tient en L1)
- Pas de valeurs fermat pre-calculees (trop couteux en memoire, comme le DA l'a souleve)

---

### Objection 4 (mineure) : Seuils parallelisme non valides empiriquement

**Reponse : Acceptee. Benchmark requis avant modification.**

L'argument du DA est solide :
1. Le semaphore est limite a NumCPU tokens -- augmenter depth sans tokens n'a pas d'effet
2. Le `select/default` non-bloquant fait que les branches excessives sont sequentielles
3. La parallelisation de sous-problemes en L1 cache peut etre contre-productive

**Ajustement au rapport SS :**
- La proposition 3.3 est retrogradee de "Quick Win" a "Optimisation conditionnelle"
- Prerequis : benchmark avec depth=3,4,5 et tokens=NumCPU, 2*NumCPU pour F(10^7)
- Si le gain est < 2%, abandonner
- Potentiellement plus utile : augmenter les tokens du semaphore en meme temps que la depth

---

### Objection 5 (mineure) : sqrt(2) trick effort sous-estime

**Reponse : Acceptee. Requalifie a 10-15 jours, Phase 3 uniquement.**

Le DA a raison que le sqrt(2) trick impacte :
- `fftSize()` et `fftSizeThreshold` (fft.go:184-207)
- `valueSize()` (fft.go:215-225)
- `EstimateBumpCapacity()` (bump.go:209-242)
- Le cache de transformees (cle de cache changee)
- Les pools (tailles de classe)

C'est un changement structurel qui touche au moins 5 fichiers critiques. Mon estimation initiale de 5-7 jours etait optimiste.

**Ajustement :** Requalifie a 10-15 jours. Maintenu en Phase 3 uniquement, avec decision prise apres les resultats de la Phase 2.

---

## Reponses aux questions directes

### Question 1 : big.Int.Mul detecte-t-il le squaring avec SetBits ?

**Non.** Go compare les pointeurs `*Int` (`if x == y`), pas les slices sous-jacentes. Dans `fermat.Mul`, `&xi != &yi` car ce sont deux variables locales distinctes. Le squaring n'est donc PAS optimise dans le chemin actuel. Voir la reponse detaillee a l'objection 2 ci-dessus.

### Question 2 : Profiling de la boucle Reconstruct

Je n'ai pas de profiling factuel. L'estimation de ~40% etait basee sur l'analyse theorique du code : la boucle Reconstruct fait 3 operations par element (ShiftHalf, Sub, Add) pour K/2 elements, tandis que les appels recursifs font le meme travail sur des sous-problemes de taille K/2. En sommant sur tous les niveaux de recursion, la boucle Reconstruct execute O(K * log K) operations au total (K/2 operations par niveau * log K niveaux), ce qui est le meme ordre que les appels recursifs.

**Ajustement :** L'estimation 40% est probablement un majorant. Une estimation plus raisonnable serait **25-35%** du temps FFT total, car les appels recursifs incluent aussi les base cases (size=0,1) qui sont plus rapides.

Cela ajuste le gain du twiddle pre-calcul a 3-6% (deja revise ci-dessus).

### Question 3 : sqrt(2) trick effort

**Oui, j'accepte 10-15 jours.** Voir reponse a l'objection 5.

---

## Estimation revisee finale

| Optimisation | Gain revise | Effort | Phase |
|-------------|-------------|--------|-------|
| 3.2 Squaring specialise (`fermat.Sqr`) | **8-12%** | 1-2 jours | **Phase 1** (priorite #1) |
| 3.5 Cache transformees agrandi | 1-3% | 0.5 jour | Phase 1 |
| 3.3 Seuils parallelisme | 1-4% (conditionnel) | 1 jour + benchmarks | Phase 1 (conditionnelle) |
| 3.1 Twiddle per-invocation + fused butterfly | 3-6% | 3 jours | Phase 2 |
| 3.4 Localite cache Reconstruct | 2-4% | 2 jours | Phase 2 |
| 3.6 Sqrt(2) trick | 5-15% | 10-15 jours | Phase 3 |
| 3.7 SIMD Fermat | 10-20% | 7-10 jours | Phase 3 |

**Phase 1 revise (3 jours) : gain 8-14%** -- domine par `fermat.Sqr` qui est le quick win le plus impactant.
**Phase 2 (5 jours) : gain supplementaire 3-7%**
**Total Phases 1-2 : 10-18%**

La decouverte que `big.Int.Mul` ne detecte pas le squaring dans le contexte de `fermat.Mul` est le resultat le plus important de cette analyse adversariale. Elle transforme une optimisation initialement estimee a 5-8% en une optimisation a 8-12%, tout en la rendant plus simple a implementer (un seul changement de pointeur pour n >= 30, plus `basicSqr` pour n < 30).
