package rest

import (
	"encoding/xml"
)

/**
 *
 * WebDav协议文档
 * https://tools.ietf.org/html/rfc4918
 * http://www.webdav.org/specs/rfc4918.html
 *
 */

const (
	//有多少层展示多少层
	INFINITE_DEPTH = -1
)

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_propfind
//PROPFIND方法请求时POST BODY入参
type Propfind struct {
	XMLName xml.Name `xml:"D:propfind"`
	XmlNS   string   `xml:"xmlns:D,attr"`

	Allprop  *struct{}     `xml:"D:allprop"`
	Propname *struct{}     `xml:"D:propname"`
	Prop     PropfindProps `xml:"D:prop"`
	Include  PropfindProps `xml:"D:include"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_prop (for propfind)
type PropfindProps []xml.Name
