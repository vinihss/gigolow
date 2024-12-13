package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Program struct {
	Name        string
	Path        string
	Args        []string
	RequireSudo bool
}

type ServiceManager struct {
	queue []Program
	mutex sync.Mutex
}

func (sm *ServiceManager) AddProgram() {
	fmt.Println("Select a program to add:")
	fmt.Println("1. SystemUpdate")
	fmt.Println("2. ArchInstall")
	fmt.Println("3. ArchRemove")
	fmt.Println("4. Custom")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter choice: ")
	scanner.Scan()
	choice := scanner.Text()

	var program Program
	switch choice {
	case "1":
		program = Program{
			Name:        "SystemUpdate",
			Path:        "/usr/bin/pacman",
			Args:        []string{"-Syu"},
			RequireSudo: true,
		}
	case "2":
		fmt.Print("Enter package to install: ")
		scanner.Scan()
		pkg := scanner.Text()
		program = Program{
			Name:        "ArchInstall",
			Path:        "/usr/bin/pacman",
			Args:        []string{"-S", "--noconfirm", pkg},
			RequireSudo: true,
		}
	case "3":
		fmt.Print("Enter package to remove: ")
		scanner.Scan()
		pkg := scanner.Text()
		program = Program{
			Name:        "ArchRemove",
			Path:        "/usr/bin/pacman",
			Args:        []string{"-R", pkg},
			RequireSudo: true,
		}
	case "4":
		fmt.Print("Enter custom program name: ")
		scanner.Scan()
		name := scanner.Text()

		fmt.Print("Enter custom program path: ")
		scanner.Scan()
		path := scanner.Text()

		fmt.Print("Enter custom program arguments (space-separated): ")
		scanner.Scan()
		argsInput := scanner.Text()
		args := []string{}
		if argsInput != "" {
			args = splitArgs(argsInput)
		}

		fmt.Print("Does this program require sudo? (yes/no): ")
		scanner.Scan()
		sudoInput := scanner.Text()
		requireSudo := strings.ToLower(sudoInput) == "yes"

		program = Program{Name: name, Path: path, Args: args, RequireSudo: requireSudo}
	default:
		fmt.Println("Invalid choice.")
		return
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.queue = append(sm.queue, program)
	fmt.Printf("Program '%s' added to queue.\n", program.Name)
}

func (sm *ServiceManager) ExecuteQueue() {
	for {
		sm.mutex.Lock()
		if len(sm.queue) == 0 {
			sm.mutex.Unlock()
			break
		}

		program := sm.queue[0]
		sm.queue = sm.queue[1:]
		sm.mutex.Unlock()

		fmt.Printf("Executing program: %s\n", program.Name)
		var cmd *exec.Cmd
		if program.RequireSudo {
			cmd = exec.Command("sudo", append([]string{program.Path}, program.Args...)...)
		} else {
			cmd = exec.Command(program.Path, program.Args...)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error executing %s: %v\n", program.Name, err)
			if isResourceLockedError(err) {
				fmt.Printf("Resource locked for '%s', retrying later...\n", program.Name)
				sm.mutex.Lock()
				sm.queue = append(sm.queue, program)
				sm.mutex.Unlock()
				time.Sleep(5 * time.Second) // Wait before retrying
			} else {
				fmt.Printf("Program '%s' failed permanently.\n", program.Name)
			}
		} else {
			fmt.Printf("Program '%s' finished successfully.\n", program.Name)
		}
	}
}

func (sm *ServiceManager) ShowQueue() {
	queue := sm.queue
	if len(queue) == 0 {
		fmt.Println("Queue is empty.")
		return
	}
	for i, program := range queue {
		fmt.Printf("%d. %s: %s %s\n", i+1, program.Name, program.Path, strings.Join(program.Args, " "))
	}
}

func isResourceLockedError(err error) bool {
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.ExitCode()
		if exitCode == 1 {
			return true // Example: Pacman lock errors often return exit code 1
		}
	}

	lockPatterns := []string{
		"unable to lock database",
		"failed to lock",
		"resource temporarily unavailable",
	}

	for _, pattern := range lockPatterns {
		if strings.Contains(err.Error(), pattern) {
			return true
		}
	}

	lockFilePatterns := []string{
		`/var/lib/pacman/db\.lck`,
		`/tmp/lockfile`,
		`/var/lock/.*`,
	}

	for _, regexPattern := range lockFilePatterns {
		matched, _ := regexp.MatchString(regexPattern, err.Error())
		if matched {
			return true
		}
	}

	return false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	serviceManager := &ServiceManager{}

	fmt.Println("Service Manager started. Type 'add' to add programs or 'run' to execute the queue.")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		command := scanner.Text()

		switch command {
		case "add":
			serviceManager.AddProgram()
		case "run":
			fmt.Println("Executing queue...")
			serviceManager.ExecuteQueue()
		case "list":
			fmt.Println("Executing queue...")
			serviceManager.ShowQueue()
		case "exit":
			fmt.Println("Exiting Service Manager.")
			return
		default:
			fmt.Println("Unknown command. Available commands: add, run, exit.")
		}
	}
}

func splitArgs(input string) []string {
	return strings.Fields(input)
}
