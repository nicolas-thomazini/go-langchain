package main

import (
	"langchaingo/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routes.GetVacationRouter(r)
	r.Run()
}
