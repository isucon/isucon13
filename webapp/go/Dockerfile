FROM golang:1.21.0-bullseye

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y wget gcc g++ make sqlite3 && \
    wget -q https://dev.mysql.com/get/mysql-apt-config_0.8.22-1_all.deb && \
    apt-get -y install ./mysql-apt-config_0.8.22-1_all.deb && \
    apt-get -y update && \
    apt-get -y install mysql-client

RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/go
WORKDIR /home/isucon/webapp/go
COPY --chown=isucon:isucon ./ /home/isucon/webapp/go/
RUN go build -o isupipe .

ENV GOPATH=/home/isucon/tmp/go
ENV GOCACHE=/home/isucon/tmp/go/.cache

EXPOSE 12345
CMD ["/home/isucon/webapp/go/isupipe"]