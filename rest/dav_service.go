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

//从一个matter中获取其 []dav.Propstat
func (this *DavService) PropstatsFromXmlNames(matter *Matter, xmlNames []xml.Name) []dav.Propstat {

	propstats := make([]dav.Propstat, 0)

	var properties []dav.Property

	for _, xmlName := range xmlNames {
		//TODO: deadprops尚未考虑

		// Otherwise, it must either be a live property or we don't know it.
		if liveProp := LivePropMap[xmlName]; liveProp.findFn != nil && (liveProp.dir || !matter.Dir) {
			innerXML := liveProp.findFn(matter)

			properties = append(properties, dav.Property{
				XMLName:  xmlName,
				InnerXML: []byte(innerXML),
			})
		} else {
			this.logger.Info("%s的%s无法完成", matter.Path, xmlName.Local)
		}
	}

	if len(properties) == 0 {
		this.PanicBadRequest("请求的属性项无法解析！")
	}

	okPropstat := dav.Propstat{Status: http.StatusOK, Props: properties}

	propstats = append(propstats, okPropstat)

	return propstats

}

//从一个matter中获取所有的propsNames
func (this *DavService) AllPropXmlNames(matter *Matter) []xml.Name {

	pnames := make([]xml.Name, 0)
	for pn, prop := range LivePropMap {
		if prop.findFn != nil && (prop.dir || !matter.Dir) {
			pnames = append(pnames, pn)
		}
	}

	return pnames
}

//从一个matter中获取其 []dav.Propstat
func (this *DavService) Propstats(matter *Matter, propfind dav.Propfind) []dav.Propstat {

	propstats := make([]dav.Propstat, 0)
	if propfind.Propname != nil {
		this.PanicBadRequest("propfind.Propname != nil 尚未处理")
	} else if propfind.Allprop != nil {

		//TODO: 如果include中还有内容，那么包含进去。
		xmlNames := this.AllPropXmlNames(matter)

		propstats = this.PropstatsFromXmlNames(matter, xmlNames)

	} else {
		propstats = this.PropstatsFromXmlNames(matter, propfind.Prop)
	}

	return propstats

}

//处理 方法
func (this *DavService) HandlePropfind(writer http.ResponseWriter, request *http.Request, subPath string) {

	fmt.Printf("列出文件/文件夹 %s\n", subPath)

	//获取请求者
	user := this.checkUser(writer, request)

	//读取请求参数。按照用户的参数请求返回内容。
	propfind, _, err := dav.ReadPropfind(request.Body)
	this.PanicError(err)

	//寻找符合条件的matter.
	var matter *Matter
	//如果是空或者/就是请求根目录
	if subPath == "" || subPath == "/" {
		matter = NewRootMatter(user)
	} else {
		matter = this.matterDao.checkByUserUuidAndPath(user.Uuid, subPath)
	}

	matters := this.matterDao.List(matter.Uuid, user.Uuid, nil)

	if len(matters) == 0 {
		this.PanicNotFound("%s不存在", subPath)
	}

	//将当前的matter添加到头部
	matters = append([]*Matter{matter}, matters...)

	//准备一个输出结果的Writer
	multiStatusWriter := dav.MultiStatusWriter{Writer: writer}

	for _, matter := range matters {

		fmt.Printf("开始分析 %s\n", matter.Path)

		propstats := this.Propstats(matter, propfind)
		path := fmt.Sprintf("%s%s", WEBDAV_PREFFIX, matter.Path)
		response := this.makePropstatResponse(path, propstats)

		err = multiStatusWriter.Write(response)
		this.PanicError(err)
	}

	//闭合
	err = multiStatusWriter.Close()
	this.PanicError(err)

	fmt.Printf("%v %v \n", subPath, propfind.Prop)

}
