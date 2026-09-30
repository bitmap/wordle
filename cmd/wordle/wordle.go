// Wordle - a word game - for the command line.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/bitmap/wordle-cli/internal/color"
	"github.com/bitmap/wordle-cli/internal/prompt"
	"github.com/bitmap/wordle-cli/internal/words"
	"golang.org/x/term"
)

const wordLength = 5
const totalGuesses = 6
const emptySpaceRune = '•'

// Indents that center the title and grid over the keyboard block.
const (
	gridIndent = 4
	gridWidth  = wordLength * 3
)

// Keyboard rows, laid out and ordered like a QWERTY keyboard.
var keyboard = [...]string{
	"qwertyuiop",
	"asdfghjkl",
	"zxcvbnm",
}

type letterState int

const (
	_ letterState = iota
	isGuessed
	isInWord
	isCorrect
)

type guess struct {
	value rune
	state letterState
}

// Prints guess rune in color.
func (g guess) Render() {
	var keyColor string

	switch g.state {
	case isCorrect:
		keyColor = color.Green
	case isInWord:
		keyColor = color.Yellow
	case isGuessed:
		keyColor = color.Gray
	default:
		keyColor = color.White
	}

	print(keyColor + strings.ToUpper(string(g.value)) + color.Reset)
}

type gameGrid [totalGuesses][wordLength]guess

// Game is a x * y grid.
var game gameGrid

// Print the current state of the game.
func (g gameGrid) render() {
	indent := centerLine() - gridWidth/2
	for i := range g {
		fmt.Print(strings.Repeat(" ", indent))
		for j := range g[i] {
			currentChar := g[i][j]
			fmt.Print(" ")
			currentChar.Render()
			fmt.Print(" ")
		}
		fmt.Println()
	}
	fmt.Println()
}

// Initialize the game grid.
func (g gameGrid) init() {
	for i := range game {
		for j := range game[i] {
			game[i][j] = guess{
				value: emptySpaceRune,
				state: 0,
			}
		}
	}
}

type letterMap map[rune]guess

// Map of all guessed letters.
var guessedLetters = letterMap{}

// Print the map of guessed letters and their state.
func (l letterMap) render() {
	// The widest row sets the block, and each row below it is staggered.
	indent := centerLine() - len(keyboard[0]) + 1
	for i, row := range keyboard {
		fmt.Print(strings.Repeat(" ", indent+i))

		for _, v := range row {
			l[v].Render()
			fmt.Print(" ")
		}

		fmt.Println()
	}
}

// Initialize the letters map. Keys from the guessedLetters are unsorted,
// so we just use the slice for display
func (g letterMap) init() {
	for _, row := range keyboard {
		for _, key := range row {
			guessedLetters[key] = guess{
				value: key,
				state: 0,
			}
		}
	}
}

// The column everything is centered on. Falls back to the grid's own
// middle when the terminal size isn't available, as when output is piped.
func centerLine() int {
	if cols, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && cols > 0 {
		return cols / 2
	}

	return gridIndent + gridWidth/2
}

// Print a line centered on the screen. Color escapes take up no space on
// screen, and emoji take up two columns, so neither can be measured by
// counting runes alone.
func printCentered(line string) {
	var width int
	var inEscape bool

	for _, r := range line {
		switch {
		case r == '\033':
			inEscape = true
		case inEscape:
			// Escapes run until their final letter.
			if unicode.IsLetter(r) {
				inEscape = false
			}
		case r > 0x2000:
			width += 2
		default:
			width++
		}
	}

	if pad := centerLine() - width/2; pad > 0 {
		fmt.Print(strings.Repeat(" ", pad))
	}

	fmt.Println(line)
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func main() {
	var (
		answer     = words.RandomAnswer()
		winFlag    = false
		guessCount = 0
		lastError  string
	)

	game.init()
	guessedLetters.init()

	// Draw everything above the prompt, showing the guess being typed in
	// the current row. Called on every keystroke and on resize.
	drawScreen := func(typing string) {
		for i := range game[guessCount] {
			value := emptySpaceRune
			if i < len([]rune(typing)) {
				value = []rune(typing)[i]
			}

			game[guessCount][i].value = value
		}

		clearScreen()
		fmt.Println()
		printCentered("Welcome to Wordle")
		game.render()
		guessedLetters.render()
	}

	// Loop until we're out of guesses.
	for guessCount < totalGuesses {
		drawScreen("")

		// Get user input
		currentGuess, err := prompt.Guess(lastError, drawScreen)

		if err != nil {
			lastError = err.Error()
			continue
		}
		lastError = ""

		for i := range game[guessCount] {
			charValue := rune(currentGuess[i])
			var charState letterState

			switch {
			// Check if it's the same character at that index...
			case rune(answer[i]) == charValue:
				charState = isCorrect
			// ...or string contains the character elsewhere
			case strings.ContainsRune(answer, charValue):
				charState = isInWord
			default:
				charState = isGuessed
			}

			// Update the values
			game[guessCount][i].value = charValue
			game[guessCount][i].state = charState

			// For letters map, check previous state if placement differs
			currentCharState := guessedLetters[charValue].state
			if charState > currentCharState {
				currentCharState = charState
			}

			// Update the character in the letters map
			guessedLetters[charValue] = guess{
				value: charValue,
				state: currentCharState,
			}
		}

		// Increment the guess counter
		guessCount++

		// Stop looping if we found the answer
		if currentGuess == answer {
			winFlag = true
			break
		}
	}

	// Print final game state
	clearScreen()
	fmt.Println()
	printCentered("Game Over")
	game.render()

	if winFlag {
		if guessCount == 1 {
			printCentered("🫨 Woah! You got it right on the first try!")
		} else {
			printCentered("🎉 Correct! You won in " + fmt.Sprint(guessCount) + " guesses.")
		}
	} else {
		printCentered("😓 Sorry, the answer was " + color.Green + strings.ToUpper(answer) + color.Reset + ".")
	}

	// Ask user to play again
	if prompt.Retry() {
		main()
	}
}
