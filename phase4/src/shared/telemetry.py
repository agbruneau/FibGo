"""
Phase 4 - OpenTelemetry Integration for Distributed Tracing
============================================================
Module de telemetrie pour le tracing distribue du mesh d'agents.

Fonctionnalites:
- Tracing distribue entre agents via Kafka
- Metriques de performance (latence, tokens, throughput)
- Context propagation pour correlation end-to-end
- Integration avec les 3 agents (Intake, Risk, Decision)
"""

import os
import time
import functools
from typing import Any, Callable, Dict, Optional, TypeVar, ParamSpec
from dataclasses import dataclass, field
from contextlib import contextmanager
from enum import Enum

from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    BatchSpanProcessor,
    ConsoleSpanExporter,
)
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import (
    PeriodicExportingMetricReader,
    ConsoleMetricExporter,
)
from opentelemetry.sdk.resources import Resource, SERVICE_NAME
from opentelemetry.trace import Status, StatusCode, SpanKind
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
from opentelemetry.context import Context


# =============================================================================
# CONFIGURATION
# =============================================================================

class AgentName(str, Enum):
    """Noms des agents du systeme."""
    INTAKE = "agent-intake"
    RISK = "agent-risk"
    DECISION = "agent-decision"
    ORCHESTRATOR = "orchestrator"


@dataclass
class TelemetryConfig:
    """Configuration de la telemetrie."""

    service_name: str = "agent-mesh-kafka"
    service_version: str = "1.0.0"
    environment: str = "development"

    # Export settings
    enable_console_export: bool = True
    enable_otlp_export: bool = False
    otlp_endpoint: str = "http://localhost:4317"

    # Sampling
    sample_rate: float = 1.0  # 1.0 = 100% des traces

    # Metrics
    metrics_export_interval_ms: int = 60000

    @classmethod
    def from_env(cls) -> "TelemetryConfig":
        """Charge la configuration depuis les variables d'environnement."""
        return cls(
            service_name=os.getenv("OTEL_SERVICE_NAME", "agent-mesh-kafka"),
            service_version=os.getenv("SERVICE_VERSION", "1.0.0"),
            environment=os.getenv("ENVIRONMENT", "development"),
            enable_console_export=os.getenv("OTEL_CONSOLE_EXPORT", "true").lower() == "true",
            enable_otlp_export=os.getenv("OTEL_OTLP_EXPORT", "false").lower() == "true",
            otlp_endpoint=os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
            sample_rate=float(os.getenv("OTEL_SAMPLE_RATE", "1.0")),
        )


# =============================================================================
# TELEMETRY SINGLETON
# =============================================================================

