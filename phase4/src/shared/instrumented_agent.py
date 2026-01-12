"""
Phase 4 - Instrumented Agent Base Class
========================================
Classe de base pour agents avec instrumentation OpenTelemetry integree.
"""

import time
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional, TypeVar, Generic
from dataclasses import dataclass, field

from .telemetry import (
    AgentTelemetry,
    AgentName,
    TelemetryConfig,
    trace_agent_operation,
    trace_llm_call,
    trace_rag_query,
    KafkaTracingMiddleware,
)
from .logging import AgentLogger


# =============================================================================
# DATA MODELS
# =============================================================================

@dataclass
class AgentContext:
    """Contexte d'execution d'un agent."""

    application_id: str
    request_id: str
    trace_headers: Dict[str, str] = field(default_factory=dict)
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class AgentResult:
    """Resultat d'une operation d'agent."""

    success: bool
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    latency_ms: float = 0.0
    tokens_used: int = 0


T = TypeVar("T")


# =============================================================================
# INSTRUMENTED AGENT BASE
# =============================================================================

class InstrumentedAgent(ABC, Generic[T]):
    """
    Classe de base pour agents avec instrumentation complete.

    Fournit:
    - Tracing OpenTelemetry automatique
    - Logging structure avec correlation
    - Metriques de performance
    - Integration Kafka tracee

    Usage:
        class RiskAgent(InstrumentedAgent[RiskEvaluation]):
            def __init__(self):
                super().__init__(AgentName.RISK)

            def _execute_core(self, context, input_data):
                # Core logic here
                return RiskEvaluation(...)
    """

    def __init__(
        self,
        agent_name: AgentName,
        config: Optional[TelemetryConfig] = None,
        log_level: str = "INFO",
    ):
        self.agent_name = agent_name

        # Initialize telemetry
        self.telemetry = AgentTelemetry.initialize(agent_name, config)

        # Initialize logging
        self.logger = AgentLogger(
            agent_name.value,
            level=log_level,
        )

        # Kafka middleware
        self.kafka_middleware = KafkaTracingMiddleware()

        self.logger.info(
            "agent_initialized",
            agent=agent_name.value,
        )

    # -------------------------------------------------------------------------
    # ABSTRACT METHODS
    # -------------------------------------------------------------------------

    @abstractmethod
    def _execute_core(
        self,
        context: AgentContext,
        input_data: Dict[str, Any],
    ) -> T:
        """
        Logique metier principale de l'agent.

        A implementer par les sous-classes.

        Args:
            context: Contexte d'execution avec IDs de correlation
            input_data: Donnees d'entree a traiter

        Returns:
            Resultat type de l'agent
        """
        pass

    @abstractmethod
    def _validate_input(self, input_data: Dict[str, Any]) -> bool:
        """
        Valide les donnees d'entree.

        Args:
            input_data: Donnees a valider

        Returns:
            True si valide
        """
        pass

    # -------------------------------------------------------------------------
    # PUBLIC API
    # -------------------------------------------------------------------------

    def execute(
        self,
        input_data: Dict[str, Any],
        context: Optional[AgentContext] = None,
    ) -> AgentResult:
        """
        Execute l'agent avec instrumentation complete.

        Args:
            input_data: Donnees d'entree
            context: Contexte optionnel (genere si absent)

        Returns:
            AgentResult avec statut et donnees
        """
        # Generate context if not provided
        if context is None:
            context = self._create_context(input_data)

        start_time = time.time()

        # Start traced operation
        with self.telemetry.trace_operation(
            f"{self.agent_name.value}.execute",
            attributes={
                "application.id": context.application_id,
                "request.id": context.request_id,
            },
        ) as span:
            self.logger.operation_start(
                "execute",
                application_id=context.application_id,
            )

            try:
                # Validate input
                with self.telemetry.trace_operation("validate_input"):
                    if not self._validate_input(input_data):
                        raise ValueError("Input validation failed")

                # Execute core logic
                result_data = self._execute_core(context, input_data)

                # Calculate metrics
                latency_ms = (time.time() - start_time) * 1000

                self.logger.operation_end(
                    "execute",
                    success=True,
                    application_id=context.application_id,
                )

                self.telemetry.record_request(
                    success=True,
                    application_id=context.application_id,
                )

                return AgentResult(
                    success=True,
                    data=self._serialize_result(result_data),
                    latency_ms=latency_ms,
                )

            except Exception as e:
                self.logger.error(
                    "execute_failed",
                    application_id=context.application_id,
                    error=str(e),
                )

                self.telemetry.record_request(
                    success=False,
                    application_id=context.application_id,
                )

                return AgentResult(
                    success=False,
                    error=str(e),
                    latency_ms=(time.time() - start_time) * 1000,
                )

    # -------------------------------------------------------------------------
    # KAFKA INTEGRATION
    # -------------------------------------------------------------------------

    def consume_and_process(
        self,
        topic: str,
        message_value: Dict[str, Any],
        message_headers: Optional[Dict[str, str]] = None,
    ) -> AgentResult:
        """
        Consomme et traite un message Kafka avec tracing.

        Args:
            topic: Topic source
            message_value: Valeur du message
            message_headers: Headers Kafka (avec trace context)

        Returns:
            Resultat du traitement
        """
        application_id = message_value.get("application_id", "unknown")

        with self.kafka_middleware.trace_consumption(
            topic=topic,
            headers=message_headers,
            application_id=application_id,
        ):
            self.logger.log_kafka_message(
                topic=topic,
                operation="consume",
                application_id=application_id,
            )

            # Create context from headers
            context = AgentContext(
                application_id=application_id,
                request_id=message_value.get("request_id", f"req-{time.time()}"),
                trace_headers=message_headers or {},
            )

            return self.execute(message_value, context)

    def prepare_output_message(
        self,
        result: AgentResult,
        context: AgentContext,
    ) -> tuple[Dict[str, Any], Dict[str, str]]:
        """
        Prepare un message de sortie avec trace context.

        Args:
            result: Resultat de l'agent
            context: Contexte d'execution

        Returns:
            (message_value, message_headers)
        """
        headers = self.kafka_middleware.prepare_headers(
            application_id=context.application_id,
        )

        value = {
            "application_id": context.application_id,
            "request_id": context.request_id,
            "agent": self.agent_name.value,
            "success": result.success,
            "data": result.data,
            "error": result.error,
            "latency_ms": result.latency_ms,
        }

        return value, headers

    # -------------------------------------------------------------------------
    # LLM INTEGRATION
    # -------------------------------------------------------------------------

    @trace_llm_call(model="claude-3-5-sonnet")
    def call_llm(
        self,
        prompt: str,
        context: AgentContext,
        **kwargs,
    ) -> Dict[str, Any]:
        """
        Appel LLM instrumente.

        A surcharger pour implementation reelle.
        """
        # Placeholder - implement actual LLM call
        self.logger.debug(
            "llm_call_placeholder",
            prompt_length=len(prompt),
            application_id=context.application_id,
        )
        return {"response": "placeholder"}

    # -------------------------------------------------------------------------
    # RAG INTEGRATION
    # -------------------------------------------------------------------------

    @trace_rag_query(collection="policies")
    def query_rag(
        self,
        query: str,
        context: AgentContext,
        n_results: int = 5,
    ) -> list:
        """
        Requete RAG instrumentee.

        A surcharger pour implementation reelle.
        """
        # Placeholder - implement actual RAG query
        self.logger.debug(
            "rag_query_placeholder",
            query_length=len(query),
            n_results=n_results,
            application_id=context.application_id,
        )
        return []

    # -------------------------------------------------------------------------
    # HELPERS
    # -------------------------------------------------------------------------

    def _create_context(self, input_data: Dict[str, Any]) -> AgentContext:
        """Cree un contexte depuis les donnees d'entree."""
        return AgentContext(
            application_id=input_data.get("application_id", f"app-{time.time()}"),
            request_id=input_data.get("request_id", f"req-{time.time()}"),
        )

    def _serialize_result(self, result: T) -> Dict[str, Any]:
        """Serialise le resultat en dict."""
        if hasattr(result, "__dict__"):
            return result.__dict__
        if hasattr(result, "model_dump"):  # Pydantic
            return result.model_dump()
        return {"value": result}


