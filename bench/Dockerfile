FROM golang:1.21.0-bookworm

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y gcc g++ make locales locales-all

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home bench
USER bench

RUN mkdir -p /home/bench/bench
WORKDIR /home/bench/bench
COPY --chown=bench:bench ./ /home/bench/bench/
RUN go build -o bench ./cmd/bench

ENV GOPATH=/home/bench/tmp/go
ENV GOCACHE=/home/bench/tmp/go/.cache

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENTRYPOINT ["/home/bench/bench/bench"]
CMD ["supervise", "--production"]
