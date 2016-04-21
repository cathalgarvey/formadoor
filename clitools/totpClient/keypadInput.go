package main

import (
	"bufio"
	"bytes"
	"os"
)

var (
	// Used to detect keys
	keypadEscape = []byte{27, 91}
	// Represent numbers on keypad when numlock disabled
	arrowDown      = []byte{27, 91, 66}      // 2
	arrowUp        = []byte{27, 91, 65}      // 8
	arrowLeft      = []byte{27, 91, 68}      // 4
	arrowRight     = []byte{27, 91, 67}      // 6
	keypadFive     = []byte{27, 91, 69}      // 5 (on my laptop?)
	keypadFiveAlt  = []byte{27, 91, 71}      // 5 (on pi?)
	keypadPageDown = []byte{27, 91, 54, 126} // 3
	keypadPageUp   = []byte{27, 91, 53, 126} // 9
	keypadEnd      = []byte{27, 91, 70}      // 1 (on my laptop?)
	keypadEndAlt   = []byte{27, 91, 52, 126} // 1 (on pi?)
	keypadHome     = []byte{27, 91, 72}      // 7 (on my laptop?)
	keypadHomeAlt  = []byte{27, 91, 49, 126} // 7 (on pi?)
	keypadInsert   = []byte{27, 91, 50, 126} // 0
)

func getKeypadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	codeAttempt, err := reader.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	// If numlock is disabled then USB keypads will return escaped garbage.
	// So, generously assume that either the garbage or digits are equivalent.
	if bytes.Contains(codeAttempt, keypadEscape) {
		codeAttempt = bytes.Replace(codeAttempt, keypadEnd, []byte("1"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadEndAlt, []byte("1"), -1)
		codeAttempt = bytes.Replace(codeAttempt, arrowDown, []byte("2"), -1)
		codeAttempt = bytes.Replace(codeAttempt, arrowLeft, []byte("4"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadPageDown, []byte("3"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadFive, []byte("5"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadFiveAlt, []byte("5"), -1)
		codeAttempt = bytes.Replace(codeAttempt, arrowRight, []byte("6"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadHome, []byte("7"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadHomeAlt, []byte("7"), -1)
		codeAttempt = bytes.Replace(codeAttempt, arrowUp, []byte("8"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadPageUp, []byte("9"), -1)
		codeAttempt = bytes.Replace(codeAttempt, keypadInsert, []byte("0"), -1)
	}
	codeAttempt = bytes.Trim(codeAttempt, "\n")
	return string(codeAttempt), nil
}
