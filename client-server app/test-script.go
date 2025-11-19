package main

import (
	"fmt"
)

func main() {
	const s string = "1G11o1L"
	var cnt int = -1
	var res string
	cnt = 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			cnt = cnt*10 + int(r-'0')
			continue
		}
		for j := 0; j < cnt; j++ {
			res += string(r)
		}
		cnt = 0
	}
	fmt.Println(res)

}
