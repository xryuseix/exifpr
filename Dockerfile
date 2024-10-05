FROM golang:1.23

RUN apt-get update && \
    apt-get install -y git libimage-exiftool-perl curl

WORKDIR /work

COPY . /work

RUN go build .

ENTRYPOINT ["/work/exifpr"]