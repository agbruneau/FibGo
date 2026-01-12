"""
Phase 4 - Structured Logging with OpenTelemetry Correlation
============================================================
Integration de structlog avec OpenTelemetry pour logs correles.
"""

import logging
import sys
from typing import Any, Dict, Optional

import structlog
from opentelemetry import trace


def add_trace_context(
    logger: logging.Logger,
    method_name: str,
    event_dict: Dict[str, Any],
) -> Dict[str, Any]:
    """
    Processeur structlog pour ajouter le contexte de trace.

    Ajoute automatiquement trace_id et span_id aux logs.
    """
    span = trace.get_current_span()
    if span.is_recording():
        ctx = span.get_span_context()
        event_dict["trace_id"] = format(ctx.trace_id, "032x")
        event_dict["span_id"] = format(ctx.span_id, "016x")
    return event_dict


def add_agent_context(agent_name: str):
    """
    Factory pour processeur qui ajoute le contexte d'agent.
    """
    def processor(
        logger: logging.Logger,
        method_name: str,
        event_dict: Dict[str, Any],
    ) -> Dict[str, Any]:
        event_dict["agent"] = agent_name
        return event_dict
    return processor


def configure_logging(
    agent_name: str,
    level: str = "INFO",
    json_format: bool = False,
) -> structlog.BoundLogger:
    """
    Configure le logging structure avec correlation OpenTelemetry.

    Args:
        agent_name: Nom de l'agent (pour contexte)
        level: Niveau de log (DEBUG, INFO, WARNING, ERROR)
        json_format: True pour format JSON (production)

    Returns:
        Logger configure

    Usage:
        logger = configure_logging("agent-risk", level="DEBUG")
        logger.info("Processing application", application_id="APP-123")
    """
    # Processors chain
    processors = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        add_agent_context(agent_name),
        add_trace_context,
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
    ]

    if json_format:
        processors.append(structlog.processors.JSONRenderer())
    else:
        processors.append(structlog.dev.ConsoleRenderer(colors=True))

    structlog.configure(
        processors=processors,
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(logging, level.upper())
        ),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )

    return structlog.get_logger()


class AgentLogger:
    """
    Logger specialise pour les agents avec metriques integrees.

    Usage:
        logger = AgentLogger("agent-risk")
        logger.info("Starting analysis", application_id="APP-123")
        logger.operation_start("validate_input")
        logger.operation_end("validate_input", success=True)
    """

    def __init__(
        self,
        agent_name: str,
        level: str = "INFO",
        json_format: bool = False,
    ):
        self.agent_name = agent_name
        self._logger = configure_logging(agent_name, level, json_format)
        self._operation_starts: Dict[str, float] = {}

    def _log(self, level: str, event: str, **kwargs):
        """Log avec le niveau specifie."""
        log_method = getattr(self._logger, level)
        log_method(event, **kwargs)

    def debug(self, event: str, **kwargs):
        self._log("debug", event, **kwargs)

    def info(self, event: str, **kwargs):
        self._log("info", event, **kwargs)

    def warning(self, event: str, **kwargs):
        self._log("warning", event, **kwargs)

    def error(self, event: str, **kwargs):
        self._log("error", event, **kwargs)

    def exception(self, event: str, **kwargs):
        self._log("exception", event, **kwargs)

    # Operation tracking
    def operation_start(self, operation: str, **context):
        """Log le debut d'une operation."""
        import time
        self._operation_starts[operation] = time.time()
        self.info(f"operation_start:{operation}", **context)

    def operation_end(
        self,
        operation: str,
        success: bool = True,
        **context,
    ):
        """Log la fin d'une operation avec duree."""
        import time
        duration_ms = None
        if operation in self._operation_starts:
            duration_ms = (time.time() - self._operation_starts[operation]) * 1000
            del self._operation_starts[operation]

        self.info(
            f"operation_end:{operation}",
            success=success,
            duration_ms=duration_ms,
            **context,
        )

    # Specialized logging methods
    def log_llm_call(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
        latency_ms: float,
        **context,
    ):
        """Log un appel LLM."""
        self.info(
            "llm_call",
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            total_tokens=input_tokens + output_tokens,
            latency_ms=latency_ms,
            **context,
        )

    def log_kafka_message(
        self,
        topic: str,
        operation: str,  # "produce" or "consume"
        application_id: Optional[str] = None,
        **context,
    ):
        """Log une operation Kafka."""
        self.info(
            f"kafka_{operation}",
            topic=topic,
            application_id=application_id,
            **context,
        )

    def log_rag_query(
        self,
        collection: str,
        query_length: int,
        results_count: int,
        latency_ms: float,
        **context,
    ):
        """Log une requete RAG."""
        self.info(
            "rag_query",
            collection=collection,
            query_length=query_length,
            results_count=results_count,
            latency_ms=latency_ms,
            **context,
        )

    def log_decision(
        self,
        application_id: str,
        decision: str,
        risk_score: float,
        **context,
    ):
        """Log une decision de pret."""
        self.info(
            "loan_decision",
            application_id=application_id,
            decision=decision,
            risk_score=risk_score,
            **context,
        )


# =============================================================================
# EXAMPLE USAGE
# =============================================================================

if __name__ == "__main__":
    # Initialize telemetry first (for trace context)
    from .telemetry import AgentTelemetry, AgentName

    telemetry = AgentTelemetry.initialize(
        agent_name=AgentName.RISK,
    )

    # Create logger
    logger = AgentLogger("agent-risk", level="DEBUG")

    # Log with trace correlation
    with telemetry.trace_operation("example_operation"):
        logger.info(
            "Processing loan application",
            application_id="APP-12345",
            amount=50000,
        )

        logger.operation_start("validate_input", application_id="APP-12345")
        # ... validation logic ...
        logger.operation_end("validate_input", success=True)

        logger.log_llm_call(
            model="claude-3-5-sonnet",
            input_tokens=500,
            output_tokens=200,
            latency_ms=1500,
        )

        logger.log_decision(
            application_id="APP-12345",
            decision="APPROVED",
            risk_score=25.5,
        )

    print("\nLogging example completed!")
