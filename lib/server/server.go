package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	libhttp "weather/lib/http"
)

//RequestWeatherForecast represent return from owm service
type RequestWeatherForecast struct {
	Condition   string  `json:"Condition"`
	Temperature float64 `json:"Temperature"`
	Humidity    int32   `json:"Humidity"`
}

func GetWeatherForecast(ctx context.Context, owmHost string, city string, tracer trace.Tracer) (*RequestWeatherForecast, error) {
	//owmHost := os.Getenv("OWM_HOST")
	requestPath := "getweather/owm"
	requestParam := city

	spanCtx, span := tracer.Start(ctx, "call_GetWeatherForecast", trace.WithAttributes(attribute.Key("GetWeatherForecast").String("returning_your_city_weather")))
	defer span.End()

	span.SetAttributes(attribute.Key("GetWeatherForecast").String("inside server GetWeatherForecast"))

	url := fmt.Sprintf("http://%s/%s/%s", owmHost, requestPath, requestParam)

	req, _ := http.NewRequest("GET", url, nil)

	resp, err := libhttp.Do(spanCtx, req, tracer)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rwf, err := parseResponse(spanCtx, body, tracer)
	if err != nil {
		return nil, err
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestGetWeatherForecastsSuccessfull")
	return rwf, nil
}

func parseResponse(ctx context.Context, body []byte, tracer trace.Tracer) (*RequestWeatherForecast, error) {
	rwf := &RequestWeatherForecast{}

	spanCtx, span := tracer.Start(ctx, "call_parseResponse", trace.WithAttributes(attribute.Key("parseResponse").String("parse GetWeatherForecast Response")))
	defer span.End()

	err := json.Unmarshal(body, rwf)
	if err != nil {
		return nil, err
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestparseResponseSuccessfull")
	return rwf, nil

}
