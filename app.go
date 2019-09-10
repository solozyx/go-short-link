package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/validator.v2"
)

type App struct {
	Router     *mux.Router
	Middleware *Middleware
	Config     *Env
}

// 短连接请求
type shortenReq struct {
	// json tag 做json和struct转换
	// 需要验证 规则非零 通过 gopkg.in/validator.v2 完成验证操作
	Url string `json:"url" validate:"nonzero"`
	// 过期时间 验证规则 最小值为0 不能小于0
	ExpirationInMinutes int64 `json:"expiration_in_minutes" validate:"min=0"`
}

// 短连接响应
type shortLinkResp struct {
	ShortLink string `json:"short_link"`
}

// 初始化
func (a *App) Initialize(env *Env) {
	// 定义 log 格式
	// log.LstdFlags 打印日志发生日期时间
	// log.Lshortfile 打印日志文件名和行号
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Config = env
	a.Router = mux.NewRouter()
	a.Middleware = &Middleware{}
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	// a.Router.HandleFunc("/api/shorten", a.createShortLink).Methods("POST")
	// a.Router.HandleFunc("/api/info", a.getShortLinkInfo).Methods("GET")
	// TODO:NOTICE 重定向接口
	//  指定短连接是 1-11位 的字母和数字组成
	// a.Router.HandleFunc("/{shortLink:[a-zA-z0-9]{1,11}}", a.redirect).Methods("GET")

	// middleware 中间件
	m := alice.New(a.Middleware.LoggingHandler, a.Middleware.RecoverHandler)
	a.Router.Handle("/api/shorten", m.ThenFunc(a.createShortLink)).Methods("POST")
	a.Router.Handle("/api/info", m.ThenFunc(a.getShortLinkInfo)).Methods("GET")
	a.Router.Handle("/{shortLink:[a-zA-z0-9]{1,11}}", m.ThenFunc(a.redirect)).Methods("GET")
}

// 长地址 --生成--> 短地址
func (a *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	// 用户请求参数是 json 格式,需要解析出具体参数
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, NewBadReqErr(fmt.Errorf("parsing json form failed %v", r.Body)), nil)
		return
	}
	// 校验 Url string `validate:"nonzero"` 和 ExpirationInMinutes int64 `validate:"min=0"`
	if err := validator.Validate(req); err != nil {
		responseWithError(w, NewBadReqErr(fmt.Errorf("validate parameters failed %v", req)), nil)
		return
	}
	defer r.Body.Close()

	s, err := a.Config.S.Shorten(req.Url, req.ExpirationInMinutes)
	if err != nil {
		responseWithError(w, err, nil)
	} else {
		responseWithJson(w, http.StatusCreated, shortLinkResp{ShortLink: s})
	}
}

// 短地址解析
func (a *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	// 参数从 url string 中解析得到
	values := r.URL.Query()
	sl := values.Get("shortlink")

	info, err := a.Config.S.ShortLinkInfo(sl)
	if err != nil {
		responseWithError(w, err, nil)
	} else {
		responseWithJson(w, http.StatusOK, info)
	}
}

// 重定向
func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	// 重定向函数的 shortlink 是从变量中获取
	vars := mux.Vars(r)
	url, err := a.Config.S.UnShorten(vars["shortLink"])
	if err != nil {
		responseWithError(w, err, nil)
	} else {
		// 为了统计用户行为,使用临时重定向
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func responseWithError(w http.ResponseWriter, err error, payload interface{}) {
	// 断言 判断错误类型
	switch e := err.(type) {
	case MiError:
		log.Printf("HTTP %d - %s", e.Status(), e)
		responseWithJson(w, e.Status(), e.Error())
	default:
		responseWithJson(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func responseWithJson(w http.ResponseWriter, status int, payload interface{}) {
	resp, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(resp)
}
