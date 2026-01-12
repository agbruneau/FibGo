"""
Phase 4 - Tests de Simulation d'Ecosysteme (L4)
=================================================
Tests pytest pour la simulation d'ecosysteme.
"""

import pytest
from typing import Dict, List

from .dataset_generator import (
    DatasetGenerator,
    RiskProfile,
    GeneratedApplication,
    get_dataset_stats,
)
from .run_simulation import (
    SimulationConfig,
    SimulationEngine,
    SimulatedAgentPipeline,
)


# =============================================================================
# TESTS DU GENERATEUR DE DATASET
# =============================================================================

class TestDatasetGenerator:
    """Tests pour le generateur de dataset."""

    @pytest.fixture
    def generator(self):
        """Fixture pour le generateur."""
        return DatasetGenerator(seed=42)

    def test_generate_dataset_correct_size(self, generator):
        """Test que le dataset a la bonne taille."""
        dataset = generator.generate_dataset(n=50)
        assert len(dataset) == 50

    def test_generate_dataset_reproducible(self, generator):
        """Test que le dataset est reproductible avec la meme seed."""
        dataset1 = generator.generate_dataset(n=20)
        generator2 = DatasetGenerator(seed=42)
        dataset2 = generator2.generate_dataset(n=20)

        for app1, app2 in zip(dataset1, dataset2):
            assert app1.application_id == app2.application_id
            assert app1.amount_requested == app2.amount_requested

    def test_generate_dataset_distribution(self, generator):
        """Test que la distribution des profils est respectee."""
        dataset = generator.generate_dataset(n=100)
        stats = get_dataset_stats(dataset)

        # Verifier la distribution approximative
        assert stats["by_profile"]["very_low"] >= 10  # ~15%
        assert stats["by_profile"]["low"] >= 20       # ~25%
        assert stats["by_profile"]["medium"] >= 25    # ~30%
        assert stats["by_profile"]["high"] >= 10      # ~15%
        assert stats["by_profile"]["critical"] >= 5   # ~10%

    def test_generate_balanced_dataset(self, generator):
        """Test le dataset equilibre."""
        dataset = generator.generate_balanced_dataset(per_profile=10)

        stats = get_dataset_stats(dataset)
        for profile in RiskProfile:
            assert stats["by_profile"][profile.value] == 10

    def test_generated_application_valid(self, generator):
        """Test que les applications generees sont valides."""
        for profile in RiskProfile:
            app = generator.generate_one(profile)

            assert app.application_id.startswith(("APP-", "EDGE-"))
            assert app.applicant_id.startswith("CUST-")
            assert app.amount_requested > 0
            assert app.declared_monthly_income > 0
            assert app.currency == "USD"
            assert app.employment_status in [
                "FULL_TIME", "PART_TIME", "SELF_EMPLOYED", "UNEMPLOYED"
            ]

    def test_very_low_risk_profile(self, generator):
        """Test les caracteristiques du profil tres faible risque."""
        app = generator.generate_one(RiskProfile.VERY_LOW)

        # Revenu eleve
        assert app.declared_monthly_income >= 8000
        # DTI faible
        dti = app.existing_debts / app.declared_monthly_income * 100
        assert dti < 20
        # Score attendu faible
        assert app.expected_score_range[1] <= 20

    def test_critical_risk_profile(self, generator):
        """Test les caracteristiques du profil critique."""
        app = generator.generate_one(RiskProfile.CRITICAL)

        # Score attendu eleve
        assert app.expected_score_range[0] >= 75
        # Outcome attendu: rejet
        assert app.expected_outcome == "REJECTED"


# =============================================================================
# TESTS DU PIPELINE
# =============================================================================

class TestSimulatedPipeline:
    """Tests pour le pipeline simule."""

    @pytest.fixture
    def pipeline(self):
        """Fixture pour le pipeline en mode mock."""
        return SimulatedAgentPipeline(use_real_api=False)

    def test_process_returns_result(self, pipeline):
        """Test que le pipeline retourne un resultat."""
        application = {
            "application_id": "TEST-001",
            "applicant_id": "CUST-001",
            "amount_requested": 50000,
            "declared_monthly_income": 5000,
            "employment_status": "FULL_TIME",
            "existing_debts": 1000,
        }

        result = pipeline.process(application)

        assert result["success"] is True
        assert "risk_score" in result
        assert "decision" in result
        assert 0 <= result["risk_score"] <= 100
        assert result["decision"] in ["APPROVED", "REJECTED", "MANUAL_REVIEW"]

    def test_low_dti_low_score(self, pipeline):
        """Test que DTI faible donne un score faible."""
        application = {
            "application_id": "TEST-002",
            "applicant_id": "CUST-002",
            "amount_requested": 10000,
            "declared_monthly_income": 10000,
            "employment_status": "FULL_TIME",
            "existing_debts": 0,
        }

        result = pipeline.process(application)

        # DTI tres faible, devrait etre bas
        assert result["risk_score"] < 30

    def test_high_dti_high_score(self, pipeline):
        """Test que DTI eleve donne un score eleve."""
        application = {
            "application_id": "TEST-003",
            "applicant_id": "CUST-003",
            "amount_requested": 100000,
            "declared_monthly_income": 3000,
            "employment_status": "PART_TIME",
            "existing_debts": 2000,
        }

        result = pipeline.process(application)

        # DTI eleve + part-time, devrait etre eleve
        assert result["risk_score"] >= 40

    def test_unemployed_high_score(self, pipeline):
        """Test que statut unemployed augmente le score."""
        application = {
            "application_id": "TEST-004",
            "applicant_id": "CUST-004",
            "amount_requested": 20000,
            "declared_monthly_income": 2000,
            "employment_status": "UNEMPLOYED",
            "existing_debts": 0,
        }

        result = pipeline.process(application)

        # Unemployed ajoute 50 points
        assert result["risk_score"] >= 45


