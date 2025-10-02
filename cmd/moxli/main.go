package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/lelopez-io/moxli/internal/session"
	"github.com/urfave/cli/v2"
)

var (
	buildVersion = "dev"
	commit       = "none"
	date         = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "moxli",
		Usage:   "Bookmark management CLI",
		Version: buildVersion,
		Commands: []*cli.Command{
			versionCommand(),
			sessionTestCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("Version:    %s\n", buildVersion)
			fmt.Printf("Commit:     %s\n", commit)
			fmt.Printf("Build Date: %s\n", date)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}

func sessionTestCommand() *cli.Command {
	return &cli.Command{
		Name:  "session-test",
		Usage: "Test session manager functionality",
		Action: func(c *cli.Context) error {
			fmt.Println("🧪 Testing Session Manager")

			// Create manager
			manager, err := session.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			fmt.Printf("✅ Manager created (config dir: ~/.moxli/)\n\n")

			// Check if session exists
			if manager.Exists() {
				fmt.Println("📂 Found existing session:")
				existing, err := manager.Load()
				if err != nil {
					return fmt.Errorf("failed to load session: %w", err)
				}
				fmt.Printf("   Working Dir:  %s\n", existing.WorkingDir)
				fmt.Printf("   Current File: %s\n", existing.CurrentFile)
				fmt.Printf("   Last Modified: %s\n", existing.LastModified.Format("2006-01-02 15:04:05"))
				fmt.Printf("   Merge History: %d records\n\n", len(existing.MergeHistory))

				for i, record := range existing.MergeHistory {
					fmt.Printf("   Merge %d:\n", i+1)
					fmt.Printf("     Base: %s\n", record.BaseFile)
					fmt.Printf("     Sources: %v\n", record.SourceFiles)
					fmt.Printf("     Enhanced: %d bookmarks\n", record.Enhanced)
					fmt.Printf("     Date: %s\n\n", record.Date.Format("2006-01-02 15:04:05"))
				}
			} else {
				fmt.Println("📭 No existing session found")
			}

			// Create a test session
			fmt.Println("💾 Creating test session...")
			testSession := &session.Session{
				WorkingDir:  "/Users/example/bookmarks",
				CurrentFile: "/Users/example/bookmarks/enhanced.json",
			}

			// Add a merge record
			testSession.AddMergeRecord(
				"/Users/example/bookmarks/anybox.json",
				[]string{
					"/Users/example/bookmarks/firefox.html",
					"/Users/example/bookmarks/safari.html",
				},
				25,
			)

			// Save session
			err = manager.Save(testSession)
			if err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Println("✅ Session saved to ~/.moxli/session.yaml")

			// Load it back
			fmt.Println("📖 Loading session back...")
			loaded, err := manager.Load()
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			fmt.Printf("✅ Session loaded successfully\n")
			fmt.Printf("   Working Dir:  %s\n", loaded.WorkingDir)
			fmt.Printf("   Current File: %s\n", loaded.CurrentFile)
			fmt.Printf("   Last Modified: %s\n", loaded.LastModified.Format("2006-01-02 15:04:05"))
			fmt.Printf("   Merge History: %d records\n\n", len(loaded.MergeHistory))

			// Verify data matches
			if loaded.WorkingDir == testSession.WorkingDir &&
				loaded.CurrentFile == testSession.CurrentFile &&
				len(loaded.MergeHistory) == 1 &&
				loaded.MergeHistory[0].Enhanced == 25 {
				fmt.Println("✅ All data verified correctly!")
			} else {
				return fmt.Errorf("data mismatch after load")
			}

			fmt.Println("💡 You can inspect the session file at: ~/.moxli/session.yaml")
			fmt.Println("💡 Run this command again to see the existing session loaded")
			fmt.Println("💡 To clear: rm ~/.moxli/session.yaml")

			return nil
		},
	}
}
