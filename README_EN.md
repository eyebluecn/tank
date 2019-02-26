[![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/logo.png)](https://github.com/eyebluecn/tank)

[中文版](./README_EN.md)

# Eyeblue Cloud Disk（2.0.0）

[Online Demo](https://tank.eyeblue.cn) (demo@tank.eyeblue.cn 123456)

![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [tank backend](https://github.com/eyebluecn/tank)

![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [tank-front frontend](https://github.com/eyebluecn/tank-front)

### Description
Eyeblue Cloud Disk is one of [Eyeblue Softwares](https://github.com/eyebluecn), and it is the basic software for other eyeblue softwares.Eyeblue Cloud Disk has the following features.

**1. Easy to use**

- Docker available

- [Install package](https://github.com/eyebluecn/tank/releases) is about 10M. Cross platform.

- Unpack the install zip, and run it with one command.

- Installation Guide

- Dashboard to view UV and PV.

- Preview usual files (doc,ppt,xls,pdf,mp3,mp4,jpg,png)

**2. Core features all in readiness**

- Manage your files on cloud.

- Responsive on PC and mobile.

- Normal user and administrator 

**3. Custom Api supported**

- Support [Custom Api](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)

- Manage files by api.

If this project is helpful for you, star ⭐ it.

### Screen Shot

#### PC

![](./build/doc/img/tank0.png)

![](./build/doc/img/tank1.png)

![](./build/doc/img/tank2.png)

![](./build/doc/img/tank3.png)

![](./build/doc/img/tank4.png)

#### Mobile

![](./build/doc/img/mobile.png)


### Installation by Docker

##### a) Run
```
docker run --name tank -p 6010:6010 -d eyeblue/tank:2.0.0
```
If you'd like to start you mysql by docker, use the following command.
```
docker run --name mysql4tank -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_DATABASE=tank -e MYSQL_USER=tank -e MYSQL_PASSWORD=tank123 -v ~/data/mysqldemo1:/var/lib/mysql -d mysql:5.7
```

##### b) Validation
Open http://127.0.0.1:6010 it succeed if you see the installation guide.

### Installation by installation package (windows/linux almost the same)

##### a) Prepare

1. [Mysql](https://www.mysql.com/downloads/) (Use utf8 charset)

2. [Here](https://github.com/eyebluecn/tank/releases) download installation package.

3. Unzip on server

##### b) Run

- Double click `tank.exe` on windows

- Run the following command on linux 

```
cd ApplicationDirectory
./tank
```

If you'd like to run in daemon mode.
```shell
# Run the application
cd ApplicationDirectory/service
./startup.sh

# Stop the application
cd ApplicationDirectory/service
./shutdown.sh

```

##### c) Validation

Open http://127.0.0.1:6010 it succeed if you see the installation guide.

### Reference

[api](https://github.com/eyebluecn/tank/blob/master/build/doc/api_zh.md)

[Custom API](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)

### Contribution

Thanks these authors [@zicla](https://github.com/zicla)，[@seaheart](https://github.com/seaheart)，[@yemuhe](https://github.com/yemuhe)，[@hxsherry](https://github.com/hxsherry)

You can join us by scanning the following qcode.

![](./build/doc/img/dingding.jpg)

### License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2017-present, eyeblue.cn
