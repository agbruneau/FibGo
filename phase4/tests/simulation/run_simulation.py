#!/usr/bin/env python3
"""
Phase 4 - Script de Simulation d'Ecosysteme (L4)
=================================================
Execute une simulation complete du maillage agentique avec 50+ demandes
et genere un rapport detaille.

Usage:
    python run_simulation.py --n 50 --report
    python run_simulation.py --n 100 --seed 42 --output results.json
"""

import argparse
import json
import os
import sys
import time
from dataclasses import dataclass, asdict
from datetime import datetime
from typing import Dict, Any, List, Optional
from statistics import mean, median, stdev
from collections import Counter

from dotenv import load_dotenv

# Ajouter le path pour les imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from tests.simulation.dataset_generator import (
    DatasetGenerator,
    GeneratedApplication,
    RiskProfile,
    get_dataset_stats,
)

load_dotenv()


# =============================================================================
# CONFIGURATION
# =============================================================================

@dataclass
class SimulationConfig:
    """Configuration de la simulation."""
    n_requests: int = 50
    seed: int = 42
    timeout_per_request: float = 30.0
    parallel_requests: int = 1  # Pour l'instant sequentiel
    dry_run: bool = False  # Si True, pas d'appels API
    verbose: bool = True


@dataclass
class RequestResult:
    """Resultat d'une requete individuelle."""
    application_id: str
    risk_profile: str
    expected_outcome: str
    expected_score_range: tuple
    actual_outcome: Optional[str] = None
    actual_score: Optional[int] = None
    latency_ms: float = 0.0
    success: bool = False
    error: Optional[str] = None
    coherent: bool = False  # Resultat coherent avec attentes


@dataclass
class SimulationReport:
    """Rapport de simulation."""
    timestamp: str
    config: Dict[str, Any]
    duration_seconds: float
    total_requests: int
    successful_requests: int
    failed_requests: int
    coherence_rate: float
    latencies: Dict[str, float]
    outcomes_distribution: Dict[str, int]
    coherence_by_profile: Dict[str, float]
    anomalies: List[str]
    cost_estimate: float


# =============================================================================
# AGENT MOCK (A remplacer par le vrai agent en production)
# =============================================================================

