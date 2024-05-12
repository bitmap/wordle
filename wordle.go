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

type letterState int

type guess struct {
	value rune
	state letterState
}

const (
	_ letterState = iota
	guessed
	containsLetter
	isCorrect
)

const wordLength uint8 = 5
const totalGuesses uint8 = 6
const emptySpaceRune = '•'

// Game is a x * y grid.
var guessesGrid [totalGuesses][wordLength]guess

// Map of all guessed letters.
var guessedLetters = make(map[rune]guess)

// This slice is used to init the map and display the guess map.
const lettersSlice = "abcdefghijklmnopqrstuvwxyz"

// Prints guess rune in color.
func printGuessCharacter(guess guess) {
	var keyColor string

	switch guess.state {
	case isCorrect:
		keyColor = color.Green
	case containsLetter:
		keyColor = color.Yellow
	case guessed:
		keyColor = color.Gray
	default:
		keyColor = color.White
	}

	print(keyColor + strings.ToUpper(string(guess.value)) + color.Reset)
}

// Print the current state of the game.
func printGuessesGrid() {
	for i := range guessesGrid {
		fmt.Print(" ")
		for j := range guessesGrid[i] {
			currentChar := guessesGrid[i][j]
			fmt.Print(" ")
			printGuessCharacter(currentChar)
			fmt.Print(" ")
		}
		fmt.Println()
	}
	fmt.Println()
}

// Print the map of guessed letters and their state.
func printGuessedLettersMap() {
	fmt.Print("  ")

	// Print used keys a-m.
	for _, v := range lettersSlice[0:13] {
		printGuessCharacter(guessedLetters[v])
	}

	fmt.Println()
	fmt.Print("  ")

	// Print used keys n-z.
	for _, v := range lettersSlice[13:] {
		printGuessCharacter(guessedLetters[v])
	}

	fmt.Println()
}

func initializeGame() {
	// Initialize the game grid.
	for i := range guessesGrid {
		for j := range guessesGrid[i] {
			guessesGrid[i][j] = guess{
				value: emptySpaceRune,
				state: 0,
			}
		}
	}

	// Initialize the letters map. Keys from the guessedLetters are unsorted,
	// so we just use the slice for display
	for _, key := range lettersSlice {
		guessedLetters[key] = guess{
			value: key,
			state: 0,
		}
	}
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	var correctAnswer = words.RandomAnswer()
	var guessCount uint8 = 0
	var winFlag = false

	initializeGame()
	clearScreen()

	// Loop until we're out of guesses.
	for guessCount < totalGuesses {
		fmt.Println("\nWelcome to Wordle")

		// Print state of the game
		printGuessesGrid()
		printGuessedLettersMap()

		// Get user input
		currentGuess, err := prompt.Guess()

		if err != nil {
			clearScreen()
			fmt.Print(color.Red + err.Error() + color.Reset)
			continue
		}

		for i := range guessesGrid[guessCount] {
			charValue := rune(currentGuess[i])
			var charState letterState

			switch {
			// Check if it's the same character at that index...
			case rune(correctAnswer[i]) == charValue:
				charState = isCorrect
			// ...or string contains the character elsewhere
			case strings.ContainsRune(correctAnswer, charValue):
				charState = containsLetter
			default:
				charState = guessed
			}

			// Update the values
			guessesGrid[guessCount][i].value = charValue
			guessesGrid[guessCount][i].state = charState

			// For letters map, check previous state if placement differs
			currentCharState := guessedLetters[charValue].state
			if charState > currentCharState {
				currentCharState = charState
			}

			// Update the character in the letters map
			guessedLetters[charValue] = guess{
				value: charValue,
				state: charState,
			}
		}

		// Increment the guess counter
		guessCount++

		// Stop looping if we found the answer
		if currentGuess == correctAnswer {
			winFlag = true
			break
		}
		clearScreen()
	}

	// Print final game state
	clearScreen()
	fmt.Println("\n    Game Over")
	printGuessesGrid()

	if winFlag {
		if guessCount == 1 {
			fmt.Println("🫨 Woah! You got it right on the first try! Nice!")
		} else {
			fmt.Println("🎉 Correct! You won in " + fmt.Sprint(guessCount) + " guesses.")
		}
	} else {
		fmt.Println("😓 Sorry, the answer was " + color.Green + correctAnswer + color.Reset + ".")
	}

	// Ask user to play again
	if prompt.Retry() {
		main()
	}
}
