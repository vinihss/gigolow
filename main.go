package main

import (
	"encoding/json"
	"fmt"
	"gigolow/configs"
	"gigolow/pkg/cli"
	"gigolow/pkg/logging"
	"gigolow/pkg/monitoring"
	"github.com/gen2brain/beeep"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// EnviaNotificacao envia uma notificação com título e mensagem
func EnviaNotificacao(titulo, mensagem string) error {
	switch runtime.GOOS {
	case "windows", "linux":
		// Usando a biblioteca beeep para notificações
		err := beeep.Notify(titulo, mensagem, "")
		if err != nil {
			return fmt.Errorf("erro ao enviar notificação: %w", err)
		}
	default:
		return fmt.Errorf("sistema operacional não suportado: %s", runtime.GOOS)
	}
	return nil
}

// go build gigolowGetDefaultPath retorna o caminho padrão com base no tipo e no sistema operacional
func GetDefaultPath(tipo string) string {
	var basePath string

	if runtime.GOOS == "windows" {
		// Obtem APPDATA ou define um caminho padrão
		basePath = os.Getenv("APPDATA")
		if basePath == "" {
			basePath = `C:\Users\Public\AppData`
		}
	} else {
		// Obtem HOME ou define um caminho padrão
		basePath = os.Getenv("HOME")
		if basePath == "" {
			basePath = "/tmp"
		}
	}

	// Adiciona o subdiretório com base no tipo
	switch tipo {
	case "config":
		return filepath.Join(basePath, "gigolow", "config")
	case "log":
		if runtime.GOOS == "windows" {
			return filepath.Join(`C:\Temp\gigolow\Logs`)
		}
		return "/var/log/gigolow"
	case "data":
		if runtime.GOOS == "windows" {
			return filepath.Join(`C:\ProgramData\gigolow`)
		}
		return "/var/lib/gigolow"
	default:
		return filepath.Join(basePath, "gigolow")
	}
}

// CreateDir garante que o diretório existe
func CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
func loadRepositories(filePath string) ([]configs.Repository, error) {
	var repos []configs.Repository
	file, err := os.Open(filePath)
	if err != nil {
		return repos, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return repos, err
	}

	var config map[string][]configs.Repository
	if err := json.Unmarshal(byteValue, &config); err != nil {
		return repos, err
	}

	return config["Repository"], nil
}

func loadEnvConfig() (string, int, bool, bool, string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return "", 0, false, false, "", err
	}

	defaultConfigPath := GetDefaultPath("config")

	// Cria o diretório, se necessário
	err = CreateDir(defaultConfigPath)
	if err != nil {
		return "", 0, false, false, "", err
	}
	configPath := defaultConfigPath
	workDir := os.Getenv("WORKDIR")
	threads, err := strconv.Atoi(os.Getenv("THREADS"))
	if err != nil {
		return "", 0, false, false, "", err
	}
	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		return "", 0, false, false, "", err
	}
	verbose, err := strconv.ParseBool(os.Getenv("VERBOSE"))
	if err != nil {
		return "", 0, false, false, "", err
	}

	return workDir, threads, debug, verbose, configPath, nil
}

func main() {
	workDir, threads, debug, verbose, configPath, err := loadEnvConfig()
	if err != nil {
		log.Fatalf("Error loading environment config: %v", err)
	}
	logger := logging.NewLogger("application.log")
	defer logger.Close()

	configFilePath := filepath.Join(configPath, "config.json")
	repos, err := loadRepositories(configFilePath)
	if err != nil {
		log.Fatalf("Error loading repositories: %v", err)
	}

	config := configs.Config{
		Repositories: repos,
		Threads:      threads,
		LogFile:      "application.log",
		WorkDir:      workDir,
		Debug:        debug,
		Verbose:      verbose,
	}
	err = EnviaNotificacao("gigolow alert", "Starting job processing")
	if err != nil {
		fmt.Printf("Erro ao enviar notificação: %v\n", err)
		return
	}
	logger.Log("Starting job processing")
	results := cli.RunJobs(config, logger)

	logger.Log("Job processing completed")
	for _, result := range results {
		fmt.Printf("Repository: %s, Stage: %s, Success: %t, Message: %s\n",
			result.Repository, result.Stage, result.Success, result.Message)
	}
	monitoring.MonitorStatus()
}
