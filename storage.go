package main

type Storage interface {
	// 长地址 --> 短地址
	Shorten(url string, exp int64) (string, error)
	// 传入短地址 返回短地址对应的详细信息
	ShortLinkInfo(eid string) (interface{}, error)
	// 长地址 <-- 短地址
	UnShorten(eid string) (string, error)
}
