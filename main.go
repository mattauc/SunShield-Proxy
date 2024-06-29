package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "log"
    "os"
    "github.com/gin-gonic/gin"
)

func Weather(c *gin.Context) {

	lat := c.Query("lat")
    lon := c.Query("lon")
	exclude := c.Query("exclude")
	units := c.Query("units")
    
    fmt.Println("Latitude:", lat)
    fmt.Println("Longitude:", lon)
	fmt.Println("Exclude:", exclude)
	fmt.Println("Units:", units)

	uri := "https://api.openweathermap.org/data/3.0/onecall?lat=" + lat + "&lon=" + lon + "&exclude=" + exclude + "&units=" + units +"&appid=" + os.Getenv("OPEN_WEATHER_TOKEN")
	//Get Request to URI
    response, error := http.Get(uri)
    if error != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": error.Error()})
    }
    defer response.Body.Close()


    body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
        return
	}

    c.Header("Access-Control-Allow-Origin", "https://sunshield.mattauc.com")
    c.String(http.StatusOK, string(body))
}

func main() {
    r := gin.Default()
	//r.Use(secure.TrustProxy(true))

	r.Use(func(c *gin.Context) {
        // Handle HTTPS detection for proxy servers
        if c.Request.Header.Get("X-Forwarded-Proto") == "http" {
            c.Redirect(http.StatusMovedPermanently, "https://"+c.Request.Host+c.Request.RequestURI)
            return
        }
        // Set secure headers
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("Referrer-Policy", "same-origin")

        // Set CORS headers
        c.Header("Access-Control-Allow-Origin", "https://sunshield.mattauc.com")

        c.Next()
    })
	
    r.POST("/api/weather", Weather)
    //r.Run(":8000")
	
    certFile := "/etc/letsencrypt/live/sunshield.mattauc.com/fullchain.pem"
    keyFile := "/etc/letsencrypt/live/sunshield.mattauc.com/privkey.pem"


    err := r.RunTLS(":443", certFile, keyFile)
    if err != nil {
        log.Fatalf("Failed to run server with TLS: %v", err)
    }
}
