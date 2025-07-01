package webdav

import (
	"encoding/xml"
	"time"
)

// WebDAV XML structures
type Multistatus struct {
	XMLName   xml.Name   `xml:"DAV: multistatus"`
	Responses []Response `xml:"response"`
}

type Response struct {
	XMLName  xml.Name `xml:"DAV: response"`
	Href     string   `xml:"href"`
	Propstat Propstat `xml:"propstat"`
}

type Propstat struct {
	XMLName xml.Name `xml:"DAV: propstat"`
	Prop    Prop     `xml:"prop"`
	Status  string   `xml:"status"`
}

type Prop struct {
	XMLName       xml.Name      `xml:"DAV: prop"`
	DisplayName   string        `xml:"displayname,omitempty"`
	ResourceType  *ResourceType `xml:"resourcetype,omitempty"`
	ContentLength *int64        `xml:"getcontentlength,omitempty"`
	ContentType   string        `xml:"getcontenttype,omitempty"`
	LastModified  string        `xml:"getlastmodified,omitempty"`
	CreationDate  string        `xml:"creationdate,omitempty"`
	ETag          string        `xml:"getetag,omitempty"`
}

type ResourceType struct {
	XMLName    xml.Name    `xml:"DAV: resourcetype"`
	Collection *Collection `xml:"collection,omitempty"`
}

type Collection struct {
	XMLName xml.Name `xml:"DAV: collection"`
}

type PropFind struct {
	XMLName xml.Name  `xml:"DAV: propfind"`
	Prop    *PropReq  `xml:"prop,omitempty"`
	AllProp *struct{} `xml:"allprop,omitempty"`
}

type PropReq struct {
	XMLName       xml.Name  `xml:"DAV: prop"`
	DisplayName   *struct{} `xml:"displayname,omitempty"`
	ResourceType  *struct{} `xml:"resourcetype,omitempty"`
	ContentLength *struct{} `xml:"getcontentlength,omitempty"`
	ContentType   *struct{} `xml:"getcontenttype,omitempty"`
	LastModified  *struct{} `xml:"getlastmodified,omitempty"`
	CreationDate  *struct{} `xml:"creationdate,omitempty"`
	ETag          *struct{} `xml:"getetag,omitempty"`
}

// Helper functions for WebDAV responses

// FormatTime formats a time for WebDAV responses
func FormatTime(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

// GenerateETag generates an ETag from a URL and modification time
func GenerateETag(url string, modTime time.Time) string {
	return `"` + url + "-" + modTime.Format("20060102150405") + `"`
}
