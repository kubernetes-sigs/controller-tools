package alpha

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// yesno reads from stdin looking for one of "y", "yes", "n", "no" and returns
// true for "y" and false for "n"
func yesno() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		switch readstdin(reader) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}

// readstdin reads a line from stdin trimming spaces, and returns the value.
// log.Fatal's if there is an error.
func readstdin(reader *bufio.Reader) string {
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(text)
}
