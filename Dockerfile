FROM golang:1.13

WORKDIR /app
COPY . /app
RUN make build-ethereum-sender

FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/bin/* /bin/

CMD ["/bin/app"]
