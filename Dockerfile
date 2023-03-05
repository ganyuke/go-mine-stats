FROM alpine:edge AS build
RUN apk add --no-cache --update go gcc g++

WORKDIR /app

COPY server/go.mod ./
COPY server/go.sum ./

RUN go mod download

COPY server/src ./src

RUN CGO_ENABLED=1 GOOS=linux go build -o /go-mine-stats ./src

EXPOSE 3000

COPY example-config.json /data/config.json

CMD [ "/go-mine-stats", "-config", "/data/config.json", "-dbpath", "/data/stats.db" ]