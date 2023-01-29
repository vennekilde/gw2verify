# Build args
ARG MODULE_NAME=gw2verify

FROM gcr.io/distroless/static AS final
# Build args
ARG MODULE_NAME
# Copy our static executable
# Note: bin is renamed to "app" as we cannot use build args or env vars in ENTRYPOINT when using scratch image
COPY bin/${MODULE_NAME} /go/bin/app
# Use an unprivileged user.
USER nonroot:nonroot

EXPOSE 5000/tcp
WORKDIR /go/bin/
# Port on which the service will be exposed.
# EXPOSE 5005
# Run the hello binary.
ENTRYPOINT ["/go/bin/app", "-logtostderr=true"] 