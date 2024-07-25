FROM golang:alpine

ARG USER=bumblebee

WORKDIR /bumblebee
RUN apk --no-cache update && apk add --no-cache git sudo
RUN adduser --disabled-password $USER && \
    chown -R $USER:$USER /bumblebee && \
    chmod -R 755 /bumblebee
RUN mkdir -p /etc/sudoers.d \
    && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
    && chmod 0440 /etc/sudoers.d/$USER
USER $USER


# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./go.mod ./
RUN go mod download && go mod verify
RUN go install -v golang.org/x/tools/gopls@latest
RUN go install github.com/spf13/cobra-cli@latest

# ENV GOCACHE=/bumblebee/tmp/build_cache