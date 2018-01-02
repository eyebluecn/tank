![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/logo.png)

# 蓝眼云盘

##### [在线Demo](http://tank.eyeblue.cn) (体验账号： demo@tank.eyeblue.cn 密码：123456)

##### [配套前端tank-front](https://github.com/eyebluecn/tank-front)

### 简介
蓝眼云盘是 [蓝眼系列开源软件](https://github.com/eyebluecn) 中的第一个

- 主要用于快速搭建私人云盘，可以简单理解为部署在自己服务器上的[百度云盘](https://pan.baidu.com/)。
- 蓝眼云盘提供了编程接口，可以使用接口上传文件，作为其他网站、系统、app的资源存储器，可以当作单机版的[七牛云](https://www.qiniu.com)或[阿里云OSS](https://www.aliyun.com/product/oss)使用。
- 蓝眼云盘还提供了账号管理系统，超级管理员可以管理用户，查看用户文件，普通用户只能查看自己的文件，修改自己的资料。上面提供的体验账号就是一个普通用户的账号。

蓝眼云盘可以作为团队内部或个人私有的云盘使用，亦可当作专门处理图片，音频，视频等二进制文件的第三方编程辅助工具。

如果您觉得蓝眼云盘对您有帮助，请不要吝惜您的star :smile:

### 使用安装包安装

#### a) 准备工作

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

#### b) 运行

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

#### c) 验证

浏览器中打开 http://127.0.0.1:6010 (127.0.0.1请使用服务器所在ip，6010请使用`tank.json`中配置的`ServerPort`) 可以看到以下登录页面：

![蓝眼云盘登录页面](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/login.png)

使用上方配置文件中的邮箱和密码登录后可以看到如下界面：
![蓝眼云盘登录页面](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/matters.png)

### 使用源代码自行打包

#### a) 准备工作

1. 一台windows/linux服务器，当然你可以使用自己的电脑充当这台服务器

2. Mysql数据库

3. clone本项目

4. 安装Golang，环境变量 `GOPATH`配置到工程目录，建议工程目录结构如下：

```
golang                   #环境变量GOPATH所在路径
....bin                  #编译生成的可执行文件目录
....pkg                  #编译生成第三方库
....src                  #golang工程源代码
........github.com       #来自github的第三方库
........golang.org       #来自golang.org的第三方库
........tank             #clone下来的tank根目录
............build        #用来辅助打包的文件夹
................conf     #默认的配置文件
................doc      #文档
................html     #前端静态资源，从项目tank-front编译获得
................pack     #打包的脚本
................service  #将tank当作服务启动的脚本
............dist         #运行打包脚本后获得的安装包目录
............rest         #golang源代码
```

5. 准备项目依赖的第三方库

- golang.org/x
- github.com/disintegration/imaging
- github.com/json-iterator/go
- github.com/go-sql-driver/mysql
- github.com/jinzhu/gorm
- github.com/nu7hatch/gouuid

其中`golang.org/x`国内无法下载，请从[这里](https://github.com/eyebluecn/golang.org)下载，并按上文推荐的目录结构放置。其余依赖项均可通过安装脚本自动下载。

#### b) 打包

- windows平台双击运行 `tank/build/pack/build.bat`，成功之后可在`tank/dist`下看到`tank-x.x.x`文件夹，该文件夹即为最终安装包。

- linux平台运行如下命令：
```
cd tank/build/pack/
./build.sh
```
成功之后可在`tank/dist`下看到`tank-x.x.x.linux-amd64.tar.gz`

利用得到的安装包即可参考上文的`使用安装包安装`。


### Contribution

感谢所有蓝眼云盘的贡献者 [@zicla](https://github.com/zicla)，[@seaheart](https://github.com/seaheart)，@heying，[@hxsherry](https://github.com/hxsherry)

如果您也想参与进来，请尽情的fork, star, post issue, pull requests

### License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2017-present, eyeblue.cn