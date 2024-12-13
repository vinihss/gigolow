// pkg/cli/command.go

package cli

import (
	"gigolow/configs"
	"gigolow/pkg/logging"
	"gigolow/pkg/repository"
	
	"runtime/debug"
	"sync"
)

type JobStatus struct {
	Repository string
	Stage      string
	Success    bool
	Message    string
}

func RunJobs(config configs.Config, logger *logging.Logger) []JobStatus {
	var wg sync.WaitGroup
	jobs := make(chan configs.Repository, len(config.Repositories))
	statuses := make(chan JobStatus, len(config.Repositories))

	// Start worker goroutines
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				if config.Verbose {
					logger.Log("Starting to process repository: " + repo.Url)
				}
				statuses <- processRepository(repo, logger)
				if config.Verbose {
					logger.Log("Finished processing repository: " + repo.Url)
				}
			}
		}()
	}

	// Dispatch jobs
	go func() {
		for _, repo := range config.Repositories {
			jobs <- repo
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(statuses)
	}()

	// Collect results
	var results []JobStatus
	for status := range statuses {
		results = append(results, status)
	}
	return results
}

func processRepository(repo configs.Repository, logger *logging.Logger) JobStatus {
	logger.Log("Processing repository: " + repo.Url)

	if err := repository.Clone(repo, logger); err != nil {
		logger.Log("Error processing repository: " + repo.Url + ", branch: " + repo.Branch + ". Error: " + err.Error())
		logger.Log(string(debug.Stack()))
		return JobStatus{Repository: repo.Url, Stage: "Cloning", Success: false, Message: err.Error()}
	}
	out, err := repository.Status(repo, logger)
	if err != nil {
		logger.Log("Error processing repository: " + repo.Url + ", branch: " + repo.Branch + ". Error: " + err.Error())
		logger.Log(string(debug.Stack()))
		return JobStatus{Repository: repo.Url, Stage: "Cloning", Success: false, Message: err.Error()}
	}

	logger.Log(out)

	logger.Log("Successfully processed repository: " + repo.Url)
	return JobStatus{Repository: repo.Url, Stage: "Completed", Success: true}
}

// pkg/repository/repo.go
