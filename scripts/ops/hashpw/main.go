// Command hashpw prints the argon2id hash of the password in the PASSWORD environment variable, in the format the
// app stores in users.password. It reads the environment, not argv, so the password stays out of the shell history
// and the process list.
package main

import (
	"fmt"
	"os"

	"github.com/AlexGalhardo/vukem/auth/hash"
)

func main() {
	password := os.Getenv("PASSWORD")
	if password == "" {
		fmt.Fprintln(os.Stderr, "set PASSWORD in the environment")
		os.Exit(2)
	}

	out, err := hash.Default().Make(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hash:", err)
		os.Exit(1)
	}

	fmt.Println(out)
}
