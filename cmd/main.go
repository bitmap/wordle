// Wordle - a word game - for the command line
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bitmap/wordle-cli/internal/color"
	"github.com/bitmap/wordle-cli/internal/words"
)

const wordLength = 5
const totalGuesses int8 = 6
const emptySpaceRune = '*'

var correctAnswer = words.RandomAnswer()

type Guess struct {
	value      rune
	isCorrect  bool
	isIncluded bool
}

// Game is a x * y grid
type GameGrid [totalGuesses][wordLength]Guess
type KeyboardMap map[rune]bool

var gameGrid = GameGrid{[wordLength]Guess{}}
var keyboardMap = KeyboardMap{}
var keySlice = []rune("abcdefghijklmnopqrstuvwxyz")

// Prompt the user to enter a string
func guessPrompt(input string) string {
	var prompt string
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, input+" ")
		prompt, _ = reader.ReadString('\n')
		if prompt != "" {
			break
		}
	}
	return strings.TrimSpace(strings.ToLower(prompt))
}

// Print the current state of the game
func printGameGrid() {
	for i := range gameGrid {
		fmt.Print("   ")

		for j := range gameGrid[i] {
			currentChar := gameGrid[i][j]

			// Set the color of the character
			runeColor := color.White
			if currentChar.isCorrect {
				runeColor = color.Green
			} else if currentChar.isIncluded {
				runeColor = color.Yellow
			}

			fmt.Print(runeColor + " " + string(currentChar.value) + " " + color.Reset)
		}
		fmt.Println()

	}
	fmt.Println()
}

func printKeyboard() {
	fmt.Print("    ")

	// Print used keys a-m
	for _, v := range keySlice[0:13] {
		keyColor := color.White
		if keyboardMap[v] {
			keyColor = color.Gray
		}
		print(keyColor + string(v) + color.Reset)
	}

	fmt.Println()
	fmt.Print("    ")

	// Print used keys n-z
	for _, v := range keySlice[13:] {
		keyColor := color.White
		if keyboardMap[v] {
			keyColor = color.Gray
		}
		print(keyColor + string(v) + color.Reset)
	}

	fmt.Println()
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
		keyboardMap[key] = false
	}
}

func main() {
	fmt.Println("\n  Welcome to Wordle")
	winFlag := false

	// Loop until we're out of guesses
	var guessCount int8 = 0
	for guessCount < totalGuesses {
		// Print state of the game
		printGameGrid()
		printKeyboard()

		// Get user input
		currentGuess := guessPrompt("\n    Guess?>")

		// Display an error if the user doesn't input enough chars
		if len(currentGuess) != 5 {
			fmt.Println(color.Red + "\nYour guess must be 5 letters long" + color.Reset)
			continue
		}

		// Check to see if word is allowed
		if !words.IsValidWord(currentGuess) {
			fmt.Println(color.Red + "\n    Invalid word!" + color.Reset)
			continue
		}

		for index := range gameGrid[guessCount] {
			char := currentGuess[index]
			// Set the value to the corresponding rune char from the guess
			gameGrid[guessCount][index].value = rune(char)

			// Mark the character as used in the keyboard map
			keyboardMap[rune(char)] = true

			// Check if it's the same character or string contains rune at all
			if correctAnswer[index] == char {
				gameGrid[guessCount][index].isCorrect = true
			} else if strings.ContainsRune(correctAnswer, rune(char)) {
				gameGrid[guessCount][index].isIncluded = true
			}
		}

		// Increment the guess counter
		guessCount++

		// Stop looping if we found the answer
		if currentGuess == correctAnswer {
			winFlag = true
			break
		}
	}

	// Print final game state
	fmt.Println("      Game Over")
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
