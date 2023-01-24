package main

import (
	"bytes"
	"fmt"
	"image"
	"io"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"
	"github.com/gabriel-vasile/mimetype"
)

func EncodeWebP(reader io.Reader, mime *mimetype.MIME, conf WebPConfig) (io.Reader, *mimetype.MIME, error) {
	if !conf.Enabled {
		fmt.Println("WebP encoding is disabled")
		return reader, mime, nil
	}
	encode := false
	for _, mt := range conf.MimeTypes {
		if mime.Is(mt) {
			encode = true
			break
		}
	}
	if !encode {
		fmt.Println("WebP encoding is not enabled for this MIME type ", mime.String())
		return reader, mime, nil
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, nil, err
	}
	buf := bytes.NewBuffer(nil)
	webp.Encode(buf, img, &webp.Options{Lossless: false, Quality: float32(conf.Quality)})
	fmt.Println("WebP encoding successful")
	return buf, mimetype.Lookup("image/webp"), nil
}
