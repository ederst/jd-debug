FROM golang:1.19 AS builder

WORKDIR /work

COPY . /work

RUN apt-get update \
    && apt-get install -y --no-install-recommends libsystemd-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN go build

# when using focal image on, for example a "jammy" host, then it will probably not work
# due to the different systemd versions (v249 vs v245 or so)
# FROM ubuntu:focal
FROM ubuntu:jammy

COPY --from=builder /work/jd-debug /jd-debug

ENTRYPOINT [ "/jd-debug" ]

