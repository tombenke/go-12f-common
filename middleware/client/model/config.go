package model

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/tombenke/go-12f-common/utils"
)

type CaptureTransportMode int

const (
	CaptureTransportModeNone   CaptureTransportMode = iota
	CaptureTransportModeRecord CaptureTransportMode = iota
	CaptureTransportModeFake   CaptureTransportMode = iota
)

type DelayedReaderPersisterer interface {
	io.ReadCloser
	Done() <-chan struct{}
	Payload() []byte
	IsNil() bool
}

type Request struct {
	Method           string
	URL              string
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	Body             *string
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	//TLS              *tls.ConnectionState
	TestTimestamp time.Time `yaml:"testTimestamp"`
}

func NewRequest(r0 *http.Request) Request {
	r := r0.Clone(context.Background())
	return Request{
		Method:           r.Method,
		URL:              utils.SafeUrl(r.URL).String(),
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header,
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close:            r.Close,
		Host:             r.Host,
		Form:             r.Form,
		PostForm:         r.PostForm,
		MultipartForm:    r.MultipartForm,
		Trailer:          r.Trailer,
		RemoteAddr:       r.RemoteAddr,
		RequestURI:       r.RequestURI,
		//TLS:              r.TLS,
		TestTimestamp: time.Now(),
	}
}

type Response struct {
	Status           string
	StatusCode       int
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	Body             *string
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Uncompressed     bool
	Trailer          http.Header
	//TLS              *tls.ConnectionState
	TestTimestamp time.Time `yaml:"testTimestamp"`
}

func NewResponse(r *http.Response) Response {
	return Response{
		Status:           r.Status,
		StatusCode:       r.StatusCode,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header.Clone(),
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close:            r.Close,
		Uncompressed:     r.Uncompressed,
		Trailer:          r.Trailer.Clone(),
		//TLS:              r.TLS,
		TestTimestamp: time.Now(),
	}
}

func (r Response) ToHttpResponse() *http.Response {
	return &http.Response{
		Status:           r.Status,
		StatusCode:       r.StatusCode,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header.Clone(),
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close:            r.Close,
		Uncompressed:     r.Uncompressed,
		Trailer:          r.Trailer.Clone(),
	}
}

type CaptureItem struct {
	Request     Request                  `yaml:"request"`
	ReqPayload  DelayedReaderPersisterer `yaml:"-"`
	Response    Response                 `yaml:"response"`
	RespPayload DelayedReaderPersisterer `yaml:"-"`
}

type CaptureMatcher func(*http.Request, *CaptureItem) bool
