package selfqr

import (
	"bytes"
	"fmt"
	"github.com/skip2/go-qrcode"
	"image"
	"io/ioutil"
	"testing"
)

func TestSelfqr(t *testing.T) {
	code := New("https://github.com/caozhigang515/self-qrcode", 500, qrcode.High)

	bjFile, _ := ioutil.ReadFile("bg.png")
	bjImg, _, _ := image.Decode(bytes.NewReader(bjFile))

	code.ForeImage(bjImg, true)

	code.CodePoint(CpsCircular, 0.8)
	code.BullEye(BesCircular, BesCircular)

	file, _ := ioutil.ReadFile("logo.png")
	logoImg, _, _ := image.Decode(bytes.NewReader(file))

	code.Logo(logoImg, false)
	result, err := code.Result(true)
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = ioutil.WriteFile("qr.png", result, 0777)
}
