package ping

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	libhttp "weather/lib/http"
)

func Ping(ctx context.Context, owmHost string, tracer trace.Tracer) (string, error) {

	requestPath := "ping"

	spanCtx, span := tracer.Start(ctx, "call_Ping", trace.WithAttributes(attribute.Key("Ping").String("returning_ping")))
	defer span.End()

	url := fmt.Sprintf("http://%s/%s", owmHost, requestPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := libhttp.Do(spanCtx, req, tracer)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("StatusCode: %d, Body: %s", resp.StatusCode, body)
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestPingSuccessfull")
	return string(body), nil
}
