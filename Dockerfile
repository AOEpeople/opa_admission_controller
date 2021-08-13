FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates tzdata git && update-ca-certificates
COPY ./ /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -o opa-admission


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/opa-admission /opa-admission

EXPOSE 8443

CMD ["/opa-admission"]
