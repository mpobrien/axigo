package main

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type TokenType int

// eof is a placeholder rune indicating the end of the input stream
const eof = rune(0)

// EOF is a token emitted by the lexer when it's done producing tokens.
const EOF TokenType = 0

const (
	Whitespace TokenType = iota + 1
)

// Token describes a lexeme scanned from the input, defined by its type and source text
type Token struct {
	TokenType TokenType
	Text      string
}

type Command interface{}

type runeFilter func(rune) bool

func isRune(in rune) runeFilter {
	return func(ch rune) bool { return in == ch }
}

func anyOf(in string) runeFilter {
	return func(ch rune) bool {
		return strings.ContainsRune(in, ch)
	}
}

func isBase64(ch rune) bool {
	return ch == '+' || ch == '/' || ch == '=' || isDecimalDigit(ch) || isAlpha(ch)
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDecimalDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isAlpha(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isLetter(ch rune) bool {
	return isAlpha(ch) || ch == '_' || ch == '$'
}

func isIDChar(ch rune) bool {
	return isLetter(ch) || isDecimalDigit(ch) || ch == '-'
}

// Scanner represents a lexical scanner.
type Scanner struct {
	input      string
	pos        int
	start      int
	widthStack []int
}

func NewScanner(input string) *Scanner {
	return &Scanner{input: input}
}

func (s *Scanner) read() rune {
	if s.pos >= len(s.input) {
		s.widthStack = append(s.widthStack, 0)
		return eof
	}
	nextRune, width := utf8.DecodeRuneInString(s.input[s.pos:])
	s.pos += width
	s.widthStack = append(s.widthStack, width)
	return nextRune
}

// unread backs up the scanner's position in the input stream by one character.
// May only be called following a complementary call to read(), otherwise this will panic.
func (s *Scanner) unread() {
	if len(s.widthStack) == 0 {
		panic("unreading character from empty stack")
	}

	lastPosition := len(s.widthStack) - 1
	lastWidth := s.widthStack[lastPosition]
	s.pos -= lastWidth
	s.widthStack = s.widthStack[:len(s.widthStack)-1]
}

// peek returns the next character in the input stream without consuming it.
func (s *Scanner) peek() rune {
	next := s.read()
	s.unread()
	return next
}

// reset moves the scanner's position in the input back to the start of the current token.
func (s *Scanner) reset() {
	for len(s.widthStack) > 0 {
		s.unread()
	}
}

// itemText returns the text of the current position spanned by the lexer.
func (s *Scanner) itemText() string {
	return s.input[s.start:s.pos]
}

// accept consumes one character from the input stream if it matches the given filter.
// If the filter does not match, it returns false without consuming any character.
func (s *Scanner) accept(filter runeFilter) bool {
	if filter(s.read()) {
		return true
	}
	s.unread()
	return false
}

// acceptRunToLength consumes up to length characters from the input string
// while they match the given filter.
// Returns true if the number of characters consumed is exactly equal to expectedLength.
func (s *Scanner) acceptRunToLength(filter runeFilter, expectedLength int) bool {
	numChars := 0
	for i := 0; i < expectedLength; i++ {
		if s.accept(filter) {
			numChars++
		} else {
			break
		}
	}
	return numChars == expectedLength
}

// acceptRun consumes a run of runes from the valid set. Returns the number of characters consumed.
func (s *Scanner) acceptRun(filter runeFilter) int {
	count := 0
	for filter(s.read()) {
		count++
	}
	s.unread()
	return count
}

// newToken creates a token of the given type whose contents is the text
// from the input betweeen the token start offset up to the current position
func (s *Scanner) newToken(tokType TokenType) Token {
	return Token{tokType, s.itemText()}
}

// Scan consumes and returns the next token from the input stream
func (s *Scanner) Scan() (Token, error) {
	defer func() {
		s.start = s.pos
		s.widthStack = []int{}
	}()

	nextRune := s.read()

	// Consume whitespace until the next character in the stream
	// is anything other than whitespace.
	if isWhitespace(nextRune) {
		s.acceptRun(isWhitespace)
		return s.newToken(Whitespace), nil
	}

	// Otherwise, branch based on what token start sequence is matching.
	switch nextRune {
	case eof:
		return Token{0, ""}, nil
	default:
		panic("TODO")
	}

	return Token{}, errors.New("unrecognized token")
}
