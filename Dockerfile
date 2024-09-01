FROM golang:alpine

ARG USER=bumblebee
ENV GOOS=linux
ENV GOARCH=arm64

WORKDIR /bumblebee
RUN apk --no-cache update && apk add --no-cache git
RUN adduser --disabled-password $USER && \
    chown -R $USER:$USER /bumblebee && \
    chmod -R 755 /bumblebee

USER $USER

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./go.mod ./
RUN go mod download && go mod verify

CMD ["sh", "-c", "GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/bumblebee main.go" ]