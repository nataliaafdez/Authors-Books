# syntax=docker/dockerfile:1

###############
# build-http
###############
FROM golang:1.23 AS build-books-http
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Compila SOLO el HTTP. Si tu main http está roto, este stage fallará (está bien).
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/books-http ./books/cmd/api

###############
# build-grpc
###############
FROM golang:1.23 AS build-books-grpc
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Compila SOLO el gRPC. Este stage no depende del http.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/books-grpc ./books/cmd/api/grpcmain

###############
# runtime HTTP
###############
FROM gcr.io/distroless/base-debian12:nonroot AS books-http
ENV ADDR=":8081" \
    CONSUL_ADDR="consul:8500" \
    SERVICE_HOST="books" \
    REGISTER_ADDRESS="books" \
    HEALTH_HOST="books"
COPY --from=build-books-http /out/books-http /books
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/books"]

###############
# runtime gRPC
###############
FROM gcr.io/distroless/base-debian12:nonroot AS books-grpc
ENV ADDR_GRPC=":9091" \
    CONSUL_ADDR="consul:8500" \
    SERVICE_HOST="books-grpc"
COPY --from=build-books-grpc /out/books-grpc /books-grpc
EXPOSE 9091
USER nonroot:nonroot
ENTRYPOINT ["/books-grpc"]
