package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"macsetup/internal/installer"
	"macsetup/internal/tui"
	"macsetup/internal/utils"

	"github.com/spf13/cobra"
)

func Execute(version, commit, date string) {
	root := &cobra.Command{
		Use:           "macsetup",
		Short:         "Team macOS onboarding setup tool",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			headless, _ := cmd.Flags().GetBool("headless")
			workers, _ := cmd.Flags().GetInt("workers")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			logFile, _ := cmd.Flags().GetString("log-file")
			verbose, _ := cmd.Flags().GetBool("verbose")

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			var out io.Writer = os.Stdout
			var closeOut func() error
			if logFile != "" {
				f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					return err
				}
				closeOut = f.Close
				out = io.MultiWriter(os.Stdout, f)
			}
			if closeOut != nil {
				defer closeOut()
			}

			if dryRun {
				return runDryRun(ctx, out)
			}

			if err := utils.PreflightChecks(ctx); err != nil {
				return err
			}

			if headless {
				selection := installer.DefaultSelection()
				summary, err := installer.RunInstallPlan(ctx, selection, workers, out, installer.RunOptions{Verbose: verbose})
				if err != nil {
					return err
				}
				if summary.FailedCount() > 0 {
					return fmt.Errorf("%d steps failed", summary.FailedCount())
				}
				return nil
			}

			return tui.Run(ctx, workers, verbose)
		},
	}

	root.Version = fmt.Sprintf("%s (commit %s, date %s)", version, commit, date)
	root.SetVersionTemplate("macsetup {{.Version}}\n")

	root.Flags().Bool("headless", false, "Run without TUI using default selections")
	root.Flags().Int("workers", 5, "Max parallel formula installs")
	root.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")
	root.Flags().String("log-file", "", "Write detailed logs to this file (headless mode)")
	root.Flags().BoolP("verbose", "v", false, "Verbose output (more details for debugging)")

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), root.Version)
		},
	})

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runDryRun(ctx context.Context, out io.Writer) error {
	selection := installer.DefaultSelection()
	plan := installer.DryRunPlan(ctx, selection)
	_, _ = fmt.Fprintln(out, "Dry run: planned steps")
	for _, line := range plan {
		_, _ = fmt.Fprintln(out, "- "+line)
	}
	return nil
}
