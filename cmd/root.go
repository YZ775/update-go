package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	showAll    bool
	noInteract bool
)

var rootCmd = &cobra.Command{
	Use:   "update-go",
	Short: "Go version management tool",
	Long:  `A CLI tool to list available Go versions and install them interactively.`,
	RunE:  runMain,
}

func init() {
	rootCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all versions (default: stable only)")
	rootCmd.Flags().BoolVarP(&noInteract, "no-interact", "n", false, "Disable interactive mode, only show list")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching available Go versions...")

	versions, err := FetchVersions()
	if err != nil {
		return fmt.Errorf("failed to fetch version list: %w", err)
	}

	// Filter stable versions unless --all is specified
	if !showAll {
		versions = GetStableVersions(versions)
	}

	// Get unique major versions (latest patch of each)
	versions = GetMajorVersions(versions)

	if len(versions) == 0 {
		fmt.Println("No available versions found")
		return nil
	}

	// Show current Go version
	currentVersion := runtime.Version()
	fmt.Printf("\nCurrent Go version: %s (%s/%s)\n\n", currentVersion, runtime.GOOS, runtime.GOARCH)

	if noInteract {
		fmt.Println("Available versions:")
		for _, v := range versions {
			marker := "  "
			if v.Version == currentVersion {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, FormatVersion(v.Version))
		}
		return nil
	}

	// Interactive mode using promptui
	items := make([]string, len(versions))
	for i, v := range versions {
		label := FormatVersion(v.Version)
		if v.Version == currentVersion {
			label = fmt.Sprintf("%s (current)", label)
		}
		if v.Stable {
			label = fmt.Sprintf("%s [stable]", label)
		}
		items[i] = label
	}

	prompt := promptui.Select{
		Label: "Select a version to install",
		Items: items,
		Size:  15,
		Searcher: func(input string, index int) bool {
			return strings.Contains(items[index], input)
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			fmt.Println("\nCancelled")
			return nil
		}
		return fmt.Errorf("selection failed: %w", err)
	}

	selectedVersion := versions[index]

	// Ask for confirmation
	confirmPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("Install Go %s", FormatVersion(selectedVersion.Version)),
		IsConfirm: true,
	}

	_, err = confirmPrompt.Run()
	if err != nil {
		fmt.Println("Cancelled")
		return nil
	}

	return InstallVersion(selectedVersion)
}
