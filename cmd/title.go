package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/magarcia/ccsession-viewer/discovery"
)

var titleCmd = &cobra.Command{
	Use:   "title <transcript>",
	Short: "Print the derived human-readable title for a session transcript",
	Long: "Print the title cc2md would put in a session note's frontmatter.\n" +
		"Exposed so external tooling (such as the archive backfill) can reuse\n" +
		"the title rules rather than reimplementing them and drifting.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTitleTo(cmd.OutOrStdout(), args)
	},
}

func init() {
	rootCmd.AddCommand(titleCmd)
}

// runTitleTo writes the derived title for args[0] to w. Split from the cobra
// RunE so tests can capture output without touching os.Stdout.
func runTitleTo(w io.Writer, args []string) error {
	path := args[0]
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("transcript %s: %w", path, err)
	}
	_, err := fmt.Fprintln(w, discovery.DeriveTitle(path))
	return err
}
