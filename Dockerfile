FROM hub.noths.com/alpine:3.5

RUN apk --no-cache add ca-certificates

USER user

COPY autoscaler /service/

ENTRYPOINT ["/bin/dumb-init"]
CMD ["/service/autoscaler"]
