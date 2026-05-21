package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebServerApi struct{}

var web_server_engine *gin.Engine
var web_server_api *WebServerApi

func (api *WebServerApi) Index(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Rich Chat!")
}

func initialize() {
	web_server_engine = gin.Default()
	web_server_api = &WebServerApi{}
	web_server_engine.GET("/", web_server_api.Index)
}

func main() {
	initialize()
	web_server_engine.Run(":2316")
}
