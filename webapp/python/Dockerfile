FROM python:3.12-bookworm

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y curl wget gcc g++ make sqlite3 locales locales-all && \
    wget -q https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb && \
    apt-get -y install ./mysql-apt-config_0.8.29-1_all.deb && \
    apt-get -y update && \
    apt-get -y install default-mysql-client pdns-server pdns-backend-mysql && \
    pip install pipenv
RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/python
WORKDIR /home/isucon/webapp/python
COPY --chown=isucon:isucon Pipfile /home/isucon/webapp/python/
COPY --chown=isucon:isucon Pipfile.lock /home/isucon/webapp/python/
RUN pipenv install
COPY --chown=isucon:isucon models.py /home/isucon/webapp/python/
COPY --chown=isucon:isucon app.py /home/isucon/webapp/python/

# ENV GOPATH=/home/isucon/tmp/go
# ENV GOCACHE=/home/isucon/tmp/go/.cache

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

EXPOSE 8080
CMD ["pipenv", "run", "python", "app.py"]
