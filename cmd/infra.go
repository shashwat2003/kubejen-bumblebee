/*
Copyright © 2024 Shashwat <shashwat13.8@gmail.com>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"kubejen/bumblebee/utils"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// infraCmd represents the infra command
var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Controls for infra repo.",
	Long: `Helps in controlling infra repo
for both production and non-production environments.`,
	Run: callInfra,
	Args: func(cmd *cobra.Command, args []string) error {
		isForAll, _ := cmd.Flags().GetBool("all")
		if isForAll && len(args) < 2 {
			return errors.New("requires at least two args with -a flag")
		} else if !isForAll && len(args) < 3 {
			return errors.New("requires at least three args")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infraCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infraCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	infraCmd.Flags().BoolP("dev", "d", false, "Runs in non-production environment.")
	infraCmd.Flags().BoolP("all", "a", false, "Execute for all compose file in the infra app.")
}

func callInfra(cmd *cobra.Command, args []string) {
	isDev, err := cmd.Flags().GetBool("dev")
	if err != nil {
		log.Fatal("Error in loading dev flag.")
	}
	isForAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		log.Fatal("Error in loading dev err.")
	}

	var environment string
	if isDev {
		environment = "non-production"
	} else {
		environment = "production"
	}

	config, config_file := utils.RetrieveOrGenerateConfig("infra")
	if config.Infra.Path == "" {
		promptUser(&config, config_file)
	}

	infra_app := path.Join(config.Infra.Path, args[0], environment)
	if !utils.FolderExists(infra_app) {
		log.Fatalf("invalid app: %s with environment: %s not found in infra", args[0], environment)
	}

	if isForAll {
		runAllCompose(infra_app, args[1:])
	} else {
		infra_app_compose := checkComposeFile(infra_app, args[1])
		runComposeWithArgs(infra_app_compose, args[2:])
	}
}

func promptUser(config *utils.Config, config_file string) {

	// create config
	validate := func(input string) error {
		// if !FolderExists(input) {
		// 	return errors.New("invalid folder specified")
		// }
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "Please enter the path of infra repo.",
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	config.Infra.Path = result
	utils.SaveConfig(config, config_file)
}

func checkComposeFile(infra_app string, service string) string {
	// first check if compose file is present
	// --> postgres.yml
	compose_file := fmt.Sprintf("%s-compose.yml", service)
	var compose_file_path = path.Join(infra_app, compose_file)
	if utils.FileExists(compose_file_path) {
		return compose_file_path
	}
	// then check if it is inside the folder of the same name
	// --> postgres/postgres-compose.yml
	compose_file_path = path.Join(infra_app, service, compose_file)
	if utils.FileExists(compose_file_path) {
		return compose_file_path
	}
	log.Fatalf("invalid compose file: %s not found in %s", compose_file, infra_app)
	return ""
}

func runAllCompose(infra_app string, args []string) {
	// list all items in infra_app
	entries, err := os.ReadDir(infra_app)
	if err != nil {
		log.Fatal(err)
	}

	// read the priority config
	fmt.Printf("Checking for priority config(%s)...\n", utils.INFRA_PRIORITY_FILE)
	priority_config := path.Join(infra_app, utils.INFRA_PRIORITY_FILE)
	priority := map[string]int{}
	if utils.FileExists((priority_config)) {
		fmt.Println("Priority config found!")
		fh, _ := os.OpenFile(priority_config, os.O_RDONLY, 0777)
		defer fh.Close()
		sc := bufio.NewScanner(fh)
		sc.Split(bufio.ScanLines)
		i := 999 // random high value to set the highest priority
		for sc.Scan() {
			data := sc.Text()
			priority[data] = i
			i -= 1
		}
	}

	// sort the entries slice according to priority
	sort.Slice(entries, func(i, j int) bool {
		p1, ok := priority[strings.ReplaceAll(entries[i].Name(), "-compose.yml", "")]
		if !ok {
			p1 = 0
		}
		p2, ok := priority[strings.ReplaceAll(entries[j].Name(), "-compose.yml", "")]
		if !ok {
			p2 = 0
		}
		return p1 > p2
	})

	for _, e := range entries {
		if e.IsDir() {
			// if it is a directory then check for postgres/postgres-compose.yml
			compose_file := path.Join(infra_app, e.Name(), fmt.Sprintf("%s-compose.yml", e.Name()))
			if utils.FileExists(compose_file) {
				runComposeWithArgs(compose_file, args)
			}
		} else {
			// else if it is a compose then run it
			if isCompose(e.Name()) {
				runComposeWithArgs(path.Join(infra_app, e.Name()), args)
			}
		}
	}
}

func isCompose(name string) bool {
	match, _ := regexp.MatchString("\\w.yml", name)
	return match
}

func runComposeWithArgs(compose_file string, args []string) {
	fmt.Println("\n----------------------------------------------------------------------")
	fmt.Printf("Running docker file -> %s\n", compose_file)
	// static check for up command
	if args[0] == "up" {
		args = append(args, "-d")
	}
	command := fmt.Sprintf("docker compose -f %s %s", compose_file, strings.Join(args, " "))
	utils.RunInShell(command)
	fmt.Printf("----------------------------------------------------------------------\n")
}
