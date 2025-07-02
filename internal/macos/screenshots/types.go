package screenshots

import (
	"fmt"
	"strings"
)

type Format string

const (
	FormatPng  Format = "png"
	FormatJpg  Format = "jpg"
	FormatJpeg Format = "jpeg"
	FormatPdf  Format = "pdf"
	FormatPsd  Format = "psd"
	FormatGif  Format = "gif"
	FormatTga  Format = "tga"
	FormatTiff Format = "tiff"
	FormatBmp  Format = "bmp"
	FormatHeic Format = "heic"
)

func (f Format) IsValid() bool {
	switch f {
	case FormatPng, FormatJpg, FormatJpeg, FormatPdf, FormatPsd, FormatGif, FormatTga, FormatBmp, FormatTiff, FormatHeic:
		return true
	default:
		return false
	}
}

func (f Format) String() string {
	return string(f)
}

func ParseFormat(s string) (Format, error) {
	format := Format(strings.ToLower(s))
	if !format.IsValid() {
		return "", fmt.Errorf("invalid screenshot format %q, must be one of: png, jpg, jpeg, pdf, psd, gif, tga, bmp, tiff, heic", s)
	}
	return format, nil
}

func AllFormats() []Format {
	return []Format{FormatPng, FormatJpg, FormatJpeg, FormatPdf, FormatPsd, FormatGif, FormatTga, FormatBmp, FormatTiff, FormatHeic}
}
