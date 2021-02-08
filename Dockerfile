FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY cortex-alert-operator /
USER nonroot:nonroot

ENTRYPOINT ["/cortex-alert-operator"]
