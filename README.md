![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/logo.png)

# 蓝眼云盘

##### [在线Demo](http://tank.eyeblue.cn) (体验账号： demo@tank.eyeblue.cn 密码：123456)

##### [配套前端tank-front](https://github.com/eyebluecn/tank-front)


### 简介
蓝眼云盘是 [蓝眼系列开源软件](https://github.com/eyebluecn) 之一，也是蓝眼系列其他软件的基础服务软件。蓝眼博客具有以下几大特色：

**1. 软件体积小，跨平台，运行简单**

- 蓝眼云盘[安装包]((https://github.com/eyebluecn/tank/releases)6M左右，在`windows`  `linux`  `mac OS`操作系统中均可安装运行；

- 蓝眼云盘为绿色软件，将安装包解压缩，修改配置文件后即可一行命令立即运行。

**2. 使用方便，核心功能齐全**

- 蓝眼云盘主要支持pc web端，同时手机web也具有不错的响应性支持。

- 蓝眼云盘实现了云盘（如：[百度云盘](https://pan.baidu.com/)，[微云](https://www.weiyun.com/))的核心功能：文件夹管理，文件管理，上传文件，下载文件，文件分享等。

- 蓝眼云盘提供能账号管理系统，超级管理员可以管理用户，查看用户文件，普通用户只能查看自己的文件，修改自己的资料。上面提供的体验账号就是一个普通用户的账号。

**3. 支持接口编程**

- 蓝眼云盘提供了[编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)，实现了云存储（如：[七牛云](https://www.qiniu.com)，[阿里云OSS](https://www.aliyun.com/product/oss)）的核心功能，可以使用编程接口上传文件，作为其他网站、系统、app的资源存储器。可以在下载图片时对图片做缩放裁剪处理，可以有效地节省客户端流量。

- 蓝眼系列开源软件之二的[《蓝眼博客》](https://github.com/eyebluecn/blog)正是使用蓝眼博客作为第三方资源存储器。蓝眼博客中的所有图片，附件均是存储在蓝眼云盘中。

**4. 前后端分离，文档丰富**

- 蓝眼云盘有详细的[后台api文档](https://github.com/eyebluecn/tank/blob/master/build/doc/api_zh.md)，对于学习前端的童鞋而言可以快速部署一个正式而又具有丰富接口的后端供学习使用。

- 后端技术栈使用 `golang`，没有使用任何web框架；初始化安装，生命周期控制，路由管理，路由匹配，日志管理，依赖注入，错误处理，panic拦截，打包，安装脚本等功能全采用纯手动打造，绿色环保，因此代码更精炼，更具有探讨学习的价值。

如果您觉得蓝眼云盘对您有帮助，请不要吝惜您的star <i class="fa fa-star"></i>

### 软件截图

![](http://tank.eyeblue.cn/api/alien/download/3e71763a-99a2-42a4-5718-6a198c9fa486/tank0.png "tank0.png")

![](http://tank.eyeblue.cn/api/alien/download/779e3a82-579a-41e6-40ee-c9c47545f02f/tank1.png "tank1.png")

![](http://tank.eyeblue.cn/api/alien/download/ebf44ac7-9fa0-4191-7e8c-821870d5fb96/tank2.png "tank2.png")

### 安装

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
1. clone ![](http://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [tank-front](https://github.com/eyebluecn/tank-front)

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

1. clone ![](http://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [tank](https://github.com/eyebluecn/tank)

2. 安装Golang，环境变量`GOPATH`配置到工程目录，建议工程目录结构如下：

```
golang                                          #环境变量GOPATH所在路径
├── bin                                         #编译生成的可执行文件目录
├── pkg                                         #编译生成第三方库
├── src                                         #golang工程源代码
│   ├── github.com                              #来自github的第三方库
│   ├── golang.org                              #来自golang.org的第三方库
│   ├── tank                                    #clone下来的tank根目录
│   │   ├── build                               #用来辅助打包的文件夹
│   │   │   ├── conf                            #默认的配置文件
│   │   │   ├── doc                             #文档
│   │   │   ├── html                            #前端静态资源，从项目tank-front编译获得
│   │   │   ├── pack                            #打包的脚本
│   │   │   ├── service                         #将tank当作服务启动的脚本
│   │   ├── dist                                #运行打包脚本后获得的安装包目录
│   │   ├── rest                                #golang源代码
      
```

3. 准备项目依赖的第三方库

- golang.org/x
- github.com/disintegration/imaging
- github.com/json-iterator/go
- github.com/go-sql-driver/mysql
- github.com/jinzhu/gorm
- github.com/nu7hatch/gouuid

其中`golang.org/x`国内无法下载，请从[这里](https://github.com/eyebluecn/golang.org)下载，并按上文推荐的目录结构放置。其余依赖项均可通过安装脚本自动下载。

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

### Contribution

感谢所有蓝眼云盘的贡献者 [@zicla](https://github.com/zicla)，[@seaheart](https://github.com/seaheart)，[@yemuhe](https://github.com/yemuhe)，[@hxsherry](https://github.com/hxsherry)

如果您也想参与进来，请尽情的fork, star, post issue, pull requests

### License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2017-present, eyeblue.cn