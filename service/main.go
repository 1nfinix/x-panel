package main

import (
	"log"
	"net"
	"strings"

	"x-panel/global"
	"x-panel/initialize"
	"x-panel/router"
)

func main() {
	err := initialize.InitApp()
	if err != nil {
		log.Println("初始化错误:", err.Error())
		panic(err)
	}
	httpPort := global.Config.GetValueStringOrDefault("base", "http_port")
	httpAddress := global.Config.GetValueString("base", "http_address")

	if err := router.InitRouters(listenAddress(httpAddress, httpPort)); err != nil {
		panic(err)
	}
}

func listenAddress(address, port string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		return ":" + port
	}
	return net.JoinHostPort(address, port)
}
