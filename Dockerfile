FROM golang:1.15-alpine

RUN apk add --virtual .build-dependencies git  
WORKDIR /$GOPATH/src/github.com/vennekilde/gw2verify
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
RUN apk del .build-dependencies  
RUN adduser -S -D -H -h . gw2verify
USER gw2verify
EXPOSE 5000/tcp
CMD ["gw2verify"] 