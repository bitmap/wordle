package prompt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/bitmap/wordle-cli/internal/color"
	"golang.org/x/term"
)

// Control bytes we handle directly.
const (
	ctrlC     = 0x03
	ctrlD     = 0x04
	keyBksp   = 0x08
	keyEnter  = '\r'
	keyNwline = '\n'
	keyEscape = 0x1b
	keyDelete = 0x7f
)

// One byte read from the terminal.
type read struct {
	char byte
	err  error
}

// Bytes from the terminal, read by a single goroutine. Starting one per
// prompt would leave the old one holding bytes meant for the next.
var keys = sync.OnceValue(func() <-chan read {
	out := make(chan read)

	go func() {
		for {
			char, err := reader.ReadByte()
			out <- read{char, err}
			if err != nil {
				close(out)
				return
			}
		}
	}()

	return out
})

// errInterrupted is returned when the user presses Ctrl-C.
var errInterrupted = errors.New("interrupted")

// Returned by readLine when the terminal can't be put into raw mode.
var errNoRawMode = errors.New("raw mode unavailable")

// readLine reads a single line in raw mode, allowing the cursor to move
// within the text so the user can insert and delete anywhere. Input is
// limited to letters, and to maxLen of them.
func readLine(label string, maxLen int, errMsg string, repaint func(string)) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errNoRawMode
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", errNoRawMode
	}
	defer term.Restore(fd, oldState)

	var (
		buf []rune
		pos int
	)

	// The prompt starts on its own line.
	fmt.Fprint(os.Stderr, "\r\n")

	// Redraw on resize so the prompt stays centered with the rest of
	// the screen.
	resized := make(chan os.Signal, 1)
	signal.Notify(resized, syscall.SIGWINCH)
	defer signal.Stop(resized)

	// Reading blocks, so it happens alongside the resize signal.
	keys := keys()

	draw(label, buf, pos, errMsg)

	for {
		var next read

		select {
		case <-resized:
			// The caller prints with plain newlines, which need the
			// terminal back in its normal mode to land in column 0.
			redraw(fd, oldState, repaint, string(buf))
			draw(label, buf, pos, errMsg)
			continue
		case next = <-keys:
		}

		if next.err != nil {
			return "", next.err
		}
		char := next.char

		switch char {
		case keyEnter, keyNwline:
			// Leave the cursor on a fresh line for whatever prints next.
			fmt.Fprint(os.Stderr, "\r\n")
			return string(buf), nil

		case ctrlC:
			// Raw mode suppresses SIGINT, so quitting is up to us.
			fmt.Fprint(os.Stderr, "\r\n")
			return "", errInterrupted

		case ctrlD:
			return "", io.EOF

		case keyDelete, keyBksp:
			// Delete the rune behind the cursor.
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
			}

		case keyEscape:
			switch readEscape(keys) {
			case keyLeft:
				if pos > 0 {
					pos--
				}
			case keyRight:
				if pos < len(buf) {
					pos++
				}
			case keyDel:
				// Delete the rune under the cursor.
				if pos < len(buf) {
					buf = append(buf[:pos], buf[pos+1:]...)
				}
			}

		default:
			// Only letters make a guess, and only up to maxLen of them.
			value := rune(char)
			if !unicode.IsLetter(value) || len(buf) >= maxLen {
				continue
			}

			// Insert at the cursor rather than appending.
			value = unicode.ToLower(value)
			buf = append(buf, 0)
			copy(buf[pos+1:], buf[pos:])
			buf[pos] = value
			pos++
		}

		redraw(fd, oldState, repaint, string(buf))
		draw(label, buf, pos, errMsg)
	}
}

// Repaint the screen above the prompt. The caller prints with plain
// newlines, which need the terminal back in its normal mode to land in
// column 0, so raw mode is lifted for the duration.
func redraw(fd int, oldState *term.State, repaint func(string), guess string) {
	if repaint == nil {
		return
	}

	term.Restore(fd, oldState)
	repaint(guess)
	term.MakeRaw(fd)

	fmt.Fprint(os.Stderr, "\r\n")
}

