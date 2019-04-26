// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dav

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/eyebluecn/tank/code/tool/dav/xml"
	"github.com/eyebluecn/tank/code/tool/result"
	"io"
	"net/http"
)

// Proppatch describes a property update instruction as defined in RFC 4918.
// See http://www.webdav.org/specs/rfc4918.html#METHOD_PROPPATCH
type Proppatch struct {
	// Remove specifies whether this patch removes properties. If it does not
	// remove them, it sets them.
	Remove bool
	// Props contains the properties to be set or removed.
	Props []Property
}

// Propstat describes a XML propstat element as defined in RFC 4918.
// See http://www.webdav.org/specs/rfc4918.html#ELEMENT_propstat
type Propstat struct {
	// Props contains the properties for which Status applies.
	Props []Property

	// Status defines the HTTP status code of the properties in Prop.
	// Allowed values include, but are not limited to the WebDAV status
	// code extensions for HTTP/1.1.
	// http://www.webdav.org/specs/rfc4918.html#status.code.extensions.to.http11
	Status int

	// XMLError contains the XML representation of the optional error element.
	// XML content within this field must not rely on any predefined
	// namespace declarations or prefixes. If empty, the XML error element
	// is omitted.
	XMLError string

	// ResponseDescription contains the contents of the optional
	// responsedescription field. If empty, the XML element is omitted.
	ResponseDescription string
}

// DeadPropsHolder holds the dead properties of a resource.
//
// Dead properties are those properties that are explicitly defined. In
// comparison, live properties, such as DAV:getcontentlength, are implicitly
// defined by the underlying resource, and cannot be explicitly overridden or
// removed. See the Terminology section of
// http://www.webdav.org/specs/rfc4918.html#rfc.section.3
//
// There is a whitelist of the names of live properties. This package handles
// all live properties, and will only pass non-whitelisted names to the Patch
// method of DeadPropsHolder implementations.
type DeadPropsHolder interface {
	// DeadProps returns a copy of the dead properties held.
	DeadProps() (map[xml.Name]Property, error)

	// Patch patches the dead properties held.
	//
	// Patching is atomic; either all or no patches succeed. It returns (nil,
	// non-nil) if an internal server error occurred, otherwise the Propstats
	// collectively contain one Property for each proposed patch Property. If
	// all patches succeed, Patch returns a slice of length one and a Propstat
	// element with a 200 OK HTTP status code. If none succeed, for reasons
	// other than an internal server error, no Propstat has status 200 OK.
	//
	// For more details on when various HTTP status codes apply, see
	// http://www.webdav.org/specs/rfc4918.html#PROPPATCH-status
	Patch([]Proppatch) ([]Propstat, error)
}

func EscapeXML(s string) string {
	for i := 0; i < len(s); i++ {
		// As an optimization, if s contains only ASCII letters, digits or a
		// few special characters, the escaped value is s itself and we don't
		// need to allocate a buffer and convert between string and []byte.
		switch c := s[i]; {
		case c == ' ' || c == '_' ||
			('+' <= c && c <= '9') || // Digits as well as + , - . and /
			('A' <= c && c <= 'Z') ||
			('a' <= c && c <= 'z'):
			continue
		}
		// Otherwise, go through the full escaping process.
		var buf bytes.Buffer
		xml.EscapeText(&buf, []byte(s))
		return buf.String()
	}
	return s
}

// http://www.webdav.org/specs/rfc4918.html#status.code.extensions.to.http11
const (
	StatusMulti               = 207
	StatusUnprocessableEntity = 422
	StatusLocked              = 423
	StatusFailedDependency    = 424
	StatusInsufficientStorage = 507
)

func StatusText(code int) string {
	switch code {
	case StatusMulti:
		return "Multi-Status"
	case StatusUnprocessableEntity:
		return "Unprocessable Entity"
	case StatusLocked:
		return "Locked"
	case StatusFailedDependency:
		return "Failed Dependency"
	case StatusInsufficientStorage:
		return "Insufficient Storage"
	}
	return http.StatusText(code)
}

