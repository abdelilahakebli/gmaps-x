FROM golang:1.21.5

WORKDIR /app 

COPY ./ ./

RUN go mod download
RUN make build

EXPOSE 3030

CMD ["./bin/api", '-headless']