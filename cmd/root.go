package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"seattle-commute-cli/config"
	"seattle-commute-cli/distance"
	"seattle-commute-cli/location"
	"seattle-commute-cli/transit"
)

var (
	fromAddress string
	atHome      bool
	atWork      bool
	workFlag    bool
)

var rootCmd = &cobra.Command{
	Use:   "commute [from] [to]",
	Short: "Get Seattle transit directions between locations",
	Long:  "A CLI tool to get optimal Seattle commute routes using real-time transit data.\n\nUsage:\n  commute                           # Home from work\n  commute -w                        # Work from home\n  commute \"U District\" \"Capitol Hill\"  # Arbitrary routing",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Determine destination and current location
		var destination, currentLoc string
		var destinationType string

		if len(args) == 2 {
			// Arbitrary routing: commute "from" "to"
			currentLoc = args[0]
			destination = args[1]
			destinationType = fmt.Sprintf("destination (%s)", destination)
			fmt.Printf("📍 Route from %s to %s\n", currentLoc, destination)
		} else if len(args) == 1 {
			fmt.Printf("❌ Please provide both from and to locations, or use no arguments for home/work routing\n")
			fmt.Printf("Example: commute \"U District Station\" \"Capitol Hill\"\n")
			os.Exit(1)
		} else {
			// Home/work routing mode - requires config
			if !cfg.IsValid() {
				fmt.Println("❌ Configuration not found. Run 'commute init' to set up.")
				os.Exit(1)
			}

			if workFlag {
				// -w flag: going to work (assume at home)
				if cfg.WorkAddress == "" {
					fmt.Println("❌ Work address not configured. Run 'commute init' to set up.")
					os.Exit(1)
				}
				destination = cfg.WorkAddress
				currentLoc = cfg.HomeAddress
				destinationType = "work"
				fmt.Printf("📍 Going to work (assuming you're at home)\n")
			} else {
				// Default: going home (assume at work)
				destination = cfg.HomeAddress
				if cfg.WorkAddress != "" {
					currentLoc = cfg.WorkAddress
					destinationType = "home"
					fmt.Printf("📍 Going home (assuming you're at work)\n")
				} else {
					// No work address configured, fall back to IP detection
					fmt.Print("📍 Getting your current location... ")
					currentLoc, err = location.GetCurrentLocation()
					if err != nil {
						fmt.Printf("\nError: %v\n", err)
						os.Exit(1)
					}
					fmt.Printf("✅ (detected: %s)\n", currentLoc)
					destinationType = "home"
				}
			}

			// Override logic for explicit location flags (only for home/work mode)
			if atHome {
				currentLoc = cfg.HomeAddress
				fmt.Printf("📍 Override: using home as current location\n")
			} else if atWork && cfg.WorkAddress != "" {
				currentLoc = cfg.WorkAddress
				fmt.Printf("📍 Override: using work as current location\n")
			} else if fromAddress != "" {
				currentLoc = fromAddress
				fmt.Printf("📍 Override: using specified location: %s\n", fromAddress)
			}
		}

		// Check if already within walking distance
		fmt.Print("📏 Checking distance... ")
		distanceChecker, err := distance.NewDistanceChecker(cfg.GoogleAPIKey)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}


		isWalkable, walkTime, walkDistance, err := distanceChecker.IsWithinWalkingDistance(currentLoc, destination)
		if err == nil && isWalkable {
			fmt.Println("✅")

			if walkTime <= 2*time.Minute {
				fmt.Printf("\n🏠 You're already at %s!\n", destinationType)
				fmt.Printf("📍 Current location matches your %s address\n", destinationType)
				return
			} else {
				fmt.Printf("\n🚶‍♂️ You're already close to %s!\n", destinationType)
				fmt.Printf("Walking time: %s (%s)\n", formatDuration(walkTime), walkDistance)
				fmt.Printf("💡 No transit needed - just walk!\n")
				return
			}
		}
		fmt.Println("✅")

		fmt.Print("🚌 Finding transit routes... ")
		service, err := transit.NewTransitService(cfg.GoogleAPIKey)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}

		routes, err := service.GetNextRoutes(currentLoc, destination, 2)
		if err != nil {
			if strings.Contains(err.Error(), "no routes found") {
				fmt.Printf("\n❌ No transit routes found. This could mean:\n")
				fmt.Printf("   • No transit service at this time (most Seattle buses run 5 AM - 2 AM)\n")
				fmt.Printf("   • Your location is too far from Seattle transit\n")
				fmt.Printf("   • Try running 'commute init' to update your addresses\n")
			} else if strings.Contains(err.Error(), "ZERO_RESULTS") {
				fmt.Printf("\n❌ No routes found between these locations.\n")
				fmt.Printf("   • Check that both addresses are in the Seattle area\n")
				fmt.Printf("   • Transit might not be available at this time\n")
			} else {
				fmt.Printf("\n❌ Transit service error: %v\n", err)
			}
			os.Exit(1)
		}
		fmt.Println("✅")

		if len(routes) == 0 {
			fmt.Println("❌ No transit routes found")
			os.Exit(1)
		}

		fmt.Printf("\n🏠 Routes to %s (%s)\n", destination, destinationType)
		fmt.Println("=" + strings.Repeat("=", len(destination)+12))

		printRoutes(routes[:min(5, len(routes))])
	},
}

func printRoutes(routes []transit.Route) {
	now := time.Now()

	for i, route := range routes {
		timeUntil := route.DepartureTime.Sub(now)
		status := ""
		if timeUntil < 0 {
			continue
		} else if timeUntil < 5*time.Minute {
			status = " 🏃‍♂️ LEAVING SOON"
		} else if timeUntil < 15*time.Minute {
			status = " ⚡ GOOD TIMING"
		}

		fmt.Printf("\n%d. Depart: %s (%s)%s\n",
			i+1,
			route.DepartureTime.Format("3:04 PM"),
			formatDuration(timeUntil),
			status)

		fmt.Printf("   Arrive: %s (Travel: %s)\n",
			route.ArrivalTime.Format("3:04 PM"),
			formatDuration(route.Duration))

		fmt.Printf("   Distance: %s\n", route.Distance)

		var transitSteps []transit.Step
		for _, step := range route.Steps {
			if step.Mode == "TRANSIT" {
				transitSteps = append(transitSteps, step)
			}
		}

		if len(transitSteps) > 0 {
			fmt.Print("   🚌 ")
			lineInfos := make([]string, 0, len(transitSteps))
			for _, step := range transitSteps {
				if step.LineInfo != "" {
					lineInfos = append(lineInfos, step.LineInfo)
				}
			}
			fmt.Printf("%s\n", strings.Join(lineInfos, " → "))
		}
	}

	fmt.Printf("\n📱 Tip: Add this tool to your PATH for quick access anywhere!\n")
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "now"
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remainingMinutes := minutes % 60
	if remainingMinutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remainingMinutes)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&workFlag, "work", "w", false, "Get routes to work (default: routes to home)")
	rootCmd.Flags().StringVarP(&fromAddress, "from", "f", "", "Specify your current location instead of auto-detection")
	rootCmd.Flags().BoolVar(&atHome, "at-home", false, "Override: you're currently at your home address")
	rootCmd.Flags().BoolVar(&atWork, "at-work", false, "Override: you're currently at your work address")
}