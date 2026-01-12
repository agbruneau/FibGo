"""
Phase 4 - Generateur de Dataset pour Simulation L4
====================================================
Genere des demandes de pret realistes et variees pour les tests
de simulation d'ecosysteme.
"""

import random
import uuid
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from typing import List, Dict, Any, Optional


class RiskProfile(Enum):
    """Profils de risque pour la generation."""
    VERY_LOW = "very_low"       # Score attendu: 0-15
    LOW = "low"                 # Score attendu: 15-30
    MEDIUM = "medium"           # Score attendu: 30-55
    HIGH = "high"               # Score attendu: 55-80
    CRITICAL = "critical"       # Score attendu: 80-100
    EDGE_CASE = "edge_case"     # Cas limites a tester


class EmploymentStatus(Enum):
    """Statuts d'emploi."""
    FULL_TIME = "FULL_TIME"
    PART_TIME = "PART_TIME"
    SELF_EMPLOYED = "SELF_EMPLOYED"
    UNEMPLOYED = "UNEMPLOYED"


@dataclass
class GeneratedApplication:
    """Demande de pret generee."""
    application_id: str
    applicant_id: str
    amount_requested: float
    currency: str
    declared_monthly_income: float
    employment_status: str
    existing_debts: float
    loan_purpose: str
    credit_score: int  # Score de credit simule (300-850)

    # Metadonnees pour validation
    risk_profile: RiskProfile
    expected_outcome: str  # APPROVED, REJECTED, MANUAL_REVIEW
    expected_score_range: tuple  # (min, max)

    def to_dict(self) -> Dict[str, Any]:
        """Convertit en dictionnaire pour les agents."""
        return {
            "application_id": self.application_id,
            "applicant_id": self.applicant_id,
            "amount_requested": self.amount_requested,
            "currency": self.currency,
            "declared_monthly_income": self.declared_monthly_income,
            "employment_status": self.employment_status,
            "existing_debts": self.existing_debts,
            "loan_purpose": self.loan_purpose,
        }


