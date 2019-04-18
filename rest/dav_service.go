package rest

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"tank/rest/dav"
)

/**
 *
 * WebDav协议文档
 * https://tools.ietf.org/html/rfc4918
 * 主要参考 golang.org/x/net/webdav
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

func (this *DavService) makePropstatResponse(href string, pstats []dav.Propstat) *dav.Response {
	resp := dav.Response{
		Href:     []string{(&url.URL{Path: href}).EscapedPath()},
		Propstat: make([]dav.SubPropstat, 0, len(pstats)),
	}
	for _, p := range pstats {
		var xmlErr *dav.XmlError
		if p.XMLError != "" {
			xmlErr = &dav.XmlError{InnerXML: []byte(p.XMLError)}
		}
		resp.Propstat = append(resp.Propstat, dav.SubPropstat{
			Status:              fmt.Sprintf("HTTP/1.1 %d %s", p.Status, dav.StatusText(p.Status)),
			Prop:                p.Props,
			ResponseDescription: p.ResponseDescription,
			Error:               xmlErr,
		})
	}
	return &resp
}

//从一个matter中获取其propnames，每个propname都是一个xml标签。
func (this *DavService) PropNames(matter *Matter) []xml.Name {

	return nil

}

//处理 方法
func (this *DavService) HandlePropfind(writer http.ResponseWriter, request *http.Request, subPath string) {

	fmt.Printf("列出文件/文件夹 %s\n", subPath)

	//获取请求者
	user := this.checkUser(writer, request)

	//找寻请求的目录
	matter := this.matterDao.checkByUserUuidAndPath(user.Uuid, subPath)

	//读取请求参数。按照用户的参数请求返回内容。
	propfind, _, err := dav.ReadPropfind(request.Body)
	this.PanicError(err)

	//寻找符合条件的matter.
	matters := this.matterDao.ListByUserUuidAndPath(user.Uuid, subPath)
	if len(matters) == 0 {
		this.PanicNotFound("%s不存在", subPath)
	}

	//准备一个输出结果的Writer
	multiStatusWriter := dav.MultiStatusWriter{Writer: writer}

	for _, matter := range matters {

		fmt.Printf("开始分析 %s\n", matter.Name)

		var propstats []dav.Propstat
		var props = make([]dav.Property, 0)
		props = append(props, dav.Property{
			XMLName: xml.Name{Space: "DAV:"},
		})
		propstats = append(propstats, dav.Propstat{
			Props:               props,
			ResponseDescription: "有点问题",
		})

		response := this.makePropstatResponse("/eyeblue/ready/go", propstats)

		err = multiStatusWriter.Write(response)
		this.PanicError(err)
	}

	//闭合
	err = multiStatusWriter.Close()
	this.PanicError(err)

	fmt.Printf("%v %v \n", matter.Name, propfind.Prop)

}
