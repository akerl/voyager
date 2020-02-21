package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmText asks the user to confirm an action by typing a verification message
func ConfirmText(confirm string, prompt ...string) error {
	fmt.Println(prompt...)
	fmt.Printf("If you want to proceed, type '%s'", confirm)
	confirmReader := bufio.NewReader(os.Stdin)
	confirmInput, err := confirmReader.ReadString('\n')
	if err != nil {
		return err
	}
	cleanedInput := strings.TrimSpace(confirmInput)
	if cleanedInput != confirm {
		return fmt.Errorf("aborting")
	}
	return nil
}
