FROM rust:1.73.0-bookworm

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
  apt-get install -y lsb-release locales locales-all && \
  curl -sSfLo /tmp/mysql-apt-config.deb https://dev.mysql.com/get/mysql-apt-config_0.8.22-1_all.deb && \
  apt-get install -y /tmp/mysql-apt-config.deb && \
  apt-get update && \
  apt-get install -y default-mysql-client pdns-server pdns-backend-mysql

RUN rm -f /etc/powerdns/pdns.d/bind.conf

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/rust
WORKDIR /home/isucon/webapp/rust
COPY --chown=isucon:isucon ./Cargo.toml ./Cargo.lock ./
RUN mkdir src && echo 'fn main() {}' > ./src/main.rs && cargo build --release --locked && rm src/main.rs target/release/deps/isupipe-*
COPY --chown=isucon:isucon ./src/ ./src/
RUN cargo build --release --locked --frozen

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENV TZ utc

EXPOSE 8080
CMD ["/home/isucon/webapp/rust/target/release/isupipe"]
