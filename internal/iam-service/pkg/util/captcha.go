package util

import (
	"github.com/mojocn/base64Captcha"
)

const (
	width   = 80                                 // png wide
	height  = 40                                 // png high
	length  = 4                                  // Verification code character length
	captcha = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ" // Verification code character set, upper case, exclude 1 0 I O
)

var d *base64Captcha.DriverString

// GenerateCaptcha generates random verification codes
// return code, b64s, err
func GenerateCaptcha() (code, b64s string, err error) {
	// lazy init
	if d == nil {
		d = &base64Captcha.DriverString{
			Height: height,
			Width:  width,
			Length: length,
			Source: captcha,
			Fonts:  []string{"chromohv.ttf"},
		}
		//d.ShowLineOptions = base64Captcha.OptionShowHollowLine + base64Captcha.OptionShowSineLine
		d = d.ConvertFonts()
	}
	// generate rand code
	code = base64Captcha.RandText(d.Length, d.Source)
	// generate png base64
	item, err := d.DrawCaptcha(code)
	if err != nil {
		return "", "", err
	}
	b64s = item.EncodeB64string()
	return code, b64s, nil
}

func RandText(size int) string {
	if size <= 0 {
		size = length
	}
	return base64Captcha.RandText(size, captcha)
}
