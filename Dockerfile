FROM golang:1.11-alpine AS onionize-build
RUN apk add --no-cache git
WORKDIR /go/src/github.com/nogoegst/onionize
COPY . .
RUN CGO_ENABLED=0 go install -v github.com/nogoegst/onionize/cmd/onionize
FROM alpine
RUN apk add --no-cache tor ca-certificates
COPY --from=onionize-build /go/bin/onionize /usr/local/bin/
ENTRYPOINT ["onionize", "-start-tor"]
CMD ["-h"]
