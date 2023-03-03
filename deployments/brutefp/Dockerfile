# Environment
FROM golang:1.19

ENV BIN_FILE /app/brutefp
ENV CODE_DIR /app

RUN mkdir -p ${CODE_DIR}
WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}
# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -a -o ${BIN_FILE} cmd/brutefp/*

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="brutefp"
LABEL MAINTAINERS="student@otus.ru"

ARG CONFIG_FILE
ENV CONFIG_FILE_LOCAL /etc/brutefp/config.json
COPY ${CONFIG_FILE} ${CONFIG_FILE_LOCAL}

EXPOSE 8088

## Add the wait script to the image
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait /wait
RUN chmod +x /wait

CMD /wait && ${BIN_FILE} -config ${CONFIG_FILE_LOCAL}
