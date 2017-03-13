FROM alpine:3.5

RUN apk --no-cache add ca-certificates

COPY alice /service/

CMD ["/service/alice"]
