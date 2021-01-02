package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/dav"
	"github.com/eyebluecn/tank/code/tool/dav/xml"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/eyebluecn/tank/code/tool/webdav"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
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
	lockSystem    webdav.LockSystem
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

	// init the webdav lock system.
	this.lockSystem = webdav.NewMemLS()
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

	var okProperties []dav.Property
	var notFoundProperties []dav.Property

	for _, xmlName := range xmlNames {
		//TODO: deadprops not implement yet.

		// Otherwise, it must either be a live property or we don't know it.
		if liveProp := LivePropMap[xmlName]; liveProp.findFn != nil && (liveProp.dir || !matter.Dir) {
			innerXML := liveProp.findFn(user, matter)

			okProperties = append(okProperties, dav.Property{
				XMLName:  xmlName,
				InnerXML: []byte(innerXML),
			})
		} else {
			this.logger.Info("handle props %s %s.", matter.Path, xmlName.Local)

			propMap := matter.FetchPropMap()
			if value, isPresent := propMap[xmlName.Local]; isPresent {
				okProperties = append(okProperties, dav.Property{
					XMLName:  xmlName,
					InnerXML: []byte(value),
				})
			} else {

				//only accept Space not null.
				if xmlName.Space != "" {

					//collect not found props
					notFoundProperties = append(notFoundProperties, dav.Property{
						XMLName:  xmlName,
						InnerXML: []byte(""),
					})
				}

			}
		}
	}

	if len(okProperties) == 0 && len(notFoundProperties) == 0 {
		panic(result.BadRequest("cannot parse request properties"))
	}

	if len(okProperties) != 0 {
		okPropstat := dav.Propstat{Status: http.StatusOK, Props: okProperties}
		propstats = append(propstats, okPropstat)
	}

	if len(notFoundProperties) != 0 {
		notFoundPropstat := dav.Propstat{Status: http.StatusNotFound, Props: notFoundProperties}
		propstats = append(propstats, notFoundPropstat)
	}

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

	// read depth
	depth := this.ParseDepth(request)

	propfind := dav.ReadPropfind(request.Body)

	//find the matter, if subPath is null, means the root directory.
	matter := this.matterDao.CheckWithRootByPath(subPath, user)

	var matters []*Matter
	if depth == 0 {
		matters = []*Matter{matter}
	} else {
		// len(matters) == 0 means empty directory
		matters = this.matterDao.FindByPuuidAndUserUuidAndDeleted(matter.Uuid, user.Uuid, FALSE, nil)

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

//change the file's property
func (this *DavService) HandleProppatch(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("PROPPATCH %s\n", subPath)

	// handle the lock feature.
	reqPath, status, err := this.stripPrefix(request.URL.Path)
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	release, status, err := this.confirmLocks(request, reqPath, "")
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	if release != nil {
		defer release()
	}

	matter := this.matterDao.checkByUserUuidAndPath(user.Uuid, subPath)

	patches, status, err := webdav.ReadProppatch(request.Body)
	this.PanicError(err)

	fmt.Println("status:%v", status)

	//prepare a multiStatusWriter.
	multiStatusWriter := &dav.MultiStatusWriter{Writer: writer}

	propstats := make([]dav.Propstat, 0)
	propMap := matter.FetchPropMap()
	for _, patch := range patches {
		propStat := dav.Propstat{Status: http.StatusOK}
		if patch.Remove {

			if len(patch.Props) > 0 {
				property := patch.Props[0]
				if _, isPresent := propMap[property.XMLName.Local]; isPresent {
					//delete the prop.
					delete(propMap, property.XMLName.Local)
				}
			}
		} else {
			for _, prop := range patch.Props {
				propMap[prop.XMLName.Local] = string(prop.InnerXML)

				propStat.Props = append(propStat.Props, dav.Property{XMLName: xml.Name{Space: prop.XMLName.Space, Local: prop.XMLName.Local}})

			}
		}

		propstats = append(propstats, propStat)

	}
	matter.SetPropMap(propMap)
	// update the matter
	this.matterDao.Save(matter)

	visitPath := fmt.Sprintf("%s%s", WEBDAV_PREFIX, matter.Path)
	response := this.makePropstatResponse(visitPath, propstats)

	err1 := multiStatusWriter.Write(response)
	this.PanicError(err1)

	err2 := multiStatusWriter.Close()
	this.PanicError(err2)

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

	// handle the lock feature.
	reqPath, status, err := this.stripPrefix(request.URL.Path)
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	release, status, err := this.confirmLocks(request, reqPath, "")
	if err != nil {

		//if status == http.StatusLocked {
		//	status = http.StatusPreconditionFailed
		//}
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	if release != nil {
		defer release()
	}

	filename := util.GetFilenameOfPath(subPath)
	dirPath := util.GetDirOfPath(subPath)

	dirMatter := this.matterDao.CheckWithRootByPath(dirPath, user)

	//if exist delete it.
	srcMatter := this.matterDao.findByUserUuidAndPath(user.Uuid, subPath)
	if srcMatter != nil {
		this.matterService.AtomicDelete(request, srcMatter, user)
	}

	this.matterService.Upload(request, request.Body, user, dirMatter, filename, true)

	//set the status code 201
	writer.WriteHeader(http.StatusCreated)

}

//delete file
func (this *DavService) HandleDelete(w http.ResponseWriter, r *http.Request, user *User, subPath string) {

	fmt.Printf("DELETE %s\n", subPath)

	reqPath, status, err := this.stripPrefix(r.URL.Path)
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	release, status, err := this.confirmLocks(r, reqPath, "")
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	if release != nil {
		defer release()
	}

	matter := this.matterDao.CheckWithRootByPath(subPath, user)

	this.matterService.AtomicDelete(r, matter, user)
}

//crate a directory
func (this *DavService) HandleMkcol(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("MKCOL %s\n", subPath)

	//the body of MKCOL request MUST be empty. (RFC2518:8.3.1)
	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println("occur error when reading body: " + err.Error())
	} else {
		if len(bodyBytes) != 0 {
			//throw conflict error
			panic(result.CustomWebResult(result.UNSUPPORTED_MEDIA_TYPE, fmt.Sprintf("%s MKCOL should NO body", subPath)))
		}
	}

	thisDirName := util.GetFilenameOfPath(subPath)
	dirPath := util.GetDirOfPath(subPath)

	dirMatter := this.matterDao.FindWithRootByPath(dirPath, user)
	if dirMatter == nil {
		//throw conflict error
		panic(result.CustomWebResult(result.CONFLICT, fmt.Sprintf("%s not exist", dirPath)))
	}

	//check whether col exists. (RFC2518:8.3.1)
	dbMatter := this.matterDao.FindByUserUuidAndPuuidAndDirAndName(user.Uuid, dirMatter.Uuid, TRUE, thisDirName)
	if dbMatter != nil {
		panic(result.CustomWebResult(result.METHOD_NOT_ALLOWED, fmt.Sprintf("%s already exists", dirPath)))
	}

	//check whether file exists. (RFC2518:8.3.1)
	fileMatter := this.matterDao.FindByUserUuidAndPuuidAndDirAndName(user.Uuid, dirMatter.Uuid, FALSE, thisDirName)
	if fileMatter != nil {
		panic(result.CustomWebResult(result.METHOD_NOT_ALLOWED, fmt.Sprintf("%s file already exists", dirPath)))
	}

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

	destDirMatter = this.matterDao.FindWithRootByPath(destinationDirPath, user)
	if destDirMatter == nil {
		//throw conflict error
		panic(result.CustomWebResult(result.CONFLICT, fmt.Sprintf("%s not exist", destinationDirPath)))
	}

	return srcMatter, destDirMatter, srcDirPath, destinationDirPath, destinationName, overwrite

}

//move or rename.
func (this *DavService) HandleMove(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("MOVE %s\n", subPath)

	// handle the lock feature.
	reqPath, status, err := this.stripPrefix(request.URL.Path)
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	release, status, err := this.confirmLocks(request, reqPath, "")
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	if release != nil {
		defer release()
	}

	srcMatter, destDirMatter, srcDirPath, destinationDirPath, destinationName, overwrite := this.prepareMoveCopy(writer, request, user, subPath)

	//move to the new directory
	if destinationDirPath == srcDirPath {
		//if destination path not change. it means rename.
		this.matterService.AtomicRename(request, srcMatter, destinationName, overwrite, user)
	} else {
		this.matterService.AtomicMove(request, srcMatter, destDirMatter, overwrite, user)
	}

	this.logger.Info("finish moving %s => %s", subPath, destDirMatter.Path)

	if overwrite {
		//overwrite old. set the status code 204
		writer.WriteHeader(http.StatusNoContent)
	} else {
		//copy new. set the status code 201
		writer.WriteHeader(http.StatusCreated)
	}
}

//copy file/directory
func (this *DavService) HandleCopy(writer http.ResponseWriter, request *http.Request, user *User, subPath string) {

	fmt.Printf("COPY %s\n", subPath)

	srcMatter, destDirMatter, _, _, destinationName, overwrite := this.prepareMoveCopy(writer, request, user, subPath)

	// handle the lock feature.
	release, status, err := this.confirmLocks(request, destDirMatter.Path+"/"+destinationName, "")
	if err != nil {
		panic(result.StatusCodeWebResult(status, err.Error()))
	}
	if release != nil {
		defer release()
	}

	//copy to the new directory
	this.matterService.AtomicCopy(request, srcMatter, destDirMatter, destinationName, overwrite, user)

	this.logger.Info("finish copying %s => %s", subPath, destDirMatter.Path)

	if overwrite {
		//overwrite old. set the status code 204
		writer.WriteHeader(http.StatusNoContent)
	} else {
		//copy new. set the status code 201
		writer.WriteHeader(http.StatusCreated)
	}

}

func (h *DavService) stripPrefix(p string) (string, int, error) {
	if r := strings.TrimPrefix(p, WEBDAV_PREFIX); len(r) < len(p) {
		return r, http.StatusOK, nil
	}
	return p, http.StatusNotFound, webdav.ErrPrefixMismatch
}

func (h *DavService) lock(now time.Time, root string) (token string, status int, err error) {
	token, err = h.lockSystem.Create(now, webdav.LockDetails{
		Root:      root,
		Duration:  webdav.InfiniteTimeout,
		ZeroDepth: true,
	})
	if err != nil {
		if err == webdav.ErrLocked {
			return "", webdav.StatusLocked, err
		}
		return "", http.StatusInternalServerError, err
	}
	return token, 0, nil
}

func (h *DavService) confirmLocks(r *http.Request, src, dst string) (release func(), status int, err error) {
	hdr := r.Header.Get("If")
	if hdr == "" {
		// An empty If header means that the client hasn't previously created locks.
		// Even if this client doesn't care about locks, we still need to check that
		// the resources aren't locked by another client, so we create temporary
		// locks that would conflict with another client's locks. These temporary
		// locks are unlocked at the end of the HTTP request.
		now, srcToken, dstToken := time.Now(), "", ""
		if src != "" {
			srcToken, status, err = h.lock(now, src)
			if err != nil {
				return nil, status, err
			}
		}
		if dst != "" {
			dstToken, status, err = h.lock(now, dst)
			if err != nil {
				if srcToken != "" {
					h.lockSystem.Unlock(now, srcToken)
				}
				return nil, status, err
			}
		}

		return func() {
			if dstToken != "" {
				h.lockSystem.Unlock(now, dstToken)
			}
			if srcToken != "" {
				h.lockSystem.Unlock(now, srcToken)
			}
		}, 0, nil
	}

	ih, ok := webdav.ParseIfHeader(hdr)
	if !ok {
		return nil, http.StatusBadRequest, webdav.ErrInvalidIfHeader
	}
	// ih is a disjunction (OR) of ifLists, so any IfList will do.
	for _, l := range ih.Lists {
		lsrc := l.ResourceTag
		if lsrc == "" {
			lsrc = src
		} else {
			u, err := url.Parse(lsrc)
			if err != nil {
				continue
			}
			if u.Host != r.Host {
				continue
			}
			lsrc, status, err = h.stripPrefix(u.Path)
			if err != nil {
				return nil, status, err
			}
		}
		release, err = h.lockSystem.Confirm(time.Now(), lsrc, dst, l.Conditions...)
		if err == webdav.ErrConfirmationFailed {
			continue
		}
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return release, 0, nil
	}
	// Section 10.4.1 says that "If this header is evaluated and all state lists
	// fail, then the request must fail with a 412 (Precondition Failed) status."
	// We follow the spec even though the cond_put_corrupt_token test case from
	// the litmus test warns on seeing a 412 instead of a 423 (Locked).
	return nil, http.StatusLocked, webdav.ErrLocked
}

//lock.
func (this *DavService) HandleLock(w http.ResponseWriter, r *http.Request, user *User, subPath string) {

	duration, err := webdav.ParseTimeout(r.Header.Get("Timeout"))
	if err != nil {
		panic(result.BadRequest(err.Error()))
	}
	li, status, err := webdav.ReadLockInfo(r.Body)
	if err != nil {
		panic(result.BadRequest(fmt.Sprintf("error:%s, status=%d", err.Error(), status)))
	}

	token, ld, now, created := "", webdav.LockDetails{}, time.Now(), false
	if li == (webdav.LockInfo{}) {
		// An empty LockInfo means to refresh the lock.
		ih, ok := webdav.ParseIfHeader(r.Header.Get("If"))
		if !ok {
			panic(result.BadRequest(webdav.ErrInvalidIfHeader.Error()))
		}
		if len(ih.Lists) == 1 && len(ih.Lists[0].Conditions) == 1 {
			token = ih.Lists[0].Conditions[0].Token
		}
		if token == "" {
			panic(result.BadRequest(webdav.ErrInvalidLockToken.Error()))
		}
		ld, err = this.lockSystem.Refresh(now, token, duration)
		if err != nil {
			if err == webdav.ErrNoSuchLock {
				panic(result.StatusCodeWebResult(http.StatusPreconditionFailed, err.Error()))
			}
			panic(result.StatusCodeWebResult(http.StatusInternalServerError, err.Error()))
		}

	} else {
		// Section 9.10.3 says that "If no Depth header is submitted on a LOCK request,
		// then the request MUST act as if a "Depth:infinity" had been submitted."
		depth := webdav.InfiniteDepth
		if hdr := r.Header.Get("Depth"); hdr != "" {
			depth = webdav.ParseDepth(hdr)
			if depth != 0 && depth != webdav.InfiniteDepth {
				// Section 9.10.3 says that "Values other than 0 or infinity must not be
				// used with the Depth header on a LOCK method".
				panic(result.StatusCodeWebResult(http.StatusBadRequest, webdav.ErrInvalidDepth.Error()))
			}
		}

		reqPath, status, err := this.stripPrefix(r.URL.Path)
		if err != nil {
			panic(result.StatusCodeWebResult(status, err.Error()))
		}

		ld = webdav.LockDetails{
			Root:      reqPath,
			Duration:  duration,
			OwnerXML:  li.Owner.InnerXML,
			ZeroDepth: depth == 0,
		}
		token, err = this.lockSystem.Create(now, ld)
		if err != nil {
			if err == webdav.ErrLocked {
				panic(result.StatusCodeWebResult(http.StatusLocked, err.Error()))
			}
			panic(result.StatusCodeWebResult(http.StatusInternalServerError, err.Error()))
		}
		defer func() {
			//when error occur, rollback.
			//this.lockSystem.Unlock(now, token)
		}()

		// Create the resource if it didn't previously exist.
		// ctx := r.Context()
		//if _, err := this.FileSystem.Stat(ctx, subPath); err != nil {
		//	f, err := h.FileSystem.OpenFile(ctx, reqPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		//	if err != nil {
		//		// TODO: detect missing intermediate dirs and return http.StatusConflict?
		//		return http.StatusInternalServerError, err
		//	}
		//	f.Close()
		//	created = true
		//}

		// http://www.webdav.org/specs/rfc4918.html#HEADER_Lock-Token says that the
		// Lock-Token value is a Coded-URL. We add angle brackets.
		w.Header().Set("Lock-Token", "<"+token+">")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	if created {
		// This is "w.WriteHeader(http.StatusCreated)" and not "return
		// http.StatusCreated, nil" because we write our own (XML) response to w
		// and Handler.ServeHTTP would otherwise write "Created".
		w.WriteHeader(http.StatusCreated)
	}
	_, _ = webdav.WriteLockInfo(w, token, ld)

}

//unlock
func (this *DavService) HandleUnlock(w http.ResponseWriter, r *http.Request, user *User, subPath string) {

	// http://www.webdav.org/specs/rfc4918.html#HEADER_Lock-Token says that the
	// Lock-Token value is a Coded-URL. We strip its angle brackets.
	t := r.Header.Get("Lock-Token")
	if len(t) < 2 || t[0] != '<' || t[len(t)-1] != '>' {
		panic(result.StatusCodeWebResult(http.StatusBadRequest, webdav.ErrInvalidLockToken.Error()))
	}
	t = t[1 : len(t)-1]

	switch err := this.lockSystem.Unlock(time.Now(), t); err {
	case nil:
		panic(result.StatusCodeWebResult(http.StatusNoContent, ""))
	case webdav.ErrForbidden:
		panic(result.StatusCodeWebResult(http.StatusForbidden, err.Error()))
	case webdav.ErrLocked:
		panic(result.StatusCodeWebResult(http.StatusLocked, err.Error()))
	case webdav.ErrNoSuchLock:
		panic(result.StatusCodeWebResult(http.StatusConflict, err.Error()))
	default:
		panic(result.StatusCodeWebResult(http.StatusInternalServerError, err.Error()))
	}
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
