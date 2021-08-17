/*
Origin : https://github.com/ramsgoli/Golang-OpenWeatherMap
Hardcoded unit from imperial to metric
Added OpenTelemetry Instrumentation by tonny@segmentationfault.xyz
*/

package owm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	libhttp "weather/lib/http"
)

/*
Define API response fields
*/
type OpenWeatherMap struct {
	APIKEY string
}

/*
Return response fields of City data
*/
type City struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

/*
Geographic spherical coordinate for input
*/
type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

/*
Return weather forecast
*/
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

/*
Return wind condition
*/
type Wind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

/*
Return cloud condition
*/
type Clouds struct {
	All int `json:"all"`
}

/*
Return rain condition
*/
type Rain struct {
	Threehr int `json:"3h"`
}

/*
Return temperature data
*/
type Main struct {
	Temp     float64 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	TempMin  float64 `json:"temp_min"`
	TempMax  float64 `json:"temp_max"`
}

/*
Define API response objects (compose of the above fields)
*/
type CurrentWeatherResponse struct {
	Coord   `json:"coord"`
	Weather []Weather `json:"weather"`
	Main    `json:"main"`
	Wind    `json:"wind"`
	Rain    `json:"rain"`
	Clouds  `json:"clouds"`
	DT      int    `json:"dt"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
}

/*
Response from openweathermap
*/
type ForecastResponse struct {
	City    `json:"city"`
	Coord   `json:"coord"`
	Country string `json:"country"`
	List    []struct {
		DT      int `json:"dt"`
		Main    `json:"main"`
		Weather `json:"weather"`
		Clouds  `json:"clouds"`
		Wind    `json:"wind"`
	} `json:"list"`
}

/*
openweathermap endpoint
*/
const (
	APIURL string = "api.openweathermap.org"
)

/*
Build request to openweathermap
*/
func makeAPIRequest(ctx context.Context, url string, tracer trace.Tracer) ([]byte, error) {
	spanCtx, span := tracer.Start(ctx, "call_owm_makeAPIRequest", trace.WithAttributes(attribute.Key("owm_MakeAPIRequest").String("call_owm_data_source")))
	defer span.End()

	req, getErr := http.NewRequest("GET", url, nil)
	if getErr != nil {
		return nil, getErr
	}

	res, err := libhttp.Do(spanCtx, req, tracer)
	if err != nil {
		return nil, err
	}

	// defer the closing of the res body
	defer res.Body.Close()

	// read the http response body into a byte stream
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "makeAPIRequestSuccessfull")
	return body, nil
}

/*
Get current weather from a city - openweathermap
*/
func (owm *OpenWeatherMap) CurrentWeatherFromCity(ctx context.Context, city string, tracer trace.Tracer) (*CurrentWeatherResponse, error) {
	spanCtx, span := tracer.Start(ctx, "call_owm_CurrentWeatherFromCity", trace.WithAttributes(attribute.Key("owm_CurrentWeatherFromCity").String("returning_weather_based_on_city")))
	defer span.End()

	span.SetAttributes(attribute.Key("CurrentWeatherFromCity").String("inside owm_CurrentWeatherFromCity"))

	if owm.APIKEY == "" {
		// No API keys present, return error
		return nil, errors.New("no api keys present")
	}
	url := fmt.Sprintf("http://%s/data/2.5/weather?q=%s&units=metric&APPID=%s", APIURL, city, owm.APIKEY)

	body, err := makeAPIRequest(spanCtx, url, tracer)
	if err != nil {
		return nil, err
	}
	var cwr CurrentWeatherResponse

	// unmarshal the byte stream into a Go data type
	jsonErr := json.Unmarshal(body, &cwr)
	if jsonErr != nil {
		return nil, jsonErr
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestCurrentWeatherFromCitySuccessfull")
	return &cwr, nil
}

/*
Get current weather from a coordinate - openweathermap
*/
func (owm *OpenWeatherMap) CurrentWeatherFromCoordinates(ctx context.Context, lat, long float64, tracer trace.Tracer) (*CurrentWeatherResponse, error) {
	spanCtx, span := tracer.Start(ctx, "call_owm_CurrentWeatherFromCoordinates", trace.WithAttributes(attribute.Key("owm_CurrentWeatherFromCoordinates").String("returning_weather_based_on_coordinates")))
	defer span.End()

	span.SetAttributes(attribute.Key("CurrentWeatherFromCoordinates").String("inside owm_CurrentWeatherFromCoordinates"))

	if owm.APIKEY == "" {
		// No API keys present, return error
		return nil, errors.New("no api keys present")
	}

	url := fmt.Sprintf("http://%s/data/2.5/weather?lat=%f&lon=%f&units=metric&APPID=%s", APIURL, lat, long, owm.APIKEY)

	body, err := makeAPIRequest(spanCtx, url, tracer)
	if err != nil {
		return nil, err
	}

	var cwr CurrentWeatherResponse

	// unmarshal the byte stream into a Go data type
	jsonErr := json.Unmarshal(body, &cwr)
	if jsonErr != nil {
		return nil, jsonErr
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestCurrentWeatherFromCoordinatesSuccessfull")
	return &cwr, nil
}

/*
Return current weather from a zip code - openweathermap
*/
func (owm *OpenWeatherMap) CurrentWeatherFromZip(ctx context.Context, zip int, tracer trace.Tracer) (*CurrentWeatherResponse, error) {
	spanCtx, span := tracer.Start(ctx, "call_owm_CurrentWeatherFromZip", trace.WithAttributes(attribute.Key("owm_CurrentWeatherFromZip").String("returning_weather_based_on_zipcode")))
	defer span.End()

	if owm.APIKEY == "" {
		// No API keys present, return error
		return nil, errors.New("no api keys present")
	}
	url := fmt.Sprintf("http://%s/data/2.5/weather?zip=%d&units=metric&APPID=%s", APIURL, zip, owm.APIKEY)

	body, err := makeAPIRequest(ctx, url, tracer)
	if err != nil {
		return nil, err
	}
	var cwr CurrentWeatherResponse

	// unmarshal the byte stream into a Go data type
	jsonErr := json.Unmarshal(body, &cwr)
	if jsonErr != nil {
		return nil, jsonErr
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestCurrentWeatherFromZipSuccessful")
	return &cwr, nil
}

/*
Return current weather from a city id - openweathermap
*/
func (owm *OpenWeatherMap) CurrentWeatherFromCityId(ctx context.Context, id int, tracer trace.Tracer) (*CurrentWeatherResponse, error) {
	spanCtx, span := tracer.Start(ctx, "call_owm_CurrentWeatherFromCityID", trace.WithAttributes(attribute.Key("owm_CurrentWeatherFromCityID").String("returning_weather_based_on_city_id")))
	defer span.End()

	if owm.APIKEY == "" {
		// No API keys present, return error
		return nil, errors.New("no api keys present")
	}
	url := fmt.Sprintf("http://%s/data/2.5/weather?id=%d&units=metric&APPID=%s", APIURL, id, owm.APIKEY)

	body, err := makeAPIRequest(spanCtx, url, tracer)
	if err != nil {
		return nil, err
	}
	var cwr CurrentWeatherResponse

	// unmarshal the byte stream into a Go data type
	jsonErr := json.Unmarshal(body, &cwr)
	if jsonErr != nil {
		return nil, jsonErr
	}

	trace.WithSpanKind(trace.SpanKindInternal)
	trace.SpanFromContext(spanCtx).SetStatus(codes.Ok, "requestCurrentWeatherFromCityIDSuccessful")
	return &cwr, nil
}
