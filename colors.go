package main

import (
	"fmt"
)

func colored(text, color string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", color, text)
}

func Yellow(text string) string {
	return colored(text, "33")
}

func YellowBold(text string) string {
	return colored(text, "1;33")
}

func Green(text string) string {
	return colored(text, "32")
}

func GreenBold(text string) string {
	return colored(text, "1;32")
}

func Blue(text string) string {
	return colored(text, "34")
}

func BlueBold(text string) string {
	return colored(text, "1;34")
}

func Red(text string) string {
	return colored(text, "31")
}

func RedBold(text string) string {
	return colored(text, "1;31")
}

func Cyan(text string) string {
	return colored(text, "36")
}

func CyanBold(text string) string {
	return colored(text, "1;36")
}
