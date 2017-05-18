FROM golang:alpine
ADD . .
RUN go install .
ENTRYPOINT /go/bin/usersvc
EXPOSE 8080:8080