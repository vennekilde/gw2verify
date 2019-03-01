FROM golang:1.12.0-alpine3.9

RUN apk add --virtual .build-dependencies git  
WORKDIR /$GOPATH/src/github.com/vennekilde/gw2verify
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
RUN apk del .build-dependencies  
RUN adduser -S -D -H -h . gw2verify
USER gw2verify
CMD ["gw2verify"] 