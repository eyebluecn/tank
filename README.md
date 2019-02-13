[![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/logo.png)](https://github.com/eyebluecn/tank)

[English Version](./README_EN.md)

# 蓝眼云盘（2.0.0）

[在线Demo](https://tank.eyeblue.cn) (体验账号： demo@tank.eyeblue.cn 密码：123456)

![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [后端tank地址](https://github.com/eyebluecn/tank)

![](https://tank.eyeblue.cn/api/alien/download/df372827-ba56-415e-42d1-0e3a34fdb2a1/github20x20.png "github20x20.png") [前端tank-front地址](https://github.com/eyebluecn/tank-front)

### 简介
蓝眼云盘是 [蓝眼系列开源软件](https://github.com/eyebluecn) 之一，也是蓝眼系列其他软件的基础服务软件。蓝眼云盘具有以下几大特色：

**1. 软件氛围友好**

- 作者比较佛系，开发的软件也比较佛系，只保留最核心大家最常使用的功能，开发这个软件纯靠兴趣，不为任何盈利

- 文末有钉钉群，欢迎加入。你的任何合理需求，我都会尽量满足

**2. 软件体积小，跨平台，运行简单，自带UI引导安装**

- 支持docker安装，蓝眼云盘的docker镜像已经发布到了Docker Hub，可以一行指令运行。

- 蓝眼云盘[安装包](https://github.com/eyebluecn/tank/releases) 10M左右，在`windows`  `linux`  `mac OS`操作系统中均可安装运行；

- 蓝眼云盘为绿色软件，将安装包解压缩，一行命令立即运行。

**3. 使用方便，核心功能齐全**

- 蓝眼云盘主要支持pc web端，同时手机web也具有不错的响应性支持。

- 蓝眼云盘实现了云盘（如：[百度云盘](https://pan.baidu.com/)，[微云](https://www.weiyun.com/))的核心功能：文件夹管理，文件管理，上传文件，下载文件，文件分享等。

- 蓝眼云盘提供能账号管理系统，超级管理员可以管理用户，查看用户文件，普通用户只能查看自己的文件，修改自己的资料。上面提供的体验账号就是一个普通用户的账号。

**4. 支持接口编程**

- 蓝眼云盘提供了[编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)，实现了云存储（如：[七牛云](https://www.qiniu.com)，[阿里云OSS](https://www.aliyun.com/product/oss)）的核心功能

- 可以使用编程接口上传文件，作为其他网站、系统、app的资源存储器。可以在下载图片时对图片做缩放裁剪处理，可以有效地节省客户端流量。同时对于缩略图有缓存策略，全面提升资源访问速度。

- 蓝眼系列开源软件之二的[《蓝眼博客》](https://github.com/eyebluecn/blog)正是使用蓝眼博客作为第三方资源存储器。蓝眼博客中的所有图片，附件均是存储在蓝眼云盘中。

**5. 前后端分离，文档丰富**

- 项目后端使用golang开发，前端使用vue套件开发。

- 蓝眼云盘有详细的[后台api文档](https://github.com/eyebluecn/tank/blob/master/build/doc/api_zh.md)，对于学习前端的童鞋而言可以快速部署一个正式而又具有丰富接口的后端供学习使用。

- 后端技术栈使用 `golang`，没有使用任何web框架；初始化安装，生命周期控制，路由管理，路由匹配，日志管理，依赖注入，错误处理，panic拦截，打包，安装脚本等功能全采用纯手动打造，绿色环保，因此代码更精炼，更具有探讨学习的价值。

**6. 2.x版本人性化的升级**

- 新增了监控大盘，一个页面查看云盘访问情况，热门文件一目了然

- 常用文件(doc,ppt,xls,pdf,mp3,mp4,jpg,png等格式)一键预览，方便快捷

- UI 全面升级，颜色更加沉稳大气，布局更加紧凑，文件上传最多支持1000个同时进行

- mysql支持到5.5，后端代码架构全新迭代升级，日志异常离线任务更加强健


如果您觉得蓝眼云盘对您有帮助，请不要吝惜您的star ⭐

### 软件截图

#### PC端截图

![](./build/doc/img/tank0.png)

![](./build/doc/img/tank1.png)

![](./build/doc/img/tank2.png)

![](./build/doc/img/tank3.png)

![](./build/doc/img/tank4.png)

#### 手机端截图

![](./build/doc/img/mobile.png)


### Docker安装

##### a) 启动
```
docker run --name tank -p 6010:6010 -d eyeblue/tank:2.0.0
```
如果你的mysql也希望用docker运行，可以使用这句话
```
docker run --name mysql4tank -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_DATABASE=tank -e MYSQL_USER=tank -e MYSQL_PASSWORD=tank123 -v ~/data/mysqldemo1:/var/lib/mysql -d mysql:5.7
```
如果mysql是docker启动的，那么在填写mysql host的时候不能用 127.0.0.1，要用局域网ip。 或者你可以在tank的镜像启动中使用 --link=mysql4tank:mysql4tank 从而填写 mysql4tank 作为mysql的host.

##### b) 验证
浏览器中打开 http://127.0.0.1:6010 看到PC截图最后一张的安装引导页面即表示成功。

### 常规安装(windows/linux方法基本一致)

##### a) 准备工作

1. 一台windows/linux服务器，当然你可以使用自己的电脑充当这台服务器

2. [安装好Mysql数据库](https://www.mysql.com/downloads/) (请使用utf8编码)

3. [在这里](https://github.com/eyebluecn/tank/releases)下载服务器对应的安装包

4. 在服务器上解压缩

##### b) 运行

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

##### c) 验证

浏览器中打开 http://127.0.0.1:6010 

(127.0.0.1请使用服务器所在ip，6010请使用`tank.json`中配置的`ServerPort`) 

可以看到安装引导页面（见上文PC截图最后一张）即表示安装成功。然后按照引导逐步配置即可。


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
│   │   ├── rest             #golang源代码目录
│   │   │   ├── ...          #golang源代码 不同文件用前缀区分
│   │   ├── .gitignore       #gitignore文件
│   │   ├── CHNAGELOG        #版本变化日志
│   │   ├── DOCKERFILE       #构建Docker的文件
│   │   ├── LICENSE          #证书说明文件
│   │   ├── main.go          #程序入口文件
│   │   ├── README.md        #README文件
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

当然你可以加入钉钉群一起直接交流

![](./build/doc/img/dingding.jpg)

### License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2017-present, eyeblue.cn
