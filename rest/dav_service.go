package rest

import (
	"net/http"
	"os"
	"tank/rest/dav"
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

	basePath := "/Users/fusu/d/group/golang/src/tank/tmp/dav"

	fileSystem := dav.Dir("/Users/fusu/d/group/golang/src/tank/tmp/dav")
	lockSystem := dav.NewMemLS()

	reqPath := r.URL.Path

	ctx := r.Context()

	fi, err := os.Stat(basePath + reqPath)
	if err != nil {
		this.PanicError(err)
	}

	depth := infiniteDepth
	if hdr := r.Header.Get("Depth"); hdr != "" {
		depth = parseDepth(hdr)
		if depth == invalidDepth {
			this.PanicBadRequest("Depth指定错误！")
		}
	}

	pf, _, err := dav.ReadPropfind(r.Body)

	this.PanicError(err)

	mw := dav.MultiStatusWriter{Writer: w}

	walkFn := func(reqPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var pstats []dav.Propstat
		if pf.Propname != nil {
			pnames, err := dav.Propnames(ctx, fileSystem, lockSystem, reqPath)
			if err != nil {
				return err
			}
			pstat := dav.Propstat{Status: http.StatusOK}
			for _, xmlname := range pnames {
				pstat.Props = append(pstat.Props, dav.Property{XMLName: xmlname})
			}
			pstats = append(pstats, pstat)
		} else if pf.Allprop != nil {
			pstats, err = dav.Allprop(ctx, fileSystem, lockSystem, reqPath, pf.Prop)
		} else {
			pstats, err = dav.Props(ctx, fileSystem, lockSystem, reqPath, pf.Prop)
		}
		if err != nil {
			return err
		}
		href := reqPath
		if info.IsDir() {
			href += "/"
		}
		return mw.Write(dav.MakePropstatResponse(href, pstats))
	}

	walkErr := dav.WalkFS(ctx, fileSystem, depth, reqPath, fi, walkFn)
	closeErr := mw.Close()
	this.PanicError(walkErr)
	this.PanicError(closeErr)

}
