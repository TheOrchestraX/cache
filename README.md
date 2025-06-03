# cache module
This project is a simple cache system that can be used for Websites and API (Graphql/REST)

Quick Start: cache-module
A lightweight Go module for in-memory caching of map[string]interface{} data with periodic reloads and simple lookup/search helpers.

⸻

### Installation
#### 1.	Add as a Go module dependency
* In your project’s go.mod, require and (if developing locally) replace:
```bash
require github.com/TheOrchestraX/cache
```

#### 2.	Import in your code
```go 
import "github.com/TheOrchestraX/cache"
```


### Usage

### 1. Define a Loader Function

* Your loader should fetch/compute data and return a map[string]interface{} keyed however you like (e.g. by “slug”, “ID”, etc.), along with an error if something goes wrong. For example, in a CMS context:

```go 
func loadPages() (map[string]interface{}, error) {
pages, err := cmsClient.GetPages()
if err != nil {
return nil, err
}
m := make(map[string]interface{}, len(pages))
for _, p := range pages {
m[p.Slug] = p
}
return m, nil
}
```

### 2. Create a Cache Instance

* Pick a reload interval (e.g. one hour). Then:

```go 
import (
"time"
"github.com/TheOrchestraX/cache"
)

pageCache := cache.NewCache(loadPages, time.Hour)
```

* loadPages is your loader function.
* time.Hour sets the reload frequency to every hour.

### 3. Initial Load & Start Auto-Reload

* Before serving any requests, do an initial load and then start automatic reload:
```go 
pageCache.Load()            // fetches data right now
pageCache.StartAutoReload() // schedule future reloads every hour
```
* If you ever need to stop the ticker (for example, on shutdown), call:
```go 
pageCache.StopAutoReload()
```

### 4. Retrieving Items

* Get by Key

```go
value, found := pageCache.Get("some-page-slug")
if !found {
// handle missing key (e.g. return 404)
}
page := value.(cms.Page) // type-assert back to your domain type

Get All Items

all := pageCache.GetAll()
// 'all' is a copy of the underlying map[string]interface{}
```


### Search/Filter
*	FindOne: returns the first item matching your predicate

```go 
match, ok := pageCache.FindOne(func(item interface{}) bool {
p := item.(cms.Page)
return p.Order == 1 // for example, find the page with Order == 1
})
if ok {
firstPage := match.(cms.Page)
// …
}
```

*	Find: returns all items matching your predicate
```Go
matches := pageCache.Find(func(item interface{}) bool {
p := item.(cms.Page)
return strings.Contains(p.Title, "Home")
})
for _, v := range matches {
p := v.(cms.Page)
// do something with each matching page
}
```


### 5. Dynamic Reload Interval (Optional)

If you want to change how often the cache reloads at runtime:
```Go 
pageCache.SetInterval(30 * time.Minute) // now reload every 30 minutes
```

⸻

### Example in Context
```Go
package main

import (
"fmt"
"log"
"time"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "github.com/TheOrchestraX/cache"
    "project/website/internal/cms"
    "project/website/internal/handlers"
    "project/website/internal/config"
)

func main() {
// Load environment variables
godotenv.Load()

    // Load app config (e.g. CMS_ENDPOINT, CMS_TOKEN, PORT)
    cfg := config.Load()
    cmsClient := cms.NewClient(cfg.CMSEndpoint, cfg.CMSToken)

    // Define loader functions
    loadPages := func() (map[string]interface{}, error) {
        pages, err := cmsClient.GetPages()
        if err != nil {
            return nil, err
        }
        m := make(map[string]interface{}, len(pages))
        for _, p := range pages {
            m[p.Slug] = p
        }
        return m, nil
    }

    loadAbout := func() (map[string]interface{}, error) {
        about, err := cmsClient.GetAbout()
        if err != nil {
            return nil, err
        }
        return map[string]interface{}{about.Slug: about}, nil
    }

    // Create caches (reload every hour)
    pageCache := cache.NewCache(loadPages, time.Hour)
    aboutCache := cache.NewCache(loadAbout, time.Hour)

    // Initial load + auto-reload
    pageCache.Load()
    aboutCache.Load()
    pageCache.StartAutoReload()
    aboutCache.StartAutoReload()
    defer pageCache.StopAutoReload()
    defer aboutCache.StopAutoReload()

    // Gin router with cache-backed handlers
    router := gin.Default()
    handlers.RegisterRoutes(router, pageCache, aboutCache /* …other caches… */)

    port := fmt.Sprintf(":%d", cfg.Port)
    log.Printf("Listening on %s", port)
    router.Run(port)
}
```


### Summary
*	cache.NewCache(loader, interval): construct a new cache.
*	Load(): fetch data immediately.
*	StartAutoReload(): schedule reloads every interval.
*	StopAutoReload(): shut off auto-reload if needed.
*	Get(key): retrieve one item by its key.
*	GetAll(): copy of entire map.
*	Find(predicate), FindOne(predicate): simple search helpers.
*	SetInterval(newInterval): change reload frequency on the fly.

This module can be used with any loader function (database, REST/GraphQL API, filesystem, etc.) that returns a map[string]interface{}. 
Just pass it into NewCache and wire up your HTTP handlers or other consumers.