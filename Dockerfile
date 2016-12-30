FROM hub.noths.com/alpine:3.4

USER user

COPY autoscaler /service/

COPY config.yaml.dist /service/config.yml

ENTRYPOINT ["/bin/dumb-init"]
CMD ["/service/autoscaler"]