package interaction

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetUserApproval(commands []string) (bool, error) {
	fmt.Println("\nThe following raw defaults commands will be executed:")
	fmt.Println("=" + strings.Repeat("=", 50))

	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}

	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Print("Do you want to proceed? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}
