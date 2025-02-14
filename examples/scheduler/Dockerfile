# Build the application from source
FROM golang:1.23 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

WORKDIR /app/examples/scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -o /scheduler


# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /scheduler /scheduler

USER nonroot:nonroot

EXPOSE 8080
EXPOSE 9464

ENTRYPOINT ["/scheduler"]
