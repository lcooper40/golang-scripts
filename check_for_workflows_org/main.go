package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

const githubAPI = "https://api.github.com"

type Repo struct {
	Name  string `json:"name"`
	Owner Owner  `json:"owner"`
}

type Owner struct {
	Login string `json:"login"`
}

type Workflow struct {
	Name string `json:"name"`
}

type WorkflowsResponse struct {
	TotalCount int        `json:"total_count"`
	Workflows  []Workflow `json:"workflows"`
}

func getOrgRepos(orgName, token string) ([]Repo, error) {
	url := fmt.Sprintf("%s/orgs/%s/repos", githubAPI, orgName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repos: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func getWorkflows(owner, repoName, token string) (*WorkflowsResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/actions/workflows", githubAPI, owner, repoName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get workflows for repo %s: %s", repoName, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var workflowsResponse WorkflowsResponse
	if err := json.Unmarshal(body, &workflowsResponse); err != nil {
		return nil, err
	}
	return &workflowsResponse, nil
}

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN environment variable not set")
		return
	}

	orgName := "your-org-name"
	omitPattern := regexp.MustCompile(`your-regex-pattern`)

	repos, err := getOrgRepos(orgName, token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, repo := range repos {
		workflowsResponse, err := getWorkflows(repo.Owner.Login, repo.Name, token)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		if workflowsResponse.TotalCount > 0 {
			fmt.Printf("Repo: %s has the following workflows:\n", repo.Name)
			for _, workflow := range workflowsResponse.Workflows {
				if !omitPattern.MatchString(workflow.Name) {
					fmt.Printf("  - %s\n", workflow.Name)
				}
			}
		} else {
			fmt.Printf("Repo: %s has no workflows.\n", repo.Name)
		}
	}
}
