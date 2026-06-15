FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/donation-api ./cmd/server

FROM alpine:3.22
RUN adduser -D -H app && mkdir -p /data/uploads && chown -R app:app /data
USER app
COPY --from=build /out/donation-api /usr/local/bin/donation-api
EXPOSE 8080
CMD ["donation-api"]
