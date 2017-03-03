FROM magnetme/burrow:latest

MAINTAINER sumeet rohatgi

ADD . $GOPATH/src/github.com/srohatgi/kafka-burrow-monitor
RUN cd $GOPATH/src/github.com/srohatgi/kafka-burrow-monitor &&\
  go install &&\
  mv $GOPATH/bin/kafka-burrow-monitor /go/bin/kafka-burrow-monitor

CMD ["/go/bin/kafka-burrow-monitor"]
