FROM golang:alpine

EXPOSE 50551

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["authentication-service"]
