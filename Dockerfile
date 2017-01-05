FROM hub.noths.com/alpine:3.4

USER user

COPY autoscaler /service/

ENTRYPOINT ["/bin/dumb-init"]
CMD ["/service/autoscaler"]
