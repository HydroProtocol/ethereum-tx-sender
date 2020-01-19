FROM golang:1.13

WORKDIR /app
COPY . /app
RUN make build-ethereum-sender

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/build/bin/ethereum-sender /usr/local/bin/

ENTRYPOINT ["ethereum-sender"]
