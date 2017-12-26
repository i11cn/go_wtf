package wtf

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type (
	File interface {
		io.Reader
		io.Writer
		io.Seeker
		io.Closer
		FileInfo() os.FileInfo
		ContentType() string
	}

	FileSystem interface {
		SetFileMapper(m func(*http.Request) []string)
		SetDefaultPages([]string)
		Open(*http.Request, ...int) (File, Error)
		OpenPath(string, ...int) (File, Error)
		Read(*http.Request) ([]byte, Error)
		Write([]byte, *http.Request) Error
		WriteStream(io.Reader, *http.Request) Error
		Append([]byte, *http.Request) Error
		AppendStream(io.Reader, *http.Request) Error
	}

	wtf_file struct {
		file *os.File
		fi   os.FileInfo
		ct   string
	}

	wtf_file_system struct {
		mapper    func(*http.Request) []string
		def_pages []string
	}
)

func (f *wtf_file) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *wtf_file) Write(p []byte) (int, error) {
	return f.file.Write(p)
}

func (f *wtf_file) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *wtf_file) Close() error {
	return f.file.Close()
}

func (f *wtf_file) FileInfo() os.FileInfo {
	return f.fi
}

func (f *wtf_file) ContentType() string {
	return f.ct
}

func (f *wtf_file) check_content_type() {
	if len(f.ct) == 0 {
		ext := filepath.Ext(f.fi.Name())
		f.ct = mime.TypeByExtension(ext)
		if len(f.ct) == 0 {
			var buf [512]byte
			n, _ := io.ReadFull(f.file, buf[:])
			f.ct = http.DetectContentType(buf[:n])
			f.file.Seek(0, io.SeekStart)
		}
	}
}

func NewFileSystem(base ...string) FileSystem {
	p := "."
	if len(base) > 0 {
		p = base[0]
	}
	return &wtf_file_system{func(r *http.Request) []string {
		path := p + r.URL.Path
		return []string{path}
	}, []string{}}
}

func (fs *wtf_file_system) SetFileMapper(m func(*http.Request) []string) {
	fs.mapper = m
}

func (fs *wtf_file_system) get_file_info(paths []string) (string, os.FileInfo, Error) {
	var we Error
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			we = NewError(http.StatusNotFound, err.Error(), err)
		} else {
			if stat.IsDir() {
				for _, f := range fs.def_pages {
					fn := path + "/" + f
					fi, err := os.Stat(fn)
					if err == nil {
						return fn, fi, nil
					}
				}
				we = NewError(http.StatusNotFound, "File Not Found")
			} else {
				return path, stat, nil
			}
		}
	}
	return "", nil, we
}

func (fs *wtf_file_system) SetDefaultPages(dp []string) {
	fs.def_pages = dp
}

func (fs *wtf_file_system) open_file(paths []string, flags ...int) (File, Error) {
	name, stat, werr := fs.get_file_info(paths)
	if werr != nil {
		return nil, werr
	}
	flag := os.O_RDONLY
	if len(flags) > 0 {
		flag = 0
		for _, f := range flags {
			flag |= f
		}
	}
	file, err := os.OpenFile(name, flag, 0644)
	if err != nil {
		return nil, NewError(http.StatusInternalServerError, err.Error(), err)
	}
	ret := &wtf_file{file, stat, ""}
	ret.check_content_type()
	return ret, nil
}

func (fs *wtf_file_system) Open(req *http.Request, flags ...int) (File, Error) {
	paths := fs.mapper(req)
	f, err := fs.open_file(paths, flags...)
	if err != nil && err.Code() == 404 {
		return nil, NewError(404, req.URL.Path+": no such file or directory", err)
	}
	return f, err
}

func (fs *wtf_file_system) OpenPath(path string, flags ...int) (File, Error) {
	return fs.open_file([]string{path}, flags...)
}

func (fs *wtf_file_system) Read(req *http.Request) ([]byte, Error) {
	file, err := fs.Open(req)
	if err != nil {
		return nil, err
	}
	size := file.FileInfo().Size()
	buf := make([]byte, size)
	_, e := file.Read(buf)
	if e != nil {
		return nil, NewError(http.StatusInternalServerError, e.Error(), e)
	}
	return buf, nil
}

func (fs *wtf_file_system) write_stream(r io.Reader, req *http.Request, flags int) Error {
	file, err := fs.Open(req, flags)
	if err != nil {
		return err
	}
	_, e := io.Copy(file, r)
	if e != nil {
		return NewError(http.StatusInternalServerError, e.Error(), e)
	}
	return nil

}

func (fs *wtf_file_system) Write(d []byte, req *http.Request) Error {
	return fs.write_stream(bytes.NewReader(d), req, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
}

func (fs *wtf_file_system) WriteStream(r io.Reader, req *http.Request) Error {
	return fs.write_stream(r, req, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
}

func (fs *wtf_file_system) Append(d []byte, req *http.Request) Error {
	return fs.write_stream(bytes.NewReader(d), req, os.O_APPEND|os.O_WRONLY|os.O_CREATE)
}

func (fs *wtf_file_system) AppendStream(r io.Reader, req *http.Request) Error {
	return fs.write_stream(r, req, os.O_APPEND|os.O_WRONLY|os.O_CREATE)
}
