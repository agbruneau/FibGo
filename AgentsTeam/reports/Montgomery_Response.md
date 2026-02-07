# Reponse aux objections Montgomery

**Auteur** : montgomery-researcher
**Date** : 2026-02-07
**En reponse a** : Counter_Report.md (devils-advocate)

---

## Reponse a l'Objection 1 (Majeure) : L'analyse du gap Karatsuba-FFT est incomplete

### L'objection

Le devils-advocate souligne que le rapport identifie un "gap" entre Karatsuba (~10K bits) et FFT (~500K bits) sans le quantifier, et met en evidence une incoherence entre `fftThreshold` dans bigfft (1800 mots = ~115K bits) et `DefaultFFTThreshold` dans fibonacci (500K bits). Il demande si abaisser le seuil serait un "quick win".

### Ma reponse

**Je suis entierement d'accord.** L'ecart entre les deux seuils est significatif et merite investigation. Voici mon analyse detaillee :

1. **Les deux seuils operent a des niveaux differents** :
   - `fftThreshold` (bigfft, 1800 mots = ~115K bits) : Seuil interne a `bigfft.Mul()` qui decide si FFT est utilisee pour une multiplication isolee. Ce seuil est calibre par `TestCalibrate` au niveau du package bigfft, ou la FFT est comparee directement a `math/big.Mul`.
   - `DefaultFFTThreshold` (fibonacci, 500K bits) : Seuil au niveau de `smartMultiply()` qui decide si la strategie adaptive appelle `bigfft.MulTo()` vs `math/big.Mul`. Ce seuil est plus conservateur car il prend en compte l'overhead d'appel a bigfft, la gestion d'erreur defer/recover, et le pattern d'utilisation en chaine du Fast Doubling.

2. **L'ecart de ~385K bits est probablement excessif** :
   - `bigfft.Mul()` ajoute un overhead de ~200ns (defer/recover, conversion nat, verification seuil) qui est negligeable pour des operandes > 115K bits.
   - `smartMultiply()` appelle `bigfft.MulTo()` qui est encore plus efficace car il reutilise le buffer de destination.
   - Le facteur decisif est que `math/big.Mul` utilise deja Karatsuba (O(n^1.585)) puis Toom-Cook 3-way (~O(n^1.465)) en interne pour les tailles intermediaires. La FFT SS a O(n log n log log n) ne devient avantageuse qu'au-dessus d'un certain seuil qui depend des constantes d'implementation.

3. **Recommandation concrete** : Le seuil optimal est probablement entre 200K et 400K bits. Le systeme de calibration existant (`calibration/runner.go:78`, `findBestFFTThreshold`) est concu exactement pour determiner cette valeur empiriquement. La fonction `GenerateQuickFFTThresholds()` dans `adaptive.go:73` genere deja une grille de test.

   **Action recommandee** : Lancer une calibration (`fibcalc --calibrate`) sur la machine cible et verifier le seuil optimal. C'est le "quick win" a meilleur ROI (0.5 jour, potentiellement 2-5% de gain) identifie par le devils-advocate, et j'approuve pleinement cette priorite.

4. **Pourquoi mon rapport ne l'a pas approfondi** : Mon mandat etait d'evaluer Montgomery comme strategie de multiplication, pas d'optimiser les seuils existants. L'identification du gap servait a montrer que l'espace entre Karatsuba et FFT n'est PAS le terrain de Montgomery (qui est modulaire) mais celui du tuning de seuils et potentiellement de NTT/Toom-Cook. Le devils-advocate a raison de souligner que cette recommandation aurait du etre plus prominente.

**Verdict** : Objection **acceptee**. Le rapport aurait du quantifier le gap et prioriser l'optimisation du seuil FFT comme recommandation #1.

---

## Reponse a l'Objection 2 (Mineure) : Utilite de Montgomery pour un mode Fibonacci modulaire futur

### L'objection

Le rapport ecarte trop rapidement le scenario `F(n) mod M` et devrait recommander de preserver la possibilite architecturale d'ajouter Montgomery.

### Ma reponse

**D'accord sur le fond.** Je precise et complete :

1. **L'architecture actuelle est deja prete** : L'interface `Multiplier` dans `strategy.go:29-59` et le pattern Strategy/Factory (`registry.go`) permettent deja l'ajout d'une `MontgomeryStrategy` sans aucune modification architecturale. C'est un point fort existant du design.

2. **Montgomery serait optimal pour F(n) mod M** : Pour le calcul de Fibonacci modulaire avec un modulus fixe M (ex: 10^9+7), la chaine de ~log2(n) multiplications modulaires dans le Fast Doubling permettrait d'amortir les conversions Montgomery form sur toute la boucle. Le gain serait significatif : Montgomery REDC evite la division par M a chaque step, remplacee par un shift (O(n) vs O(n^2) pour la division naive, ou vs le cout de Barrett reduction).

