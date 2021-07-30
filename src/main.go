package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/xanzy/go-gitlab"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// Function for configure helm from envVars (set NS)
func configHelm(getConfigEnv *config) action.Configuration {
    settings := cli.New()
    actionConfig := new(action.Configuration)
    if err := actionConfig.Init(settings.RESTClientGetter(), getConfigEnv.Namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        log.Printf("%+v", err)
        os.Exit(1)
    }
	return *actionConfig
}

// Function for authentication in gitlab
func gitClient(getConfigEnv *config) gitlab.Client{
	gitlabClient, err := gitlab.NewClient(getConfigEnv.GitlabToken, gitlab.WithBaseURL(getConfigEnv.GitUrl))
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	return *gitlabClient
}

// Function get deploed release in K8s features NS
func helmList() []string{
	kubeConfig := configHelm(getEnvVars())
    client := action.NewList(&kubeConfig)
    client.Deployed = true
    results, err := client.Run()
    if err != nil {
        log.Printf("%+v", err)
        os.Exit(1)
    }
	fmt.Println("Features branches in K8s NS features")
	fmt.Println("____________________________________")
	helmBranchesList := make([]string, 0)
    for i, _ := range results {
		helmBranchesList = append(helmBranchesList, results[i].Name)
		fmt.Println(results[i].Name)
	}
	fmt.Println("------------------------------------")
	return helmBranchesList
}

// Function get gitlab branches list
func gitList(getConfigEnv *config)[]string {
	git := gitClient(getEnvVars())
	fmt.Printf("Features branches in GitLab for project with ProjectID`s %s\n", getConfigEnv.Pids)
	fmt.Println("__________________________________________________________\n")
	mstvBranches := make([]string,0)
	for i, _ := range getConfigEnv.Pids {
		gitBranches, _, err := git.Branches.ListBranches(getConfigEnv.Pids[i], &gitlab.ListBranchesOptions{})
		if err != nil {
			fmt.Println("Fatal", err)
		}
		for j, _ := range gitBranches {
			mstvBranches = append(mstvBranches, strings.ToLower(gitBranches[j].Name))
		}
	}
	for k, _ := range mstvBranches {
		fmt.Println(mstvBranches[k])
	}
	fmt.Println("----------------------------------------------------------")
	return mstvBranches
}

// Main function for compare releasse in K8s and branches in gitlab
func main() {
	flag.Parse()
	configHelm(getEnvVars())
	kubeConfig := configHelm(getEnvVars())
	gitList := gitList(getEnvVars())
	helmList := helmList()

	// Get deffirence between source and destination
	absentFeatures := make([]string, 0)
	sort.Strings(gitList)
	for _, featuresBranches := range helmList {
		position := sort.SearchStrings(gitList, featuresBranches)
		if position < len(gitList) {
				if gitList[position] != featuresBranches {
						absentFeatures = append(absentFeatures, featuresBranches)
				}
		}
	}
	fmt.Println("Next release was removed from NS features", absentFeatures)

	// Helm delete release fron features NS
	uninstall := action.NewUninstall(&kubeConfig)
	for i, _ := range absentFeatures {
		uninstall.Description = absentFeatures[i]
		results, err := uninstall.Run(uninstall.Description)
		if err != nil {
			log.Printf("%+v", err)
			os.Exit(1)
		}
		fmt.Println(results)
	}
}