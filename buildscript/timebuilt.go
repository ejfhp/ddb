package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	out, _ := os.Create("timebuilt.go")
	out.Write([]byte("package main\n\n"))
	out.Write([]byte(fmt.Sprintf("var buildTimestamp string = `%s`\n", time.Now().Format("2006-01-02 15:04 MST"))))
}
