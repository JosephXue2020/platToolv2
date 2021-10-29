package main

import (
	"fmt"
	"projects/platToolv2/internal/entrypage"
	"projects/platToolv2/internal/head"
)

func main() {

	inpath := "./URLs.xlsx"
	// outpath := "./data/entry_meta.xlsx"
	// docxpath := "./data/entries(文本).docx"

	suffix := head.TimeSuffix()
	outpath := fmt.Sprintf("./data/entry_meta(%s).xlsx", suffix)
	docxpath := fmt.Sprintf("./data/entries文本(%s).xlsx", suffix)

	timeout := 5
	sleep := 500
	entrypage.TaskRun(inpath, outpath, docxpath, timeout, sleep)

	fmt.Println("\npess any key to exit:")
	fmt.Scanln()
}
