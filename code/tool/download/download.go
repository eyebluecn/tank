package download

import (
	"errors"
	"fmt"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// HttpRange specifies the byte range to be sent to the client.
type HttpRange struct {
	start  int64
	length int64
}

func (r HttpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

func (r HttpRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Range": {r.contentRange(size)},
		"Content-Type":  {contentType},
	}
}

// CountingWriter counts how many bytes have been written to it.
type CountingWriter int64

func (w *CountingWriter) Write(p []byte) (n int, err error) {
	*w += CountingWriter(len(p))
	return len(p), nil
}

//检查Last-Modified头。返回true: 请求已经完成了。（言下之意，文件没有修改过） 返回false：文件修改过。
func CheckLastModified(w http.ResponseWriter, r *http.Request, modifyTime time.Time) bool {
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

// handle ETag
// CheckETag implements If-None-Match and If-Range checks.
//
// The ETag or modtime must have been previously set in the
// ResponseWriter's headers.  The modtime is only compared at second
// granularity and may be the zero value to mean unknown.
//
// The return value is the effective request "Range" header to use and
// whether this request is now considered done.
func CheckETag(w http.ResponseWriter, r *http.Request, modtime time.Time) (rangeReq string, done bool) {
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

// ParseRange parses a Range header string as per RFC 2616.
func ParseRange(s string, size int64) ([]HttpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []HttpRange
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
		var r HttpRange
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

// RangesMIMESize returns the number of bytes it takes to encode the
// provided ranges as a multipart response.
func RangesMIMESize(ranges []HttpRange, contentType string, contentSize int64) (encSize int64) {
	var w CountingWriter
	mw := multipart.NewWriter(&w)
	for _, ra := range ranges {
		_, e := mw.CreatePart(ra.mimeHeader(contentType, contentSize))

		PanicError(e)

		encSize += ra.length
	}
	e := mw.Close()
	PanicError(e)
	encSize += int64(w)
	return
}

func SumRangesSize(ranges []HttpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

//file download. https://github.com/Masterminds/go-fileserver
func DownloadFile(
	writer http.ResponseWriter,
	request *http.Request,
	filePath string,
	filename string,
	withContentDisposition bool) {

	diskFile, err := os.Open(filePath)
	PanicError(err)

	defer func() {
		e := diskFile.Close()
		PanicError(e)
	}()

	//content-disposition tell browser to download rather than preview.
	if withContentDisposition {
		fileName := url.QueryEscape(filename)
		writer.Header().Set("content-disposition", "attachment; filename=\""+fileName+"\"")
	}

	fileInfo, err := diskFile.Stat()
	if err != nil {
		panic("cannot load fileInfo from disk." + filePath)
	}

	modifyTime := fileInfo.ModTime()

	if CheckLastModified(writer, request, modifyTime) {
		return
	}
	rangeReq, done := CheckETag(writer, request, modifyTime)
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
		//get mime
		ctype = util.GetFallbackMimeType(filename, "")
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [sniffLen]byte
			n, _ := io.ReadFull(diskFile, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err := diskFile.Seek(0, os.SEEK_SET) // rewind to output whole file
			if err != nil {
				panic("cannot seek file")
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
		ranges, err := ParseRange(rangeReq, size)
		if err != nil {
			panic(result.CustomWebResult(result.RANGE_NOT_SATISFIABLE, "range header出错"))
		}
		if SumRangesSize(ranges) > size {
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
				panic(result.CustomWebResult(result.RANGE_NOT_SATISFIABLE, "range header出错"))
			}
			sendSize = ra.length
			code = http.StatusPartialContent
			writer.Header().Set("Content-Range", ra.contentRange(size))
		case len(ranges) > 1:
			sendSize = RangesMIMESize(ranges, ctype, size)
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
