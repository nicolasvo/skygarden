FROM golang:alpine

WORKDIR /go/src/skygarden
COPY . /go/src/skygarden

RUN go get ./...
RUN apk add --update nodejs npm
RUN npm install -g serverless

ENTRYPOINT ["/bin/sh"]
