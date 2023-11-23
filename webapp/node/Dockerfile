FROM node:20.9.0-bullseye-slim

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

RUN mkdir -p /home/isucon/webapp/node
WORKDIR /home/isucon/webapp/node
COPY --chown=isucon:isucon ./package.json ./package-lock.json /home/isucon/webapp/node/
RUN npm install

COPY --chown=isucon:isucon ./ /home/isucon/webapp/node/

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

EXPOSE 8080
CMD ["npm", "run", "start"]
