package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bitmap/wordle/internal/words"
)

// Prompt the user to enter a string
func Guess() (string, error) {
	var prompt string
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "\n  Guess?> ")
		prompt, _ = reader.ReadString('\n')
		if prompt != "" {
			break
		}
	}

	prompt = strings.TrimSpace(strings.ToLower(prompt))

	// Display an error if the user doesn't input enough chars
	if len(prompt) != 5 {
		return "", errors.New("your guess must be 5 letters long")
	}

	// Check to see if word is allowed
	if !words.IsValidWord(prompt) {
		return "", errors.New("invalid word")
	}

	return prompt, nil
}