class AgentTelemetry:
    """
    Gestionnaire de telemetrie pour le mesh d'agents.

    Pattern Singleton pour assurer une seule instance par processus.

    Usage:
        telemetry = AgentTelemetry.initialize(agent_name=AgentName.RISK)

        with telemetry.trace_operation("analyze_risk") as span:
            span.set_attribute("application_id", "APP-123")
            result = do_analysis()
    """

    _instance: Optional["AgentTelemetry"] = None
    _initialized: bool = False

    def __new__(cls, *args, **kwargs):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(
        self,
        agent_name: AgentName,
        config: Optional[TelemetryConfig] = None,
    ):
        if self._initialized:
            return

        self.agent_name = agent_name
        self.config = config or TelemetryConfig.from_env()

        # Setup resource
        self.resource = Resource.create({
            SERVICE_NAME: f"{self.config.service_name}-{agent_name.value}",
            "service.version": self.config.service_version,
            "deployment.environment": self.config.environment,
            "agent.name": agent_name.value,
        })

        # Setup tracing
        self._setup_tracing()

        # Setup metrics
        self._setup_metrics()

        # Propagator for distributed context
        self.propagator = TraceContextTextMapPropagator()

        self._initialized = True

    @classmethod
    def initialize(
        cls,
        agent_name: AgentName,
        config: Optional[TelemetryConfig] = None,
    ) -> "AgentTelemetry":
        """Initialise et retourne l'instance de telemetrie."""
        return cls(agent_name, config)

    @classmethod
    def get_instance(cls) -> "AgentTelemetry":
        """Retourne l'instance existante."""
        if cls._instance is None:
            raise RuntimeError("Telemetry not initialized. Call initialize() first.")
        return cls._instance

    def _setup_tracing(self) -> None:
        """Configure le tracing OpenTelemetry."""
        provider = TracerProvider(resource=self.resource)

        # Console exporter (dev)
        if self.config.enable_console_export:
            processor = BatchSpanProcessor(ConsoleSpanExporter())
            provider.add_span_processor(processor)

        # OTLP exporter (production - Jaeger, Zipkin, etc.)
        if self.config.enable_otlp_export:
            try:
                from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import (
                    OTLPSpanExporter,
                )
                otlp_exporter = OTLPSpanExporter(endpoint=self.config.otlp_endpoint)
                provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
            except ImportError:
                print("Warning: OTLP exporter not available. Install opentelemetry-exporter-otlp")

        trace.set_tracer_provider(provider)
        self.tracer = trace.get_tracer(
            instrumenting_module_name=f"agent.{self.agent_name.value}",
            instrumenting_library_version=self.config.service_version,
        )

    def _setup_metrics(self) -> None:
        """Configure les metriques OpenTelemetry."""
        readers = []

        if self.config.enable_console_export:
            readers.append(
                PeriodicExportingMetricReader(
                    ConsoleMetricExporter(),
                    export_interval_millis=self.config.metrics_export_interval_ms,
                )
            )

        provider = MeterProvider(resource=self.resource, metric_readers=readers)
        metrics.set_meter_provider(provider)

        self.meter = metrics.get_meter(
            name=f"agent.{self.agent_name.value}",
            version=self.config.service_version,
        )

        # Create standard metrics
        self._create_metrics()

    def _create_metrics(self) -> None:
        """Cree les metriques standard pour les agents."""
        # Latency histogram
        self.latency_histogram = self.meter.create_histogram(
            name="agent.operation.duration",
            description="Duration of agent operations in milliseconds",
            unit="ms",
        )

        # Token counter
        self.token_counter = self.meter.create_counter(
            name="agent.llm.tokens",
            description="Number of tokens used in LLM calls",
            unit="tokens",
        )

        # Request counter
        self.request_counter = self.meter.create_counter(
            name="agent.requests",
            description="Number of requests processed",
            unit="requests",
        )

        # Error counter
        self.error_counter = self.meter.create_counter(
            name="agent.errors",
            description="Number of errors encountered",
            unit="errors",
        )

        # Active operations gauge
        self.active_operations = self.meter.create_up_down_counter(
            name="agent.active_operations",
            description="Number of currently active operations",
            unit="operations",
        )

    # -------------------------------------------------------------------------
    # CONTEXT PROPAGATION (for Kafka)
    # -------------------------------------------------------------------------

    def inject_context(self, carrier: Dict[str, str]) -> None:
        """
        Injecte le contexte de trace dans un carrier (headers Kafka).

        Usage:
            headers = {}
            telemetry.inject_context(headers)
            producer.send(topic, value=data, headers=headers)
        """
        self.propagator.inject(carrier)

    def extract_context(self, carrier: Dict[str, str]) -> Context:
        """
        Extrait le contexte de trace d'un carrier (headers Kafka).

        Usage:
            context = telemetry.extract_context(message.headers)
            with telemetry.trace_operation("process", context=context):
                process(message)
        """
        return self.propagator.extract(carrier=carrier)

    # -------------------------------------------------------------------------
    # TRACING HELPERS
    # -------------------------------------------------------------------------

    @contextmanager
    def trace_operation(
        self,
        operation_name: str,
        kind: SpanKind = SpanKind.INTERNAL,
        attributes: Optional[Dict[str, Any]] = None,
        context: Optional[Context] = None,
    ):
        """
        Context manager pour tracer une operation.

        Usage:
            with telemetry.trace_operation("validate_input") as span:
                span.set_attribute("field_count", len(data))
                result = validate(data)
        """
        self.active_operations.add(1)
        start_time = time.time()

        with self.tracer.start_as_current_span(
            name=operation_name,
            kind=kind,
            attributes=attributes,
            context=context,
        ) as span:
            try:
                yield span
                span.set_status(Status(StatusCode.OK))
            except Exception as e:
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                self.error_counter.add(1, {"operation": operation_name})
                raise
            finally:
                duration_ms = (time.time() - start_time) * 1000
                self.latency_histogram.record(
                    duration_ms,
                    {"operation": operation_name, "agent": self.agent_name.value},
                )
                self.active_operations.add(-1)

    def record_llm_call(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
        latency_ms: float,
    ) -> None:
        """Enregistre les metriques d'un appel LLM."""
        span = trace.get_current_span()
        span.set_attributes({
            "llm.model": model,
            "llm.input_tokens": input_tokens,
            "llm.output_tokens": output_tokens,
            "llm.latency_ms": latency_ms,
        })

        self.token_counter.add(
            input_tokens + output_tokens,
            {"model": model, "type": "total"},
        )
        self.token_counter.add(input_tokens, {"model": model, "type": "input"})
        self.token_counter.add(output_tokens, {"model": model, "type": "output"})

    def record_request(self, success: bool, application_id: str) -> None:
        """Enregistre une requete traitee."""
        self.request_counter.add(
            1,
            {
                "success": str(success).lower(),
                "agent": self.agent_name.value,
            },
        )

        span = trace.get_current_span()
        span.set_attributes({
            "application.id": application_id,
            "request.success": success,
        })