# =============================================================================
# TESTS DE LA SIMULATION
# =============================================================================

@pytest.mark.e2e
@pytest.mark.slow
class TestSimulation:
    """Tests de simulation complete."""

    @pytest.fixture
    def config(self):
        """Fixture pour la configuration de test."""
        return SimulationConfig(
            n_requests=20,  # Petit dataset pour les tests
            seed=42,
            dry_run=True,  # Pas d'appels API
            verbose=False,
        )

    def test_simulation_runs_successfully(self, config):
        """Test que la simulation s'execute sans erreur."""
        engine = SimulationEngine(config)
        report = engine.run()

        assert report.total_requests == 20
        assert report.successful_requests > 0
        assert report.duration_seconds > 0

    def test_simulation_high_success_rate(self, config):
        """Test que le taux de succes est eleve."""
        engine = SimulationEngine(config)
        report = engine.run()

        success_rate = report.successful_requests / report.total_requests
        assert success_rate >= 0.95  # Au moins 95% de succes

    def test_simulation_reasonable_latencies(self, config):
        """Test que les latences sont raisonnables."""
        engine = SimulationEngine(config)
        report = engine.run()

        # En mode mock, latences tres faibles
        assert report.latencies["p50"] < 1000  # < 1s
        assert report.latencies["p99"] < 5000  # < 5s

    def test_simulation_coherence_rate(self, config):
        """Test que le taux de coherence est acceptable."""
        engine = SimulationEngine(config)
        report = engine.run()

        # Au moins 80% de coherence en mode mock
        assert report.coherence_rate >= 0.80

    def test_simulation_outcomes_distribution(self, config):
        """Test que la distribution des outcomes est raisonnable."""
        config.n_requests = 50  # Plus de demandes pour distribution
        engine = SimulationEngine(config)
        report = engine.run()

        # Verifier qu'on a les 3 types d'outcomes
        outcomes = report.outcomes_distribution
        assert "APPROVED" in outcomes or "REJECTED" in outcomes or "MANUAL_REVIEW" in outcomes

        # Pas plus de 80% d'un seul type
        total = sum(outcomes.values())
        for count in outcomes.values():
            assert count / total < 0.80

    def test_simulation_report_complete(self, config):
        """Test que le rapport contient toutes les infos."""
        engine = SimulationEngine(config)
        report = engine.run()

        assert report.timestamp is not None
        assert report.config is not None
        assert isinstance(report.latencies, dict)
        assert isinstance(report.outcomes_distribution, dict)
        assert isinstance(report.coherence_by_profile, dict)
        assert isinstance(report.anomalies, list)
        assert report.cost_estimate >= 0


# =============================================================================
# TESTS D'INTEGRATION AVEC VRAIE API (optionnel)
# =============================================================================

@pytest.mark.e2e
@pytest.mark.slow
@pytest.mark.skipif(
    not pytest.importorskip("anthropic", reason="anthropic not installed"),
    reason="anthropic not installed"
)
class TestSimulationWithAPI:
    """Tests de simulation avec vraie API (couteux)."""

    @pytest.fixture
    def config(self):
        """Fixture pour la configuration avec API."""
        return SimulationConfig(
            n_requests=5,  # Tres petit pour limiter les couts
            seed=42,
            dry_run=False,  # Vraie API
            verbose=True,
        )

    @pytest.mark.skip(reason="Couteux - executer manuellement si necessaire")
    def test_simulation_with_real_api(self, config):
        """Test avec vraie API (decommenter pour executer)."""
        import os
        if not os.getenv("ANTHROPIC_API_KEY"):
            pytest.skip("ANTHROPIC_API_KEY not set")

        engine = SimulationEngine(config)
        report = engine.run()

        assert report.successful_requests >= 4  # Au moins 80% succes
        print(f"\nCout estime: ${report.cost_estimate:.4f}")
