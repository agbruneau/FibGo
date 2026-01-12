# Constitution du Maillage Agentique

> **Version :** 2.0.0 | **Statut :** Approuve | **Derniere revision :** Janvier 2026
>
> **Documents connexes :** [05-ThreatModel.md](./05-ThreatModel.md) | [06-Plan.md](./06-Plan.md)

Ce document definit la **Constitution** du projet AgentMeshKafka : les principes fondamentaux, les regles inviolables et les standards d'ingenierie qui gouvernent le developpement et l'operation du Maillage Agentique.

---

## Preambule

Le Maillage Agentique est un systeme ou des agents autonomes prennent des decisions qui impactent des personnes reelles. Cette Constitution etablit les principes qui garantissent que ces decisions sont prises de maniere **juste**, **transparente** et **securisee**.

---

## ARTICLE I : Les Trois Lois de la Robotique Bancaire

### Premiere Loi : Integrite du Contrat

> **Un agent ne doit jamais emettre un evenement qui viole le schema defini.**

**Explication :**
- Tout message publie doit etre valide selon le schema Avro enregistre
- Si un agent ne peut pas produire un output valide, il doit echouer proprement
- Aucun "best effort" : soit le contrat est respecte, soit l'erreur est signalée

**Implementation :**
```python
def publish_safely(event: dict, schema: AvroSchema) -> bool:
    if not schema.validate(event):
        log.error("Event violates schema", event=event)
        send_to_dead_letter_queue(event)
        return False
    return producer.send(event)
```

**Consequence :** Un agent qui hallucine un JSON malformé ne polluera jamais le systeme.

---

### Deuxieme Loi : Transparence Cognitive

> **Un agent doit toujours expliciter son raisonnement avant de produire une action.**

**Explication :**
- Chaque decision doit etre accompagnee d'une justification (Chain of Thought)
- Le raisonnement doit etre trace et auditable
- Aucune "boite noire" : les decisions doivent pouvoir etre expliquées

**Implementation :**
```python
class RiskAssessment(BaseModel):
    risk_score: int
    risk_category: RiskCategory
    rationale: str  # OBLIGATOIRE - Jamais vide
    checked_policies: List[str]  # OBLIGATOIRE - Politiques consultées
```

**Consequence :** On peut toujours savoir POURQUOI une decision a été prise.

---

### Troisieme Loi : Securite et Confidentialite

> **Un agent doit proteger ses instructions contre les injections et sanitiser les données personnelles.**

**Explication :**
- Les prompts systeme ne doivent jamais etre reveles
- Les donnees personnelles (PII) doivent etre protegees
- Les tentatives d'injection doivent etre detectees et bloquees

**Implementation :**
```python
def sanitize_for_logging(data: dict) -> dict:
    """Masque les PII avant logging."""
    sensitive_fields = ["applicant_id", "email", "phone", "ssn"]
    return {
        k: mask_pii(v) if k in sensitive_fields else v
        for k, v in data.items()
    }
```

**Consequence :** La confidentialite est preservee, meme en cas d'attaque.

---

## ARTICLE II : Principes Architecturaux

### II.1 Schema First

> **Les schemas sont definis AVANT le code.**

| Etape | Action |
|-------|--------|
| 1 | Definir le schema Avro (.avsc) |
| 2 | Generer les modeles Pydantic |
| 3 | Implementer la logique |
| 4 | Ecrire les tests |

**Justification :** Le schema est le contrat entre les agents. Le modifier en cours de route brise la compatibilite.

### II.2 Event-Driven First

> **Les agents communiquent par evenements, jamais par appels directs.**

```
INTERDIT:
Agent A ────HTTP────> Agent B

AUTORISE:
Agent A ────> [Kafka Topic] ────> Agent B
```

**Justification :** Le decouplage permet la resilience et la scalabilite.

### II.3 Fail Fast, Fail Loud

> **En cas d'erreur, echouer immediatement et signaler clairement.**

```python
# BON
if not validate(data):
    raise ValidationError("Invalid data", details=data)

# MAUVAIS
if not validate(data):
    return None  # Echec silencieux
```

**Justification :** Les erreurs silencieuses sont des bombes a retardement.

---

## ARTICLE III : Standards de Developpement

### III.1 Stack Technologique Officielle

| Composant | Technologie | Version Minimum |
|-----------|-------------|-----------------|
| **Langage** | Python | 3.10+ |
| **LLM** | Anthropic Claude | Haiku/Sonnet/Opus 4.5 |
| **Framework IA** | LangChain / LangGraph | 0.3.0+ |
| **Messaging** | Apache Kafka | 3.x |
| **Serialisation** | Apache Avro | 1.11+ |
| **Base Vectorielle** | ChromaDB | 0.4+ |
| **Validation** | Pydantic | 2.5+ |
| **Tests** | pytest | 8.0+ |

### III.2 Matrice des Modeles LLM

| Agent | Modele Recommande | Alternative | Temperature |
|-------|-------------------|-------------|-------------|
| **Intake** | Claude 3.5 Haiku | - | 0.0 |
| **Risk** | Claude Sonnet 4 | Claude Opus 4.5 | 0.2 |
| **Decision** | Claude 3.5 Sonnet | - | 0.1 |
| **LLM-Juge** | Claude Opus 4.5 | - | 0.0 |

**Justification :** Chaque agent a le "cerveau" adapte a sa complexite cognitive.

### III.3 Configuration Externalisee

> **Aucune valeur magique dans le code.**

```yaml
# config.yaml - TOUTE la configuration ici
thresholds:
  auto_approve_score: 20
  auto_reject_score: 80
  high_value_amount: 100000

kafka:
  bootstrap_servers: "${KAFKA_BOOTSTRAP_SERVERS:localhost:9092}"
```

