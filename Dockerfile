FROM golang:1.12.5

ENV GOPATH /app
RUN mkdir /app 
RUN go get github.com/gin-gonic/gin
RUN go get github.com/globalsign/mgo
RUN go get github.com/streamrail/concurrent-map
RUN go get github.com/DeanThompson/ginpprof
RUN go get github.com/satori/go.uuid
RUN go get github.com/zsais/go-gin-prometheus

WORKDIR /app
ADD ./ /app 
ADD ./entrypoint.sh /

ENV GIN_MODE release

RUN go build -o acmeair .
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/acmeair"]
