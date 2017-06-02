FROM golang:alpine

RUN apk add --no-cache git
RUN go get github.com/Masterminds/glide
WORKDIR /go/src/github.com/studiously/classsvc

ADD ./glide.yaml ./glide.yaml
ADD ./glide.lock ./glide.lock
RUN glide install --skip-test -v

ADD . .
RUN go install .

ENTRYPOINT /go/bin/classsvc host
EXPOSE 8080:8080, 8081:8081