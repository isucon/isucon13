FROM ruby:3.2.2-bookworm

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
  apt-get install -y lsb-release locales locales-all && \
  curl -sSfLo /tmp/mysql-apt-config.deb https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb && \
  apt-get install -y /tmp/mysql-apt-config.deb && \
  apt-get update && \
  apt-get install -y default-mysql-client pdns-server pdns-backend-mysql

RUN rm -f /etc/powerdns/pdns.d/bind.conf

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/ruby
WORKDIR /home/isucon/webapp/ruby

COPY --chown=isucon:isucon ./Gemfile ./Gemfile.lock ./
ENV BUNDLE_DEPLOYMENT=1 BUNDLE_JOBS=8
RUN bundle install
COPY ./app.rb ./config.ru ./

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENV TZ utc

ENV RUBY_YJIT_ENABLE=1

EXPOSE 8080
CMD ["bundle", "exec", "puma", "--bind", "tcp://0.0.0.0:8080", "--workers", "8", "--threads", "0:8", "--environment", "production"]
