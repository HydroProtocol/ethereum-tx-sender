FROM golang:1.12

WORKDIR /app
COPY . /app
RUN go mod download
RUN go build -ldflags "-w -linkmode external -extldflags -static" -v -o build/app

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/build/app /bin/

CMD ["/bin/app"]
