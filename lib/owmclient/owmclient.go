package owmclient

import (
	"context"
	"os"
	openweathermap "weather/lib/owm"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type StrippedWeatherData struct {
	Condition   string
	Temperature float64
	Humidity    int
}

func GetOwmForecastByCity(ctx context.Context, city string, tracer trace.Tracer) (*StrippedWeatherData, error) {

	spanCtx, span := tracer.Start(ctx, "call_owmclient_GetOwmForecastByCity", trace.WithAttributes(attribute.Key("owmclient_GetOwmForecastByCity").String("get_owmforecast_by_city")))
	defer span.End()

	owm := openweathermap.OpenWeatherMap{APIKEY: os.Getenv("OWM_APP_ID")}
	currentWeather, err := owm.CurrentWeatherFromCity(spanCtx, city, tracer)

	if err != nil {
		return nil, err
	}

	swd := &StrippedWeatherData{
		Condition:   currentWeather.Weather[0].Main,
		Temperature: currentWeather.Main.Temp,
		Humidity:    currentWeather.Main.Humidity,
	}

	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "GetOwmForecastByCitySuccessfull")
	return swd, nil

}
