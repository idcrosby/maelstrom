FROM golang

ADD . /go/src/github.com/idcrosby/maelstrom
RUN go install github.com/idcrosby/maelstrom
ENTRYPOINT /go/bin/maelstrom

EXPOSE 8123