package main

import (
	"fmt"
	"projects/platToolv2/internal/entrypage"
)

func main() {

	inpath := "./URLs.xlsx"
	outpath := "./data/entry_meta.xlsx"
	docxpath := "./data/entries(文本).docx"
	timeout := 5
	sleep := 500
	entrypage.TaskRun(inpath, outpath, docxpath, timeout, sleep)

	fmt.Println("\npess any key to exit:")
	fmt.Scanln()
}
