package main // must be main for plugin entry point

import (
	"context"
	"fmt"
	"github.com/merico-dev/lake/errors"
	"github.com/merico-dev/lake/plugins/helper"
	"os"
	"strings"
	"time"

	"github.com/merico-dev/lake/config"
	lakeModels "github.com/merico-dev/lake/models"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/github/api"
	"github.com/merico-dev/lake/plugins/github/models"
	"github.com/merico-dev/lake/plugins/github/tasks"
	"github.com/merico-dev/lake/utils"
	"github.com/mitchellh/mapstructure"
)

var _ core.Plugin = (*Github)(nil)

type Github string

func (plugin Github) Init() {
	err := lakeModels.Db.AutoMigrate(
		&models.GithubRepo{},
		&models.GithubCommit{},
		&models.GithubRepoCommit{},
		&models.GithubPullRequest{},
		&models.GithubReviewer{},
		&models.GithubPullRequestComment{},
		&models.GithubPullRequestCommit{},
		&models.GithubPullRequestLabel{},
		&models.GithubIssue{},
		&models.GithubIssueComment{},
		&models.GithubIssueEvent{},
		&models.GithubIssueLabel{},
		&models.GithubUser{},
		&models.GithubPullRequestIssue{},
		&models.GithubCommitStat{},
	)
	if err != nil {
		panic(err)
	}
}

func (plugin Github) Description() string {
	return "To collect and enrich data from GitHub"
}

