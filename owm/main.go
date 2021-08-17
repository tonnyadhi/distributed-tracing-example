package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"weather/lib/owmclient"
	"weather/lib/tracing"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

const svcName = "OWMService"

func pingReceiver(w http.ResponseWriter, r *http.Request) {

	tracer := otel.GetTracerProvider().Tracer("ping_receiver_route on OWMService")
	attrs, _, receivedCtx := otelhttptrace.Extract(r.Context(), r)

	spanLabels := []attribute.KeyValue{
		attribute.String("URI", r.RequestURI),
		attribute.String("METHOD", r.Method),
		attribute.String("PROTO", r.Proto),
	}

	baggageGetWeatherByCity, _ := baggage.NewMember(string("FunctionRoute"), "pingReceiverRoute")
	baggageContents, err := baggage.New(baggageGetWeatherByCity)
	if err != nil {
		log.Fatalf("Error occurred: %s", err)
	}

	_, span := tracer.Start(
		trace.ContextWithRemoteSpanContext(r.Context(), receivedCtx),
		"ping_receiver_route has been invoked",
		trace.WithAttributes(attrs...),
		trace.WithAttributes(spanLabels...),
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	span.SetAttributes(attribute.Key("baggage").String(baggageContents.Member("FunctionRoute").Value()))
	span.SetStatus(codes.Ok, "requestPingReceiverRouteSuccessfull")
	w.Write([]byte(fmt.Sprintf("%s", svcName)))

}

func getWeatherByCity(w http.ResponseWriter, r *http.Request) {

	tracer := otel.GetTracerProvider().Tracer("getWeatherByCity_route on OWMservice")
	attrs, _, receivedCtx := otelhttptrace.Extract(r.Context(), r)

	spanLabels := []attribute.KeyValue{
		attribute.String("URI", r.RequestURI),
		attribute.String("METHOD", r.Method),
		attribute.String("PROTO", r.Proto),
	}

	baggageGetWeatherByCity, _ := baggage.NewMember(string("FunctionRoute"), "getWeatherByCity()")
	baggageContents, err := baggage.New(baggageGetWeatherByCity)
	if err != nil {
		log.Fatalf("Error occurred: %s", err)
	}

	r = r.WithContext(baggage.ContextWithBaggage(r.Context(), baggageContents))

	spanCtx, span := tracer.Start(
		trace.ContextWithRemoteSpanContext(r.Context(), receivedCtx),
		"getWeatherByCity_route has been invoked",
		trace.WithAttributes(attrs...),
		trace.WithAttributes(spanLabels...),
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	city := chi.URLParam(r, "city")
	cityWeather, err := owmclient.GetOwmForecastByCity(spanCtx, city, tracer)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Key("baggage").String(baggageContents.Member("FunctionRoute").Value()))
	span.SetStatus(codes.Ok, "requestGetWeatherByCityRouteSuccessfull")
	render.JSON(w, r, cityWeather)
}

func httpTraceWrapper(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t := otel.GetTracerProvider().Tracer("http-root-tracer")
		ctx, span := t.Start(r.Context(), r.URL.Path)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
		defer span.End()
	}
	return http.HandlerFunc(fn)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := "8082"
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

	r.Get("/ping", pingReceiver)
	r.Route("/getweather/owm", func(r chi.Router) {
		r.Get("/{city}", getWeatherByCity)
	})

	errListen := http.ListenAndServe(":"+port, httpTraceWrapper(r))

	if errListen != nil {
		fmt.Println("ListenAndServe:", err)
	}
}
