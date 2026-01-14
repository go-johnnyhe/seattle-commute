package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"seattle-commute-cli/config"
	"seattle-commute-cli/validation"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize your commute settings",
	Long:  "Set up your home address and Google Maps API key for transit directions",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🏠 Seattle Commute CLI Setup")
		fmt.Println("=============================")

		if len(args) > 0 {
			cfg.HomeAddress = strings.Join(args, " ")
		} else {
			fmt.Print("Enter your home address: ")
			homeAddr, _ := reader.ReadString('\n')
			cfg.HomeAddress = strings.TrimSpace(homeAddr)
		}

		if cfg.GoogleAPIKey == "" {
			fmt.Print("Enter your Google Maps API key: ")
			apiKey, _ := reader.ReadString('\n')
			cfg.GoogleAPIKey = strings.TrimSpace(apiKey)
		}

		validator, err := validation.NewAddressValidator(cfg.GoogleAPIKey)
		if err != nil {
			fmt.Printf("⚠️  Unable to validate addresses (API key might be invalid): %v\n", err)
		} else {
			fmt.Print("🔍 Validating home address... ")
			validatedHome, err := validator.ValidateSeattleAddress(cfg.HomeAddress)
			if err != nil {
				fmt.Printf("\n⚠️  %v\n", err)
				fmt.Print("Continue anyway? (y/N): ")
				response, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(response)) != "y" {
					fmt.Println("Setup cancelled.")
					os.Exit(1)
				}
			} else {
				cfg.HomeAddress = validatedHome
				fmt.Println("✅")
			}
		}

		fmt.Print("Enter work address (optional): ")
		workAddr, _ := reader.ReadString('\n')
		workAddr = strings.TrimSpace(workAddr)
		if workAddr != "" {
			if validator != nil {
				fmt.Print("🔍 Validating work address... ")
				validatedWork, err := validator.ValidateSeattleAddress(workAddr)
				if err != nil {
					fmt.Printf("\n⚠️  %v\n", err)
					fmt.Print("Continue anyway? (y/N): ")
					response, _ := reader.ReadString('\n')
					if strings.ToLower(strings.TrimSpace(response)) != "y" {
						workAddr = ""
					} else {
						cfg.WorkAddress = workAddr
					}
				} else {
					cfg.WorkAddress = validatedWork
					fmt.Println("✅")
				}
			} else {
				cfg.WorkAddress = workAddr
			}
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ Configuration saved!")
		fmt.Printf("Home: %s\n", cfg.HomeAddress)
		if cfg.WorkAddress != "" {
			fmt.Printf("Work: %s\n", cfg.WorkAddress)
		}
		fmt.Println("\nRun 'commute' to see routes home")
		if cfg.WorkAddress != "" {
			fmt.Println("Run 'commute -w' to see routes to work")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}