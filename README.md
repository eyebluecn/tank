[![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/logo.png)](https://github.com/eyebluecn/tank)

[English Version](./README_EN.md)

# 蓝眼云盘（3.0.0）

[在线Demo](https://tank.eyeblue.cn) (体验账号： demo 密码：123456)

[文档](https://tank-doc.eyeblue.cn/)

### 软件截图

#### PC端截图

![](./build/doc/img/tank0.png)

![](./build/doc/img/tank1.png)

![](./build/doc/img/tank2.png)

![](./build/doc/img/tank3.png)

![](./build/doc/img/tank4.png)

#### 手机端截图

![](./build/doc/img/mobile.png)


#### [安装文档](https://tank-doc.eyeblue.cn/zh/basic/install.html)

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
