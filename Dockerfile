# build
FROM golang:1.22.5 as builder
WORKDIR /app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download && go mod verify

COPY ./cmd/gophermart ./cmd/gophermart
COPY ./internal ./internal
COPY ./pkg ./pkg
RUN cd cmd/gophermart && go build -o gophermart

FROM golang:1.22.5
ARG GOPHERMART_PORT=${GOPHERMART_PORT}
ARG DATABASE_URI=${DATABASE_URI}
ARG RUN_ADDRESS=${RUN_ADDRESS}
ARG ACCRUAL_SYSTEM_ADDRESS=${ACCRUAL_SYSTEM_ADDRESS}
WORKDIR /app
COPY --from=builder /app/cmd/gophermart/gophermart ./gophermart
RUN chmod +x gophermart
EXPOSE ${GOPHERMART_PORT}

ENTRYPOINT ["./gophermart"]
CMD ["-a", "${RUN_ADDRESS}", "-r", "${ACCRUAL_SYSTEM_ADDRESS}", "-d", "${DATABASE_URI}"]
