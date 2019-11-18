package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	// 引入当前包 用 . 表示
	// "."
	"testing"

	"github.com/stretchr/testify/mock"
)

const (
	expTime       = 60
	longUrl       = "https://www.example.com"
	shortLink     = "IFHzaO"
	shortLinkInfo = `{"url":"https://www.example.com","created_at":"2017-06-09 15:50:35","expiration_in_minutes":60}`
)

type storageMock struct {
	mock.Mock
}

var (
	app   App
	mockR *storageMock
)

// 每次运行测试用例 都会执行该 init 函数
func init() {
	app = App{}
	mockR = new(storageMock)
	app.Initialize(&Env{S: mockR})
}

func (s *storageMock) Shorten(url string, exp int64) (string, error) {
	// 调用 Mock 的 Called 方法,Called表示方法调用完毕  返回赋值给args
	args := s.Called(url, exp)
	return args.String(0), args.Error(1)
}

func (s *storageMock) UnShorten(short string) (string, error) {
	args := s.Called(short)
	return args.String(0), args.Error(1)
}

func (s *storageMock) ShortLinkInfo(short string) (interface{}, error) {
	args := s.Called(short)
	return args.String(0), args.Error(1)
}

func TestCreateShortlink(t *testing.T) {
	var jsonStr = []byte(`{
		"url":"https://www.example.com",
		"expiration_in_minutes":60
	}`)
	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewReader(jsonStr))
	if err != nil {
		t.Fatal("创建请求失败 err = ", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 调用 Shorten 入参 longUrl int64(expTime)
	// 期望返回 shortLink
	// 执行1次 Once
	mockR.On("Shorten", longUrl, int64(expTime)).Return(shortLink, nil).Once()

	// 发送请求
	// 构造resp
	rw := httptest.NewRecorder()
	// 调用路由 对请求进行响应
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("期望响应 %d 但是得到 %d", http.StatusCreated, rw.Code)
	}

	resp := struct {
		Shortlink string `json:"shortlink"`
	}{}
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("响应JSON数据解析失败 err = %v", err)
	}

	fmt.Println(resp.Shortlink)

	if resp.Shortlink != shortLink {
		t.Fatalf("期望短地址 %s 实际短地址 %s", shortLink, resp.Shortlink)
	}
}
