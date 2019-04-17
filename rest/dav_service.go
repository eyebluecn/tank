package rest

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

/**
 *
 * WebDav协议文档
 * https://tools.ietf.org/html/rfc4918
 *
 */
//@Service
type DavService struct {
	Bean
	matterDao *MatterDao
}

//初始化方法
func (this *DavService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}
}

//从request中读取深度
func (this *DavService) readDepth(request *http.Request) int {

	depth := INFINITE_DEPTH
	if hdr := request.Header.Get("Depth"); hdr != "" {
		if hdr == "0" {
			depth = 0
		} else if hdr == "1" {
			depth = 1
		} else if hdr == "infinity" {
			depth = INFINITE_DEPTH
		} else {
			panic("Depth格式错误！")
		}
	}
	return depth
}

//处理 方法
func (this *DavService) HandlePropfind(writer http.ResponseWriter, request *http.Request, subPath string) {

	//获取请求者
	user := this.checkUser(writer, request)

	//读取希望访问的深度。
	depth := this.readDepth(request)

	//找寻请求的目录
	matter := this.matterDao.checkByUserUuidAndPath(user.Uuid, subPath)

	//TODO: 读取请求参数。按照用户的参数请求返回内容。
	propfind := &Propfind{}
	body, err := ioutil.ReadAll(request.Body)
	this.PanicError(err)

	//从xml中解析内容到struct
	err = xml.Unmarshal(body, &propfind)
	this.PanicError(err)

	//从struct还原到xml
	output, err := xml.MarshalIndent(propfind, "  ", "    ")
	this.PanicError(err)
	fmt.Println(string(output))



	fmt.Printf("%v %v \n", depth, matter.Name)

}
