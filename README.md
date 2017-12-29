![蓝眼云盘logo](https://raw.githubusercontent.com/eyebluecn/tank/dev/build/doc/logo.png)

# 蓝眼云盘

##### [在线Demo](http://tank.eyeblue.cn) 

##### [配套前端tank-front](https://github.com/eyebluecn/tank-front)

### 简介
蓝眼云盘是 [蓝眼系列开源软件](https://github.com/eyebluecn) 中的第一个

- 主要用于快速搭建私人云盘，可以简单理解为部署在自己服务器上的百度云盘。
- 蓝眼云盘提供了编程接口，可以使用接口上传文件，作为其他网站、系统、app的资源存储器，可以当作单机版的七牛云或阿里云OSS使用。

蓝眼云盘可以作为团队内部或个人私有的云盘使用，亦可当作专门处理图片，音频，视频等二进制文件的第三方编程辅助工具。

### 安装

1. 一台windows/linux服务器，当然你可以使用自己的电脑充当这台服务器

2. Mysql数据库

3. [在这里](https://github.com/eyebluecn/tank/releases)下载服务器对应的安装包

4. 在服务器上解压缩，修改配置文件`tank.json`，各项说明如下：
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

5. 运行

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

6. 验证

浏览器中打开 http://127.0.0.1:6010 (127.0.0.1请使用服务器所在ip，6010请使用`tank.json`中配置的`ServerPort`) 可以看到以下登录页面：


### 使用源代码自行打包

