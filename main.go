package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	if err := do(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
	os.Exit(0)
}

func do() error {
	// Step 1 get old resources
	kubectlExec := CreateCommandExecutor("kubectl")
	oldResources, err := kubectlExec.getOldResources()
	if err != nil {
		return err
	}

	// Step 2 delete them
	if err := kubectlExec.DeleteResources(oldResources); err != nil {
		return err
	}

	return nil
}

// CommandExecutor contains command execution information.
type CommandExecutor struct {
	// BinPath is the path the binary being used for the command (e.g. the path
	// to the kubectl binary if the kubectl command is to be used).
	binPath string
}

// CreateCommandExecutor returns a CommandExecutor for the given binary
func CreateCommandExecutor(binPath string) *CommandExecutor {
	ce := &CommandExecutor{
		binPath: binPath,
	}
	return ce
}

// execCommand runs the given command and returns the output.
func (ce CommandExecutor) execCommand(args []string) (string, error) {
	fmt.Printf("Running the following command: %s %s\n", ce.binPath, args)
	cmd := exec.Command(ce.binPath, args...)
	// By default set locations to standard error and output (visible in cloud build logs)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Write error output to two locations simultaneously. Unless disabled by a CommandOption, this will allow
	// us to see the error output as the execution is happening (by writing to standard error) and also allow us
	// to gather all stderr at the end (by also writing to var stderr).
	var stderr bytes.Buffer
	errWriter := io.MultiWriter(&stderr, cmd.Stderr)
	cmd.Stderr = errWriter

	// Write output to two locations simultaneously. Unless disabled by a CommandOption, this will allow
	// us to see the output as the execution is happening (by writing to standard out) and also allow us
	// to gather all stdout at the end (by also writing to var stdout).
	var stdout bytes.Buffer
	outWriter := io.MultiWriter(&stdout, cmd.Stdout)
	cmd.Stdout = outWriter

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command, err is %w", err)
	}

	// Wait for everything to finish.
	if err := cmd.Wait(); err != nil {
		// Read the stdErr output
		errorOutput := stderr.Bytes()
		fullErr := fmt.Errorf("error running command: %w\n%s", err, errorOutput)
		return "", fullErr
	}
	return stdout.String(), nil
}

// diffSlices returns the elements in slice1 that aren't in slice2.
func diffSlices(slice1, slice2 []string) []string {
	var diff []string
	set := make(map[string]bool)

	for _, val := range slice2 {
		set[val] = true
	}

	for _, val := range slice1 {
		// If they're not in slice 2
		if !set[val] {
			diff = append(diff, val)
		}
	}

	return diff
}
