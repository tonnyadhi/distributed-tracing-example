# Distributed Tracing Example

## Structure

   This example contains  two services 
   - HTTP Weather Service
   - HTTP OWM Service, fetching weather forecast from [openweathermap](https://openweathermap.org)

## Stack

   - Golang v1.16
   - [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go) v1.0.0-RC2
   - [Go Chi](https://github.com/go-chi/chi) v1.5.4

## Running

## Running Without Kubernetes

   - Get yourself an API Key from openweathermap
   - Run using docker-compose
     - `$docker-compose up`
   - Try to curl into Weather Service
     - `$curl localhost:8080/forecast/depok`
     - `$curl localhost:8080/ping`
   - Go to Jaeger All In UI at port 16686 for observing traces
     - `$firefox localhost:16686`
    
## Running Inside Kubernetes
   - You can use provided helm chart on k8s directory
   - Assumed that you already installed Istio on your k8s for convenience on receiving tracing via Jaeger
   - Modify provided helm chart base on your need
   - If you want to use the OpenTelemetry Collector, you can use and adjust `sentry-jaeger-pipeline.yaml` manifest. Make sure OpenTelemetry Operator installed on your kubernetes

