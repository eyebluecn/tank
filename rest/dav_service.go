package rest

import (
	"net/http"
)

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

const (
	infiniteDepth = -1
	invalidDepth  = -2
)

// parseDepth maps the strings "0", "1" and "infinity" to 0, 1 and
// infiniteDepth. Parsing any other string returns invalidDepth.
//
// Different WebDAV methods have further constraints on valid depths:
//	- PROPFIND has no further restrictions, as per section 9.1.
//	- COPY accepts only "0" or "infinity", as per section 9.8.3.
//	- MOVE accepts only "infinity", as per section 9.9.2.
//	- LOCK accepts only "0" or "infinity", as per section 9.10.3.
// These constraints are enforced by the handleXxx methods.
func parseDepth(s string) int {
	switch s {
	case "0":
		return 0
	case "1":
		return 1
	case "infinity":
		return infiniteDepth
	}
	return invalidDepth
}


//处理 方法
func (this *DavService) HandlePropfind(w http.ResponseWriter, r *http.Request) {

	//basePath := "/Users/fusu/d/group/golang/src/tank/tmp/dav"
	//
	//reqPath := r.URL.Path
	//
	//ctx := r.Context()
	//
	//fi, err := os.Stat(basePath + reqPath)
	//if err != nil {
	//	this.PanicError(err)
	//}
	//
	//depth := infiniteDepth
	//if hdr := r.Header.Get("Depth"); hdr != "" {
	//	depth = parseDepth(hdr)
	//	if depth == invalidDepth {
	//		this.PanicBadRequest("Depth指定错误！")
	//	}
	//}
	//
	//pf, status, err := readPropfind(r.Body)
	//if err != nil {
	//	return status, err
	//}
	//
	//mw := multistatusWriter{w: w}
	//
	//walkFn := func(reqPath string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//	var pstats []Propstat
	//	if pf.Propname != nil {
	//		pnames, err := propnames(ctx, h.FileSystem, h.LockSystem, reqPath)
	//		if err != nil {
	//			return err
	//		}
	//		pstat := Propstat{Status: http.StatusOK}
	//		for _, xmlname := range pnames {
	//			pstat.Props = append(pstat.Props, Property{XMLName: xmlname})
	//		}
	//		pstats = append(pstats, pstat)
	//	} else if pf.Allprop != nil {
	//		pstats, err = allprop(ctx, h.FileSystem, h.LockSystem, reqPath, pf.Prop)
	//	} else {
	//		pstats, err = props(ctx, h.FileSystem, h.LockSystem, reqPath, pf.Prop)
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	href := path.Join(h.Prefix, reqPath)
	//	if info.IsDir() {
	//		href += "/"
	//	}
	//	return mw.write(makePropstatResponse(href, pstats))
	//}
	//
	//walkErr := walkFS(ctx, h.FileSystem, depth, reqPath, fi, walkFn)
	//closeErr := mw.close()
	//if walkErr != nil {
	//	return http.StatusInternalServerError, walkErr
	//}
	//if closeErr != nil {
	//	return http.StatusInternalServerError, closeErr
	//}
	//return 0, nil

}
