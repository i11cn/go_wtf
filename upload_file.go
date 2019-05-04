package wtf

import (
	"mime/multipart"
	"net/textproto"
)

type (
	wtf_upload_file struct {
		file   multipart.File
		header *multipart.FileHeader
	}
)

func (up *wtf_upload_file) Read(p []byte) (n int, err error) {
	return up.file.Read(p)
}

func (up *wtf_upload_file) ReadAt(p []byte, off int64) (n int, err error) {
	return up.file.ReadAt(p, off)
}

func (up *wtf_upload_file) Seek(offset int64, whence int) (int64, error) {
	return up.file.Seek(offset, whence)
}

func (up *wtf_upload_file) Close() error {
	return up.file.Close()
}

func (up *wtf_upload_file) Filename() string {
	return up.header.Filename
}

func (up *wtf_upload_file) Size() int64 {
	return up.header.Size
}

func (up *wtf_upload_file) ContentType() string {
	return up.header.Header.Get("ContentType")
}

func (up *wtf_upload_file) Header() textproto.MIMEHeader {
	return up.header.Header
}
