# syntax=docker/dockerfile:1

################
# build-http
################
FROM golang:1.23 AS build-metadata-http
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/metadata-http ./metadata/cmd/api

################
# build-grpc
################
FROM golang:1.23 AS build-metadata-grpc
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/metadata-grpc ./metadata/cmd/api/grpcmain

################
# runtime HTTP
################
FROM gcr.io/distroless/base-debian12:nonroot AS metadata-http
ENV PORT="8082" \
    CONSUL_ADDR="consul:8500" \
    SERVICE_HOST="metadata"
COPY --from=build-metadata-http /out/metadata-http /metadata
EXPOSE 8082
USER nonroot:nonroot
ENTRYPOINT ["/metadata"]

################
# runtime gRPC
################
FROM gcr.io/distroless/base-debian12:nonroot AS metadata-grpc
ENV ADDR_GRPC=":9092" \
    CONSUL_ADDR="consul:8500" \
    SERVICE_HOST="metadata-grpc"
COPY --from=build-metadata-grpc /out/metadata-grpc /metadata-grpc
EXPOSE 9092
USER nonroot:nonroot
ENTRYPOINT ["/metadata-grpc"]
