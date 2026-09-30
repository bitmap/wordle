package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/bitmap/wordle-cli/internal/words"
	"golang.org/x/term"
)

const wordLength = 5

// Shared across reads so bytes buffered by one call aren't dropped by the next.
var reader = bufio.NewReader(os.Stdin)

// The indent that centers something of the given width, measured fresh
// each time so the layout follows the terminal when it's resized.
func centerIndent(width int) int {
	cols, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || cols <= 0 {
		return 0
	}

	return max(cols/2-width/2, 0)
}

// The indent that centers a prompt, leaving room for the word typed
// after the label.
func promptIndent(label string) int {
	return centerIndent(len(label) + wordLength)
}

// Returns trimmed & lowercase response to user input. An empty response is
// returned as-is so callers can treat it as their default answer.
func promptString(str string) (string, error) {
	fmt.Fprint(os.Stderr, str)

	prompt, err := readString()

	return strings.TrimSpace(strings.ToLower(prompt)), err
}

// Prompt the user to guess a word. Any errMsg from the previous attempt
// is shown below the prompt. repaint redraws the screen above it, and is
// given the guess so far so it can show the letters as they're typed.
func Guess(errMsg string, repaint func(string)) (string, error) {
	const label = "Guess?> "

	prompt, err := readLine(label, wordLength, errMsg, repaint)

	// Fall back to a plain line read when stdin isn't an interactive
	// terminal, so piped input still works.
	if errors.Is(err, errNoRawMode) {
		prompt, err = promptString("\n" + label)
	}

	// Ctrl-C, Ctrl-D, or a closed stdin all mean we're done.
	if errors.Is(err, errInterrupted) || errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr)
		os.Exit(0)
	}

	if err != nil {
		panic(err)
	}

	// Display an error if the user doesn't input enough chars
	if len([]rune(prompt)) != wordLength {
		return "", errors.New("your guess must be 5 letters long")
	}

	// Check to see if word is allowed
	if !words.IsValidWord(prompt) {
		return "", errors.New("invalid word")
	}

	return prompt, nil
}

// Prompt the user to play again. A single keypress answers it.
func Retry() bool {
	const question = "Play again? [Y/n] "

	fmt.Fprint(os.Stderr, "\n"+strings.Repeat(" ", promptIndent(question))+question)

	char, err := readKey("yn\r\n")

	// Without a terminal there are no single keypresses, so read a line.
	if errors.Is(err, errNoRawMode) {
		var line string
		line, err = readString()
		char = byte('y')
		if line != "" {
			char = line[0]
		}
	}

	if errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr)
		return false
	}
	if err != nil {
		panic(err)
	}

	// Anything but an explicit no starts another game.
	return unicode.ToLower(rune(char)) != 'n'
}
