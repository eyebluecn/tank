package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/dav"
	"github.com/eyebluecn/tank/code/tool/dav/xml"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

/**
 *
 * WebDav document
 * https://tools.ietf.org/html/rfc4918
 * refer: golang.org/x/net/webdav
 * test machine: http://www.webdav.org/neon/litmus/
 */
//@Service
type DavService struct {
	BaseBean
	matterDao     *MatterDao
	matterService *MatterService
}

func (this *DavService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

}

//get the depth in header. Not support infinity yet.
func (this *DavService) ParseDepth(request *http.Request) int {

	depth := 1
	if hdr := request.Header.Get("Depth"); hdr != "" {
		switch hdr {
		case "0":
			return 0
		case "1":
			return 1
		case "infinity":
			return 1
		}
	} else {
		panic(result.BadRequest("Header Depth cannot be null"))
	}
	return depth
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

//fetch a matter's []dav.Propstat
func (this *DavService) PropstatsFromXmlNames(user *User, matter *Matter, xmlNames []xml.Name) []dav.Propstat {

	propstats := make([]dav.Propstat, 0)

	var properties []dav.Property

	for _, xmlName := range xmlNames {
		//TODO: deadprops not implement yet.

		// Otherwise, it must either be a live property or we don't know it.
		if liveProp := LivePropMap[xmlName]; liveProp.findFn != nil && (liveProp.dir || !matter.Dir) {
			innerXML := liveProp.findFn(user, matter)

			properties = append(properties, dav.Property{
				XMLName:  xmlName,
				InnerXML: []byte(innerXML),
			})
		} else {
			this.logger.Info("%s %s cannot finish.", matter.Path, xmlName.Local)
		}
	}

	if len(properties) == 0 {
		panic(result.BadRequest("cannot parse request properties"))
	}

	okPropstat := dav.Propstat{Status: http.StatusOK, Props: properties}

	propstats = append(propstats, okPropstat)

	return propstats

}

func (this *DavService) AllPropXmlNames(matter *Matter) []xml.Name {

	pnames := make([]xml.Name, 0)
	for pn, prop := range LivePropMap {
		if prop.findFn != nil && (prop.dir || !matter.Dir) {
			pnames = append(pnames, pn)
		}
	}

	return pnames
}

func (this *DavService) Propstats(user *User, matter *Matter, propfind *dav.Propfind) []dav.Propstat {

	propstats := make([]dav.Propstat, 0)
	if propfind.Propname != nil {
		panic(result.BadRequest("TODO: propfind.Propname != nil "))
	} else if propfind.Allprop != nil {

		//TODO: if include other things. add to it.
		xmlNames := this.AllPropXmlNames(matter)

		propstats = this.PropstatsFromXmlNames(user, matter, xmlNames)

	} else {
		propstats = this.PropstatsFromXmlNames(user, matter, propfind.Prop)
	}

	return propstats

}

//list the directory.
func (this *DavService) HandlePropfind(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("PROPFIND %s\n", subPath)

	depth := this.ParseDepth(request)

	propfind := dav.ReadPropfind(request.Body)

	//find the matter, if subPath is null, means the root directory.
	matter := this.matterDao.CheckWithRootByPath(subPath, user)

	var matters []*Matter
	if depth == 0 {
		matters = []*Matter{matter}
	} else {
		// len(matters) == 0 means empty directory
		matters = this.matterDao.FindByPuuidAndUserUuid(matter.Uuid, user.Uuid, nil)

		//add this matter to head.
		matters = append([]*Matter{matter}, matters...)
	}

	//prepare a multiStatusWriter.
	multiStatusWriter := &dav.MultiStatusWriter{Writer: writer}

	for _, matter := range matters {

		fmt.Printf("handle Matter %s\n", matter.Path)

		propstats := this.Propstats(user, matter, propfind)
		visitPath := fmt.Sprintf("%s%s", WEBDAV_PREFIX, matter.Path)
		response := this.makePropstatResponse(visitPath, propstats)

		err := multiStatusWriter.Write(response)
		this.PanicError(err)
	}

	err := multiStatusWriter.Close()
	this.PanicError(err)

	fmt.Printf("%v %v \n", subPath, propfind.Prop)

}

//handle download
func (this *DavService) HandleGetHeadPost(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("GET %s\n", subPath)

	matter := this.matterDao.CheckWithRootByPath(subPath, user)

	//if this is a Directory, it means Propfind
	if matter.Dir {
		this.HandlePropfind(writer, request, user, subPath)
		return
	}

	//download a file.
	this.matterService.DownloadFile(writer, request, matter.AbsolutePath(), matter.Name, false)

}

//upload a file
func (this *DavService) HandlePut(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("PUT %s\n", subPath)

	filename := util.GetFilenameOfPath(subPath)
	dirPath := util.GetDirOfPath(subPath)

	dirMatter := this.matterDao.CheckWithRootByPath(dirPath, user)

	//if exist delete it.
	srcMatter := this.matterDao.findByUserUuidAndPath(user.Uuid, subPath)
	if srcMatter != nil {
		this.matterService.AtomicDelete(request, srcMatter, user)
	}

	this.matterService.Upload(request, request.Body, user, dirMatter, filename, true)

}

//delete file
func (this *DavService) HandleDelete(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("DELETE %s\n", subPath)

	matter := this.matterDao.CheckWithRootByPath(subPath, user)

	this.matterService.AtomicDelete(request, matter, user)
}

//crate a directory
func (this *DavService) HandleMkcol(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("MKCOL %s\n", subPath)

	thisDirName := util.GetFilenameOfPath(subPath)
	dirPath := util.GetDirOfPath(subPath)

	dirMatter := this.matterDao.CheckWithRootByPath(dirPath, user)

	this.matterService.AtomicCreateDirectory(request, dirMatter, thisDirName, user)

}

//cors options
func (this *DavService) HandleOptions(w http.ResponseWriter, r *http.Request, user *User, subPath string) {

	fmt.Printf("OPTIONS %s\n", subPath)

	matter := this.matterDao.CheckWithRootByPath(subPath, user)

	allow := "OPTIONS, LOCK, PUT, MKCOL"
	if matter.Dir {
		allow = "OPTIONS, LOCK, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND"
	} else {
		allow = "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND, PUT"
	}

	w.Header().Set("Allow", allow)
	// http://www.webdav.org/specs/rfc4918.html#dav.compliance.classes
	w.Header().Set("DAV", "1, 2")
	// http://msdn.microsoft.com/en-au/library/cc250217.aspx
	w.Header().Set("MS-Author-Via", "DAV")

}

//prepare for moving or copying
func (this *DavService) prepareMoveCopy(
	writer http.ResponseWriter,
	request *http.Request,
	user *User, subPath string) (
	srcMatter *Matter,
	destDirMatter *Matter,
	srcDirPath string,
	destinationDirPath string,
	destinationName string,
	overwrite bool) {

	//parse the destination.
	destinationStr := request.Header.Get("Destination")

	//parse Overwrite。
	overwriteStr := request.Header.Get("Overwrite")

	//destination path with prefix
	var fullDestinationPath string
	//destination path without prefix
	var destinationPath string

	if destinationStr == "" {
		panic(result.BadRequest("Header Destination cannot be null"))
	}

	//if rename. not start with http
	if strings.HasPrefix(destinationStr, WEBDAV_PREFIX) {
		fullDestinationPath = destinationStr
	} else {
		destinationUrl, err := url.Parse(destinationStr)
		this.PanicError(err)
		if destinationUrl.Host != request.Host {
			panic(result.BadRequest("Destination Host not the same. %s  %s != %s", destinationStr, destinationUrl.Host, request.Host))
		}
		fullDestinationPath = destinationUrl.Path
	}

	//clean the relative path. eg. /a/b/../ => /a/
	fullDestinationPath = path.Clean(fullDestinationPath)

	//clean the prefix
	pattern := fmt.Sprintf(`^%s(.*)$`, WEBDAV_PREFIX)
	reg := regexp.MustCompile(pattern)
	strs := reg.FindStringSubmatch(fullDestinationPath)
	if len(strs) == 2 {
		destinationPath = strs[1]
	} else {
		panic(result.BadRequest("destination prefix must be %s", WEBDAV_PREFIX))
	}

	destinationName = util.GetFilenameOfPath(destinationPath)
	destinationDirPath = util.GetDirOfPath(destinationPath)
	srcDirPath = util.GetDirOfPath(subPath)

	overwrite = false
	if overwriteStr == "T" {
		overwrite = true
	}

	//if not change return.
	if destinationPath == subPath {
		return
	}

	//source matter
	srcMatter = this.matterDao.CheckWithRootByPath(subPath, user)

	//if source matter is root.
	if srcMatter.Uuid == MATTER_ROOT {
		panic(result.BadRequest("you cannot move the root directory"))
	}

	destDirMatter = this.matterDao.CheckWithRootByPath(destinationDirPath, user)

	return srcMatter, destDirMatter, srcDirPath, destinationDirPath, destinationName, overwrite

}

//move or rename.
func (this *DavService) HandleMove(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("MOVE %s\n", subPath)

	srcMatter, destDirMatter, srcDirPath, destinationDirPath, destinationName, overwrite := this.prepareMoveCopy(writer, request, user, subPath)
	//move to the new directory
	if destinationDirPath == srcDirPath {
		//if destination path not change. it means rename.
		this.matterService.AtomicRename(request, srcMatter, destinationName, user)
	} else {
		this.matterService.AtomicMove(request, srcMatter, destDirMatter, overwrite, user)
	}

	this.logger.Info("finish moving %s => %s", subPath, destDirMatter.Path)
}

//copy file/directory
func (this *DavService) HandleCopy(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("COPY %s\n", subPath)

	srcMatter, destDirMatter, _, _, destinationName, overwrite := this.prepareMoveCopy(writer, request, user, subPath)

	//copy to the new directory
	this.matterService.AtomicCopy(request, srcMatter, destDirMatter, destinationName, overwrite, user)

	this.logger.Info("finish copying %s => %s", subPath, destDirMatter.Path)

}

//lock.
func (this *DavService) HandleLock(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	panic(result.BadRequest("not support LOCK yet."))
}

//unlock
func (this *DavService) HandleUnlock(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	panic(result.BadRequest("not support UNLOCK yet."))
}

//change the file's property
func (this *DavService) HandleProppatch(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	panic(result.BadRequest("not support PROPPATCH yet."))
}

//hanle all the request.
func (this *DavService) HandleDav(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	method := request.Method
	if method == "OPTIONS" {

		//cors option
		this.HandleOptions(writer, request, user, subPath)

	} else if method == "GET" || method == "HEAD" || method == "POST" {

		//get the detail of file. download
		this.HandleGetHeadPost(writer, request, user, subPath)

	} else if method == "DELETE" {

		//delete file
		this.HandleDelete(writer, request, user, subPath)

	} else if method == "PUT" {

		//upload file
		this.HandlePut(writer, request, user, subPath)

	} else if method == "MKCOL" {

		//crate directory
		this.HandleMkcol(writer, request, user, subPath)

	} else if method == "COPY" {

		//copy file/directory
		this.HandleCopy(writer, request, user, subPath)

	} else if method == "MOVE" {

		//move/rename a file or directory
		this.HandleMove(writer, request, user, subPath)

	} else if method == "LOCK" {

		//lock
		this.HandleLock(writer, request, user, subPath)

	} else if method == "UNLOCK" {

		//unlock
		this.HandleUnlock(writer, request, user, subPath)

	} else if method == "PROPFIND" {

		//list a directory
		this.HandlePropfind(writer, request, user, subPath)

	} else if method == "PROPPATCH" {

		//change file's property.
		this.HandleProppatch(writer, request, user, subPath)

	} else {

		panic(result.BadRequest("not support %s yet.", method))

	}

}
