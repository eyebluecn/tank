# use golang 1.12
FROM golang:1.12

# maintainer. Author's name and email.
MAINTAINER eyeblue "eyebluecn@126.com"

# work directory.
WORKDIR /data

# Copy this project to /data
COPY . /data

# in order to make docker stateless. Prepare a volumn
VOLUME /data/build/matter

# proxy
ENV GOPROXY=https://athens.azurefd.net

# prepare the config file
RUN go build -mod=readonly \
    && cp -r /data/tank /data/build

# use 6010 as default.
EXPOSE 6010

# tank as execute file.
ENTRYPOINT ["/data/build/tank"]
