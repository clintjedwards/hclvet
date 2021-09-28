package utils

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

// CreateDir creates a directory path if it does not exist and returns nil if the path already exists.
// Will return the underlying os.Stat error if there were any other errors
func CreateDir(dirPath string) error {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	} else if os.IsExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

// ExecuteCmd wraps context around a given command and executes it.
func ExecuteCmd(path string, args []string, env []string, dir string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = env
	cmd.Dir = dir

	// Execute command
	return cmd.CombinedOutput()
}

// ReadLine grabs a specific line from the provided corpus and returns it as a string
func ReadLine(r io.Reader, lineNum int) (line string, lastLine int, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lastLine++
		if lastLine == lineNum {
			// you can return sc.Bytes() if you need output in []bytes
			return sc.Text(), lastLine, sc.Err()
		}
	}
	return line, lastLine, io.EOF
}
