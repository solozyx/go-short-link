package main

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	// Storage接口变量
	S Storage
}

func getEnv() *Env {
	redisAddr := os.Getenv("APP_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "192.168.174.134:6379"
	}
	redisPwd := os.Getenv("APP_REDIS_PASSWORD")
	if redisPwd == "" {
		redisPwd = ""
	}
	redisDb := os.Getenv("APP_REDIS_DB")
	if redisDb == "" {
		redisDb = "0"
	}
	db, err := strconv.Atoi(redisDb)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connect to redis (addr:%s password: %s db: %d", redisAddr, redisPwd, db)
	cli := NewRedisCli(redisAddr, redisPwd, db)
	return &Env{S: cli}
}