// Repaint the input line and leave the cursor at pos. label must not
// contain a newline, or the line will scroll on every keystroke.
func draw(label string, buf []rune, pos int, errMsg string) {
	// Recomputed every time so a resize recenters the prompt.
	indent := strings.Repeat(" ", promptIndent(label))
	label = indent + label

	// Guesses are stored lowercase to match the word list, but shown
	// uppercase to match the grid.
	fmt.Fprint(os.Stderr, "\r\033[K"+label+strings.ToUpper(string(buf)))

	// Any error sits on the line below, centered on its own width rather
	// than the prompt's, so drop down to rewrite it and come back up
	// before placing the cursor.
	if errMsg != "" {
		errIndent := strings.Repeat(" ", centerIndent(len(errMsg)))
		fmt.Fprint(os.Stderr, "\r\n\033[K"+errIndent+color.Red+errMsg+color.Reset+"\033[A")
	}

	// Walk the cursor back to column 0, then out to its real position.
	fmt.Fprint(os.Stderr, "\r")
	if offset := len([]rune(label)) + pos; offset > 0 {
		fmt.Fprintf(os.Stderr, "\033[%dC", offset)
	}
}

// The escape sequences we act on.
type escapeKey int

const (
	keyUnknown escapeKey = iota
	keyLeft
	keyRight
	keyDel
)

// Wait briefly for the next byte of an escape sequence. Terminals send
// the whole sequence at once, so nothing arriving means the user pressed
// ESC on its own.
func nextKey(keys <-chan read) (byte, bool) {
	select {
	case next := <-keys:
		return next.char, next.err == nil
	case <-time.After(20 * time.Millisecond):
		return 0, false
	}
}

// Read a single keypress from accept, without waiting for Enter. Anything
// else is ignored, so a stray key can't answer the prompt. Returns
// errNoRawMode when stdin isn't a terminal, so callers can fall back to
// reading a line.
func readKey(accept string) (byte, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return 0, errNoRawMode
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, errNoRawMode
	}
	defer term.Restore(fd, oldState)

	for next := range keys() {
		if next.err != nil {
			return 0, next.err
		}

		// Ctrl-C has to be handled here, since raw mode suppresses SIGINT.
		if next.char == ctrlC || next.char == ctrlD {
			return 0, io.EOF
		}

		char := byte(unicode.ToLower(rune(next.char)))
		if !strings.ContainsRune(accept, rune(char)) {
			continue
		}

		// Echo it so the answer is visible, then move off the prompt line.
		if char == keyEnter || char == keyNwline {
			fmt.Fprint(os.Stderr, "\r\n")
		} else {
			fmt.Fprintf(os.Stderr, "%c\r\n", next.char)
		}

		return char, nil
	}

	return 0, io.EOF
}

// Read a whole line. The terminal is in its normal mode here, so this is
// only reached between raw-mode prompts — but the reader goroutine owns
// stdin either way, so the bytes have to come from it.
func readString() (string, error) {
	var line strings.Builder

	for next := range keys() {
		if next.err != nil {
			return line.String(), next.err
		}

		if next.char == '\n' || next.char == '\r' {
			return line.String(), nil
		}

		line.WriteByte(next.char)
	}

	return line.String(), io.EOF
}

// Read the rest of an escape sequence, having already consumed the ESC.
// Unrecognized sequences are consumed whole so their bytes can't leak into
// the guess. A lone ESC is reported as soon as nothing follows it.
func readEscape(keys <-chan read) escapeKey {
	// Arrows arrive as either CSI (ESC [ D) or, when the terminal is in
	// application cursor mode, SS3 (ESC O D). Both end in the same byte.
	intro, ok := nextKey(keys)
	if !ok || (intro != '[' && intro != 'O') {
		return keyUnknown
	}

	// CSI carries parameter bytes before a final byte in the range @ to ~.
	// SS3 has none, so its next byte is already the final one.
	var params []byte
	var final byte
	for {
		char, ok := nextKey(keys)
		if !ok {
			return keyUnknown
		}
		if intro == 'O' || (char >= 0x40 && char <= 0x7e) {
			final = char
			break
		}
		params = append(params, char)
	}

	switch final {
	case 'D':
		return keyLeft
	case 'C':
		return keyRight
	case '~':
		if string(params) == "3" {
			return keyDel
		}
	}

	return keyUnknown
}
