FROM golang
ADD . /go/src/github.com/studiously/usersvc
RUN go install github.com/studiously/usersvc
ENTRYPOINT /go/bin/usersvc
EXPOSE 8080