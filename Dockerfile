FROM golang

ADD . /go/src/github.com/idcrosby/maelstrom
RUN go install github.com/idcrosby/maelstrom
RUN go get google.golang.org/cloud/compute/metadata
ENTRYPOINT /go/bin/maelstrom

EXPOSE 8123