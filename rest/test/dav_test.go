package test

import (
	"bytes"
	"tank/rest/dav"
	"tank/rest/dav/xml"
	"testing"
	"time"
)

func TestReadPropfind(t *testing.T) {

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
			t.Errorf("index = %s error", k)
		}
	}

	t.Logf("[%v] pass!", time.Now())

}
