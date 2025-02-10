FROM gcr.io/distroless/static:latest
WORKDIR /
COPY grafana-dups grafana-dups

ENTRYPOINT ["/grafana-dups"]
