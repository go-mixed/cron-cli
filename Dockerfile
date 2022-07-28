FROM golang:1.18 as builder

COPY . /home/src

RUN /bin/sh -c set -eux \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && cd /home/src \
    && go build

FROM ubuntu:20.04

ENV CRON_VERSION v3.0.1
ENV PATH /cron/:${PATH}

RUN /bin/sh -c set -eux \
    && mkdir -p /cron

COPY --from=builder /home/src/cron /usr/bin/cron

WORKDIR /cron

ENTRYPOINT ["cron"]

CMD ["-h"]