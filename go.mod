module weather

go 1.16

require (
	github.com/ernesto-jimenez/httplogger v0.0.0-20150224132909-86cc44f6150a
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/render v1.0.1
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.22.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/exporters/jaeger v1.0.0-RC2
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC2
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.0.0-RC2
	go.opentelemetry.io/otel/sdk v1.0.0-RC2
	go.opentelemetry.io/otel/trace v1.0.0-RC2
	google.golang.org/grpc v1.39.1
)
