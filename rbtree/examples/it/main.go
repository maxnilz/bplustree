package main

import (
	"bufio"
	"fmt"
	"github.com/maxnilz/tree/rbtree"
	"log"
	"os"
	"strings"
)

func main() {
	compare := func(a, b int) int {
		if a == b {
			return 0
		}
		if a < b {
			return -1
		}
		return 1
	}
	tree := rbtree.New[int](compare)
	value := 0
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		if len(text) == 0 {
			continue
		}
		instruction := []byte(text)
		cmd := instruction[0]
		switch cmd {
		case 'd':
			_, err := fmt.Sscanf(string(instruction[1:]), "%d", &value)
			if err != nil {
				log.Println(err)
			}
			var ok bool
			value, ok = tree.Remove(value)
			if ok {
				fmt.Printf("removed %d with value %d\n", value, value)
			}
			_ = tree.Print(os.Stdout)
		case 'i':
			_, err := fmt.Sscanf(string(instruction[1:]), "%d", &value)
			if err != nil {
				log.Println(err)
			}
			tree.Insert(value)
			_ = tree.Print(os.Stdout)
		}
	}
}
