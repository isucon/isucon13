FROM perl:5.38.0-bookworm

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y curl wget gcc g++ make sqlite3 locales locales-all && \
    wget -q https://dev.mysql.com/get/mysql-apt-config_0.8.22-1_all.deb && \
    apt-get -y install ./mysql-apt-config_0.8.22-1_all.deb && \
    apt-get -y update && \
    apt-get -y install default-mysql-client pdns-server pdns-backend-mysql

RUN rm -f /etc/powerdns/pdns.d/bind.conf

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/perl
WORKDIR /home/isucon/webapp/perl

COPY cpanfile ./
RUN cpm install --show-build-log-on-failure

COPY --chown=isucon:isucon ./ /home/isucon/webapp/perl/
ENV PERL5LIB=/home/isucon/webapp/perl/local/lib/perl5
ENV PATH=/home/isucon/webapp/perl/local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENV TZ utc

EXPOSE 8080
CMD ["./local/bin/plackup", "-s", "Starlet", "-p", "8080", "-Ilib", "app.psgi"]
