##########################################################
# Dockerfile which builds a simple Go restservicce
##########################################################
FROM golang

RUN apt-get update && \
	apt-get install -y pkg-config lxc && \
	apt-get install -y libexif-dev && \
	go get github.com/xiam/exif

COPY msortweb /bin/
COPY web/msortweb.d /bin/msortweb.d

EXPOSE 8081

WORKDIR /bin

CMD ./msortweb -config-file=./msortweb.d/appconfig.json