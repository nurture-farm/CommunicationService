FROM golang:1.14.2-alpine3.11


RUN apk update && apk --no-cache add ca-certificates

RUN apk update && apk add tzdata
RUN cp /usr/share/zoneinfo/Asia/Kolkata /etc/localtime
RUN echo "Asia/Kolkata" > /etc/timezone

RUN apk update && apk add -f git librdkafka-dev pkgconf build-base
RUN mkdir -p /core/communication_service/
ADD . /core/communication_service/

ENV SERVICE=communication_service
ENV NAMESPACE=core
ENV CONFIG_DIR=/core/communication_service/config
WORKDIR /core/communication_service/

RUN go build -o main .

EXPOSE 8020
EXPOSE 8025
CMD ["/core/communication_service/main"]
