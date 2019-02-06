FROM golang:1.11

RUN go get github.com/globalsign/mgo
RUN go get github.com/gorilla/mux
WORKDIR /go/src/github.com/sashayakovtseva/bookshelf
COPY book.go .
COPY db.go .
COPY app/ app/
RUN go build --ldflags '-linkmode "external" -extldflags "-static"' -o shelf ./app

FROM scratch
COPY --from=0 /go/src/github.com/sashayakovtseva/bookshelf/shelf .
EXPOSE 8080
CMD ["./shelf"]
