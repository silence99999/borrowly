FROM golang:1.24-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GONOSUMDB=* go mod tidy && go build -o /bin/service .

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache tzdata ca-certificates
COPY --from=build /bin/service /app/service
EXPOSE 5004
CMD ["/app/service"]
