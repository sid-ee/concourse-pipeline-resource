# builder image
# ============================================================================
FROM golang:alpine as builder
ENV CGO_ENABLED 0

RUN mkdir -p /assets/ /app/

ARG CONCOURSE_VERSION=6.5.1

# install fly-cli
ARG FLY_DOWNLOAD_URL="https://github.com/concourse/concourse/releases/download/v${CONCOURSE_VERSION}/fly-${CONCOURSE_VERSION}-linux-amd64.tgz"
ADD ${FLY_DOWNLOAD_URL} /assets/

COPY ./go.mod ./go.sum /app/

WORKDIR /app/

RUN go mod download

COPY ./ /app/

RUN go build -o /assets/in ./cmd/in \
	&& go build -o /assets/out ./cmd/out \
	&& go build -o /assets/check ./cmd/check \
	&& build_timestamp=$(date +%s) \
	&& set -e; for pkg in $(go list ./... | grep -v "acceptance"); do \
		go test -o "/tests/$(basename $pkg).${build_timestamp}.test" -c $pkg; \
	done

# runtime image
# ============================================================================
FROM alpine:edge AS resource
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder assets/ /opt/resource/
RUN cd /opt/resource/ \
    && tar xzvf /opt/resource/fly*.tgz \
    && rm /opt/resource/fly*.tgz
RUN chmod +x /opt/resource/*
# test image
# ============================================================================
FROM resource AS tests
COPY --from=builder /tests /go-tests
RUN set -e; for test in /go-tests/*.test; do \
		$test; \
	done
RUN /opt/resource/fly --version

# export runtime image
# ============================================================================
FROM resource
