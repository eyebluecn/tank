[![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/logo.png)](https://github.com/eyebluecn/tank)

# 蓝眼云盘

[在线Demo](https://tank.eyeblue.cn) (体验账号： demo@tank.eyeblue.cn 密码：123456)

![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [后端tank地址](https://github.com/eyebluecn/tank)

![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [前端tank-front地址](https://github.com/eyebluecn/tank-front)

### 简介
蓝眼云盘是 [蓝眼系列开源软件](https://github.com/eyebluecn) 之一，也是蓝眼系列其他软件的基础服务软件。蓝眼博客具有以下几大特色：

**1. 支持docker**

- 蓝眼云盘的docker镜像已经发布到了Docker Hub，可以一行指令运行。

**2. 软件体积小，跨平台，运行简单**

- 蓝眼云盘[安装包](https://github.com/eyebluecn/tank/releases) 6M左右，在`windows`  `linux`  `mac OS`操作系统中均可安装运行；

- 蓝眼云盘为绿色软件，将安装包解压缩，修改配置文件后即可一行命令立即运行。

**3. 使用方便，核心功能齐全**

- 蓝眼云盘主要支持pc web端，同时手机web也具有不错的响应性支持。

- 蓝眼云盘实现了云盘（如：[百度云盘](https://pan.baidu.com/)，[微云](https://www.weiyun.com/))的核心功能：文件夹管理，文件管理，上传文件，下载文件，文件分享等。

- 蓝眼云盘提供能账号管理系统，超级管理员可以管理用户，查看用户文件，普通用户只能查看自己的文件，修改自己的资料。上面提供的体验账号就是一个普通用户的账号。

**4. 支持接口编程**

- 蓝眼云盘提供了[编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)，实现了云存储（如：[七牛云](https://www.qiniu.com)，[阿里云OSS](https://www.aliyun.com/product/oss)）的核心功能，可以使用编程接口上传文件，作为其他网站、系统、app的资源存储器。可以在下载图片时对图片做缩放裁剪处理，可以有效地节省客户端流量。

- 蓝眼系列开源软件之二的[《蓝眼博客》](https://github.com/eyebluecn/blog)正是使用蓝眼博客作为第三方资源存储器。蓝眼博客中的所有图片，附件均是存储在蓝眼云盘中。

**5. 前后端分离，文档丰富**

- 项目后端使用golang开发，前端使用vue套件开发。

- 蓝眼云盘有详细的[后台api文档](https://github.com/eyebluecn/tank/blob/master/build/doc/api_zh.md)，对于学习前端的童鞋而言可以快速部署一个正式而又具有丰富接口的后端供学习使用。

- 后端技术栈使用 `golang`，没有使用任何web框架；初始化安装，生命周期控制，路由管理，路由匹配，日志管理，依赖注入，错误处理，panic拦截，打包，安装脚本等功能全采用纯手动打造，绿色环保，因此代码更精炼，更具有探讨学习的价值。

如果您觉得蓝眼云盘对您有帮助，请不要吝惜您的star <i class="fa fa-star"></i>

### 软件截图

#### PC端截图

![](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/tank0.png)

![](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/tank1.png)

![](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/tank2.png)

#### 手机端截图

![](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/mobile.png)

### Docker方式运行

蓝眼云盘在运行时依赖`mysql`数据库，因此最好的方式是使用`docker-compose`来运行，如果你对`docker-compose`不熟悉，可以参考这篇文章：[《Docker Compose 项目》](https://yeasy.gitbooks.io/docker_practice/content/compose/introduction.html) 

** 1.准备 docker-compose.yml 文件 **

`docker-compose.yml`描述了镜像启动的数据卷，环境变量，启动方式，依赖项等。该文件位于项目的根目录下，内容如下：
```shell
#docker-compose版本，这里的3不要动
version: "3"
services:

   # 数据库的镜像信息
   # 使用mysql:5.7的镜像
   db:
     image: mysql:5.7
     volumes:
       # 数据库文件存放在宿主机的`~/data/mysql`位置，如果宿主机目录不存在，则会自动创建
       - ~/data/mysql:/var/lib/mysql
     # 如果启动失败，则总是会重启。因为镜像有依赖的情况，不停重启可以保证依赖项启动成功后自己再运行
     restart: always
     environment:
       # 指定root密码为`tank123`，并且创建一个新数据库`tank`，同时用户名和密码为`tank` `tank123`
       MYSQL_ROOT_PASSWORD: tank123
       MYSQL_DATABASE: tank
       MYSQL_USER: tank
       MYSQL_PASSWORD: tank123

   # 蓝眼云盘的镜像信息
   # 依赖 mysql:5.7 的镜像
   tank:
     image: eyeblue/tank:1.0.2
     depends_on:
       - db
     ports:
       # 端口映射关系，宿主机端口:镜像端口
       - "6010:6010"
     # 如果启动失败，则总是会重启。因为镜像有依赖的情况，不停重启可以保证依赖项启动成功后自己再运行
     restart: always
     environment:
       # mysql的端口
       TANK_MYSQL_PORT: 3306
       # Mysql的主机，和services的第一个节点一致。
       TANK_MYSQL_HOST: db
       # 数据库
       TANK_MYSQL_SCHEMA: tank
       # 数据库的用户名
       TANK_MYSQL_USERNAME: tank
       # 数据库的密码
       TANK_MYSQL_PASSWORD: tank123
       # 超级管理员的昵称。只能是英文或数字
       TANK_ADMIN_USERNAME: admin
       # 超级管理员邮箱，作为登录账号
       TANK_ADMIN_EMAIL: admin@tank.eyeblue.cn
       # 超级管理员密码，作为登录密码
       TANK_ADMIN_PASSWORD: 123456
     volumes:
       # 日志文件存放在宿主机的`~/data/eyeblue/tank/log`位置，如果宿主机目录不存在，则会自动创建
       - ~/data/eyeblue/tank/log:/data/log
       # 上传文件存放在宿主机的`~/data/eyeblue/tank/matter`位置，如果宿主机目录不存在，则会自动创建
       - ~/data/eyeblue/tank/matter:/data/matter
```

** 2.运行项目 **

首先保证当前目录是`docker-compose.yml`所在的目录，然后执行以下指令即可运行蓝眼云盘：
```shell
$ docker-compose up -d
```

** 3.验证 **

由于数据库启动需要一定的时间，因此大约20s后，打开浏览器访问`http://127.0.0.1:6010`，如果看到登录界面则表示运行成功。

** 4.停止项目**

方法一：使用以下命令来停止蓝眼云盘
``` shell
$ docker-compose stop
```

方法二：当然你也可以用停止容器的方式来停止蓝眼云盘
``` shell
$ docker container ls

CONTAINER ID        IMAGE                COMMAND                  CREATED             STATUS              PORTS                    NAMES
f5f64735fc53        eyeblue/tank:1.0.2   "/go/bin/tank"           20 minutes ago      Up 13 seconds       0.0.0.0:6010->6010/tcp   tank_tank_1
3a859cad3e7e        mysql:5.7            "docker-entrypoint.s…"   20 minutes ago      Up 14 seconds       3306/tcp                 tank_db_1

$ docker container stop f5
$ docker container stop 3a
```

如果你比较关心如何使用docker来构建蓝眼云盘，请参考 [《Docker 化你的开源项目》](https://blog.eyeblue.cn/home/article/510f9316-9ca1-40fe-b1b3-5285505a527d) 

### 常规安装

** a) 准备工作 **

1. 一台windows/linux服务器，当然你可以使用自己的电脑充当这台服务器

2. [安装好Mysql数据库](https://www.mysql.com/downloads/)

3. [在这里](https://github.com/eyebluecn/tank/releases)下载服务器对应的安装包

4. 在服务器上解压缩，修改配置文件`conf/tank.json`，各项说明如下：

```
{
   //服务器运行的端口，默认6010。如果配置为80，则可直接用http打开
  "ServerPort": 6010,
  //日志是否需要打印到控制台，默认false，主要用于调试
  "LogToConsole": false,
  //Mysql端口，默认3306
  "MysqlPort": 3306,
  //Mysql主机
  "MysqlHost": "127.0.0.1",
  //Mysql数据库名称
  "MysqlSchema": "tank",
  //Mysql用户名，建议为蓝眼云盘创建一个用户，不建议使用root
  "MysqlUserName": "tank",
  //Mysql密码
  "MysqlPassword": "tank123",
  //超级管理员用户名，只能是字母和数字
  "AdminUsername": "admin",
  //超级管理员邮箱，作为登录的账号
  "AdminEmail": "admin@tank.eyeblue.cn",
  //超级管理员密码
  "AdminPassword": "123456"
}

```

** b) 运行 **

- windows平台直接双击应用目录下的`tank.exe`。

- linux平台执行 

```
cd 应用目录路径
./tank
```

如果你希望关闭shell窗口后，应用依然运行，请使用以下脚本启动和停止
```shell
# 启动应用
cd 应用目录路径/service
./startup.sh

# 停止应用
cd 应用目录路径/service
./shutdown.sh

```

** c) 验证 **

浏览器中打开 http://127.0.0.1:6010 (127.0.0.1请使用服务器所在ip，6010请使用`tank.json`中配置的`ServerPort`) 可以看到登录页面，并且使用配置文件中的邮箱和密码登录成功后可以看到全部文件（见上文截图）即表示安装成功。


### 使用源代码自行打包


**前端项目打包**
1. clone ![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [tank-front](https://github.com/eyebluecn/tank-front)

2. 安装依赖项
```
npm install
```
3. 执行打包命令
```
npm run build
```
4. 通过前面三步可以在`dist`文件夹下得到打包后的静态文件，将`dist`目录下的所有文件拷贝到后端项目的`build/html`文件夹下。（下文的工程目录中也有说明）

**后端项目打包**

1. clone ![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [tank](https://github.com/eyebluecn/tank)

2. 安装Golang，环境变量`GOPATH`配置到工程目录，建议工程目录结构如下：

```
golang                       #环境变量GOPATH所在路径
├── bin                      #编译生成的可执行文件目录
├── pkg                      #编译生成第三方库
├── src                      #golang工程源代码
│   ├── github.com           #来自github的第三方库
│   ├── golang.org           #来自golang.org的第三方库
│   ├── tank                 #clone下来的tank根目录
│   │   ├── build            #用来辅助打包的文件夹
│   │   │   ├── conf         #默认的配置文件
│   │   │   ├── doc          #文档
│   │   │   ├── html         #前端静态资源，从项目tank-front编译获得
│   │   │   ├── pack         #打包的脚本
│   │   │   ├── service      #将tank当作服务启动的脚本
│   │   ├── dist             #运行打包脚本后获得的安装包目录
│   │   ├── rest             #golang源代码
      
```

3. 准备项目依赖的第三方库

- golang.org/x
- github.com/disintegration/imaging
- github.com/json-iterator/go
- github.com/go-sql-driver/mysql
- github.com/jinzhu/gorm
- github.com/nu7hatch/gouuid

其中`golang.org/x`国内无法下载，默认会通过git clone 的方式从 [这里](https://github.com/eyebluecn/golang.org)下载。其余依赖项均会通过`go get`的方式下载。

4. 打包

- windows平台双击运行 `tank/build/pack/build.bat`，成功之后可在`tank/dist`下看到`tank-x.x.x`文件夹，该文件夹即为最终安装包。

- linux平台运行如下命令：
```
cd tank/build/pack/
./build.sh
```
成功之后可在`tank/dist`下看到`tank-x.x.x.linux-amd64.tar.gz`

利用得到的安装包即可参考上文的`安装`一节进行安装。


### 相关文档

[蓝眼云盘后端api](https://github.com/eyebluecn/tank/blob/master/build/doc/api_zh.md)

[蓝眼云盘编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)

[快速使用Let's Encrypt开启个人网站的https](https://blog.eyeblue.cn/home/article/9f580b3f-5679-4a9d-be6f-4d9f0dd417af) 

 [Docker 化你的开源项目](https://blog.eyeblue.cn/home/article/510f9316-9ca1-40fe-b1b3-5285505a527d)  

### Contribution

感谢所有蓝眼云盘的贡献者 [@zicla](https://github.com/zicla)，[@seaheart](https://github.com/seaheart)，[@yemuhe](https://github.com/yemuhe)，[@hxsherry](https://github.com/hxsherry)

如果您也想参与进来，请尽情的fork, star, post issue, pull requests

### License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2017-present, eyeblue.cn
