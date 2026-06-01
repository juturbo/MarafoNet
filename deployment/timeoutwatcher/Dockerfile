# syntax=docker/dockerfile:1

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=1.25.7
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

# Download dependencies as a separate step to take advantage of Docker's caching.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# This is the architecture you're building for, which is passed in by the builder.
ARG TARGETARCH

# Build the application.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/timeout-watcher

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application.
FROM alpine:latest AS final

# Install any runtime dependencies that are needed to run your application.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

# Create a non-privileged user that the app will run under.
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

# Copy the executable from the "build" stage.
COPY --from=build /bin/server /bin/

# Run the application (no explicit CMD, uses default entrypoint)
ENTRYPOINT ["/bin/server"]