func (plugin Github) Execute(options map[string]interface{}, progress chan<- float32, ctx context.Context) error {
	var err error

	// process option from request
	var op tasks.GithubOptions
	err = mapstructure.Decode(options, &op)
	if err != nil {
		return err
	}
	if op.Owner == "" {
		return fmt.Errorf("owner is required for GitHub execution")
	}
	if op.Repo == "" {
		return fmt.Errorf("repo is required for GitHub execution")
	}
	var since *time.Time
	if op.Since != "" {
		*since, err = time.Parse("2006-01-02T15:04:05Z", op.Since)
		if err != nil {
			return fmt.Errorf("invalid value for `since`: %w", err)
		}
	}
	logger := helper.NewDefaultTaskLogger(nil, "github")

	tasksToRun := make(map[string]bool, len(op.Tasks))
	if len(op.Tasks) == 0 {
		tasksToRun = map[string]bool{
			"collectApiRepositories": false,
			"extractApiRepositories": false,
			//"collectApiCommits":
			//"extractApiCommits":
			//"collectApiCommitStats":
			//"extractApiCommitStats":        false,
			"collectIssues":                true,
			"collectIssueEvents":           true,
			"collectIssueComments":         true,
			"collectApiIssues":             false,
			"extractApiIssues":             false,
			"collectApiPullRequests":       false,
			"extractApiPullRequests":       false,
			"collectApiComments":           false,
			"extractApiComments":           false,
			"collectApiEvents":             false,
			"extractApiEvents":             false,
			"collectApiPullRequestCommits": false,
			"extractApiPullRequestCommits": false,
			"collectPullRequests":          true,
			"collectPullRequestReviews":    true,
			"collectPullRequestCommits":    true,
			"enrichIssues":                 true,
			"enrichPullRequests":           true,
			"enrichComments":               true,
			"enrichPullRequestIssues":      true,
			"convertRepos":                 true,
			"convertIssues":                true,
			"convertIssueLabels":           true,
			"convertPullRequests":          true,
			"convertCommits":               true,
			"convertPullRequestCommits":    true,
			"convertPullRequestLabels":     true,
			"convertPullRequestIssues":     true,
			"convertNotes":                 true,
			"convertUsers":                 true,
		}
	} else {
		for _, task := range op.Tasks {
			tasksToRun[task] = true
		}
	}

	v := config.GetConfig()
	// process configuration
	endpoint := v.GetString("GITHUB_ENDPOINT")
	tokens := strings.Split(v.GetString("GITHUB_AUTH"), ",")
	// setup rate limit
	tokenCount := len(tokens)
	if tokenCount == 0 {
		return fmt.Errorf("owner is required for GitHub execution")
	}
	rateLimitPerSecondInt, err := core.GetRateLimitPerSecond(options, tokenCount)
	if err != nil {
		return err
	}
	scheduler, err := utils.NewWorkerScheduler(20, rateLimitPerSecondInt, ctx)
	if err != nil {
		return err
	}
	defer scheduler.Release()
	// TODO: add endpoind, auth validation
	apiClient := tasks.NewGithubApiClient(endpoint, tokens, v.GetString("GITHUB_PROXY"), ctx, scheduler, logger)

	taskData := &tasks.GithubTaskData{
		Options:   &op,
		ApiClient: &apiClient.ApiClient,
		Since:     since,
	}
	taskCtx := helper.NewDefaultTaskContext("github", ctx, logger, taskData, tasksToRun)
	repo := models.GithubRepo{}
	err = taskCtx.GetDb().Model(&models.GithubRepo{}).
		Where("owner_login = ? and `name` = ?", taskData.Options.Owner, taskData.Options.Repo).Limit(1).Find(&repo).Error
	if err != nil {
		return err
	}

	taskData.Repo = &repo

	newTasks := []struct {
		name       string
		entryPoint core.SubTaskEntryPoint
	}{
		//{name: "collectApiRepositories", entryPoint: tasks.CollectApiRepositories},
		//{name: "extractApiRepositories", entryPoint: tasks.ExtractApiRepositories},

		//{name: "collectApiCommits", entryPoint: tasks.CollectApiCommits},
		//{name: "extractApiCommits", entryPoint: tasks.ExtractApiCommits},
		//{name: "collectApiCommitStats", entryPoint: tasks.CollectApiCommitStats},
		//{name: "extractApiCommitStats", entryPoint: tasks.ExtractApiCommitStats},

		//{name: "collectApiIssues", entryPoint: tasks.CollectApiIssues},
		//{name: "extractApiIssues", entryPoint: tasks.ExtractApiIssues},
		//{name: "collectApiPullRequests", entryPoint: tasks.CollectApiPullRequests},
		//{name: "extractApiPullRequests", entryPoint: tasks.ExtractApiPullRequests},
		//{name: "collectApiComments", entryPoint: tasks.CollectApiComments},
		//{name: "extractApiComments", entryPoint: tasks.ExtractApiComments},
		//{name: "collectApiEvents", entryPoint: tasks.CollectApiEvents},
		//{name: "extractApiEvents", entryPoint: tasks.ExtractApiEvents},
		{name: "collectApiPullRequestCommits", entryPoint: tasks.CollectApiPullRequestCommits},
		{name: "extractApiPullRequestCommits", entryPoint: tasks.ExtractApiPullRequestCommits},
		{name: "collectApiPullRequestReviews", entryPoint: tasks.CollectApiPullRequestReviews},
		{name: "extractApiPullRequestReviews", entryPoint: tasks.ExtractApiPullRequestReviews},
	}
	for _, t := range newTasks {
		c, err := taskCtx.SubTaskContext(t.name)
		if err != nil {
			return err
		}
		if c != nil {
			err = t.entryPoint(c)
			if err != nil {
				return &errors.SubTaskError{
					SubTaskName: t.name,
					Message:     err.Error(),
				}
			}
		}
	}

	repoId := 1

	if tasksToRun["enrichPullRequestIssues"] {
		progress <- 0.73
		fmt.Println("INFO >>> Enriching PullRequestIssues")
		err = tasks.EnrichPullRequestIssues(ctx, repoId, op.Owner, op.Repo)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not enrich PullRequests: %v", err).Error(),
				SubTaskName: "enrichPullRequestIssues",
			}
		}
	}
	if tasksToRun["convertRepos"] {
		progress <- 0.80
		fmt.Println("INFO >>> Converting repos")
		err = tasks.ConvertRepos(ctx)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert Repos: %v", err).Error(),
				SubTaskName: "convertRepos",
			}
		}
	}
	if tasksToRun["convertIssues"] {
		progress <- 0.85
		fmt.Println("INFO >>> Converting Issues")
		err = tasks.ConvertIssues(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert Issues: %v", err).Error(),
				SubTaskName: "convertIssues",
			}
		}
	}
	if tasksToRun["convertIssueLabels"] {

		progress <- 0.90
		fmt.Println("INFO >>> starting convertIssueLabels")
		err = tasks.ConvertIssueLabels(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert IssueLabels: %v", err).Error(),
				SubTaskName: "convertIssueLabels",
			}
		}
	}
	if tasksToRun["convertPullRequests"] {
		progress <- 0.91
		fmt.Println("INFO >>> starting convertPullRequests")
		err = tasks.ConvertPullRequests(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert PullRequests: %v", err).Error(),
				SubTaskName: "convertPullRequests",
			}
		}
	}
	if tasksToRun["convertPullRequestLabels"] {
		progress <- 0.92
		fmt.Println("INFO >>> starting convertPullRequestLabels")
		err = tasks.ConvertPullRequestLabels(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert PullRequests: %v", err).Error(),
				SubTaskName: "convertPullRequestLabels",
			}
		}
	}
	if tasksToRun["convertCommits"] {
		progress <- 0.93
		fmt.Println("INFO >>> starting convertCommits")
		err = tasks.ConvertCommits(repoId, ctx)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert Commits: %v", err).Error(),
				SubTaskName: "convertCommits",
			}
		}
	}
	if tasksToRun["convertPullRequestCommits"] {
		progress <- 0.94
		fmt.Println("INFO >>> starting convertPullRequestCommits")
		err = tasks.PrCommitConvertor(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert PullRequestCommits: %v", err).Error(),
				SubTaskName: "convertPullRequestCommits",
			}
		}
	}
	if tasksToRun["convertPullRequestIssues"] {
		progress <- 0.95
		fmt.Println("INFO >>> Converting PullRequestIssues")
		err = tasks.ConvertPullRequestIssues(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert PullRequestCommits: %v", err).Error(),
				SubTaskName: "convertPullRequestIssues",
			}
		}
	}
	if tasksToRun["convertNotes"] {
		progress <- 0.97
		fmt.Println("INFO >>> starting convertNotes")
		err = tasks.ConvertNotes(ctx, repoId)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert Notes: %v", err).Error(),
				SubTaskName: "convertNotes",
			}
		}
	}
	if tasksToRun["convertUsers"] {
		progress <- 0.98
		fmt.Println("INFO >>> starting convertUsers")
		err = tasks.ConvertUsers(ctx)
		if err != nil {
			return &errors.SubTaskError{
				Message:     fmt.Errorf("could not convert Users: %v", err).Error(),
				SubTaskName: "convertUsers",
			}
		}
	}

	progress <- 1
	return nil
}

