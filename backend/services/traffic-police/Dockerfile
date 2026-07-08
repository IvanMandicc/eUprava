# Faza 1: build
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/service ./cmd

# Faza 2: minimalna runtime slika
FROM alpine:3.20
COPY --from=build /bin/service /bin/service
EXPOSE 8082
ENTRYPOINT ["/bin/service"]