# =============================================================================
# EXAMPLE IMPLEMENTATION
# =============================================================================

class ExampleRiskAgent(InstrumentedAgent[Dict[str, Any]]):
    """
    Exemple d'implementation d'agent instrumente.
    """

    def __init__(self):
        super().__init__(AgentName.RISK)

    def _validate_input(self, input_data: Dict[str, Any]) -> bool:
        """Valide une demande de pret."""
        required_fields = [
            "application_id",
            "amount_requested",
            "declared_monthly_income",
        ]
        return all(field in input_data for field in required_fields)

    def _execute_core(
        self,
        context: AgentContext,
        input_data: Dict[str, Any],
    ) -> Dict[str, Any]:
        """Evalue le risque d'une demande."""
        self.logger.info(
            "analyzing_risk",
            application_id=context.application_id,
            amount=input_data.get("amount_requested"),
        )

        # Calculate simple DTI ratio
        amount = input_data.get("amount_requested", 0)
        income = input_data.get("declared_monthly_income", 1)
        debts = input_data.get("existing_debts", 0)

        dti = (debts + amount * 0.03) / income * 100

        # Simple risk score
        risk_score = min(100, max(0, dti * 1.5))

        # Determine category
        if risk_score < 25:
            category = "LOW"
        elif risk_score < 50:
            category = "MEDIUM"
        elif risk_score < 75:
            category = "HIGH"
        else:
            category = "CRITICAL"

        result = {
            "risk_score": round(risk_score, 2),
            "risk_category": category,
            "dti_ratio": round(dti, 2),
            "rationale": f"DTI ratio of {dti:.1f}% indicates {category} risk",
        }

        self.logger.info(
            "risk_evaluated",
            application_id=context.application_id,
            risk_score=result["risk_score"],
            category=category,
        )

        return result


# =============================================================================
# EXAMPLE USAGE
# =============================================================================

if __name__ == "__main__":
    # Create agent
    agent = ExampleRiskAgent()

    # Sample application
    application = {
        "application_id": "APP-12345",
        "applicant_id": "CUST-001",
        "amount_requested": 50000,
        "declared_monthly_income": 5000,
        "existing_debts": 1000,
        "employment_status": "FULL_TIME",
    }

    # Execute with tracing
    result = agent.execute(application)

    print(f"\nResult: {result}")
    print(f"Success: {result.success}")
    print(f"Data: {result.data}")
    print(f"Latency: {result.latency_ms:.2f}ms")