func (plugin Github) RootPkgPath() string {
	return "github.com/merico-dev/lake/plugins/github"
}

func (plugin Github) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"sources": {
			"GET":  api.ListSources,
			"POST": api.PutSource,
		},
		"sources/:sourceId": {
			"GET": api.GetSource,
			"PUT": api.PutSource,
		},
	}
}

// Export a variable named PluginEntry for Framework to search and load
var PluginEntry Github //nolint

// standalone mode for debugging
func main() {
	args := os.Args[1:]
	owner := "merico-dev"
	repo := "lake"
	if len(args) > 0 {
		owner = args[0]
	}
	if len(args) > 1 {
		repo = args[1]
	}

	err := core.RegisterPlugin("github", PluginEntry)
	if err != nil {
		panic(err)
	}
	PluginEntry.Init()
	progress := make(chan float32)

	go func() {
		err := PluginEntry.Execute(
			map[string]interface{}{
				"owner": owner,
				"repo":  repo,
				"tasks": []string{
					"collectApiRepositories",
					"extractApiRepositories",
					"collectApiCommits",
					"extractApiCommits",
					"collectApiCommitStats",
					"extractApiCommitStats",

					"collectApiIssues",
					"extractApiIssues",
					"collectApiComments",
					"extractApiComments",
					"collectApiEvents",
					"extractApiEvents",
					"collectApiPullRequests",
					"extractApiPullRequests",
					"collectApiPullRequestCommits",
					"extractApiPullRequestCommits",
					"collectApiPullRequestReviews",
					"extractApiPullRequestReviews",
					//"enrichApiIssues",
					//"collectIssueEvents",
					//"collectIssueComments",
					//"collectPullRequests",
					//"collectPullRequestReviews",
					//"collectPullRequestCommits",
					//"enrichIssues",
					//"enrichPullRequests",
					//"enrichComments",
					//"enrichPullRequestIssues",
					//"convertRepos",
					//"convertIssues",
					//"convertIssueLabels",
					//"convertPullRequests",
					//"convertCommits",
					//"convertPullRequestCommits",
					//"convertPullRequestLabels",
					//"convertPullRequestIssues",
					//"convertNotes",
					//"convertUsers",
				},
			},
			progress,
			context.Background(),
		)
		if err != nil {
			panic(err)
		}
		close(progress)
	}()
	for p := range progress {
		fmt.Println(p)
	}
}
