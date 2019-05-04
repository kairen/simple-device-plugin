FROM kairen/golang-dep:1.11-alpine AS build-env
LABEL maintainer="Kyle Bai <k2r2.bai@gmail.com>"

ENV GOPATH "/go"
ENV PROJECT_PATH "$GOPATH/src/github.com/kubedev/simple-device-plugin"

COPY . $PROJECT_PATH
RUN cd $PROJECT_PATH && \
  dep ensure && IN_DOCKER=1 make && \
  mv out/device-plugin /tmp/device-plugin

# Run stage
FROM alpine:3.7

COPY --from=build-env /tmp/device-plugin /bin/device-plugin
RUN apk add --no-cache util-linux bash && \
  rm -f /var/cache/apk/*

ENTRYPOINT ["/bin/device-plugin"]
