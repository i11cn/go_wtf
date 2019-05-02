package wtf

type (
	// 以REST方式的请求，在URI中定义的参数将会被解析成该结构
	RESTParam struct {
		name  string
		value string
	}

	// REST方式的请求，URI中定义的参数集合
	RESTParams []RESTParam
)

// 获取命名了的URI参数
//
// 例如：/test/:foo，则命名参数为foo
//
// 又如：/test/(?P<name>\d+)，则命名参数为name
func (p RESTParams) Get(name string) string {
	if len(name) > 0 {
		for _, i := range []RESTParam(p) {
			if i.name == name {
				return i.value
			}
		}
	}
	return ""
}

// 按索引获取URI参数
//
// 例如：/test/:foo/(\d+)，第一个参数命名为foo，第二个参数没有命名，只能通过索引取得
func (p RESTParams) GetIndex(i int) string {
	pa := []RESTParam(p)
	if i >= 0 && i < len(pa) {
		return pa[i].value
	}
	return ""
}

// 增加URI参数
//
// 对于重名的问题，不在此处考虑，那是使用者需要考虑的事
func (p RESTParams) Append(name, value string) RESTParams {
	ret := []RESTParam(p)
	ret = append(ret, RESTParam{name, value})
	return RESTParams(ret)
}
