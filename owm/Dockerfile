FROM golang:latest as builder
LABEL maintainer="tonny@segmentationfault.xyz"

ENV GO111MODULE=on
ENV APP OWMService
ENV PORT 8082
ENV OWM_ADDR http://localhost:8082
ENV OWM_APP_ID testingabc123
ENV TRACER_ENDPOINT http://localhost:14268/api/traces

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY owm/main.go main.go
COPY lib/ lib

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/${APP} main.go

FROM alpine:latest
COPY --from=builder /out/${APP} /app/

EXPOSE ${PORT}
ENTRYPOINT ["/app/OWMService"]