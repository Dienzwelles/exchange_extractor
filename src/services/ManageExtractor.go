package services

import (
	"github.com/gin-gonic/gin"
	"net/http"

)

var router *gin.Engine
var Ch chan<- bool


func StopWorkpool(ch chan<- bool ) {
	router = gin.Default()
	initializeRoutes(ch)
	router.Run()
}


func initializeRoutes(ch chan<- bool ) {
	Ch = ch
	router.GET("/stopExtractor", stopPool)
	router.GET("/checkStatusExtractor", checkStatusExtractor)

}


func stopPool(c *gin.Context ) {

		Ch <- true
		//da inserire verifica credenziali utente tecnico
		c.JSON(http.StatusOK, gin.H{
		})
}

func checkStatusExtractor(c *gin.Context ) {

	//da inserire verifica credenziali utente tecnico
	c.JSON(http.StatusOK, gin.H{
	})
}
