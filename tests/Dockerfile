FROM golang:1.19 as build

ENV CODE_DIR /app

RUN mkdir -p ${CODE_DIR}
WORKDIR ${CODE_DIR}
COPY . ${CODE_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test ./tests/
