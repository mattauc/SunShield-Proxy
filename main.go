package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "log"
    "os"
    "strconv"
    "math"
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

func roundToDecimalPlaces(value float64, decimalPlaces int) float64 {
    factor := math.Pow(10, float64(decimalPlaces))
    return math.Round(value*factor) / factor
}

func Weather(c *gin.Context) {

    ip := c.ClientIP()
    limiter := getRateLimiter(ip)

    if !limiter.Allow() {
        c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
        return
    }

	latStr := c.Query("lat")
    lonStr := c.Query("lon")
    exclude := c.Query("exclude")
    units := c.Query("units")

    lat, err := strconv.ParseFloat(latStr, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
        return
    }

    lon, err := strconv.ParseFloat(lonStr, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
        return
    }

    lat = roundToDecimalPlaces(lat, 3)
    lon = roundToDecimalPlaces(lon, 3)

    cacheKey := fmt.Sprintf("%f:%f:%s:%s", lat, lon, exclude, units)

    fmt.Println("CACHE KEY: ", cacheKey)

    if cachedData, found := cacheStore.Get(cacheKey); found {
        fmt.Println("CACHING")
        c.Header("Access-Control-Allow-Origin", "https://sunshield.mattauc.com")
        c.String(http.StatusOK, cachedData.(string))
        return
    }

    fmt.Println("NOT CACHING")
	//uri := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%s&lon=%s&exclude=%s&units=%s&appid=%s", latStr, lonStr, exclude, units, os.Getenv("OPEN_WEATHER_TOKEN"))
    uri := "https://api.openweathermap.org/data/3.0/onecall?lat=-33.926530492173406&lon=151.25599730755016&exclude=minutely,alerts&units=metric&appid=525ccad372318b79140c5e40030e620d"
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

    fmt.Println(string(body))

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
