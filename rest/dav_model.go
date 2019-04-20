package rest

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"tank/rest/dav"
)

//访问前缀，这个是特殊入口
var WEBDAV_PREFFIX = "/api/dav"

//动态的文件属性
type LiveProp struct {
	findFn func(matter *Matter) string
	dir    bool
}

//所有的动态属性定义及其值的获取方式
var LivePropMap = map[xml.Name]LiveProp{
	{Space: "DAV:", Local: "resourcetype"}: {
		findFn: func(matter *Matter) string {
			if matter.Dir {
				return `<D:collection xmlns:D="DAV:"/>`
			} else {
				return ""
			}
		},
		dir: true,
	},
	{Space: "DAV:", Local: "displayname"}: {
		findFn: func(matter *Matter) string {
			if dav.SlashClean(matter.Name) == "/" {
				return ""
			} else {
				return dav.EscapeXML(matter.Name)
			}
		},
		dir: true,
	},
	{Space: "DAV:", Local: "getcontentlength"}: {
		findFn: func(matter *Matter) string {
			return strconv.FormatInt(matter.Size, 10)
		},
		dir: false,
	},
	{Space: "DAV:", Local: "getlastmodified"}: {
		findFn: func(matter *Matter) string {
			return matter.UpdateTime.UTC().Format(http.TimeFormat)
		},
		// http://webdav.org/specs/rfc4918.html#PROPERTY_getlastmodified
		// suggests that getlastmodified should only apply to GETable
		// resources, and this package does not support GET on directories.
		//
		// Nonetheless, some WebDAV clients expect child directories to be
		// sortable by getlastmodified date, so this value is true, not false.
		// See golang.org/issue/15334.
		dir: true,
	},
	{Space: "DAV:", Local: "creationdate"}: {
		findFn: nil,
		dir:    false,
	},
	{Space: "DAV:", Local: "getcontentlanguage"}: {
		findFn: nil,
		dir:    false,
	},
	{Space: "DAV:", Local: "getcontenttype"}: {
		findFn: func(matter *Matter) string {
			if matter.Dir {
				return ""
			} else {
				return dav.EscapeXML(matter.Name)
			}
		},
		dir: false,
	},
	{Space: "DAV:", Local: "getetag"}: {
		findFn: func(matter *Matter) string {
			return fmt.Sprintf(`"%x%x"`, matter.UpdateTime.UnixNano(), matter.Size)
		},
		// findETag implements ETag as the concatenated hex values of a file's
		// modification time and size. This is not a reliable synchronization
		// mechanism for directories, so we do not advertise getetag for DAV
		// collections.
		dir: false,
	},
	// TODO: The lockdiscovery property requires LockSystem to list the
	// active locks on a resource.
	{Space: "DAV:", Local: "lockdiscovery"}: {},
	{Space: "DAV:", Local: "supportedlock"}: {
		findFn: func(matter *Matter) string {
			return `` +
				`<D:lockentry xmlns:D="DAV:">` +
				`<D:lockscope><D:exclusive/></D:lockscope>` +
				`<D:locktype><D:write/></D:locktype>` +
				`</D:lockentry>`
		},
		dir: true,
	},
}
