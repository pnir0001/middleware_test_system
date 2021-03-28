ARG PACKAGE_NAME=github.com/pnir0001/middleware_test_system

FROM golang:1.10.2-alpine3.7
ARG PACKAGE_NAME

ADD ./mongo_subscriber /go/src/$PACKAGE_NAME/mongo_subscriber

WORKDIR /go/src/$PACKAGE_NAME/mongo_subscriber

RUN apk update \
  && apk add --no-cache git \
  && go get -u github.com/golang/dep/cmd/dep \
  && dep ensure -v -vendor-only

RUN go build -o /go/bin/main main.go
    
CMD ["/go/bin/main"]