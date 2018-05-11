package services

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"fmt"
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

		fmt.Println("richiesta stop - popolamento chan")
		Ch <- true
		//da inserire verifica credenziali utente tecnico
		fmt.Println("richiesta stop - uscita chan")
		//da inserire verifica credenziali utente tecnico
		c.JSON(http.StatusOK, gin.H{
		})
}

func checkStatusExtractor(c *gin.Context ) {

	//da inserire verifica credenziali utente tecnico
	fmt.Println("richiesta check verifica arrivata")
	c.JSON(http.StatusOK, gin.H{
	})
}
