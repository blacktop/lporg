package background

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

// SetDesktopImage sets the desktop background image to the supplied image
func SetDesktopImage(image string) (string, error) {
	return Tell("Finder", "set desktop picture to POSIX file "+wrapInQuotes(image))
}

// Tell tells an macOS application what to do
func Tell(application string, commands ...string) (string, error) {
	return run(buildTell(application, commands...))
}

// Parse the Tell options and build the command
func buildTell(application string, commands ...string) string {
	application = wrapInQuotes(application)
	args := []string{"tell application", application, "\n"}
	for _, command := range commands {
		args = append(args, command, "\n")
	}
	args = append(args, "end", "tell")
	return build(args...)
}

// Build the AppleScript command from a set of optional parameters, return the output
func run(command string) (string, error) {
	cmd := exec.Command("osascript", "-e", command)
	output, err := cmd.CombinedOutput()
	prettyOutput := strings.Replace(string(output), "\n", "", -1)

	// Ignore errors from the user hitting the cancel button
	if err != nil && strings.Index(string(output), "User canceled.") < 0 {
		return "", errors.New(err.Error() + ": " + prettyOutput + " (" + command + ")")
	}

	return prettyOutput, nil
}

// Wrap text in quotes for proper command line formatting
func wrapInQuotes(text string) string {
	return "\"" + text + "\""
}

// Build the AppleScript command, ignoring any blank optional parameters
func build(params ...string) string {
	var validParams []string

	for _, param := range params {
		if param != "" {
			validParams = append(validParams, param)
		}
	}

	return strings.Join(validParams, " ")
}

// Parse a button response
func parseResponse(output string, buttons []string) Response {
	var clicked, text string
	var gaveUp bool

	// Find out if the notification gave up
	gaveUpRe := regexp.MustCompile("gave up:(true|false)")
	gaveUpMatches := gaveUpRe.FindStringSubmatch(output)
	if len(gaveUpMatches) > 1 && gaveUpMatches[1] == "true" {
		gaveUp = true
	}

	if !gaveUp {
		for _, button := range buttons {
			// Find which button was clicked
			button = strings.Trim(button, `"`)
			buttonStr := "button returned:" + button
			clickedRe := regexp.MustCompile(buttonStr + ",")
			if clickedRe.MatchString(output) || output == buttonStr {
				clicked = button
				break
			}
		}

		// Don't mess around with regex, just get the text returned
		if strings.Index(output, ", text returned:") > 0 {
			output = strings.Replace(output, "button returned:"+clicked+", ", "", 1)
			output = strings.Replace(output, ", gave up:false", "", 1)
			output = strings.Replace(output, "text returned:", "", 1)
			text = output
		}
	}

	// Find out if the user entered text

	response := Response{
		Clicked: clicked,
		GaveUp:  gaveUp,
		Text:    text,
	}

	return response
}

// Response the response format after a button click on an alert or dialog box
type Response struct {
	Clicked string // The name of the button clicked
	GaveUp  bool   // True if the user failed to respond in the duration specified
	Text    string // Only on Dialog boxes - The return value of the input field
}
