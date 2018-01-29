# 使用1.9的golang作为母镜像
FROM golang:1.9

# 维护者信息
MAINTAINER eyeblue "eyebluecn@126.com"

# 指定工作目录就是 tank
WORKDIR $GOPATH/src/tank

# 将tank项目下的所有文件移动到golang镜像中去
COPY . $GOPATH/src/tank

# 日志和上传文件的目录
VOLUME /data/log
VOLUME /data/matter
ENV TANK_LOG_PATH=/data/log TANK_MATTER_PATH=/data/matter


# 开始下载依赖库并且进行编译
RUN git clone https://github.com/eyebluecn/golang.org.git $GOPATH/src/golang.org \
    && go get github.com/disintegration/imaging \
    && go get github.com/json-iterator/go \
    && go get github.com/go-sql-driver/mysql \
    && go get github.com/jinzhu/gorm \
    && go get github.com/nu7hatch/gouuid \
    && go install tank \
    && cp -r $GOPATH/src/tank/build/* $GOPATH/bin

# 暴露6010端口
EXPOSE 6010

# tank作为执行文件
ENTRYPOINT ["/go/bin/tank"]