class SimulatedAgentPipeline:
    """
    Pipeline d'agents simule pour la simulation.

    En production, remplacer par les vrais agents connectes a Kafka.
    """

    def __init__(self, use_real_api: bool = True):
        self.use_real_api = use_real_api
        if use_real_api:
            try:
                from anthropic import Anthropic
                self.client = Anthropic(api_key=os.getenv("ANTHROPIC_API_KEY"))
                self.model = "claude-3-5-haiku-20241022"  # Economique pour simulation
            except ImportError:
                self.use_real_api = False
                print("Warning: anthropic not installed, using mock mode")

    def process(self, application: Dict[str, Any]) -> Dict[str, Any]:
        """
        Traite une demande a travers le pipeline.

        Args:
            application: Demande de pret

        Returns:
            Resultat avec decision et score
        """
        if not self.use_real_api:
            return self._mock_process(application)

        return self._real_process(application)

    def _mock_process(self, application: Dict[str, Any]) -> Dict[str, Any]:
        """Traitement simule (sans API)."""
        import random
        import time

        # Simule la latence
        time.sleep(random.uniform(0.1, 0.5))

        # Calcul basique du score
        income = application.get("declared_monthly_income", 5000)
        amount = application.get("amount_requested", 50000)
        debts = application.get("existing_debts", 0)
        employment = application.get("employment_status", "FULL_TIME")

        # DTI
        monthly_payment = amount / 60  # 5 ans
        dti = ((debts + monthly_payment) / income) * 100 if income > 0 else 100

        # Score basique
        score = 0
        if dti > 50:
            score += 40
        elif dti > 35:
            score += 25
        elif dti > 20:
            score += 10

        employment_scores = {
            "FULL_TIME": 0,
            "PART_TIME": 15,
            "SELF_EMPLOYED": 20,
            "UNEMPLOYED": 50,
        }
        score += employment_scores.get(employment, 25)

        # Ajout de variabilite
        score += random.randint(-5, 10)
        score = max(0, min(100, score))

        # Decision
        if score < 20:
            outcome = "APPROVED"
        elif score > 80:
            outcome = "REJECTED"
        else:
            outcome = "MANUAL_REVIEW"

        return {
            "success": True,
            "risk_score": score,
            "risk_category": self._score_to_category(score),
            "decision": outcome,
            "rationale": f"Score calcule: {score}. DTI: {dti:.1f}%",
        }

    def _real_process(self, application: Dict[str, Any]) -> Dict[str, Any]:
        """Traitement reel via API."""
        import json
        import re

        prompt = f"""Tu es un agent d'evaluation de risque bancaire.
Evalue cette demande de pret et produis un JSON avec:
- risk_score (0-100)
- risk_category (LOW/MEDIUM/HIGH/CRITICAL)
- decision (APPROVED/REJECTED/MANUAL_REVIEW)
- rationale (explication courte)

Demande:
{json.dumps(application, indent=2)}

Regles:
- Score < 20 = APPROVED
- Score > 80 = REJECTED
- Sinon = MANUAL_REVIEW
- DTI > 50% = score += 40
- UNEMPLOYED = score += 50

Reponds UNIQUEMENT en JSON."""

        try:
            message = self.client.messages.create(
                model=self.model,
                max_tokens=500,
                messages=[{"role": "user", "content": prompt}]
            )

            response_text = message.content[0].text

            # Parse JSON
            json_match = re.search(r'\{[^{}]*\}', response_text, re.DOTALL)
            if json_match:
                result = json.loads(json_match.group())
                result["success"] = True
                return result

            return {
                "success": False,
                "error": "Could not parse JSON response",
                "raw": response_text
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }

    def _score_to_category(self, score: int) -> str:
        """Convertit un score en categorie."""
        if score < 20:
            return "LOW"
        elif score < 50:
            return "MEDIUM"
        elif score < 80:
            return "HIGH"
        else:
            return "CRITICAL"


# =============================================================================
# SIMULATION ENGINE
# =============================================================================

