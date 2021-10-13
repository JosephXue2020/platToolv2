package main

import (
	"fmt"
	"projects/platToolv2/internal/querypage"
)

func main() {
	fmt.Println("Starting line...")
	// 测试：
	// q := querypage.NewQuery("光学", "1", "title", "all", 5, 500)
	// fmt.Printf("Query struct is: %+v\n", *q)
	// fmt.Print("\n")

	inpath := "./template.xlsx"
	outpath := "./data/result.xlsx"
	accuracy := "1"
	field := "title"
	scope := "all"
	timeout := 5
	sleep := 500
	querypage.TaskRun(inpath, outpath, accuracy, field, scope, timeout, sleep)

	fmt.Print("pess any key to exit:")
	fmt.Scanln()
}
