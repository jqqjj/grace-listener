# graceful
A simple package for creating socket listeners to restart servers gracefully.

## Usage

http package:
```go
ln, err := graceful.NewGraceListener(":8080")
if err != nil {
    log.Fatalln(err)
}
http.Serve(ln, nil)
```

gin framework:
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

## Restart command:
```shell
kill -HUP PID
```

## Run in background
```shell
./mygraceapp -daemon
```

## Thanks:
* [mitchellh/go-ps](https://github.com/mitchellh/go-ps) &nbsp;&nbsp; Find, list, and inspect processes from Go (golang).
