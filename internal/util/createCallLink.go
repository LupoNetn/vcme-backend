package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// convert title into a clean url format
func Slugify(input string) string {
	s := strings.ToLower(input)

	//replace spaces with dash
	s = strings.ReplaceAll(s, " ", "-")

	//remove non url-safe chars
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	s = reg.ReplaceAllString(s, "")

	return s
}

//generate a random unique string

func GnerateRandomID(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

//combine both functions to create a well formated call link

func GenerateCallLink(appName string, callTitle string) string {
	app := Slugify(appName)
	title := Slugify(callTitle)

	randomID := GnerateRandomID(3)

	link := fmt.Sprintf("%s.com/v/%s-%s", app, title, randomID)

	return link
}
