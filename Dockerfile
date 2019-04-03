FROM ubuntu:18.04
LABEL maintainer="Michail Apakhov"

RUN apt-get update

RUN apt-get install golang-go -y
RUN go get ""