# =============================================================================
# CONVENIENCE FUNCTIONS
# =============================================================================

def get_tracer(name: Optional[str] = None) -> trace.Tracer:
    """Retourne un tracer pour instrumentation manuelle."""
    return trace.get_tracer(name or "agent-mesh")


def get_meter(name: Optional[str] = None) -> metrics.Meter:
    """Retourne un meter pour metriques manuelles."""
    return metrics.get_meter(name or "agent-mesh")


# =============================================================================
# DECORATORS
# =============================================================================

P = ParamSpec("P")
T = TypeVar("T")


def trace_agent_operation(
    operation_name: Optional[str] = None,
    record_args: bool = False,
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """
    Decorateur pour tracer automatiquement les operations d'agent.

    Usage:
        @trace_agent_operation("validate_application")
        def validate(self, application: dict) -> bool:
            ...
    """
    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        @functools.wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            name = operation_name or func.__name__
            telemetry = AgentTelemetry.get_instance()

            attributes = {"function": func.__name__}
            if record_args and kwargs:
                # Record safe kwargs (not sensitive data)
                safe_keys = ["application_id", "request_id", "step"]
                for key in safe_keys:
                    if key in kwargs:
                        attributes[f"arg.{key}"] = str(kwargs[key])

            with telemetry.trace_operation(name, attributes=attributes):
                return func(*args, **kwargs)

        return wrapper
    return decorator


def trace_kafka_message(
    topic: Optional[str] = None,
    operation: str = "process",
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """
    Decorateur pour tracer le traitement d'un message Kafka.

    Usage:
        @trace_kafka_message(topic="loan-applications")
        def handle_message(self, message: dict) -> None:
            ...
    """
    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        @functools.wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            telemetry = AgentTelemetry.get_instance()

            # Extract context from message headers if available
            context = None
            if "headers" in kwargs and kwargs["headers"]:
                context = telemetry.extract_context(kwargs["headers"])

            attributes = {
                "messaging.system": "kafka",
                "messaging.operation": operation,
            }
            if topic:
                attributes["messaging.destination"] = topic

            with telemetry.trace_operation(
                f"kafka.{operation}",
                kind=SpanKind.CONSUMER,
                attributes=attributes,
                context=context,
            ):
                return func(*args, **kwargs)

        return wrapper
    return decorator


def trace_llm_call(
    model: Optional[str] = None,
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """
    Decorateur pour tracer les appels LLM.

    Usage:
        @trace_llm_call(model="claude-3-5-sonnet")
        async def call_claude(self, prompt: str) -> str:
            ...
    """
    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        @functools.wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            telemetry = AgentTelemetry.get_instance()

            attributes = {"llm.model": model or "unknown"}

            with telemetry.trace_operation(
                "llm.call",
                kind=SpanKind.CLIENT,
                attributes=attributes,
            ) as span:
                start = time.time()
                result = func(*args, **kwargs)
                latency = (time.time() - start) * 1000

                # Try to extract token counts from result
                if hasattr(result, "usage"):
                    telemetry.record_llm_call(
                        model=model or "unknown",
                        input_tokens=getattr(result.usage, "input_tokens", 0),
                        output_tokens=getattr(result.usage, "output_tokens", 0),
                        latency_ms=latency,
                    )

                return result

        return wrapper
    return decorator


def trace_rag_query(
    collection: Optional[str] = None,
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """
    Decorateur pour tracer les requetes RAG.

    Usage:
        @trace_rag_query(collection="bank-policies")
        def query_policies(self, query: str) -> List[Document]:
            ...
    """
    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        @functools.wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            telemetry = AgentTelemetry.get_instance()

            attributes = {
                "rag.collection": collection or "unknown",
                "rag.operation": "query",
            }

            with telemetry.trace_operation(
                "rag.query",
                kind=SpanKind.CLIENT,
                attributes=attributes,
            ) as span:
                result = func(*args, **kwargs)

                # Record result count if available
                if hasattr(result, "__len__"):
                    span.set_attribute("rag.results_count", len(result))

                return result

        return wrapper
    return decorator


# =============================================================================
# KAFKA INTEGRATION HELPERS
# =============================================================================

@dataclass
class TracedMessage:
    """Message avec contexte de trace pour Kafka."""

    topic: str
    key: Optional[str]
    value: Dict[str, Any]
    headers: Dict[str, str] = field(default_factory=dict)

    def inject_trace_context(self) -> "TracedMessage":
        """Injecte le contexte de trace dans les headers."""
        telemetry = AgentTelemetry.get_instance()
        telemetry.inject_context(self.headers)
        return self


class KafkaTracingMiddleware:
    """
    Middleware pour le tracing automatique des messages Kafka.

    Usage:
        middleware = KafkaTracingMiddleware()

        # Pour producer
        headers = middleware.prepare_headers(application_id="APP-123")
        producer.send(topic, value=data, headers=headers)

        # Pour consumer
        with middleware.trace_consumption(message):
            process(message)
    """

    def __init__(self):
        self.telemetry = AgentTelemetry.get_instance()

    def prepare_headers(
        self,
        application_id: Optional[str] = None,
        extra_headers: Optional[Dict[str, str]] = None,
    ) -> Dict[str, str]:
        """Prepare les headers avec contexte de trace."""
        headers = extra_headers.copy() if extra_headers else {}

        # Inject trace context
        self.telemetry.inject_context(headers)

        # Add application correlation
        if application_id:
            headers["x-application-id"] = application_id

        return headers

    @contextmanager
    def trace_consumption(
        self,
        topic: str,
        headers: Optional[Dict[str, str]] = None,
        application_id: Optional[str] = None,
    ):
        """Context manager pour tracer la consommation d'un message."""
        context = None
        if headers:
            context = self.telemetry.extract_context(headers)

        attributes = {
            "messaging.system": "kafka",
            "messaging.destination": topic,
            "messaging.operation": "receive",
        }
        if application_id:
            attributes["application.id"] = application_id

        with self.telemetry.trace_operation(
            f"kafka.consume.{topic}",
            kind=SpanKind.CONSUMER,
            attributes=attributes,
            context=context,
        ):
            yield

    @contextmanager
    def trace_production(
        self,
        topic: str,
        application_id: Optional[str] = None,
    ):
        """Context manager pour tracer la production d'un message."""
        attributes = {
            "messaging.system": "kafka",
            "messaging.destination": topic,
            "messaging.operation": "send",
        }
        if application_id:
            attributes["application.id"] = application_id

        with self.telemetry.trace_operation(
            f"kafka.produce.{topic}",
            kind=SpanKind.PRODUCER,
            attributes=attributes,
        ):
            yield


# =============================================================================
# EXAMPLE USAGE
# =============================================================================

if __name__ == "__main__":
    # Initialize telemetry for Risk Agent
    telemetry = AgentTelemetry.initialize(
        agent_name=AgentName.RISK,
        config=TelemetryConfig(
            enable_console_export=True,
            environment="development",
        ),
    )

    # Example: Trace an operation
    with telemetry.trace_operation("example_analysis") as span:
        span.set_attribute("application_id", "APP-12345")
        span.set_attribute("amount_requested", 50000)

        # Simulate work
        import time
        time.sleep(0.1)

        # Record LLM call
        telemetry.record_llm_call(
            model="claude-3-5-sonnet-20241022",
            input_tokens=500,
            output_tokens=200,
            latency_ms=1500,
        )

    # Example: Kafka middleware
    middleware = KafkaTracingMiddleware()
    headers = middleware.prepare_headers(application_id="APP-12345")
    print(f"Prepared headers: {headers}")

    with middleware.trace_production("risk-evaluations", application_id="APP-12345"):
        print("Producing message...")

    print("\nTelemetry example completed!")
