package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"weather/lib/ping"
	"weather/lib/server"
	"weather/lib/tracing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const svcName = "WeatherService"

func pingCaller(w http.ResponseWriter, r *http.Request) {
	tracer := otel.GetTracerProvider().Tracer("ping_caller_route on WeatherService")
	attrs, _, receivedCtx := otelhttptrace.Extract(r.Context(), r)

	spanLabels := []attribute.KeyValue{
		attribute.String("URI", r.RequestURI),
		attribute.String("METHOD", r.Method),
		attribute.String("PROTO", r.Proto),
	}

	baggageGetWeatherByCity, _ := baggage.NewMember(string("FunctionRoute"), "pingCallerRoute")
	baggageContents, err := baggage.New(baggageGetWeatherByCity)
	if err != nil {
		log.Fatalf("Error occurred: %s", err)
	}

	spanCtx, span := tracer.Start(
		trace.ContextWithRemoteSpanContext(r.Context(), receivedCtx),
		"ping_caller_route has been invoked",
		trace.WithAttributes(attrs...),
		trace.WithAttributes(spanLabels...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	defer span.End()

	pingServer, ok := os.LookupEnv("OWM_ADDR")
	if !ok {
		pingServer = "localhost:8082"
	}

	response, err := ping.Ping(spanCtx, pingServer, tracer)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Key("baggage").String(baggageContents.Member("FunctionRoute").Value()))
	span.SetStatus(codes.Ok, "requestPingCallerRouteSuccessfull")
	w.Write([]byte(fmt.Sprintf("%s -> %s", svcName, response)))

}

func weatherForecast(w http.ResponseWriter, r *http.Request) {
	tracer := otel.GetTracerProvider().Tracer("weatherForecast_route on WeatherService")
	attrs, _, receivedCtx := otelhttptrace.Extract(r.Context(), r)

	spanLabels := []attribute.KeyValue{
		attribute.String("URI", r.RequestURI),
		attribute.String("METHOD", r.Method),
		attribute.String("PROTO", r.Proto),
	}

	baggageWeatherForecast, _ := baggage.NewMember(string("FunctionRoute"), "weatherForecast()")
	baggageContents, err := baggage.New(baggageWeatherForecast)
	if err != nil {
		log.Fatalf("Error occurred: %s", err)
	}

	r = r.WithContext(baggage.ContextWithBaggage(r.Context(), baggageContents))

	spanCtx, span := tracer.Start(
		trace.ContextWithRemoteSpanContext(r.Context(), receivedCtx),
		"weatherForecast_route has been invoked",
		trace.WithAttributes(attrs...),
		trace.WithAttributes(spanLabels...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)

	defer span.End()

	owmAddr, ok := os.LookupEnv("OWM_ADDR")
	if !ok {
		owmAddr = "localhost:8082"
	}

	city := chi.URLParam(r, "city")
	wF, err := server.GetWeatherForecast(spanCtx, owmAddr, city, tracer)
	if err != nil {
		w.WriteHeader(500)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "requestWeatherForecastRouteSuccessfull")
	span.SetAttributes(attribute.Key("baggage").String(baggageContents.Member("FunctionRoute").Value()))
	render.JSON(w, r, wF)

}

func httpTraceWrapper(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t := otel.GetTracerProvider().Tracer("http-root-tracer")
		ctx, span := t.Start(r.Context(), r.URL.Path)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
		span.End()
	}
	return http.HandlerFunc(fn)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	tracerEndpoint := os.Getenv("TRACER_ENDPOINT")
	tracer, err := tracing.InitTracer(ctx, "oteltrace", svcName, tracerEndpoint)
	if err != nil {
		log.Fatalf("Error occurred: %s", err)
	}
	defer tracer()

	log.Printf("Starting %s", svcName)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/ping", pingCaller)
	r.Route("/forecast", func(r chi.Router) {
		r.Get("/{city}", weatherForecast)
	})

	errListen := http.ListenAndServe(":"+port, httpTraceWrapper(r))

	if errListen != nil {
		fmt.Println("ListenAndServe:", err)
	}

}
