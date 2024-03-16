package gopngquant

import (
	"os"
	"testing"
	"time"
)

func Test_FileCompression(t *testing.T) {
	exit := make(chan os.Signal)
	go func() {
		for range time.Tick(time.Second) {
			CompressFile("./images/example.png", "./images/output_gopngquant.png")
		}
	}()
	<-exit
}
