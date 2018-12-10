package rest

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//@Service
type MatterService struct {
	Bean
	matterDao *MatterDao
	userDao   *UserDao
}

//初始化方法
func (this *MatterService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//根据一个文件夹路径，找到最后一个文件夹的uuid，如果中途出错，返回err.
func (this *MatterService) GetDirUuid(userUuid string, dir string) string {

	if dir == "" {
		panic(`文件夹不能为空`)
	} else if dir[0:1] != "/" {
		panic(`文件夹必须以/开头`)
	} else if strings.Index(dir, "//") != -1 {
		panic(`文件夹不能出现连续的//`)
	} else if m, _ := regexp.MatchString(`[<>|*?\\]`, dir); m {
		panic(`文件夹中不能包含以下特殊符号：< > | * ? \`)
	}

	if dir == "/" {
		return "root"
	}

	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}

	//递归找寻文件的上级目录uuid.
	folders := strings.Split(dir, "/")

	if len(folders) > 32 {
		panic("文件夹最多32层。")
	}

	puuid := "root"
	for k, name := range folders {

		if len(name) > 200 {
			panic("每级文件夹的最大长度为200")
		}

		if k == 0 {
			continue
		}

		matter := this.matterDao.FindByUserUuidAndPuuidAndNameAndDirTrue(userUuid, puuid, name)
		if matter == nil {
			//创建一个文件夹。这里一般都是通过alien接口来创建的文件夹。
			matter = &Matter{
				Puuid:    puuid,
				UserUuid: userUuid,
				Dir:      true,
				Alien:    true,
				Name:     name,
			}
			matter = this.matterDao.Create(matter)
		}

		puuid = matter.Uuid
	}

	return puuid
}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *MatterService) Detail(uuid string) *Matter {

	matter := this.matterDao.CheckByUuid(uuid)

	//组装file的内容，展示其父组件。
	puuid := matter.Puuid
	tmpMatter := matter
	for puuid != "root" {
		pFile := this.matterDao.CheckByUuid(puuid)
		tmpMatter.Parent = pFile
		tmpMatter = pFile
		puuid = pFile.Puuid
	}

	return matter
}

//开始上传文件
//上传文件. alien表明文件是否是应用使用的文件。
func (this *MatterService) Upload(file multipart.File, user *User, puuid string, filename string, privacy bool, alien bool) *Matter {

	//文件名不能太长。
	if len(filename) > 200 {
		panic("文件名不能超过200")
	}


	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	absolutePath, relativePath := GetUserFilePath(user.Username, false)
	absolutePath = absolutePath + "/" + filename
	relativePath = relativePath + "/" + filename

	distFile, err := os.OpenFile(absolutePath, os.O_WRONLY|os.O_CREATE, 0777)
	this.PanicError(err)

	defer distFile.Close()

	written, err := io.Copy(distFile, file)
	this.PanicError(err)

	//判断用户自身上传大小的限制。
	if user.SizeLimit >= 0 {
		if written > user.SizeLimit {
			this.PanicBadRequest("文件大小超出限制 " + HumanFileSize(user.SizeLimit) + ">" + HumanFileSize(written))
		}
	}

	//查找文件夹下面是否有同名文件。
	matters := this.matterDao.ListByUserUuidAndPuuidAndDirAndName(user.Uuid, puuid, false, filename)
	//如果有同名的文件，那么我们直接覆盖同名文件。
	for _, dbFile := range matters {
		this.matterDao.Delete(dbFile)
	}

	//将文件信息存入数据库中。
	matter := &Matter{
		Puuid:    puuid,
		UserUuid: user.Uuid,
		Dir:      false,
		Alien:    alien,
		Name:     filename,
		Md5:      "",
		Size:     written,
		Privacy:  privacy,
		Path:     relativePath,
	}

	matter = this.matterDao.Create(matter)

	return matter
}

// 从指定的url下载一个文件。参考：https://golangcode.com/download-a-file-from-a-url/
func (this *MatterService) httpDownloadFile(filepath string, url string) (int64, error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Write the body to file
	size, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, err
	}

	return size, nil
}

//去指定的url中爬文件
func (this *MatterService) Crawl(url string, filename string, user *User, puuid string, privacy bool) *Matter {

	//文件名不能太长。
	if len(filename) > 200 {
		panic("文件名不能超过200")
	}

	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	absolutePath, relativePath := GetUserFilePath(user.Username, false)
	absolutePath = absolutePath + "/" + filename
	relativePath = relativePath + "/" + filename

	//使用临时文件存放
	fmt.Printf("存放于%s", absolutePath)
	size, err := this.httpDownloadFile(absolutePath, url)
	this.PanicError(err)

	//判断用户自身上传大小的限制。
	if user.SizeLimit >= 0 {
		if size > user.SizeLimit {
			panic("您最大只能上传" + HumanFileSize(user.SizeLimit) + "的文件")
		}
	}

	//查找文件夹下面是否有同名文件。
	matters := this.matterDao.ListByUserUuidAndPuuidAndDirAndName(user.Uuid, puuid, false, filename)
	//如果有同名的文件，那么我们直接覆盖同名文件。
	for _, dbFile := range matters {
		this.matterDao.Delete(dbFile)
	}

	//将文件信息存入数据库中。
	matter := &Matter{
		Puuid:    puuid,
		UserUuid: user.Uuid,
		Dir:      false,
		Alien:    false,
		Name:     filename,
		Md5:      "",
		Size:     size,
		Privacy:  privacy,
		Path:     relativePath,
	}

	matter = this.matterDao.Create(matter)

	return matter
}

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

func (r httpRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Range": {r.contentRange(size)},
		"Content-Type":  {contentType},
	}
}

// countingWriter counts how many bytes have been written to it.
type countingWriter int64

func (w *countingWriter) Write(p []byte) (n int, err error) {
	*w += countingWriter(len(p))
	return len(p), nil
}

//检查Last-Modified头。返回true: 请求已经完成了。（言下之意，文件没有修改过） 返回false：文件修改过。
func (this *MatterService) checkLastModified(w http.ResponseWriter, r *http.Request, modifyTime time.Time) bool {
	if modifyTime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modifyTime.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modifyTime.UTC().Format(http.TimeFormat))
	return false
}

// 处理ETag标签
// checkETag implements If-None-Match and If-Range checks.
//
// The ETag or modtime must have been previously set in the
// ResponseWriter's headers.  The modtime is only compared at second
// granularity and may be the zero value to mean unknown.
//
// The return value is the effective request "Range" header to use and
// whether this request is now considered done.
func (this *MatterService) checkETag(w http.ResponseWriter, r *http.Request, modtime time.Time) (rangeReq string, done bool) {
	etag := w.Header().Get("Etag")
	rangeReq = r.Header.Get("Range")

	// Invalidate the range request if the entity doesn't match the one
	// the client was expecting.
	// "If-Range: version" means "ignore the Range: header unless version matches the
	// current file."
	// We only support ETag versions.
	// The caller must have set the ETag on the response already.
	if ir := r.Header.Get("If-Range"); ir != "" && ir != etag {
		// The If-Range value is typically the ETag value, but it may also be
		// the modtime date. See golang.org/issue/8367.
		timeMatches := false
		if !modtime.IsZero() {
			if t, err := http.ParseTime(ir); err == nil && t.Unix() == modtime.Unix() {
				timeMatches = true
			}
		}
		if !timeMatches {
			rangeReq = ""
		}
	}

	if inm := r.Header.Get("If-None-Match"); inm != "" {
		// Must know ETag.
		if etag == "" {
			return rangeReq, false
		}

		// (bradfitz): non-GET/HEAD requests require more work:
		// sending a different status code on matches, and
		// also can't use weak cache validators (those with a "W/
		// prefix).  But most users of ServeContent will be using
		// it on GET or HEAD, so only support those for now.
		if r.Method != "GET" && r.Method != "HEAD" {
			return rangeReq, false
		}

		// (bradfitz): deal with comma-separated or multiple-valued
		// list of If-None-match values.  For now just handle the common
		// case of a single item.
		if inm == etag || inm == "*" {
			h := w.Header()
			delete(h, "Content-Type")
			delete(h, "Content-Length")
			w.WriteHeader(http.StatusNotModified)
			return "", true
		}
	}
	return rangeReq, false
}

// parseRange parses a Range header string as per RFC 2616.
func (this *MatterService) parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i >= size || i < 0 {
				return nil, errors.New("invalid range")
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}

// rangesMIMESize returns the number of bytes it takes to encode the
// provided ranges as a multipart response.
func (this *MatterService) rangesMIMESize(ranges []httpRange, contentType string, contentSize int64) (encSize int64) {
	var w countingWriter
	mw := multipart.NewWriter(&w)
	for _, ra := range ranges {
		mw.CreatePart(ra.mimeHeader(contentType, contentSize))
		encSize += ra.length
	}
	mw.Close()
	encSize += int64(w)
	return
}

func (this *MatterService) sumRangesSize(ranges []httpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}

//文件下载。具有进度功能。
//下载功能参考：https://github.com/Masterminds/go-fileserver
func (this *MatterService) DownloadFile(
	writer http.ResponseWriter,
	request *http.Request,
	filePath string,
	filename string,
	withContentDisposition bool) {

	diskFile, err := os.Open(filePath)
	this.PanicError(err)
	defer diskFile.Close()

	//如果是图片或者文本或者视频就直接打开。其余的一律以下载形式返回。
	if withContentDisposition {
		fileName := url.QueryEscape(filename)
		writer.Header().Set("content-disposition", "attachment; filename=\""+fileName+"\"")
	}

	//显示文件大小。
	fileInfo, err := diskFile.Stat()
	if err != nil {
		this.PanicServer("无法从磁盘中获取文件信息")
	}

	modifyTime := fileInfo.ModTime()

	if this.checkLastModified(writer, request, modifyTime) {
		return
	}
	rangeReq, done := this.checkETag(writer, request, modifyTime)
	if done {
		return
	}

	code := http.StatusOK

	// From net/http/sniff.go
	// The algorithm uses at most sniffLen bytes to make its decision.
	const sniffLen = 512

	// If Content-Type isn't set, use the file's extension to find it, but
	// if the Content-Type is unset explicitly, do not sniff the type.
	ctypes, haveType := writer.Header()["Content-Type"]
	var ctype string
	if !haveType {
		//放弃原有的判断mime的方法
		//ctype = mime.TypeByExtension(filepath.Ext(fileInfo.Name()))
		//使用mimeUtil来获取mime
		ctype = GetFallbackMimeType(filename, "")
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [sniffLen]byte
			n, _ := io.ReadFull(diskFile, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err := diskFile.Seek(0, os.SEEK_SET) // rewind to output whole file
			if err != nil {
				this.PanicServer("无法准确定位文件")
			}
		}
		writer.Header().Set("Content-Type", ctype)
	} else if len(ctypes) > 0 {
		ctype = ctypes[0]
	}

	size := fileInfo.Size()

	// handle Content-Range header.
	sendSize := size
	var sendContent io.Reader = diskFile
	if size >= 0 {
		ranges, err := this.parseRange(rangeReq, size)
		if err != nil {
			panic(CustomWebResult(CODE_WRAPPER_RANGE_NOT_SATISFIABLE, "range header出错"))
		}
		if this.sumRangesSize(ranges) > size {
			// The total number of bytes in all the ranges
			// is larger than the size of the file by
			// itself, so this is probably an attack, or a
			// dumb client.  Ignore the range request.
			ranges = nil
		}
		switch {
		case len(ranges) == 1:
			// RFC 2616, Section 14.16:
			// "When an HTTP message includes the content of a single
			// range (for example, a response to a request for a
			// single range, or to a request for a set of ranges
			// that overlap without any holes), this content is
			// transmitted with a Content-Range header, and a
			// Content-Length header showing the number of bytes
			// actually transferred.
			// ...
			// A response to a request for a single range MUST NOT
			// be sent using the multipart/byteranges media type."
			ra := ranges[0]
			if _, err := diskFile.Seek(ra.start, io.SeekStart); err != nil {
				panic(CustomWebResult(CODE_WRAPPER_RANGE_NOT_SATISFIABLE, "range header出错"))
			}
			sendSize = ra.length
			code = http.StatusPartialContent
			writer.Header().Set("Content-Range", ra.contentRange(size))
		case len(ranges) > 1:
			sendSize = this.rangesMIMESize(ranges, ctype, size)
			code = http.StatusPartialContent

			pr, pw := io.Pipe()
			mw := multipart.NewWriter(pw)
			writer.Header().Set("Content-Type", "multipart/byteranges; boundary="+mw.Boundary())
			sendContent = pr
			defer pr.Close() // cause writing goroutine to fail and exit if CopyN doesn't finish.
			go func() {
				for _, ra := range ranges {
					part, err := mw.CreatePart(ra.mimeHeader(ctype, size))
					if err != nil {
						pw.CloseWithError(err)
						return
					}
					if _, err := diskFile.Seek(ra.start, io.SeekStart); err != nil {
						pw.CloseWithError(err)
						return
					}
					if _, err := io.CopyN(part, diskFile, ra.length); err != nil {
						pw.CloseWithError(err)
						return
					}
				}
				mw.Close()
				pw.Close()
			}()
		}

		writer.Header().Set("Accept-Ranges", "bytes")
		if writer.Header().Get("Content-Encoding") == "" {
			writer.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
		}
	}

	writer.WriteHeader(code)

	if request.Method != "HEAD" {
		io.CopyN(writer, sendContent, sendSize)
	}

}
