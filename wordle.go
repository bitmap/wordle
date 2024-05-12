// Wordle - a word game - for the command line
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
const totalGuesses int8 = 6
const emptySpaceRune = '*'

var correctAnswer = words.RandomAnswer()

type Guess struct {
	value      rune
	isCorrect  bool
	isIncluded bool
	isUsed     bool
}

// Game is a x * y grid
type GameGrid [totalGuesses][wordLength]Guess
type KeyboardMap map[rune]Guess

var gameGrid = GameGrid{[wordLength]Guess{}}
var keyboardMap = KeyboardMap{}
var keySlice = []rune("abcdefghijklmnopqrstuvwxyz")

// Print the current state of the game
func printGameGrid() {
	for i := range gameGrid {
		fmt.Print(" ")

		for j := range gameGrid[i] {
			currentChar := gameGrid[i][j]

			// Set the color of the character
			runeColor := color.White
			if currentChar.isCorrect {
				runeColor = color.Green
			} else if currentChar.isIncluded {
				runeColor = color.Yellow
			}

			fmt.Print(runeColor + " " + strings.ToUpper(string(currentChar.value)) + " " + color.Reset)
		}
		fmt.Println()
	}
	fmt.Println()
}

func setKeyColor(key rune) {
	keyColor := color.White

	if keyboardMap[key].isCorrect {
		keyColor = color.Green
	} else if keyboardMap[key].isIncluded {
		keyColor = color.Yellow
	} else if keyboardMap[key].isUsed {
		keyColor = color.Gray
	}

	print(keyColor + strings.ToUpper(string(key)) + color.Reset)
}

func printKeyboard() {
	fmt.Print("  ")

	// Print used keys a-m
	for _, v := range keySlice[0:13] {
		setKeyColor(v)
	}

	fmt.Println()
	fmt.Print("  ")

	// Print used keys n-z
	for _, v := range keySlice[13:] {
		setKeyColor(v)
	}

	fmt.Println()
}

func clearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()
}

func init() {
	// Initialize the grid
	for i := range gameGrid {
		for j := range gameGrid[i] {
			gameGrid[i][j].value = emptySpaceRune
		}
	}

	// Initialize the keyboard map. Keys from the keyboardMap are unsorted,
	// so we just use the slice for display
	for _, key := range keySlice {
		keyboardMap[key] = Guess{
			value:      key,
			isUsed:     false,
			isCorrect:  false,
			isIncluded: false,
		}
	}
}

func main() {
	clearScreen()
	winFlag := false

	// Loop until we're out of guesses
	var guessCount int8 = 0
	for guessCount < totalGuesses {
		fmt.Println("\nWelcome to Wordle")

		// Print state of the game
		printGameGrid()
		printKeyboard()

		// Get user input
		currentGuess, err := prompt.Guess()

		if err != nil {
			clearScreen()
			fmt.Print(color.Red + err.Error() + color.Reset)
			continue
		}

		for index := range gameGrid[guessCount] {
			char := rune(currentGuess[index])

			// Check if it's the same character at that index...
			isCorrect := rune(correctAnswer[index]) == char
			// ...or string contains the character elsewhere
			isIncluded := strings.ContainsRune(correctAnswer, char)

			// Update the values
			gameGrid[guessCount][index].value = char
			gameGrid[guessCount][index].isUsed = true
			gameGrid[guessCount][index].isCorrect = isCorrect
			gameGrid[guessCount][index].isIncluded = isIncluded

			// Update the character in the keyboard map
			keyboardMap[char] = Guess{
				value:  char,
				isUsed: true,
				// Use previous values if new guess has different placements
				isCorrect:  keyboardMap[char].isCorrect || isCorrect,
				isIncluded: keyboardMap[char].isIncluded || isIncluded,
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
	printGameGrid()

	if winFlag {
		if guessCount == 1 {
			fmt.Println("🫨 Woah! You got it right on the first try! Nice!")
		} else {
			fmt.Println("🎉 Correct! You won in " + fmt.Sprint(guessCount) + " guesses.")
		}
	} else {
		fmt.Println("😓 Sorry, the answer was " + color.Green + correctAnswer + color.Reset + ". Try again.")
	}
}
