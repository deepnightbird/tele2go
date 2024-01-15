package main

import (
	"bufio"
	"fmt"
)

const Esc string = "\x1b"
const Comma string = "" // ;

func escape(format string, args ...interface{}) string {
    return fmt.Sprintf("%s%s", Esc, fmt.Sprintf(format, args...))
}

// FORE
var (
    FORE_BLACK      string  = Esc + "[30" + Comma + "m"
    FORE_RED        string  = Esc + "[31" + Comma + "m"
    FORE_GREEN      string  = Esc + "[32" + Comma + "m"
    FORE_YELLOW     string  = Esc + "[33" + Comma + "m"
    FORE_BLUE       string  = Esc + "[34" + Comma + "m"
    FORE_MAGENTA    string  = Esc + "[35" + Comma + "m"
    FORE_CYAN       string  = Esc + "[36" + Comma + "m"
    FORE_WHITE      string  = Esc + "[37" + Comma + "m"
    FORE_RESET      string  = Esc + "[39" + Comma + "m"

    // These are fairly well supported, but not part of the standard.
    FORE_LIGHTBLACK_EX   string  = Esc + "[90" + Comma + "m"
    FORE_LIGHTRED_EX     string  = Esc + "[91" + Comma + "m"
    FORE_LIGHTGREEN_EX   string  = Esc + "[92" + Comma + "m"
    FORE_LIGHTYELLOW_EX  string  = Esc + "[93" + Comma + "m"
    FORE_LIGHTBLUE_EX    string  = Esc + "[94" + Comma + "m"
    FORE_LIGHTMAGENTA_EX string  = Esc + "[95" + Comma + "m"
    FORE_LIGHTCYAN_EX    string  = Esc + "[96" + Comma + "m"
    FORE_LIGHTWHITE_EX   string  = Esc + "[97" + Comma + "m"
)

// BACK
const (
    BACK_BLACK           int = 40
    BACK_RED             int = 41
    BACK_GREEN           int = 42
    BACK_YELLOW          int = 43
    BACK_BLUE            int = 44
    BACK_MAGENTA         int = 45
    BACK_CYAN            int = 46
    BACK_WHITE           int = 47
    BACK_RESET           int = 49

    // These are fairly well supported, but not part of the standard.
    BACK_LIGHTBLACK_EX   int = 100
    BACK_LIGHTRED_EX     int = 101
    BACK_LIGHTGREEN_EX   int = 102
    BACK_LIGHTYELLOW_EX  int = 103
    BACK_LIGHTBLUE_EX    int = 104
    BACK_LIGHTMAGENTA_EX int = 105
    BACK_LIGHTCYAN_EX    int = 106
    BACK_LIGHTWHITE_EX   int = 107
)

// Style
const (
    STYLE_BRIGHT    int = 1
    STYLE_DIM       int = 2
    STYLE_NORMAL    int = 22
    STYLE_RESET_ALL int = 0
)

var stdOut *bufio.Writer

func MoveTo(line, col int) {
    var text string = escape("[%d;%dH", line, col)
    fmt.Fprint(stdOut, text)
}

func ShowCursor() {
    var text string = escape("[?25h")
    fmt.Fprint(stdOut, text)
}

func HideCursor() {
    var text string = escape("[?25l")
    fmt.Fprint(stdOut, text)
}

func ClearLine() {
    var text string = "\033[2K"
    fmt.Fprint(stdOut, text)
}

/*func PrintAnsi(text string, fc int, bc int) {
    //var newtext string = escape("[%d;%dm%s", fc, bc, text)
    //var newtext string = escape("[%d;%dm%s", fc, BACK_RESET, text)
    var newtext string = escape("[%d;m%s", fc, text)
    fmt.Fprint(stdOut, newtext)
}*/

func PrintAnsi(text string) {
    fmt.Fprint(stdOut, text)
}
