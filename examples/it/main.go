package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maxnilz/bplustree"
)

func main() {
	order := 4
	less := func(a, b int) bool { return a < b }
	tree := bplustree.New[int, int](order, less)
	key, value := 0, 0
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
			_, err := fmt.Sscanf(string(instruction[1:]), "%d", &key)
			if err != nil {
				log.Panicln(err)
			}
			var ok bool
			value, ok = tree.Remove(key)
			if ok {
				fmt.Printf("removed %d with value %d\n", key, value)
			}
			_ = tree.Print(os.Stdout)
		case 'i':
			_, err := fmt.Sscanf(string(instruction[1:]), "%d %d", &key, &value)
			if err != nil {
				log.Panicln(err)
			}
			tree.Insert(key, value)
			out := bytes.Buffer{}
			_ = tree.Print(&out)
			fmt.Println(out.String())
		}
	}
}