class DatasetGenerator:
    """Generateur de datasets pour simulation."""

    # Objectifs de pret realistes
    LOAN_PURPOSES = [
        "Achat immobilier - residence principale",
        "Achat immobilier - investissement locatif",
        "Refinancement hypothecaire",
        "Achat vehicule neuf",
        "Achat vehicule occasion",
        "Travaux de renovation",
        "Consolidation de dettes",
        "Etudes superieures",
        "Demarrage d'entreprise",
        "Equipement professionnel",
        "Vacances",
        "Mariage",
        "Frais medicaux",
        "Autres besoins personnels",
    ]

    # Distribution par defaut des profils de risque
    DEFAULT_DISTRIBUTION = {
        RiskProfile.VERY_LOW: 0.15,   # 15%
        RiskProfile.LOW: 0.25,         # 25%
        RiskProfile.MEDIUM: 0.30,      # 30%
        RiskProfile.HIGH: 0.15,        # 15%
        RiskProfile.CRITICAL: 0.10,    # 10%
        RiskProfile.EDGE_CASE: 0.05,   # 5%
    }

    def __init__(self, seed: Optional[int] = None):
        """
        Initialise le generateur.

        Args:
            seed: Graine pour la reproductibilite
        """
        if seed is not None:
            random.seed(seed)
        self._counter = 0

    def _generate_id(self, prefix: str) -> str:
        """Genere un ID unique."""
        self._counter += 1
        return f"{prefix}-{uuid.uuid4().hex[:8].upper()}"

    def _generate_very_low_risk(self) -> GeneratedApplication:
        """Genere une demande a tres faible risque."""
        income = random.uniform(8000, 15000)
        amount = random.uniform(5000, income * 3)  # Max 3x revenu
        debts = random.uniform(0, income * 0.15)   # Max 15% DTI existant

        return GeneratedApplication(
            application_id=self._generate_id("APP"),
            applicant_id=self._generate_id("CUST"),
            amount_requested=round(amount, 2),
            currency="USD",
            declared_monthly_income=round(income, 2),
            employment_status=EmploymentStatus.FULL_TIME.value,
            existing_debts=round(debts, 2),
            loan_purpose=random.choice(self.LOAN_PURPOSES[:5]),  # Objectifs stables
            credit_score=random.randint(750, 850),
            risk_profile=RiskProfile.VERY_LOW,
            expected_outcome="APPROVED",
            expected_score_range=(0, 20),
        )

    def _generate_low_risk(self) -> GeneratedApplication:
        """Genere une demande a faible risque."""
        income = random.uniform(5000, 10000)
        amount = random.uniform(10000, income * 5)
        debts = random.uniform(income * 0.1, income * 0.25)

        return GeneratedApplication(
            application_id=self._generate_id("APP"),
            applicant_id=self._generate_id("CUST"),
            amount_requested=round(amount, 2),
            currency="USD",
            declared_monthly_income=round(income, 2),
            employment_status=random.choice([
                EmploymentStatus.FULL_TIME.value,
                EmploymentStatus.FULL_TIME.value,  # Plus probable
                EmploymentStatus.PART_TIME.value,
            ]),
            existing_debts=round(debts, 2),
            loan_purpose=random.choice(self.LOAN_PURPOSES[:8]),
            credit_score=random.randint(680, 750),
            risk_profile=RiskProfile.LOW,
            expected_outcome="APPROVED",
            expected_score_range=(15, 35),
        )

    def _generate_medium_risk(self) -> GeneratedApplication:
        """Genere une demande a risque moyen."""
        income = random.uniform(3500, 7000)
        amount = random.uniform(20000, income * 7)
        debts = random.uniform(income * 0.2, income * 0.4)

        return GeneratedApplication(
            application_id=self._generate_id("APP"),
            applicant_id=self._generate_id("CUST"),
            amount_requested=round(amount, 2),
            currency="USD",
            declared_monthly_income=round(income, 2),
            employment_status=random.choice([
                EmploymentStatus.FULL_TIME.value,
                EmploymentStatus.PART_TIME.value,
                EmploymentStatus.SELF_EMPLOYED.value,
            ]),
            existing_debts=round(debts, 2),
            loan_purpose=random.choice(self.LOAN_PURPOSES),
            credit_score=random.randint(620, 680),
            risk_profile=RiskProfile.MEDIUM,
            expected_outcome="MANUAL_REVIEW",
            expected_score_range=(35, 60),
        )

    def _generate_high_risk(self) -> GeneratedApplication:
        """Genere une demande a haut risque."""
        income = random.uniform(2500, 5000)
        amount = random.uniform(30000, income * 10)
        debts = random.uniform(income * 0.35, income * 0.55)

        return GeneratedApplication(
            application_id=self._generate_id("APP"),
            applicant_id=self._generate_id("CUST"),
            amount_requested=round(amount, 2),
            currency="USD",
            declared_monthly_income=round(income, 2),
            employment_status=random.choice([
                EmploymentStatus.PART_TIME.value,
                EmploymentStatus.SELF_EMPLOYED.value,
                EmploymentStatus.SELF_EMPLOYED.value,  # Plus risque
            ]),
            existing_debts=round(debts, 2),
            loan_purpose=random.choice(self.LOAN_PURPOSES[-5:]),  # Objectifs risques
            credit_score=random.randint(550, 620),
            risk_profile=RiskProfile.HIGH,
            expected_outcome="MANUAL_REVIEW",
            expected_score_range=(55, 80),
        )

    def _generate_critical_risk(self) -> GeneratedApplication:
        """Genere une demande a risque critique."""
        income = random.uniform(1500, 3500)
        amount = random.uniform(50000, 150000)  # Montant eleve
        debts = random.uniform(income * 0.5, income * 0.8)

        return GeneratedApplication(
            application_id=self._generate_id("APP"),
            applicant_id=self._generate_id("CUST"),
            amount_requested=round(amount, 2),
            currency="USD",
            declared_monthly_income=round(income, 2),
            employment_status=random.choice([
                EmploymentStatus.UNEMPLOYED.value,
                EmploymentStatus.PART_TIME.value,
                EmploymentStatus.SELF_EMPLOYED.value,
            ]),
            existing_debts=round(debts, 2),
            loan_purpose=random.choice(self.LOAN_PURPOSES[-3:]),
            credit_score=random.randint(300, 550),
            risk_profile=RiskProfile.CRITICAL,
            expected_outcome="REJECTED",
            expected_score_range=(75, 100),
        )

    def _generate_edge_case(self) -> GeneratedApplication:
        """Genere un cas limite pour tester les bornes."""
        edge_cases = [
            # Montant exactement a la limite
            lambda: {
                "amount": 100000.00,
                "income": 10000,
                "debts": 2000,
                "status": EmploymentStatus.FULL_TIME.value,
                "outcome": "MANUAL_REVIEW",  # High value
            },
            # DTI exactement a 40%
            lambda: {
                "amount": 50000,
                "income": 5000,
                "debts": 2000,  # DTI = 40%
                "status": EmploymentStatus.FULL_TIME.value,
                "outcome": "MANUAL_REVIEW",
            },
            # Score limite 20 (auto-approve threshold)
            lambda: {
                "amount": 15000,
                "income": 6000,
                "debts": 500,
                "status": EmploymentStatus.FULL_TIME.value,
                "outcome": "APPROVED",
            },
            # Unemployed avec petit montant
            lambda: {
                "amount": 5000,
                "income": 2000,  # Allocations
                "debts": 100,
                "status": EmploymentStatus.UNEMPLOYED.value,
                "outcome": "REJECTED",
            },
            # Self-employed avec excellent profil
            lambda: {
                "amount": 75000,
                "income": 12000,
                "debts": 1000,
                "status": EmploymentStatus.SELF_EMPLOYED.value,
                "outcome": "MANUAL_REVIEW",
            },
        ]

        case_fn = random.choice(edge_cases)
        case = case_fn()

        return GeneratedApplication(
            application_id=self._generate_id("EDGE"),
            applicant_id=self._generate_id("CUST"),
            amount_requested=case["amount"],
            currency="USD",
            declared_monthly_income=case["income"],
            employment_status=case["status"],
            existing_debts=case["debts"],
            loan_purpose="Cas de test limite",
            credit_score=random.randint(600, 700),
            risk_profile=RiskProfile.EDGE_CASE,
            expected_outcome=case["outcome"],
            expected_score_range=(15, 85),  # Large range for edge cases
        )

    def generate_one(self, profile: RiskProfile) -> GeneratedApplication:
        """Genere une demande selon un profil specifique."""
        generators = {
            RiskProfile.VERY_LOW: self._generate_very_low_risk,
            RiskProfile.LOW: self._generate_low_risk,
            RiskProfile.MEDIUM: self._generate_medium_risk,
            RiskProfile.HIGH: self._generate_high_risk,
            RiskProfile.CRITICAL: self._generate_critical_risk,
            RiskProfile.EDGE_CASE: self._generate_edge_case,
        }
        return generators[profile]()

    def generate_dataset(
        self,
        n: int = 50,
        distribution: Optional[Dict[RiskProfile, float]] = None
    ) -> List[GeneratedApplication]:
        """
        Genere un dataset de demandes.

        Args:
            n: Nombre de demandes a generer
            distribution: Distribution des profils (optionnel)

        Returns:
            Liste de demandes generees
        """
        distribution = distribution or self.DEFAULT_DISTRIBUTION

        dataset = []
        for profile, ratio in distribution.items():
            count = int(n * ratio)
            for _ in range(count):
                dataset.append(self.generate_one(profile))

        # Completer si necessaire
        while len(dataset) < n:
            profile = random.choice(list(RiskProfile))
            dataset.append(self.generate_one(profile))

        # Melanger
        random.shuffle(dataset)

        return dataset[:n]

    def generate_balanced_dataset(self, per_profile: int = 10) -> List[GeneratedApplication]:
        """
        Genere un dataset equilibre avec le meme nombre par profil.

        Args:
            per_profile: Nombre de demandes par profil

        Returns:
            Liste de demandes generees
        """
        dataset = []
        for profile in RiskProfile:
            for _ in range(per_profile):
                dataset.append(self.generate_one(profile))

        random.shuffle(dataset)
        return dataset