3. **Cas d'usage concrets** :
   - Competitive programming : F(n) mod 10^9+7 (tres courant)
   - Cryptographie : F(n) mod p pour des primes specifiques
   - Pisano period : Calcul de pi(m) qui necessite F(n) mod m iterativement

4. **Note ajoutee a la recommandation** : Le Strategy pattern existant (`Multiplier`/`DoublingStepExecutor` + `CalculatorFactory`) rend l'ajout futur de Montgomery trivial du point de vue architectural. Si un mode `F(n) mod M` est ajoute, Montgomery devrait etre la strategie par defaut pour ce mode. Aucune action immediate n'est requise car l'architecture est deja extensible.

**Verdict** : Objection **acceptee**. Le rapport devrait explicitement mentionner cette ouverture architecturale dans sa conclusion.

---

## Reponse a l'Objection 3 (Mineure) : Les estimations d'overhead Montgomery sont imprecises

### L'objection

Le rapport utilise un ratio uniforme de ~25% de ralentissement pour toutes les tailles, alors que l'overhead REDC varie asymptotiquement.

### Ma reponse

**D'accord.** Voici l'analyse affinee :

1. **Analyse asymptotique du ratio overhead** :
   - Le cout REDC pour un nombre de n mots est O(n) (une multiplication par un seul mot + addition).
   - Le cout de la multiplication sous-jacente M(n) varie :
     - Petits n (< ~500 mots) : M(n) = O(n^1.585) (Karatsuba)
     - Moyens n (~500-8000 mots) : M(n) = O(n^1.465) (Toom-Cook 3)
     - Grands n (> ~8000 mots) : M(n) = O(n log n log log n) (FFT SS)
   - Ratio overhead = REDC / M(n) :
     - Petits n : O(n) / O(n^1.585) = O(n^{-0.585}) -- overhead proportionnellement **eleve** (~40-60%)
     - Moyens n : O(n) / O(n^1.465) = O(n^{-0.465}) -- overhead significatif (~20-35%)
     - Grands n : O(n) / O(n log n log log n) = O(1 / (log n log log n)) -- overhead **decroissant** (~5-15%)

2. **Estimations revisees par taille** :

   | N | Bits F(n) | Overhead REDC estime | Ratio ajuste |
   |---|-----------|---------------------|--------------|
   | 10^3 | ~700 | ~50% (setup domine) | **0.67x** |
   | 10^4 | ~7,000 | ~35% (Karatsuba + REDC eleve) | **0.74x** |
   | 10^5 | ~70,000 | ~25% (Toom-Cook + REDC moyen) | **0.80x** |
   | 10^6 | ~700,000 | ~15% (FFT + REDC faible) | **0.87x** |
   | 10^7 | ~7,000,000 | ~10% (FFT + REDC negligeable) | **0.91x** |
   | 10^8 | ~70,000,000 | ~7% (FFT domine) | **0.93x** |
   | 10^9 | ~700,000,000 | ~5% (FFT domine completement) | **0.95x** |

3. **Conclusion inchangee** : Meme avec des estimations affinee, Montgomery est **toujours plus lent** que les methodes actuelles pour toute taille. L'overhead decroit avec la taille mais ne disparait jamais (il y a toujours le cout de conversion vers/depuis Montgomery form a chaque iteration, car pas de modulus fixe). Le fait que l'overhead diminue pour les grands N ne change pas le verdict : Montgomery est inapplicable sans modulus.

**Verdict** : Objection **acceptee**. Les estimations du rapport original etaient trop uniformes. L'analyse asymptotique ci-dessus est plus rigoureuse, mais la conclusion reste inchangee : REJETER.

---

## Reponses aux questions directes du devils-advocate

### Question 1 : L'ecart fftThreshold vs DefaultFFTThreshold merite-t-il investigation ?

**Oui, absolument.** C'est le meilleur quick win identifie dans toute cette revue. Mes observations :

- `bigfft.fftThreshold` = 1800 mots = ~115,200 bits : calibre par le package bigfft contre `math/big.Mul`
- `fibonacci.DefaultFFTThreshold` = 500,000 bits : seuil conservateur dans `smartMultiply()`
- L'ecart de ~385K bits signifie que pour des operandes entre 115K et 500K bits, `smartMultiply` utilise `math/big.Mul` alors que `bigfft.Mul` utiliserait deja la FFT (qui est potentiellement plus rapide a cette taille)

