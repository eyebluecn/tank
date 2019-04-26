package test

import (
	"bytes"
	"github.com/eyebluecn/tank/code/tool/dav"
	"github.com/eyebluecn/tank/code/tool/dav/xml"
	"testing"
	"time"
)

func TestXmlDecoder(t *testing.T) {

	propfind := &dav.Propfind{}

	str := `
		<?xml version="1.0" encoding="utf-8" ?>
		<D:propfind xmlns:D="DAV:">
			<D:prop>
				<D:resourcetype />
				<D:getcontentlength />
				<D:creationdate />
				<D:getlastmodified />
			</D:prop>
		</D:propfind>
		`

	reader := bytes.NewReader([]byte(str))

	err := xml.NewDecoder(reader).Decode(propfind)
	if err != nil {
		t.Error(err.Error())
	}

	resultMap := make(map[string]bool)

	resultMap[`propfind.XMLName.Space == "DAV:"`] = propfind.XMLName.Space == "DAV:"

	resultMap[`propfind.XMLName.Local == "propfind"`] = propfind.XMLName.Local == "propfind"

	resultMap[`len(propfind.Prop) == 4`] = len(propfind.Prop) == 4

	resultMap[`propfind.Prop[0]`] = propfind.Prop[0].Space == "DAV:" && propfind.Prop[0].Local == "resourcetype"
	resultMap[`propfind.Prop[1]`] = propfind.Prop[1].Space == "DAV:" && propfind.Prop[1].Local == "getcontentlength"
	resultMap[`propfind.Prop[2]`] = propfind.Prop[2].Space == "DAV:" && propfind.Prop[2].Local == "creationdate"
	resultMap[`propfind.Prop[3]`] = propfind.Prop[3].Space == "DAV:" && propfind.Prop[3].Local == "getlastmodified"

	for k, v := range resultMap {
		if !v {
			t.Errorf(" %s error", k)
		}
	}

	t.Logf("[%v] pass!", time.Now())

}

func TestXmlEncoder(t *testing.T) {

	writer := &bytes.Buffer{}

	response := &dav.Response{
		XMLName: xml.Name{Space: "DAV:", Local: "response"},
		Href:    []string{"/api/dav"},
		Propstat: []dav.SubPropstat{
			{
				Prop: []dav.Property{
					{
						XMLName:  xml.Name{Space: "DAV:", Local: "resourcetype"},
						InnerXML: []byte(`<D:collection xmlns:D="DAV:"/>`),
					},
					{
						XMLName:  xml.Name{Space: "DAV:", Local: "getlastmodified"},
						InnerXML: []byte(`Mon, 22 Apr 2019 06:38:36 GMT`),
					},
				},
				Status: "HTTP/1.1 200 OK",
			},
		},
	}

	err := xml.NewEncoder(writer).Encode(response)

	if err != nil {
		t.Error(err.Error())
	}

	bs := writer.Bytes()

	str := string(bs)

	resultMap := make(map[string]bool)

	resultMap["equal"] = str == `<D:response><D:href>/api/dav</D:href><D:propstat><D:prop><D:resourcetype><D:collection xmlns:D="DAV:"/></D:resourcetype><D:getlastmodified>Mon, 22 Apr 2019 06:38:36 GMT</D:getlastmodified></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response>`

	for k, v := range resultMap {
		if !v {
			t.Errorf("%s error", k)
		}
	}

	t.Logf("[%v] pass!", time.Now())

}
