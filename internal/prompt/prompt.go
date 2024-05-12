package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bitmap/wordle/internal/words"
)

// Returns trimmed & lowercase response to user input
func promptString(str string) (string, error) {
	var prompt string
	var err error

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, str)
		prompt, err = reader.ReadString('\n')
		if prompt != "" {
			break
		}
	}
	prompt = strings.TrimSpace(strings.ToLower(prompt))

	return prompt, err
}

// Prompt the user to guess a word.
func Guess() (string, error) {
	prompt, err := promptString("\n  Guess?> ")
	if err != nil {
		panic(err)
	}

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

// Prompt the user to play again.
func Retry() bool {
	prompt, err := promptString("\nPlay again? [y/N]")
	if err != nil {
		panic(err)
	}

	if prompt == "y" || prompt == "Y" {
		return true
	}

	return false
}
