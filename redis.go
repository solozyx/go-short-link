package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
	"github.com/speps/go-hashids"
)

const (
	// 本项目所用数据库
	Db = "db_shortlink:"
	// 全局计数器 自增ID 每次自增1 保证永远不重复
	UrlIdKey = "next.url.id"
	// 映射 长地址 和 短地址
	ShortLinkKey = "shortlink:%s:url"
	// 映射 短地址 和 哈希值
	UrlHashKey = "urlhash:%s:url"
	// 映射 短地址 和 它的详细信息
	ShortLinkDetailKey = "shortlink:%s:detail"
)

// redis客户端
type RedisCli struct {
	Cli *redis.Client
}

// Url长地址 详细信息
type UrlDetail struct {
	Url string `json:"url"`
	// 短地址创建时间
	CreateAt string `json:"create_at"`
	// 过期时间
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr string, pwd string, db int) *RedisCli {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
	// 验证redis客户端是否可用
	if _, err := client.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisCli{Cli: client}
}

// TODO : NOTICE 长地址 --> 短地址 算法
// param  url 长地址 exp 过期时间
// return 短地址
func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	// 计算长地址哈希值
	hd := hashids.NewData()
	hd.Salt = url
	hd.MinLength = 0
	h, _ := hashids.NewWithData(hd)
	urlHash, _ := h.Encode([]int{45, 434, 1313, 99})

	// 从redis缓存获取长地址哈希值
	d, err := r.Cli.Get(fmt.Sprintf(Db+UrlHashKey, urlHash)).Result()
	if err == redis.Nil {
		// no exist , nothing to do
	} else if err != nil {
		return "", err
	} else {
		// 过期 返回值是零值填充
		if d == "{}" {
			// expiration, nothing to do
		} else {
			// 返回redis缓存中的 短地址
			return d, nil
		}
	}

	// 长地址 第1次 转换为 短地址
	// 全局自增ID + 1
	err = r.Cli.Incr(Db + UrlIdKey).Err()
	if err != nil {
		return "", err
	}
	// 全局自增ID 进行 base62 编码 得到长地址对应的短地址
	id, err := r.Cli.Get(Db + UrlIdKey).Int64()
	if err != nil {
		return "", err
	}
	// 从redis取出的全局自增ID是整型值,把它编码为 既包含整型 又包含大小写字母 的短地址
	shortLink := base62.EncodeInt64(id)

	// 映射: 短地址 和 长地址 做映射
	err = r.Cli.Set(fmt.Sprintf(Db+ShortLinkKey, shortLink), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// 映射: 长地址哈希值 和 短地址 做映射
	err = r.Cli.Set(fmt.Sprintf(Db+UrlHashKey, urlHash), shortLink, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", nil
	}
	// 长地址详细信息
	detail, err := json.Marshal(&UrlDetail{
		Url:                 url,
		CreateAt:            time.Now().String(),
		ExpirationInMinutes: time.Duration(exp),
	})
	if err != nil {
		return "", err
	}

	// 映射 : 短地址 和 短地址详细信息 做映射
	err = r.Cli.Set(fmt.Sprintf(Db+ShortLinkDetailKey, shortLink), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return shortLink, nil
}

// 传入短地址 返回短地址对应的详细信息
func (r *RedisCli) ShortLinkInfo(shortLink string) (interface{}, error) {
	// 从redis缓存获取 短地址 对应详细信息
	detail, err := r.Cli.Get(fmt.Sprintf(Db+ShortLinkDetailKey, shortLink)).Result()
	if err == redis.Nil {
		return "", NewNotFindErr(errors.New("unknown short url"))
	} else if err != nil {
		return "", err
	} else {
		return detail, nil
	}
}

// 长地址 <-- 短地址
func (r *RedisCli) UnShorten(shortLink string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(Db+ShortLinkKey, shortLink)).Result()
	if err == redis.Nil {
		return "", NewNotFindErr(errors.New("unknown short url"))
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}
