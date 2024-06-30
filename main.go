package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "log"
    "os"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/patrickmn/go-cache"
    "golang.org/x/time/rate"
)

var (
    rateLimiters = make(map[string]*rate.Limiter)
    cacheStore   = cache.New(30*time.Minute, 10*time.Minute)
    mu           sync.Mutex
)

func getRateLimiter(ip string) *rate.Limiter {
    mu.Lock()
    defer mu.Unlock()

    limiter, exists := rateLimiters[ip]
    if !exists {
        limiter = rate.NewLimiter(1, 1)
        rateLimiters[ip] = limiter
    }
    return limiter
}

func Weather(c *gin.Context) {

    ip := c.ClientIP()
    limiter := getRateLimiter(ip)

    if !limiter.Allow() {
        c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
        return
    }

	lat := c.Query("lat")
    lon := c.Query("lon")
	exclude := c.Query("exclude")
	units := c.Query("units")
    
    fmt.Println("Latitude:", lat)
    fmt.Println("Longitude:", lon)
	fmt.Println("Exclude:", exclude)
	fmt.Println("Units:", units)
    cacheKey := fmt.Sprintf("%s:%s:%s:%s", lat, lon, exclude, units)

    if cachedData, found := cacheStore.Get(cacheKey); found {
        c.Header("Access-Control-Allow-Origin", "https://sunshield.mattauc.com")
        c.String(http.StatusOK, cachedData.(string))
        return
    }

	uri := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%s&lon=%s&exclude=%s&units=%s&appid=%s", lat, lon, exclude, units, os.Getenv("OPEN_WEATHER_TOKEN"))
    response, err := http.Get(uri)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer response.Body.Close()


    body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
        return
	}

    cacheStore.Set(cacheKey, string(body), cache.DefaultExpiration)
    c.Header("Access-Control-Allow-Origin", "https://sunshield.mattauc.com")
    c.String(http.StatusOK, string(body))
}

func main() {
    r := gin.Default()
    r.POST("/api/weather", Weather)
	
    certFile := "/etc/letsencrypt/live/sunshield.mattauc.com/fullchain.pem"
    keyFile := "/etc/letsencrypt/live/sunshield.mattauc.com/privkey.pem"


    err := r.RunTLS(":443", certFile, keyFile)
    if err != nil {
        log.Fatalf("Failed to run server with TLS: %v", err)
    }
}