var (
	errInvalidPropfind = errors.New("webdav: invalid propfind")
	errInvalidResponse = errors.New("webdav: invalid response")
)

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_lockinfo
type LockInfo struct {
	XMLName   xml.Name  `xml:"lockinfo"`
	Exclusive *struct{} `xml:"lockscope>exclusive"`
	Shared    *struct{} `xml:"lockscope>shared"`
	Write     *struct{} `xml:"locktype>write"`
	Owner     Owner     `xml:"owner"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_owner
type Owner struct {
	InnerXML string `xml:",innerxml"`
}

//这是一个带字节计数器的Reader，可以知道总共读取了多少个字节。
type CountingReader struct {
	n      int
	reader io.Reader
}

func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.n += n
	return n, err
}

// Next returns the next token, if any, in the XML stream of d.
// RFC 4918 requires to ignore comments, processing instructions
// and directives.
// http://www.webdav.org/specs/rfc4918.html#property_values
// http://www.webdav.org/specs/rfc4918.html#xml-extensibility
func next(d *xml.Decoder) (xml.Token, error) {
	for {
		t, err := d.Token()
		if err != nil {
			return t, err
		}
		switch t.(type) {
		case xml.Comment, xml.Directive, xml.ProcInst:
			continue
		default:
			return t, nil
		}
	}
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_prop (for propfind)
type PropfindProps []xml.Name

// UnmarshalXML appends the property names enclosed within start to pn.
//
// It returns an error if start does not contain any properties or if
// properties contain values. Character data between properties is ignored.
func (pn *PropfindProps) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		t, err := next(d)
		if err != nil {
			return err
		}
		switch t.(type) {
		case xml.EndElement:
			if len(*pn) == 0 {
				return fmt.Errorf("%s must not be empty", start.Name.Local)
			}
			return nil
		case xml.StartElement:
			name := t.(xml.StartElement).Name
			t, err = next(d)
			if err != nil {
				return err
			}
			if _, ok := t.(xml.EndElement); !ok {
				return fmt.Errorf("unexpected token %T", t)
			}
			*pn = append(*pn, xml.Name(name))
		}
	}
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_propfind
// <!ELEMENT propfind ( propname | (allprop, include?) | prop ) >
type Propfind struct {
	XMLName  xml.Name      `xml:"DAV: propfind"`
	Allprop  *struct{}     `xml:"DAV: allprop"`
	Propname *struct{}     `xml:"DAV: propname"`
	Prop     PropfindProps `xml:"DAV: prop"`
	Include  PropfindProps `xml:"DAV: include"`
}

//从request中读出需要的属性。比如：getcontentlength 大小 creationdate 创建时间
func ReadPropfind(reader io.Reader) (propfind *Propfind) {
	propfind = &Propfind{}

	c := CountingReader{reader: reader}
	if err := xml.NewDecoder(&c).Decode(&propfind); err != nil {
		if err == io.EOF {
			if c.n == 0 {
				// An empty body means to propfind allprop.
				// http://www.webdav.org/specs/rfc4918.html#METHOD_PROPFIND
				return &Propfind{Allprop: new(struct{})}
			}
			err = errInvalidPropfind
		}

		panic(result.BadRequest(err.Error()))
	}

	if propfind.Allprop == nil && propfind.Include != nil {
		panic(result.BadRequest(errInvalidPropfind.Error()))
	}
	if propfind.Allprop != nil && (propfind.Prop != nil || propfind.Propname != nil) {
		panic(result.BadRequest(errInvalidPropfind.Error()))
	}
	if propfind.Prop != nil && propfind.Propname != nil {
		panic(result.BadRequest(errInvalidPropfind.Error()))
	}
	if propfind.Propname == nil && propfind.Allprop == nil && propfind.Prop == nil {
		panic(result.BadRequest(errInvalidPropfind.Error()))
	}

	return propfind
}

// Property represents a single DAV resource property as defined in RFC 4918.
// See http://www.webdav.org/specs/rfc4918.html#data.model.for.resource.properties
type Property struct {
	// XMLName is the fully qualified name that identifies this property.
	XMLName xml.Name

	// Lang is an optional xml:lang attribute.
	Lang string `xml:"xml:lang,attr,omitempty"`

	// InnerXML contains the XML representation of the property value.
	// See http://www.webdav.org/specs/rfc4918.html#property_values
	//
	// Property values of complex type or mixed-content must have fully
	// expanded XML namespaces or be self-contained with according
	// XML namespace declarations. They must not rely on any XML
	// namespace declarations within the scope of the XML document,
	// even including the DAV: namespace.
	InnerXML []byte `xml:",innerxml"`
}

// ixmlProperty is the same as the Property type except it holds an xml.Name
// instead of an xml.Name.
type IxmlProperty struct {
	XMLName  xml.Name
	Lang     string `xml:"xml:lang,attr,omitempty"`
	InnerXML []byte `xml:",innerxml"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_error
// See MultiStatusWriter for the "D:" namespace prefix.
type XmlError struct {
	XMLName  xml.Name `xml:"D:error"`
	InnerXML []byte   `xml:",innerxml"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_propstat
// See MultiStatusWriter for the "D:" namespace prefix.
type SubPropstat struct {
	Prop                []Property `xml:"D:prop>_ignored_"`
	Status              string     `xml:"D:status"`
	Error               *XmlError  `xml:"D:error"`
	ResponseDescription string     `xml:"D:responsedescription,omitempty"`
}

// ixmlPropstat is the same as the propstat type except it holds an xml.Name
// instead of an xml.Name.
type IxmlPropstat struct {
	Prop                []IxmlProperty `xml:"D:prop>_ignored_"`
	Status              string         `xml:"D:status"`
	Error               *XmlError      `xml:"D:error"`
	ResponseDescription string         `xml:"D:responsedescription,omitempty"`
}

// MarshalXML prepends the "D:" namespace prefix on properties in the DAV: namespace
// before encoding. See MultiStatusWriter.
func (ps SubPropstat) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Convert from a propstat to an ixmlPropstat.
	ixmlPs := IxmlPropstat{
		Prop:                make([]IxmlProperty, len(ps.Prop)),
		Status:              ps.Status,
		Error:               ps.Error,
		ResponseDescription: ps.ResponseDescription,
	}
	for k, prop := range ps.Prop {
		ixmlPs.Prop[k] = IxmlProperty{
			XMLName:  xml.Name(prop.XMLName),
			Lang:     prop.Lang,
			InnerXML: prop.InnerXML,
		}
	}

	for k, prop := range ixmlPs.Prop {
		if prop.XMLName.Space == "DAV:" {
			prop.XMLName = xml.Name{Space: "", Local: "D:" + prop.XMLName.Local}
			ixmlPs.Prop[k] = prop
		}
	}
	// Distinct type to avoid infinite recursion of MarshalXML.
	type newpropstat IxmlPropstat
	return e.EncodeElement(newpropstat(ixmlPs), start)
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_response
// See MultiStatusWriter for the "D:" namespace prefix.
type Response struct {
	XMLName             xml.Name      `xml:"D:response"`
	Href                []string      `xml:"D:href"`
	Propstat            []SubPropstat `xml:"D:propstat"`
	Status              string        `xml:"D:status,omitempty"`
	Error               *XmlError     `xml:"D:error"`
	ResponseDescription string        `xml:"D:responsedescription,omitempty"`
}

// MultistatusWriter marshals one or more Responses into a XML
// multistatus response.
// See http://www.webdav.org/specs/rfc4918.html#ELEMENT_multistatus
// TODO(rsto, mpl): As a workaround, the "D:" namespace prefix, defined as
// "DAV:" on this element, is prepended on the nested response, as well as on all
// its nested elements. All property names in the DAV: namespace are prefixed as
// well. This is because some versions of Mini-Redirector (on windows 7) ignore
// elements with a default namespace (no prefixed namespace). A less intrusive fix
// should be possible after golang.org/cl/11074. See https://golang.org/issue/11177
type MultiStatusWriter struct {
	// ResponseDescription contains the optional responsedescription
	// of the multistatus XML element. Only the latest content before
	// close will be emitted. Empty response descriptions are not
	// written.
	ResponseDescription string

	Writer  http.ResponseWriter
	Encoder *xml.Encoder
}

// Write validates and emits a DAV response as part of a multistatus response
// element.
//
// It sets the HTTP status code of its underlying http.ResponseWriter to 207
// (Multi-Status) and populates the Content-Type header. If r is the
// first, valid response to be written, Write prepends the XML representation
// of r with a multistatus tag. Callers must call close after the last response
// has been written.
func (this *MultiStatusWriter) Write(response *Response) error {
	switch len(response.Href) {
	case 0:
		return errInvalidResponse
	case 1:
		if len(response.Propstat) > 0 != (response.Status == "") {
			return errInvalidResponse
		}
	default:
		if len(response.Propstat) > 0 || response.Status == "" {
			return errInvalidResponse
		}
	}
	err := this.writeHeader()
	if err != nil {
		return err
	}
	return this.Encoder.Encode(response)
}

// writeHeader writes a XML multistatus start element on w's underlying
// http.ResponseWriter and returns the result of the write operation.
// After the first write attempt, writeHeader becomes a no-op.
func (this *MultiStatusWriter) writeHeader() error {
	if this.Encoder != nil {
		return nil
	}
	this.Writer.Header().Add("Content-Type", "text/xml; charset=utf-8")
	this.Writer.WriteHeader(StatusMulti)
	_, err := fmt.Fprintf(this.Writer, `<?xml version="1.0" encoding="UTF-8"?>`)
	if err != nil {
		return err
	}
	this.Encoder = xml.NewEncoder(this.Writer)
	return this.Encoder.EncodeToken(xml.StartElement{
		Name: xml.Name{
			Space: "DAV:",
			Local: "multistatus",
		},
		Attr: []xml.Attr{{
			Name:  xml.Name{Space: "xmlns", Local: "D"},
			Value: "DAV:",
		}},
	})
}

// Close completes the marshalling of the multistatus response. It returns
// an error if the multistatus response could not be completed. If both the
// return value and field Encoder of w are nil, then no multistatus response has
// been written.
func (this *MultiStatusWriter) Close() error {
	if this.Encoder == nil {
		return nil
	}
	var end []xml.Token
	if this.ResponseDescription != "" {
		name := xml.Name{Space: "DAV:", Local: "responsedescription"}
		end = append(end,
			xml.StartElement{Name: name},
			xml.CharData(this.ResponseDescription),
			xml.EndElement{Name: name},
		)
	}
	end = append(end, xml.EndElement{
		Name: xml.Name{Space: "DAV:", Local: "multistatus"},
	})
	for _, t := range end {
		err := this.Encoder.EncodeToken(t)
		if err != nil {
			return err
		}
	}
	return this.Encoder.Flush()
}

var xmlLangName = xml.Name{Space: "http://www.w3.org/XML/1998/namespace", Local: "lang"}

func xmlLang(s xml.StartElement, d string) string {
	for _, attr := range s.Attr {
		if attr.Name == xmlLangName {
			return attr.Value
		}
	}
	return d
}

type XmlValue []byte

func (v *XmlValue) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// The XML value of a property can be arbitrary, mixed-content XML.
	// To make sure that the unmarshalled value contains all required
	// namespaces, we encode all the property value XML tokens into a
	// buffer. This forces the encoder to redeclare any used namespaces.
	var b bytes.Buffer
	e := xml.NewEncoder(&b)
	for {
		t, err := next(d)
		if err != nil {
			return err
		}
		if e, ok := t.(xml.EndElement); ok && e.Name == start.Name {
			break
		}
		if err = e.EncodeToken(t); err != nil {
			return err
		}
	}
	err := e.Flush()
	if err != nil {
		return err
	}
	*v = b.Bytes()
	return nil
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_prop (for proppatch)
type ProppatchProps []Property

// UnmarshalXML appends the property names and values enclosed within start
// to ps.
//
// An xml:lang attribute that is defined either on the DAV:prop or property
// name XML element is propagated to the property's Lang field.
//
// UnmarshalXML returns an error if start does not contain any properties or if
// property values contain syntactically incorrect XML.
func (ps *ProppatchProps) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	lang := xmlLang(start, "")
	for {
		t, err := next(d)
		if err != nil {
			return err
		}
		switch elem := t.(type) {
		case xml.EndElement:
			if len(*ps) == 0 {
				return fmt.Errorf("%s must not be empty", start.Name.Local)
			}
			return nil
		case xml.StartElement:
			p := Property{
				XMLName: xml.Name(t.(xml.StartElement).Name),
				Lang:    xmlLang(t.(xml.StartElement), lang),
			}
			err = d.DecodeElement(((*XmlValue)(&p.InnerXML)), &elem)
			if err != nil {
				return err
			}
			*ps = append(*ps, p)
		}
	}
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_set
// http://www.webdav.org/specs/rfc4918.html#ELEMENT_remove
type SetRemove struct {
	XMLName xml.Name
	Lang    string         `xml:"xml:lang,attr,omitempty"`
	Prop    ProppatchProps `xml:"DAV: prop"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_propertyupdate
type PropertyUpdate struct {
	XMLName   xml.Name    `xml:"DAV: propertyupdate"`
	Lang      string      `xml:"xml:lang,attr,omitempty"`
	SetRemove []SetRemove `xml:",any"`
}
