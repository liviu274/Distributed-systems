package main

import (
	"fmt"

	"golang.org/x/tour/tree"
)

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	if t == nil {
		return
	}
	Walk(t.Left, ch)
	ch <- t.Value
	Walk(t.Right, ch)
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	c1 := make(chan int)
	c2 := make(chan int)
	go Walk(t1, c1)
	go Walk(t2, c2)

	for i := 0; i < 10; i++ {
		x, y := <-c1, <-c2
		if x != y {
			return false
		}
	}
	return true
}

func main() {
	t := tree.New(1)
	fmt.Println(t)

	result := Same(tree.New(1), tree.New(1))
	fmt.Println(result)

}
