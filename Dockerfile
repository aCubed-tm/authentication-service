# based on https://github.com/neo4j/neo4j-go-driver/issues/56
# alpine 3.11 doesn't seem to work
FROM golang:1.13-alpine3.10

# add our cgo dependencies
RUN apk add --update --no-cache ca-certificates cmake make g++ openssl-dev git curl pkgconfig
RUN git clone -b 1.7 https://github.com/neo4j-drivers/seabolt.git /seabolt
# invoke cmake build and install artifacts - default location is /usr/local
WORKDIR /seabolt/build
# CMAKE_INSTALL_LIBDIR=lib is a hack where we override default lib64 to lib to workaround a defect
# in our generated pkg-config file 
RUN cmake -D CMAKE_BUILD_TYPE=Release -D CMAKE_INSTALL_LIBDIR=lib .. && cmake --build . --target install

# install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# continue normal dockerfile
EXPOSE 50551

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["authentication-service"]
