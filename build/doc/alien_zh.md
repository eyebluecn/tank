# 蓝眼云盘编程接口

- 蓝眼云盘提供了[编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)，实现了云存储（如：[七牛云](https://www.qiniu.com)，[阿里云OSS](https://www.aliyun.com/product/oss)）的核心功能，可以使用编程接口上传文件，作为其他网站、系统、app的资源存储器。可以在下载图片时对图片做缩放裁剪处理，可以有效地节省客户端流量。

- 蓝眼系列开源软件之二的[《蓝眼博客》](https://github.com/eyebluecn/blog)正是使用蓝眼博客作为第三方资源存储器。蓝眼博客中的所有图片，附件均是存储在蓝眼云盘中。

所有的编程接口均定义在`alien_controller`中，主要包括以下几个接口：

`/api/alien/fetch/upload/token` 获取上传token

`/api/alien/fetch/download/token` 获取下载token

`/api/alien/confirm` 来蓝眼云盘确认文件

`/api/alien/upload` 使用form表单上传文件

`/api/alien/download/{uuid}/{filename}` 下载文件

### 上传时序图

![上传时序图](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/upload-time-line.png)

### 下载时序图

![下载时序图](https://raw.githubusercontent.com/eyebluecn/tank/master/build/doc/img/download-time-line.png)

### 接口详情

#### /api/alien/fetch/upload/token
功能：一个蓝眼云盘受信任的用户请求一个`UploadToken`，用于给另一个用户向蓝眼云盘上传文件。

一般的使用场景是`应用服务器`向`蓝眼云盘`请求`UploadToken`，然后将此`UploadToken`交由`浏览器`去向`蓝眼云盘`上传文件。

参数 | 类型 | 描述
--------- | ---- | -----------
email | `string` | 【必填】邮箱，用于确定请求者身份
password | `string` | 【必填】密码，用于确定请求者身份
filename | `string` | 【必填】即将上传的文件名，不能包含以下特殊符号：`< > \| * ? / \`
expire | `int` | 【选填】UploadToken过期时间，单位：s。默认 86400s 即24h
privacy | `bool` | 【必填】文件的共有性。`true`表示文件私有，下载时必须要DownloadToken. `false`表示文件公有，任何人可以通过下载链接直接下载
size | `int` | 【必填】文件的大小。单位：byte
dir | `string` | 【必填】文件存放的路径。不能为空，必须以`/`开头，不能出现连续的`//`,不能包含以下特殊符号：`< > \| * ? \`。举例：`/app/blog/20180101121212001`

#### /api/alien/upload

功能：浏览器拿着`UploadToken`通过FormData向蓝眼云盘上传文件。

一般的使用场景是`应用服务器`向`蓝眼云盘`请求`UploadToken`，然后将此`UploadToken`交由`浏览器`去向`蓝眼云盘`上传文件。由于在请求`UploadToken`的时候已经传入了文件元信息，因此这里的文件信息必须要和`/api/alien/fetch/upload/token`传入的参数信息一致。

参数 | 类型 | 描述
--------- | ---- | -----------
uploadTokenUuid | `string` | 【必填】uploadToken标识，`/api/alien/fetch/upload/token`请求返回对象中的`uuid`
file | `file` | 【必填】文件，在浏览器中是通过`<input type="file" name="file"/>`来选择的

#### /api/alien/confirm

功能：`应用服务器`向蓝眼云盘确认某个文件是否确实已经上传好了。

参数 | 类型 | 描述
--------- | ---- | -----------
email | `string` | 【必填】邮箱，用于确定请求者身份
password | `string` | 【必填】密码，用于确定请求者身份
matterUuid | `string` | 【必填】浏览器上传完毕后，蓝眼云盘返回给浏览器的`uuid`

#### /api/alien/fetch/download/token

功能：一个蓝眼云盘受信任的用户请求一个`DownloadToken`，用于给另一个用户下载蓝眼云盘上的私有文件。

一般的使用场景是`应用服务器`向`蓝眼云盘`请求`DownloadToken`，然后将此`DownloadToken`交由`浏览器`去向`蓝眼云盘`下载文件。

参数 | 类型 | 描述
--------- | ---- | -----------
email | `string` | 【必填】邮箱，用于确定请求者身份
password | `string` | 【必填】密码，用于确定请求者身份
matterUuid | `string` | 【必填】文件uuid，要想下载的文件`uuid`
expire | `int` | 【选填】UploadToken过期时间，单位：s。默认 86400s 即24h


#### /api/alien/download/{uuid}/{filename}

功能：在浏览器中下载文件

这个接口既可以下载公有文件，又可以下载私有文件。同时对于图片文件还可以做裁剪缩放等操作。

参数 | 类型 | 描述
--------- | ---- | -----------
uuid | `string` | 【必填】文件的uuid，该参数放在url的路径中
filename | `string` | 【必填】文件的名称，该参数放在url的路径中
downloadTokenUuid | `string` | 【选填】download的uuid，如果是私有文件该参数必须，公有文件无需填写。

该接口同时还可以对图片进行缩放预处理
> 图片缩放支持的格式有：".jpg", ".jpeg", ".png", ".tif", ".tiff", ".bmp", ".gif"

##### 额外参数

| 参数 | 类型   |  描述  | 取值范围  |
| ------------ | ---- | ------------ | ------------ |
| imageProcess | `string`  | 指定图片处理的方式，对于图片缩放固定为`resize`  |  固定为`resize` |
| imageResizeM | `string` | 指定图片缩放的策略，有三种策略，`fit` 表示固定一边，另一边按比例缩放；`fill`表示先将图片延伸出指定W与H的矩形框外，然后进行居中裁剪；`fixed`表示直接按照指定的W和H缩放图片，这种方式可能导致图片变形  | [`fit`,`fill`,`fixed`] 不填默认`fit`   |
|  imageResizeW | `int`  |  指定的宽度，对于`fit`可以不指定 |  1 ~ 4096  |
|  imageResizeH | `int`  |  指定的高度，对于`fit`可以不指定 |  1 ~ 4096  |

##### 示例

原图：

![将宽度指定为200，高度等比例缩放](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg)

1. 将宽度指定为200，高度等比例缩放

![将宽度指定为200，高度等比例缩放](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fit&imageResizeW=200)

[http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fit&imageResizeW=200](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fit&imageResizeW=200)

2. 将高度指定为200，宽度等比例缩放

![将高度指定为200，宽度等比例缩放](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fit&imageResizeH=200)

[http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fit&imageResizeH=200](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fit&imageResizeH=200)


3. 图片自动填充在200*200的大小中 （这种情况用得最多）

![图片自动填充在200*200的大小中](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fill&imageResizeW=200&imageResizeH=200)

[http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fill&imageResizeW=200&imageResizeH=200](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fill&imageResizeW=200&imageResizeH=200)

4. 图片固定大小200*200 (一般会导致变形)

![图片自动填充在200*200的大小中](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fixed&imageResizeW=200&imageResizeH=200)

[http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fixed&imageResizeW=200&imageResizeH=200](http://tank.eyeblue.cn/api/alien/download/3f4b3090-e688-4d63-7705-93a120690505/horse.jpg?imageProcess=resize&imageResizeM=fixed&imageResizeW=200&imageResizeH=200)
