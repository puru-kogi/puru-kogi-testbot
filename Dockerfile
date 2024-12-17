FROM golang:1.23.4 AS go
WORKDIR /app
COPY go.mod go.sum main.go ./
RUN go mod download \
&& go build -o main /app/main.go
CMD [ "/app/main" ]