L'explication probable : `DefaultFFTThreshold` a ete fixe de maniere conservative en anticipant l'overhead de l'appel `bigfft.MulTo()` dans le contexte specifique du Fast Doubling (ou la reutilisation de transformees via `executeDoublingStepFFT` ne s'active que pour les tres grandes tailles). Mais cet overhead est negligeable pour des operandes > 100K bits.

**Action recommandee** : Benchmark systematique avec `fibcalc --calibrate` ou benchmark manual de `smartMultiply` pour des operandes de 150K, 200K, 300K, 400K, 500K bits. Si bigfft est plus rapide a 200K bits, abaisser `DefaultFFTThreshold` a 200K -- gain gratuit.

### Question 2 : Montgomery pour l'arithmetique Fermat interne ?

**Question pertinente et interessante.** Analysons :

Le devils-advocate demande si Montgomery pourrait accelerer `fermat.Mul` (`fermat.go:149-205`) qui effectue des multiplications mod 2^(n*W)+1.

1. **Le modulus 2^(n*W)+1 EST fixe pendant toute la FFT** : Pour une FFT de taille donnee, le modulus Fermat ne change pas entre les multiplications pointwise. C'est exactement le scenario ou Montgomery excelle (chaine de multiplications avec modulus fixe).

2. **MAIS la reduction Fermat est deja quasi-gratuite** : La reduction mod 2^(n*W)+1 dans `fermat.Mul` (lignes 174-204) se fait par simple soustraction des mots hauts (`subVV`) et ajout des carries -- c'est O(n) avec des constantes tres faibles (pas de multiplication, juste des additions/soustractions vectorielles). Montgomery REDC fait aussi O(n) mais avec une multiplication par N' (l'inverse du modulus) puis une addition de m*N, ce qui est significativement plus couteux.

3. **Comparaison concrete** :
   - **Fermat reduction actuelle** : ~3-5 operations vectorielles (addVW, subVV, addVW) sur n mots. Cout : ~3-5n cycles.
   - **Montgomery REDC** : Multiplication `(T mod R) * N'` (n mots * 1 mot = n cycles) + multiplication `m * N` (n mots * n mots ou n mots * 1 mot selon la variante) + addition + comparaison conditionnelle. Cout minimal : ~5-8n cycles pour CIOS word-by-word, plus si N est multi-mots.
   - **Verdict** : Fermat reduction est **2-3x plus rapide** que Montgomery REDC pour le modulus 2^n+1, car ce modulus a une forme speciale qui rend la reduction triviale (juste des shifts et additions).

4. **Raison fondamentale** : Les nombres de Fermat 2^n+1 ont une structure binaire extreme (1 suivi de n zeros puis 1). Multiplier par 2^n+1 ou reduire modulo 2^n+1 se fait par shift + addition/soustraction. Montgomery ne peut pas exploiter cette structure car REDC travaille avec N' (l'inverse multiplicatif) qui est un nombre sans structure speciale. C'est exactement pour cela que Schonhage et Strassen ont choisi les nombres de Fermat : leur reduction est la plus rapide possible.

**Verdict** : Montgomery pour Fermat interne serait une **regression**. L'arithmetique Fermat actuelle est deja optimale pour le modulus 2^n+1. Montgomery est concu pour des modulus generiques ou la division est couteuse -- ce n'est pas le cas pour les nombres de Fermat.

---

## Synthese des ajustements au rapport Montgomery

| Objection | Verdict | Impact sur la recommandation |
|-----------|---------|------------------------------|
| #1 (gap Karatsuba-FFT) | Acceptee | Ajouter recommandation #1 : benchmarker et potentiellement abaisser DefaultFFTThreshold. Quick win a meilleur ROI. |
| #2 (mode F(n) mod M futur) | Acceptee | Ajouter note : l'architecture Strategy pattern permet l'ajout futur de Montgomery sans modification. Recommander Montgomery si mode modulaire ajoute. |
| #3 (overhead imprecis) | Acceptee | Affiner les estimations : overhead variable de ~50% (petits N) a ~5% (grands N). Conclusion REJETER inchangee. |
| Q1 (seuil FFT) | D'accord | Recommandation transversale : investiguer l'ecart 115K-500K bits comme quick win prioritaire. |
| Q2 (Montgomery Fermat interne) | Rejete | Fermat reduction est 2-3x plus rapide que Montgomery REDC pour le modulus 2^n+1 grace a la structure speciale des nombres de Fermat. |

**Recommandation finale inchangee : REJETER Montgomery pour la multiplication d'entiers libres dans FibGo.**

Les ajustements ameliorent la precision du rapport mais ne modifient pas le verdict fondamental : Montgomery requiert un modulus fixe et n'apporte aucun avantage en multiplication non-modulaire. Les "quick wins" identifies (optimisation du seuil FFT, ameliorations SS) sont les priorites reelles.
