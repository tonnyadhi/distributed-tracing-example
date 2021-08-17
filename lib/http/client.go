package xhttp

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ernesto-jimenez/httplogger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

/*
httpLogger log http request response
*/
type httpLogger struct {
	log *log.Logger
}

func newLogger() *httpLogger {
	return &httpLogger{
		log: log.New(os.Stderr, "log - ", log.LstdFlags),
	}
}

func (l *httpLogger) LogRequest(req *http.Request) {
	l.log.Printf(
		"Request %s %s\n",
		req.Method,
		req.URL.String(),
	)
	l.log.Printf(
		"Request %+q %s",
		req.Header["User-Agent"],
		req.URL.String(),
	)
}

func (l *httpLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	duration /= time.Millisecond
	if err != nil {
		l.log.Println(err)
	} else {
		l.log.Printf(
			"Response method=%s status=%d durationMs=%d %s",
			req.Method,
			res.StatusCode,
			duration,
			req.URL.String(),
		)
	}
}

func Do(ctx context.Context, req *http.Request, tracer trace.Tracer) (*http.Response, error) {

	spanCtx, span := tracer.Start(ctx, req.RequestURI, trace.WithSpanKind(trace.SpanKindClient))

	defer span.End()

	span.SetAttributes(attribute.Key("http_client_call").String("inside xhttp Do"))

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: httplogger.NewLoggedTransport(http.DefaultTransport, newLogger()),
	}

	otelCtx, req := otelhttptrace.W3C(spanCtx, req)
	otelhttptrace.Inject(otelCtx, req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(ctx).SetStatus(codes.Ok, "requestHTTPSuccessfull")
	return resp, nil

}
