FROM golang:1.21.5

WORKDIR /app 

COPY ./ ./

RUN go mod download
RUN go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4001.0 install --with-deps chromium
RUN make build

EXPOSE 3030

ENTRYPOINT [ "/app/bin/api", "-headless"]