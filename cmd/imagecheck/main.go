package main

import (
	"fmt"
	"projects/platToolv2/internal/head"
	"projects/platToolv2/internal/imageprocess"
)

func main() {
	inpath := "./template.xlsx"
	suffix := head.TimeSuffix()
	outpath := fmt.Sprintf("./data/entry_image_meta(%s).xlsx", suffix)

	imageprocess.TaskRun(inpath, outpath, suffix)

}
