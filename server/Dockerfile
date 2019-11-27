FROM golang:1.12

LABEL maintainer="Yichao Wang<wangycbeyond@gmail.com>"

RUN go get -u github.com/go-sql-driver/mysql
WORKDIR $GOPATH/src/github.com/carpark/server

COPY . .

# Install the package
RUN go install -v ./...

EXPOSE 8080/tcp

# Run the executable
CMD ["server"]
