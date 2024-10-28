package cmd

import (
	"github.com/julien040/anyquery/controller"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [database file] [sql query]",
	Short: "Run a SQL query",
	Long: `Run a SQL query on the data sources installed on the system.
The query can be specified as an argument or read from stdin.
If no query is provided, the command will launch an interactive input.`,
	RunE:    controller.Query,
	Aliases: []string{"q"},
	Example: `# Run a one-off query
anyquery query -d mydatabase.db -q "SELECT * FROM mytable"

# Open the interactive shell
anyquery query -d mydatabase.db

# Open the interactive shell in memory
anyquery query

# Query from stdin
echo "SELECT * FROM mytable" | anyquery query -d mydatabase.db`,
}

func init() {
	rootCmd.AddCommand(queryCmd)
	addFlag_commandModifiesConfiguration(queryCmd)
	addFlag_commandPrintsData(queryCmd)
	queryCmd.Flags().StringP("database", "d", "", "Database to connect to (a path or :memory:)")
	queryCmd.Flags().Bool("in-memory", false, "Use an in-memory database")
	queryCmd.Flags().Bool("readonly", false, "Start the server in read-only mode")
	queryCmd.Flags().Bool("read-only", false, "Start the server in read-only mode")
	queryCmd.Flags().StringArray("init", []string{}, "Run SQL commands in a file before the query. You can specify multiple files.")
	queryCmd.Flags().Bool("dev", false, "Run the program in developer mode")
	queryCmd.Flags().StringSlice("extension", []string{}, "Load one or more extensions by specifying their path. Separate multiple extensions with a comma.")

	// Query flags
	queryCmd.Flags().StringP("query", "q", "", "Query to run")

	// Log flags
	queryCmd.Flags().String("log-file", "", "Log file")
	queryCmd.Flags().String("log-level", "info", "Log level (trace, debug, info, warn, error, off)")
	queryCmd.Flags().String("log-format", "text", "Log format (text, json)")

	// Alternative language flags
	queryCmd.Flags().String("language", "", "Alternative language to use")
	queryCmd.Flags().Bool("prql", false, "Use the PRQL language (requires prqlc in PATH)")
	queryCmd.Flags().Bool("pql", false, "Use the PQL language")
}