class SimulationEngine:
    """Moteur de simulation."""

    def __init__(self, config: SimulationConfig):
        self.config = config
        self.pipeline = SimulatedAgentPipeline(use_real_api=not config.dry_run)
        self.results: List[RequestResult] = []

    def run(self) -> SimulationReport:
        """Execute la simulation complete."""
        start_time = time.time()

        # Generer le dataset
        generator = DatasetGenerator(seed=self.config.seed)
        dataset = generator.generate_dataset(n=self.config.n_requests)

        if self.config.verbose:
            print(f"\n{'='*60}")
            print("SIMULATION D'ECOSYSTEME - AgentMeshKafka (L4)")
            print(f"{'='*60}")
            print(f"Demandes a traiter: {len(dataset)}")
            print(f"Mode: {'Dry Run' if self.config.dry_run else 'Real API'}")
            print()

        # Traiter chaque demande
        for i, app in enumerate(dataset):
            if self.config.verbose:
                print(f"[{i+1}/{len(dataset)}] {app.application_id}...", end=" ")

            result = self._process_one(app)
            self.results.append(result)

            if self.config.verbose:
                status = "OK" if result.success else "FAIL"
                coherent = "V" if result.coherent else "X"
                print(f"{status} | Score: {result.actual_score or 'N/A'} | "
                      f"Decision: {result.actual_outcome or 'N/A'} | "
                      f"Coherent: {coherent}")

        duration = time.time() - start_time

        # Generer le rapport
        return self._generate_report(duration)

    def _process_one(self, app: GeneratedApplication) -> RequestResult:
        """Traite une demande individuelle."""
        result = RequestResult(
            application_id=app.application_id,
            risk_profile=app.risk_profile.value,
            expected_outcome=app.expected_outcome,
            expected_score_range=app.expected_score_range,
        )

        start = time.time()

        try:
            response = self.pipeline.process(app.to_dict())
            result.latency_ms = (time.time() - start) * 1000

            if response.get("success"):
                result.success = True
                result.actual_score = response.get("risk_score")
                result.actual_outcome = response.get("decision")

                # Verifier la coherence
                result.coherent = self._check_coherence(app, result)
            else:
                result.success = False
                result.error = response.get("error", "Unknown error")

        except Exception as e:
            result.latency_ms = (time.time() - start) * 1000
            result.success = False
            result.error = str(e)

        return result

    def _check_coherence(self, app: GeneratedApplication, result: RequestResult) -> bool:
        """Verifie si le resultat est coherent avec les attentes."""
        if not result.actual_outcome or result.actual_score is None:
            return False

        # Verifier le score dans la plage attendue (avec tolerance)
        min_score, max_score = app.expected_score_range
        tolerance = 15  # Points de tolerance
        score_ok = (min_score - tolerance) <= result.actual_score <= (max_score + tolerance)

        # Verifier l'outcome
        outcome_ok = False
        if app.expected_outcome == "APPROVED":
            outcome_ok = result.actual_outcome in ["APPROVED", "MANUAL_REVIEW"]
        elif app.expected_outcome == "REJECTED":
            outcome_ok = result.actual_outcome in ["REJECTED", "MANUAL_REVIEW"]
        else:  # MANUAL_REVIEW
            outcome_ok = True  # Toute reponse est acceptable

        return score_ok and outcome_ok

    def _generate_report(self, duration: float) -> SimulationReport:
        """Genere le rapport de simulation."""
        successful = [r for r in self.results if r.success]
        failed = [r for r in self.results if not r.success]
        coherent = [r for r in successful if r.coherent]

        # Latences
        latencies = [r.latency_ms for r in successful]
        latency_stats = {
            "p50": median(latencies) if latencies else 0,
            "p90": sorted(latencies)[int(len(latencies) * 0.9)] if latencies else 0,
            "p99": sorted(latencies)[int(len(latencies) * 0.99)] if latencies else 0,
            "avg": mean(latencies) if latencies else 0,
            "min": min(latencies) if latencies else 0,
            "max": max(latencies) if latencies else 0,
        }

        # Distribution des outcomes
        outcomes = Counter(r.actual_outcome for r in successful if r.actual_outcome)

        # Coherence par profil
        coherence_by_profile = {}
        for profile in RiskProfile:
            profile_results = [r for r in successful if r.risk_profile == profile.value]
            if profile_results:
                coherent_count = sum(1 for r in profile_results if r.coherent)
                coherence_by_profile[profile.value] = coherent_count / len(profile_results)

        # Anomalies
        anomalies = []
        for r in successful:
            if not r.coherent:
                anomalies.append(
                    f"{r.application_id}: Expected {r.expected_outcome}, "
                    f"got {r.actual_outcome} (score: {r.actual_score})"
                )

        # Estimation du cout (Haiku: ~$0.00025 per 1K input, $0.00125 per 1K output)
        # Estimation tres approximative
        cost_per_request = 0.002  # ~$0.002 par requete
        cost_estimate = len(successful) * cost_per_request

        return SimulationReport(
            timestamp=datetime.now().isoformat(),
            config=asdict(self.config),
            duration_seconds=duration,
            total_requests=len(self.results),
            successful_requests=len(successful),
            failed_requests=len(failed),
            coherence_rate=len(coherent) / len(successful) if successful else 0,
            latencies=latency_stats,
            outcomes_distribution=dict(outcomes),
            coherence_by_profile=coherence_by_profile,
            anomalies=anomalies[:10],  # Top 10
            cost_estimate=cost_estimate,
        )


# =============================================================================
# CLI & MAIN
# =============================================================================

