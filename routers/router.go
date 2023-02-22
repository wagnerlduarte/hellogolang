package routers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wagnerlduarte/hellogolang/midlewares/jwt"
	"github.com/wagnerlduarte/hellogolang/routers/api"
)

func ConfigureEndpoints() http.Handler {

	r := gin.Default()

	auth := r.Group("/")

	auth.Use(jwt.JwtTokenCheck)
	auth.Use(jwt.PrivateACLCheck)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/hello", func(c *gin.Context) {
		c.Writer.Write([]byte("Hello World!!!"))
	})

	auth.GET("/serie/:id", func(c *gin.Context) {

		userId := c.GetString("userId")

		id := c.Param("id")

		serie, err := api.FindSerie(userId, id)

		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, serie)
	})

	auth.GET("/series", func(c *gin.Context) {

		userId := c.GetString("userId")

		page := c.Request.URL.Query().Get("page")
		limit := c.Request.URL.Query().Get("limit")
		genre := c.Request.URL.Query().Get("genre")
		rate := c.Request.URL.Query().Get("rate")

		series, err := api.ListSeries(userId, page, limit, genre, rate)

		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, series)
	})

	return r
}
