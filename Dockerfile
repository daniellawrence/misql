FROM golang:1.26.1-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /misql .

FROM scratch
COPY --from=builder /misql /misql

# Standard MySQL env vars
ENV MYSQL_HOST=0.0.0.0
ENV MYSQL_TCP_PORT=3306
ENV MYSQL_DATA_DIR=/data

EXPOSE 3306

VOLUME ["/data"]

ENTRYPOINT ["/misql"]
