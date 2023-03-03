FROM golang:1.19

ENV DATA_DIR /data

RUN mkdir -p ${DATA_DIR}
WORKDIR ${DATA_DIR}

COPY . .
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

## Add the wait script to the image
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait /wait
RUN chmod +x /wait

CMD /wait && /go/bin/goose -dir=migrations up
#ENTRYPOINT ["/go/bin/goose", "-dir=migrations", "up"]
