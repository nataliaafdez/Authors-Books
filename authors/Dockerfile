# syntax=docker/dockerfile:1


FROM golang:1.23 AS build-authors-http
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/authors-http ./authors/cmd/api


FROM golang:1.23 AS build-authors-grpc
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/authors-grpc ./authors/cmd/api/grpcmain


FROM gcr.io/distroless/base-debian12:nonroot AS authors-http
ENV ADDR=":8080" \
    CONSUL_ADDR="consul:8500" \
    SERVICE_HOST="authors" \
    REGISTER_ADDRESS="authors" \
    HEALTH_HOST="authors"
COPY --from=build-authors-http /out/authors-http /authors
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/authors"]

FROM gcr.io/distroless/base-debian12:nonroot AS authors-grpc
ENV ADDR_GRPC=":9090" \
    BOOKS_URL="http://books:8081" \
    METADATA_URL="http://metadata:8082" \
    CONSUL_ADDR="consul:8500" \
    SERVICE_HOST="authors-grpc"
COPY --from=build-authors-grpc /out/authors-grpc /authors-grpc
EXPOSE 9090
USER nonroot:nonroot
ENTRYPOINT ["/authors-grpc"]