# =============================================================================
# FONCTIONS UTILITAIRES
# =============================================================================

def generate_simulation_dataset(n: int = 50, seed: int = 42) -> List[Dict[str, Any]]:
    """
    Fonction helper pour generer un dataset de simulation.

    Args:
        n: Nombre de demandes
        seed: Graine pour reproductibilite

    Returns:
        Liste de dictionnaires (applications + metadata)
    """
    generator = DatasetGenerator(seed=seed)
    applications = generator.generate_dataset(n)

    return [
        {
            "application": app.to_dict(),
            "metadata": {
                "risk_profile": app.risk_profile.value,
                "expected_outcome": app.expected_outcome,
                "expected_score_range": app.expected_score_range,
                "credit_score": app.credit_score,
            }
        }
        for app in applications
    ]


def get_dataset_stats(dataset: List[GeneratedApplication]) -> Dict[str, Any]:
    """Calcule les statistiques d'un dataset."""
    stats = {
        "total": len(dataset),
        "by_profile": {},
        "by_expected_outcome": {},
        "amounts": {
            "min": min(a.amount_requested for a in dataset),
            "max": max(a.amount_requested for a in dataset),
            "avg": sum(a.amount_requested for a in dataset) / len(dataset),
        },
    }

    for profile in RiskProfile:
        count = sum(1 for a in dataset if a.risk_profile == profile)
        stats["by_profile"][profile.value] = count

    for outcome in ["APPROVED", "REJECTED", "MANUAL_REVIEW"]:
        count = sum(1 for a in dataset if a.expected_outcome == outcome)
        stats["by_expected_outcome"][outcome] = count

    return stats


# =============================================================================
# MAIN
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("DATASET GENERATOR - AgentMeshKafka")
    print("=" * 60)

    generator = DatasetGenerator(seed=42)
    dataset = generator.generate_dataset(n=50)
    stats = get_dataset_stats(dataset)

    print(f"\nDataset genere: {stats['total']} demandes")
    print(f"\nPar profil de risque:")
    for profile, count in stats["by_profile"].items():
        print(f"  {profile}: {count} ({count/stats['total']*100:.1f}%)")

    print(f"\nPar resultat attendu:")
    for outcome, count in stats["by_expected_outcome"].items():
        print(f"  {outcome}: {count} ({count/stats['total']*100:.1f}%)")

    print(f"\nMontants:")
    print(f"  Min: ${stats['amounts']['min']:,.2f}")
    print(f"  Max: ${stats['amounts']['max']:,.2f}")
    print(f"  Moy: ${stats['amounts']['avg']:,.2f}")

    print("\n" + "=" * 60)
