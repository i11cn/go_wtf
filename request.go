package wtf

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

type (
	wtf_request struct {
		req              *http.Request
		buf              []byte
		url_form         url.Values
		form_parsed      bool
		multiform_parsed bool
	}
)

func NewRequest(log Logger, req *http.Request) Request {
	return &wtf_request{
		req:      req,
		url_form: make(url.Values),
	}
}

func (r *wtf_request) BasicAuth() (username, password string, ok bool) {
	return r.req.BasicAuth()
}

func (r *wtf_request) Cookie(name string) (*http.Cookie, error) {
	return r.req.Cookie(name)
}

func (r *wtf_request) Cookies() []*http.Cookie {
	return r.req.Cookies()
}

func (r *wtf_request) MultipartReader() (*multipart.Reader, error) {
	return r.req.MultipartReader()
}

func (r *wtf_request) ParseMultipartForm(maxMemory ...int64) error {
	if r.multiform_parsed {
		return nil
	}
	var mm int64 = 16 << 20
	if len(maxMemory) > 0 {
		mm = maxMemory[0]
	}
	if err := r.req.ParseMultipartForm(mm); err != nil {
		return err
	}
	r.multiform_parsed = true
	return nil
}

func (r *wtf_request) Proto() (int, int) {
	return r.req.ProtoMajor, r.req.ProtoMinor
}

func (r *wtf_request) Referer() string {
	return r.req.Referer()
}

func (r *wtf_request) UserAgent() string {
	return r.req.UserAgent()
}

func (r *wtf_request) Method() string {
	return r.req.Method
}

func (r *wtf_request) URL() *url.URL {
	return r.req.URL
}

func (r *wtf_request) GetHeader(key string) string {
	return r.req.Header.Get(key)
}

func (r *wtf_request) ContentLength() int64 {
	return r.req.ContentLength
}

func (r *wtf_request) Host() string {
	return r.req.Host
}

func unescape_form(src url.Values) (ret url.Values, err error) {
	for k, va := range src {
		for i, v := range va {
			if va[i], err = url.QueryUnescape(v); err != nil {
				return make(url.Values), nil
			}
		}
		src[k] = va
	}
	return src, nil
}

func (r *wtf_request) Forms() (url.Values, url.Values, *multipart.Form) {
	if !r.form_parsed && r.req.ParseForm() == nil {
		if use, err := url.ParseQuery(r.req.URL.RawQuery); err == nil {
			if r.url_form, err = unescape_form(use); err == nil {
				r.form_parsed = true
			}
		}
	}
	r.ParseMultipartForm()
	return r.url_form, r.req.PostForm, r.req.MultipartForm
}

func (r *wtf_request) RemoteAddr() string {
	return r.req.RemoteAddr
}

func (r *wtf_request) GetBodyData() ([]byte, Error) {
	if r.buf != nil {
		return r.buf, nil
	}
	buf, err := ioutil.ReadAll(r.req.Body)
	if err != nil {
		return buf, NewError(500, "读取Body失败", err)
	}
	r.buf = buf
	return buf, nil
}

func (r *wtf_request) GetBody() (io.Reader, Error) {
	if buf, err := r.GetBodyData(); err != nil {
		return nil, err
	} else {
		return bytes.NewReader(buf), nil
	}
}

func (r *wtf_request) GetJsonBody(obj interface{}) Error {
	d, err := r.GetBodyData()
	if err != nil {
		return err
	}
	e := json.Unmarshal(d, obj)
	if e != nil {
		return NewError(500, "解析Json数据失败", e)
	}
	return nil
}

func (r *wtf_request) GetUploadFile(key string) ([]UploadFile, Error) {
	if !r.multiform_parsed {
		if err := r.req.ParseMultipartForm(16 << 20); err != nil {
			return nil, NewError(500, "解析Body的Multipart数据失败", err)
		}
		r.multiform_parsed = true
	}
	f, h, err := r.req.FormFile(key)
	if err != nil {
		return nil, NewError(500, "获取上传文件失败", err)
	}
	ret := make([]UploadFile, 0, 5)
	ret = append(ret, &wtf_upload_file{f, h})
	return ret, nil
}
