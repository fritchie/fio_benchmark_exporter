FROM golang:1.16
MAINTAINER Frank R <12985912+fritchie@users.noreply.github.com>

RUN apt-get update
RUN apt-get -y install fio

WORKDIR /go/src/fio_benchmark_exporter
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9996

CMD [ "fio_benchmark_exporter" ]
