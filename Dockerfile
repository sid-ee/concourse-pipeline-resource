FROM gliderlabs/alpine:3.2

RUN apk --update add \
  ca-certificates \
  jq

ADD cmd/check/check /opt/resource/check
ADD cmd/in/in /opt/resource/in
ADD cmd/out/out /opt/resource/out
ADD fly-bin /opt/resource/fly

RUN chmod +x /opt/resource/*
