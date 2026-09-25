package tui

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

func RenderAlbumArt(imgData []byte, widthCols, heightRows int) string {
	if len(imgData) == 0 {
		return emptyArtBox(widthCols, heightRows)
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return emptyArtBox(widthCols, heightRows)
	}

	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	targetPxW := widthCols
	targetPxH := heightRows * 2

	var sb strings.Builder

	for y := 0; y < targetPxH; y += 2 {
		for x := 0; x < targetPxW; x++ {
			srcX := (x * imgW) / targetPxW
			srcYTop := (y * imgH) / targetPxH
			rt, gt, bt, _ := img.At(srcX, srcYTop).RGBA()

			srcYBot := ((y + 1) * imgH) / targetPxH
			rb, gb, bb, _ := img.At(srcX, srcYBot).RGBA()

			rt, gt, bt = rt>>8, gt>>8, bt>>8
			rb, gb, bb = rb>>8, gb>>8, bb>>8

			sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀\033[0m", rt, gt, bt, rb, gb, bb))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func emptyArtBox(width, height int) string {
	var sb strings.Builder
	for y := 0; y < height; y++ {
		if y == height/2 {
			pad := (width - 10) / 2
			sb.WriteString(strings.Repeat(" ", pad) + "[ NO ART ]" + strings.Repeat(" ", width-pad-10) + "\n")
		} else {
			sb.WriteString(strings.Repeat(" ", width) + "\n")
		}
	}
	return sb.String()
}
