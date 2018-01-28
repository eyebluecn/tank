# 使用最新的golang作为母镜像
FROM golang:latest

# 维护者信息
MAINTAINER eyeblue "eyebluecn@126.com"

WORKDIR $GOPATH/src/hellodocker
ADD . $GOPATH/src/hellodocker
RUN go build .


EXPOSE 8080

ENTRYPOINT ["./hellodocker"]
