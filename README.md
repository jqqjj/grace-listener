# graceful
A simple package for creating socket listeners to restart servers gracefully.

## Installation
Import the package by using go mod:
```go
import "github.com/jqqjj/graceful"
```


## Usage

Use in net/http:
```go
ln, err := graceful.NewGraceListener(":8080")
if err != nil {
    log.Fatalln(err)
}
http.Serve(ln, nil)
```

Use in gin framework:
```go
ln, err := graceful.NewGraceListener(":8080")
if err != nil {
    log.Fatalln(err)
}

router := gin.Default()
router.GET("/", func(c *gin.Context) {
    c.String(http.StatusOK, "this is a test")
})
router.RunListener(ln)
```

## Graceful restart:
```shell
kill -HUP PID
```

## Run Args:
* **daemon** &nbsp;&nbsp; Run app as a daemon

## Thanks:
* [mitchellh/go-ps](https://github.com/mitchellh/go-ps) &nbsp;&nbsp; Find, list, and inspect processes from Go (golang).
