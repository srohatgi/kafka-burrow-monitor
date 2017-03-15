FROM magnetme/burrow:latest

MAINTAINER sumeet rohatgi

ADD . $GOPATH/src/github.com/srohatgi/kafka-burrow-monitor
RUN cd $GOPATH/src/github.com/srohatgi/kafka-burrow-monitor &&\
  go install &&\
  mv $GOPATH/bin/kafka-burrow-monitor /go/bin/kafka-burrow-monitor

RUN go get -v github.com/srohatgi/mq-bootstrap

ENV PATH=$PATH:$GOPATH/bin

ENV PROGRAM_NAME=/go/bin/kafka-burrow-monitor

CMD ["mq-bootstrap"]
