// Wordle - a word game - for the command line.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitmap/wordle/internal/color"
	"github.com/bitmap/wordle/internal/prompt"
	"github.com/bitmap/wordle/internal/words"
)

const wordLength = 5
const totalGuesses = 6
const emptySpaceRune = '•'
const aplhabet = "abcdefghijklmnopqrstuvwxyz"

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
	for i := range g {
		fmt.Print(" ")
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
	fmt.Print("  ")

	// Print used keys a-m.
	for _, v := range aplhabet[0:13] {
		l[v].Render()
	}

	fmt.Println()
	fmt.Print("  ")

	// Print used keys n-z.
	for _, v := range aplhabet[13:] {
		l[v].Render()
	}

	fmt.Println()
}

// Initialize the letters map. Keys from the guessedLetters are unsorted,
// so we just use the slice for display
func (g letterMap) init() {
	for _, key := range aplhabet {
		guessedLetters[key] = guess{
			value: key,
			state: 0,
		}
	}
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
	)

	game.init()
	guessedLetters.init()
	clearScreen()

	// Loop until we're out of guesses.
	for guessCount < totalGuesses {
		fmt.Println("\nWelcome to Wordle")

		// Print state of the game
		game.render()
		guessedLetters.render()

		// Get user input
		currentGuess, err := prompt.Guess()

		if err != nil {
			clearScreen()
			fmt.Print(color.Red + err.Error() + color.Reset)
			continue
		}

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
		clearScreen()
	}

	// Print final game state
	clearScreen()
	fmt.Println("\n    Game Over")
	game.render()

	if winFlag {
		if guessCount == 1 {
			fmt.Println("🫨 Woah! You got it right on the first try! Nice!")
		} else {
			fmt.Println("🎉 Correct! You won in " + fmt.Sprint(guessCount) + " guesses.")
		}
	} else {
		fmt.Println("😓 Sorry, the answer was " + color.Green + answer + color.Reset + ".")
	}

	// Ask user to play again
	if prompt.Retry() {
		main()
	}
}
