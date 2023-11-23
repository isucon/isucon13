FROM golang:1.21.0-bookworm

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y gcc g++ make locales locales-all

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home benchuser
USER benchuser

RUN mkdir -p /home/benchuser
WORKDIR /home/benchuser
COPY --chown=benchuser:benchuser ./ /home/benchuser/
RUN go build -o bench ./cmd/bench

ENV GOPATH=/home/benchuser/tmp/go
ENV GOCACHE=/home/benchuser/tmp/go/.cache

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENTRYPOINT ["/home/benchuser/bench"]
CMD ["supervise"]
