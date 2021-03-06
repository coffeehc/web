package httpx

import (
	"context"
	"io"
	"net/http"

	"github.com/coffeehc/logger"
)

//Reply a wapper for http requese and response
type Reply interface {
	GetStatusCode() int
	SetStatusCode(statusCode int) Reply
	SetCookie(cookie http.Cookie) Reply
	SetHeader(key, value string) Reply
	AddHeader(key, value string) Reply
	DelHeader(key string) Reply
	GetHeader(key string) string
	Header() http.Header
	Redirect(code int, url string) Reply
	AddPathFragment(k, v string)

	With(data interface{}) Reply
	As(render Render) Reply

	GetRequest() *http.Request
	GetResponseWriter() http.ResponseWriter
	GetPathFragment() PathFragment
	AdapterHTTPHandler(adapter bool)
	//包装一层ResponseWriter,如 Gzip
	WarpResponseWriter(http.ResponseWriter)
	GetContext() context.Context
	SetContext(key string, value interface{})
}

type httpReply struct {
	statusCode         int
	data               interface{}
	header             http.Header
	cookies            []http.Cookie
	render             Render
	request            *http.Request
	responseWriter     http.ResponseWriter
	adapterHTTPHandler bool
	pathFragment       PathFragment
	cxt                context.Context
}

func newHTTPReply(request *http.Request, w http.ResponseWriter, config *Config) *httpReply {
	return &httpReply{
		statusCode:     200,
		render:         config.getDefaultRender(),
		cookies:        make([]http.Cookie, 0),
		request:        request,
		responseWriter: w,
		header:         w.Header(),
		cxt:            config.GetRootContext(),
	}
}

func (reply *httpReply) GetContext() context.Context {
	reply.request.Context()
	return reply.cxt
}
func (reply *httpReply) SetContext(key string, value interface{}) {
	reply.cxt = context.WithValue(reply.cxt, key, value)

}

func (reply *httpReply) AdapterHTTPHandler(adapter bool) {
	reply.adapterHTTPHandler = adapter
}

func (reply *httpReply) GetStatusCode() int {
	return reply.statusCode
}

func (reply *httpReply) SetStatusCode(statusCode int) Reply {
	reply.statusCode = statusCode
	return reply
}

func (reply *httpReply) SetCookie(cookie http.Cookie) Reply {
	reply.cookies = append(reply.cookies, cookie)
	return reply
}

func (reply *httpReply) SetHeader(key, value string) Reply {
	reply.header.Set(key, value)
	return reply
}
func (reply *httpReply) AddHeader(key, value string) Reply {
	reply.header.Add(key, value)
	return reply
}
func (reply *httpReply) DelHeader(key string) Reply {
	reply.header.Del(key)
	return reply
}

func (reply *httpReply) GetHeader(key string) string {
	return reply.header.Get(key)
}

func (reply *httpReply) Header() http.Header {
	return reply.header
}

func (reply *httpReply) Redirect(code int, url string) Reply {
	reply.responseWriter.Header().Set("Location", url)
	reply.statusCode = code
	return reply
}

func (reply *httpReply) With(data interface{}) Reply {
	reply.data = data
	return reply
}

func (reply *httpReply) As(render Render) Reply {
	if render != nil {
		reply.render = render
	}
	return reply
}

func (reply *httpReply) GetRequest() *http.Request {
	return reply.request
}

func (reply *httpReply) GetResponseWriter() http.ResponseWriter {
	return reply.responseWriter
}

func (reply *httpReply) GetPathFragment() PathFragment {
	return reply.pathFragment
}

func (reply *httpReply) AddPathFragment(k, v string) {
	if reply.pathFragment == nil {
		reply.pathFragment = make(PathFragment, 0)
	}
	reply.pathFragment[k] = RequestParam(v)
}

//Reply 最后的清理工作
func (reply *httpReply) finishReply() {
	if reply.adapterHTTPHandler {
		return
	}
	reply.writeWarpHeader()
	if reply.data == nil {
		reply.data = ""
	}
	reader, err := reply.render.Render(reply.data)
	if err != nil {
		reply.SetStatusCode(500)
		reader, _ = DefaultRenderText.Render(logger.Error("render error %#v", err))
	}
	if reader == nil {
		logger.Error("渲染结果为空")
		return
	}
	reply.responseWriter.WriteHeader(reply.GetStatusCode())
	io.Copy(reply.responseWriter, reader)
	reader.Close()
}

func (reply *httpReply) writeWarpHeader() {
	header := reply.Header()
	for _, cookie := range reply.cookies {
		header.Add("Set-Cookie", cookie.String())
	}
}

func (reply *httpReply) WarpResponseWriter(writwe http.ResponseWriter) {
	reply.responseWriter = writwe
}
