package main

import "os"

func main() {
	setEnv()
	a := App{}
	a.Initialize(getEnv())
	a.Run(":8000")
}

func setEnv() {
	_ = os.Setenv("APP_REDIS_ADDR", "localhost:6379")
	_ = os.Setenv("APP_REDIS_PASSWORD", "")
	_ = os.Setenv("APP_REDIS_DB", "0")
}
