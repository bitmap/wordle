// Wordle - a word game - for the command line
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bitmap/wordle-cli/internal/allowlist"
	answers "github.com/bitmap/wordle-cli/internal/answers"
	color "github.com/bitmap/wordle-cli/internal/color"
)

const wordLength = 5
const totalGuesses int8 = 6
const emptySpaceRune = '*'

var correctAnswer = answers.RandomAnswer()

type Guess struct {
	value      rune
	isCorrect  bool
	isIncluded bool
}

// Game is a x * y grid
type GameGrid [totalGuesses][wordLength]Guess

// Prompt the user to enter a string
func guessPrompt(label string) string {
	var prompt string
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		prompt, _ = reader.ReadString('\n')
		if prompt != "" {
			break
		}
	}
	return strings.TrimSpace(prompt)
}

// Print the current state of the game
func printGameGrid(grid GameGrid) {
	for i := range grid {
		fmt.Print("   ")

		for j := range grid[i] {
			currentChar := grid[i][j]

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

func main() {
	fmt.Println("\n  Welcome to Wordle")

	winFlag := false

	grid := GameGrid{[wordLength]Guess{}}

	// Initialize the grid
	for i := range grid {
		for j := range grid[i] {
			grid[i][j].value = emptySpaceRune
		}
	}

	// Loop until we're out of guesses
	var guessCount int8 = 0
	for guessCount < totalGuesses {
		printGameGrid(grid)
		currentGuess := guessPrompt("  Your guess?")

		// Display an error if the user doesn't input enough chars
		if len(currentGuess) != 5 {
			fmt.Println(color.Red + "\nYour guess must be 5 letters long" + color.Reset)
			continue
		}

		// Check to see if word is allowed
		if !allowlist.IsValidWord(currentGuess) {
			fmt.Println(color.Red + "\nInvalid word" + color.Reset)
			continue
		}

		for index := range grid[guessCount] {
			// Set the value to the corresponding rune char from the guess
			grid[guessCount][index].value = rune(currentGuess[index])

			// Check if it's the same character or string contains rune at all
			if correctAnswer[index] == currentGuess[index] {
				grid[guessCount][index].isCorrect = true
			} else if strings.ContainsRune(correctAnswer, rune(currentGuess[index])) {
				grid[guessCount][index].isIncluded = true
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

	fmt.Println("      Game Over")
	printGameGrid(grid)

	if winFlag {
		if guessCount == 1 {
			fmt.Println("🫨 Woah! You got it right on the first try! Nice!")
		} else {
			fmt.Println("🎉 Correct! You won in " + fmt.Sprint(guessCount) + " guesses.")
		}
	} else {
		fmt.Println("😓 Sorry, the answer was " + color.Cyan + correctAnswer + color.Reset + ". Try again.")
	}
}
