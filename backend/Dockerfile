FROM golang:1.25-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN apk add tzdata

COPY --from=build /bin/api /app/api

EXPOSE 5000
CMD ["/app/api"]
