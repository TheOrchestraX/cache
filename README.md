# Generic Go Cache

A lightweight, thread-safe, generic cache for Go (1.18+) that:

* Holds items of any type (`T any`) keyed by string
* Periodically reloads from a user-provided loader function
* Supports on-demand reloads
* Allows individual additions, deletions, and clearing of items
* Provides flexible search methods (`Find`, `FindOne`)

---

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
    * [Creating a Cache](#creating-a-cache)
    * [Starting and Stopping Auto-Reload](#starting-and-stopping-auto-reload)
    * [On-Demand Reload](#on-demand-reload)
    * [CRUD Operations](#crud-operations)
    * [Searching and Retrieval](#searching-and-retrieval)
* [Examples](#examples)
    * [BlogPost Cache](#blogpost-cache)
    * [Product Cache](#product-cache)
* [API Reference](#api-reference)
* [Contributing](#contributing)
* [License](#license)

---

## Installation

```bash
go get github.com/TheOrchestraX/cache
```

Import in your code:

```go
import "github.com/TheOrchestraX/cache"
```

---

## Usage

### Creating a Cache

Provide a loader function that returns a `map[string]T` and an error, plus a reload interval:

```go
// Loader signature: func() (map[string]T, error)
loader := func() (map[string]MyType, error) {
    // fetch or compute your data keyed by string
}

c := cache.NewCache(loader, 5*time.Minute)
```

### Starting and Stopping Auto-Reload

```go
c.StartAutoReload()
// ... later, when shutting down:
c.StopAutoReload()
```

### On-Demand Reload

```go
c.Reload() // immediately invoke loader and swap data
```

### CRUD Operations

* **Add** or update one item:

  ```go
  c.Add(key, value)
  ```
* **Delete** by key:

  ```go
  c.Delete(key)
  ```
* **Clear** entire cache:

  ```go
  c.Clear()
  ```

### Searching and Retrieval

* **Get** one item:

  ```go
  v, ok := c.Get(key)
  ```
* **GetAll** returns a copy of the entire map:

  ```go
  all := c.GetAll()
  ```
* **Find** multiple by predicate:

  ```go
  results := c.Find(func(item T) bool {
      // return true for matching items
  })
  ```
* **FindOne** first match:

  ```go
  item, found := c.FindOne(func(item T) bool { ... })
  ```

---

## Examples

### BlogPost Cache

```go
// Domain type

type BlogPost struct {
    Slug    string
    Title   string
    Author  string
    Content string
}

// Loader function
func loadBlogPosts() (map[string]BlogPost, error) {
    // fetch from database or API
    return map[string]BlogPost{
        "hello-world": {Slug: "hello-world", Title: "Hello, World!", Author: "Alice"},
        "go-caching":  {Slug: "go-caching", Title: "Caching in Go", Author: "Bob"},
    }, nil
}

func main() {
    blogCache := cache.NewCache(loadBlogPosts, 10*time.Minute)
    blogCache.StartAutoReload()
    defer blogCache.StopAutoReload()

    // On-demand refresh
    blogCache.Reload()

    // Lookup
    if post, ok := blogCache.Get("go-caching"); ok {
        fmt.Println(post.Title)
    }
}
```

### Product Cache

```go
// Domain type

type Product struct {
    ID    string
    Name  string
    Price float64
    Stock int
}

// Loader function
func loadProducts() (map[string]Product, error) {
    return map[string]Product{
        "p100": {ID: "p100", Name: "Mug", Price: 9.99, Stock: 100},
        "p200": {ID: "p200", Name: "T-Shirt", Price: 19.99, Stock: 50},
    }, nil
}

func main() {
    prodCache := cache.NewCache(loadProducts, 1*time.Hour)
    prodCache.StartAutoReload()
    defer prodCache.StopAutoReload()

    // Add a new product manually
    prodCache.Add("p300", Product{ID: "p300", Name: "Notebook", Price: 4.99, Stock: 200})

    // Find one under $10
    if cheap, ok := prodCache.FindOne(func(p Product) bool { return p.Price < 10 }); ok {
        fmt.Println("Cheap item:", cheap.Name)
    }
}
```