def print_report(report: SimulationReport):
    """Affiche le rapport de maniere formatee."""
    print("\n" + "=" * 60)
    print("RAPPORT DE SIMULATION")
    print("=" * 60)

    print(f"\nDate: {report.timestamp}")
    print(f"Duree: {report.duration_seconds:.2f}s")

    print(f"\n{'─'*40}")
    print("RESULTATS GLOBAUX")
    print(f"{'─'*40}")
    print(f"Total: {report.total_requests}")
    print(f"Succes: {report.successful_requests} ({report.successful_requests/report.total_requests*100:.1f}%)")
    print(f"Echecs: {report.failed_requests}")
    print(f"Coherence: {report.coherence_rate*100:.1f}%")

    print(f"\n{'─'*40}")
    print("DISTRIBUTION DES DECISIONS")
    print(f"{'─'*40}")
    for outcome, count in report.outcomes_distribution.items():
        pct = count / report.successful_requests * 100 if report.successful_requests > 0 else 0
        print(f"  {outcome}: {count} ({pct:.1f}%)")

    print(f"\n{'─'*40}")
    print("LATENCES (ms)")
    print(f"{'─'*40}")
    print(f"  P50: {report.latencies['p50']:.1f}")
    print(f"  P90: {report.latencies['p90']:.1f}")
    print(f"  P99: {report.latencies['p99']:.1f}")
    print(f"  Avg: {report.latencies['avg']:.1f}")

    print(f"\n{'─'*40}")
    print("COHERENCE PAR PROFIL")
    print(f"{'─'*40}")
    for profile, rate in report.coherence_by_profile.items():
        status = "OK" if rate >= 0.85 else "WARN" if rate >= 0.70 else "FAIL"
        print(f"  {profile}: {rate*100:.1f}% [{status}]")

    if report.anomalies:
        print(f"\n{'─'*40}")
        print(f"ANOMALIES ({len(report.anomalies)} sur top 10)")
        print(f"{'─'*40}")
        for anomaly in report.anomalies[:5]:
            print(f"  - {anomaly}")
        if len(report.anomalies) > 5:
            print(f"  ... et {len(report.anomalies) - 5} autres")

    print(f"\n{'─'*40}")
    print("COUTS")
    print(f"{'─'*40}")
    print(f"  Estimation: ${report.cost_estimate:.2f}")
    print(f"  Par requete: ${report.cost_estimate/report.successful_requests:.4f}" if report.successful_requests > 0 else "  N/A")

    print("\n" + "=" * 60)


def main():
    """Point d'entree CLI."""
    parser = argparse.ArgumentParser(
        description="Simulation d'ecosysteme AgentMeshKafka (L4)"
    )
    parser.add_argument(
        "-n", "--requests",
        type=int,
        default=50,
        help="Nombre de demandes a traiter (default: 50)"
    )
    parser.add_argument(
        "--seed",
        type=int,
        default=42,
        help="Graine aleatoire pour reproductibilite (default: 42)"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Mode simulation sans appels API"
    )
    parser.add_argument(
        "-q", "--quiet",
        action="store_true",
        help="Mode silencieux (pas de progression)"
    )
    parser.add_argument(
        "-o", "--output",
        type=str,
        help="Fichier de sortie JSON pour le rapport"
    )

    args = parser.parse_args()

    config = SimulationConfig(
        n_requests=args.requests,
        seed=args.seed,
        dry_run=args.dry_run,
        verbose=not args.quiet,
    )

    engine = SimulationEngine(config)
    report = engine.run()

    # Afficher le rapport
    print_report(report)

    # Sauvegarder si demande
    if args.output:
        with open(args.output, "w") as f:
            json.dump(asdict(report), f, indent=2)
        print(f"\nRapport sauvegarde: {args.output}")

    # Exit code base sur la coherence
    if report.coherence_rate < 0.85:
        print("\n[WARN] Coherence < 85%, verification recommandee")
        sys.exit(1)

    sys.exit(0)


if __name__ == "__main__":
    main()
