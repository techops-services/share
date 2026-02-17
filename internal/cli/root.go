package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/techops-services/share/internal/client"
	"github.com/techops-services/share/internal/clipboard"
	"github.com/techops-services/share/internal/config"
	"github.com/techops-services/share/internal/display"
)

var (
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildDate    = "unknown"
)

// SetBuildInfo sets the build-time version information.
func SetBuildInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

// flags stores persistent flag values.
type flags struct {
	ttl         string
	verbose     bool
	noClipboard bool
	configPath  string
	apiKey      string
	endpoint    string
}

// NewRootCmd creates the root cobra command with all subcommands.
func NewRootCmd() *cobra.Command {
	f := &flags{}

	root := &cobra.Command{
		Use:   "share [file or directory]",
		Short: "Upload HTML and get a short URL",
		Long: `share uploads HTML content to a Cloudflare Worker and returns a short URL.

Examples:
  share page.html                 Upload a file
  share ./dist                    Upload index.html from a directory
  echo "<h1>hi</h1>" | share     Pipe HTML from stdin
  pbpaste | share                 Upload clipboard contents
  share list                      List your shared pages
  share delete V1StGXR8           Delete a shared page`,
		Args:               cobra.MaximumNArgs(1),
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpload(cmd, args, f)
		},
	}

	// Persistent flags (available to all subcommands).
	pf := root.PersistentFlags()
	pf.StringVar(&f.ttl, "ttl", "", "Expiry duration (e.g. 1h, 7d, 30d). Default: 24h")
	pf.BoolVar(&f.verbose, "verbose", false, "Print full response details")
	pf.BoolVar(&f.noClipboard, "no-clipboard", false, "Don't copy URL to clipboard")
	pf.StringVar(&f.configPath, "config", "", "Config file path (default: ~/.config/share/config.toml)")
	pf.StringVar(&f.apiKey, "api-key", "", "API key (overrides config)")
	pf.StringVar(&f.endpoint, "endpoint", "", "API endpoint URL (overrides config)")

	// Subcommands.
	root.AddCommand(newListCmd(f))
	root.AddCommand(newDeleteCmd(f))
	root.AddCommand(newVersionCmd())
	root.AddCommand(newInitCmd(f))

	return root
}

// loadConfig loads and merges config with flag overrides.
func loadConfig(f *flags) (*config.Config, error) {
	cfg, err := config.Load(f.configPath)
	if err != nil {
		return nil, err
	}

	// Flag overrides have highest precedence.
	if f.apiKey != "" {
		cfg.APIKey = f.apiKey
	}
	if f.endpoint != "" {
		cfg.Endpoint = f.endpoint
	}
	if f.noClipboard {
		cfg.Clipboard = false
	}

	return cfg, nil
}

// newClient creates an API client from resolved config.
func newClient(cfg *config.Config) *client.Client {
	opts := []client.Option{
		client.WithEndpoint(cfg.Endpoint),
	}
	if cfg.APIKey != "" {
		opts = append(opts, client.WithAPIKey(cfg.APIKey))
	}
	return client.New(opts...)
}

// runUpload handles the default command: upload from file, directory, or stdin.
func runUpload(cmd *cobra.Command, args []string, f *flags) error {
	cfg, err := loadConfig(f)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Determine TTL.
	ttlStr := cfg.DefaultTTL
	if f.ttl != "" {
		ttlStr = f.ttl
	}
	ttlSeconds, err := config.ParseTTL(ttlStr)
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}

	var htmlReader io.Reader

	if len(args) == 0 {
		// No args: try stdin.
		htmlReader, err = readStdin()
		if err != nil {
			// Not a pipe -- show help.
			return cmd.Help()
		}
	} else {
		target := args[0]
		htmlReader, err = readTarget(target)
		if err != nil {
			return err
		}
	}

	// Upload.
	c := newClient(cfg)
	resp, uploadErr := c.Upload(htmlReader, ttlSeconds)
	if uploadErr != nil {
		return formatUploadError(uploadErr, c.LastRateLimit)
	}

	// Print URL (clean for piping).
	fmt.Println(resp.URL)

	// Copy to clipboard.
	if cfg.Clipboard {
		if clipErr := clipboard.Copy(resp.URL); clipErr != nil {
			if f.verbose {
				fmt.Fprintf(os.Stderr, "Warning: could not copy to clipboard: %v\n", clipErr)
			}
		} else if f.verbose {
			fmt.Fprintln(os.Stderr, "Copied to clipboard.")
		}
	}

	// Verbose output.
	if f.verbose {
		fmt.Fprintln(os.Stderr)
		display.PrintVerboseUpload(os.Stderr, resp, c.LastRateLimit)
	}

	return nil
}

// readStdin checks if stdin is a pipe and reads all content.
func readStdin() (io.Reader, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("checking stdin: %w", err)
	}

	if fi.Mode()&os.ModeCharDevice != 0 {
		// stdin is a terminal, not a pipe.
		return nil, fmt.Errorf("stdin is not a pipe")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("stdin is empty")
	}

	return strings.NewReader(string(data)), nil
}

// readTarget reads HTML from a file path or directory.
func readTarget(target string) (io.Reader, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("cannot access %q: %w", target, err)
	}

	if info.IsDir() {
		return readDirectory(target)
	}

	return readFile(target)
}

// readFile reads and returns an HTML file as a Reader.
func readFile(path string) (io.Reader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", path, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("file %q is empty", path)
	}
	return strings.NewReader(string(data)), nil
}

// readDirectory looks for index.html inside a directory. If not found, lists
// available .html files as a helpful error.
func readDirectory(dir string) (io.Reader, error) {
	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		return readFile(indexPath)
	}

	// No index.html -- list .html files for a helpful error.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	var htmlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			htmlFiles = append(htmlFiles, entry.Name())
		}
	}

	if len(htmlFiles) > 0 {
		return nil, fmt.Errorf(
			"no index.html found in %q\n\nHTML files found:\n  %s\n\nUpload one directly: share %s",
			dir,
			strings.Join(htmlFiles, "\n  "),
			filepath.Join(dir, htmlFiles[0]),
		)
	}

	return nil, fmt.Errorf("no HTML files found in %q", dir)
}

// formatUploadError produces user-friendly error messages.
func formatUploadError(err error, rl *client.RateLimitInfo) error {
	apiErr, ok := err.(*client.APIError)
	if !ok {
		return err
	}

	if apiErr.IsRateLimit() {
		msg := fmt.Sprintf("Rate limit exceeded.")
		if apiErr.RetryAfter > 0 {
			minutes := apiErr.RetryAfter / 60
			if minutes < 1 {
				msg += fmt.Sprintf(" Try again in %d seconds.", apiErr.RetryAfter)
			} else {
				msg += fmt.Sprintf(" Try again in %d minutes.", minutes)
			}
		}
		if rl != nil {
			msg += fmt.Sprintf("\n  Limit: %d uploads/hour", rl.Limit)
		}
		msg += "\n  Tip: Add an API key for higher limits."
		return fmt.Errorf("%s", msg)
	}

	if apiErr.IsPayloadTooLarge() {
		return fmt.Errorf("File too large (max 5 MB).\n  Tip: Minify your HTML or remove inline assets.")
	}

	return fmt.Errorf("%s", apiErr.Message)
}