**Override via environnement :**
```bash
export THRESHOLDS__AUTO_APPROVE_SCORE=15
```

---

## ARTICLE IV : Standards de Qualite

### IV.1 Couverture de Tests

| Type | Cible | Minimum |
|------|-------|---------|
| **L1 - Unitaires** | 80% | 70% |
| **L2 - Cognitifs** | Tous les agents | Score >= 7/10 |
| **L3 - Adversite** | 95% blocage | 90% |

### IV.2 Documentation

Tout composant doit avoir :
- Docstring Python
- README dans son dossier (si applicable)
- Entree dans la documentation technique

### IV.3 Code Review

Criteres de validation :
- [ ] Tests passent
- [ ] Couverture maintenue
- [ ] Schema First respecte
- [ ] Pas de secrets dans le code
- [ ] Logging structure present

---

## ARTICLE V : Operations (AgentOps)

### V.1 Logging Structure

```python
import structlog

log = structlog.get_logger()

log.info(
    "loan_decision_made",
    application_id="APP-12345",
    decision="APPROVED",
    risk_score=25,
    trace_id="abc-123"
)
```

**Format de sortie :**
```json
{
  "event": "loan_decision_made",
  "application_id": "APP-12345",
  "decision": "APPROVED",
  "risk_score": 25,
  "trace_id": "abc-123",
  "timestamp": "2026-01-12T10:30:00Z"
}
```

### V.2 Metriques Prometheus

```python
from prometheus_client import Counter, Histogram

decisions_total = Counter(
    'loan_decisions_total',
    'Total loan decisions',
    ['status']  # APPROVED, REJECTED, MANUAL_REVIEW
)

processing_time = Histogram(
    'agent_processing_seconds',
    'Time spent processing',
    ['agent']  # intake, risk, decision
)
```

### V.3 Alertes

| Condition | Severite | Action |
|-----------|----------|--------|
| Latence P99 > 30s | Warning | Notification |
| Erreur rate > 5% | Critical | PagerDuty |
| Kafka lag > 100 | Warning | Scale up |
| Injection detectee | Critical | Block + Investigate |

---

## ARTICLE VI : Securite (AgentSec)

### VI.1 Principes Zero Trust

1. **Jamais de confiance implicite** entre agents
2. **Validation a chaque etape** du pipeline
3. **Identite unique** pour chaque agent (Consumer Group)
4. **Encryption en transit** (TLS pour Kafka en production)

### VI.2 Protection des Prompts

```python
# Les prompts systeme sont dans des fichiers separes
# Jamais dans le code source ou les logs

def get_system_prompt(agent_name: str) -> str:
    path = f"prompts/{agent_name}_system.txt"
    with open(path, 'r') as f:
        return f.read()
```

### VI.3 Gestion des Secrets

| Secret | Stockage | Acces |
|--------|----------|-------|
| API Keys | Variables d'environnement | Runtime only |
| Credentials Kafka | .env (local) / Vault (prod) | Agents only |
| Prompts systeme | Fichiers proteges | Code only |

---

## ARTICLE VII : Outils de Developpement

### VII.1 Claude Code

Le developpement est assiste par **Claude Code** pour :
- Generation de code boilerplate
- Refactoring
- Documentation
- Debug

**Regles d'utilisation :**
- Toujours reviewer le code genere
- Ne jamais commiter sans comprendre
- Utiliser pour accelerer, pas remplacer la reflexion

### VII.2 Auto Claude

Pour l'auto-correction et l'amelioration continue :
- Review automatique des PRs
- Detection d'anomalies
- Suggestions d'optimisation

---

## ARTICLE VIII : Evolution de la Constitution

### VIII.1 Processus d'Amendement

1. Proposition via Issue GitHub
2. Discussion avec les stakeholders
3. Vote majoritaire (si applicable)
4. Mise a jour du document
5. Communication du changement

### VIII.2 Versionning

| Changement | Version |
|------------|---------|
| Correction mineure | 2.0.x |
| Nouvelle regle | 2.x.0 |
| Refonte majeure | x.0.0 |

---

## ARTICLE IX : Glossaire

| Terme | Definition |
|-------|------------|
| **Agent** | Entite autonome utilisant un LLM pour prendre des decisions |
| **Maillage Agentique** | Architecture decentralisee d'agents communiquant via evenements |
| **AgentOps** | Pratiques d'industrialisation et d'operation des agents |
| **AgentSec** | Pratiques de securite specifiques aux agents LLM |
| **Chain of Thought** | Trace du raisonnement d'un agent |
| **Dead Letter Queue** | File d'attente pour messages en echec |
| **DTI** | Debt-to-Income ratio (ratio dette/revenu) |
| **Event Sourcing** | Pattern ou l'etat est reconstruit depuis les evenements |
| **LLM-Juge** | LLM utilise pour evaluer les reponses d'autres LLMs |
| **PII** | Personally Identifiable Information (donnees personnelles) |
| **ReAct** | Pattern Reason + Act pour agents |
| **Schema Registry** | Service de gouvernance des schemas Avro |

---

## Signature

Cette Constitution est la reference ultime pour le developpement et l'operation du Maillage Agentique AgentMeshKafka.

**Auteur :** Andre-Guy Bruneau
**Date :** Janvier 2026
**Licence :** MIT

---

## Navigation

| Precedent | Index | Suivant |
|:----------|:-----:|--------:|
| [06-Plan.md](./06-Plan.md) | [Documentation](./00-Readme.md) | - |
