# 使用1.8的golang作为母镜像
FROM golang:1.8

# 维护者信息
MAINTAINER eyeblue "eyebluecn@126.com"

# 指定工作目录就是 tank
WORKDIR $GOPATH/src/tank

# 将tank项目下的所有文件移动到golang镜像中去
COPY . $GOPATH/src/tank

# 开始编译
RUN ./build/pack/build.sh

# 暴露6010端口
EXPOSE 6010

# tank作为执行文件
ENTRYPOINT ["./dist/tank-1.0.2/tank"]
