package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/replace"
	"github.com/pyhub/pyhub-docs/internal/ui"
	"github.com/spf13/cobra"
)

var (
	rulesFile   string
	targetPath  string
	dryRun      bool
	backup      bool
	recursive   bool
	excludeGlob string
	concurrent  bool
	maxWorkers  int
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace text in documents using rules from a YAML file",
	Long: `Replace text in Word and PowerPoint documents based on rules defined in a YAML file.

The rules file should contain replacement pairs:
  - old: "old text"
    new: "new text"
  - old: "v1.0.0"
    new: "v2.0.0"

Examples:
  # Replace text in a single file
  dox replace --rules rules.yml --path document.docx

  # Replace text in all documents in a directory
  dox replace --rules rules.yml --path ./docs

  # Dry run to preview changes
  dox replace --rules rules.yml --path ./docs --dry-run

  # Create backups before modifying
  dox replace --rules rules.yml --path ./docs --backup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate inputs
		if rulesFile == "" {
			return pkgErrors.NewValidationError("rules", rulesFile, "rules file is required")
		}
		if targetPath == "" {
			return pkgErrors.NewValidationError("path", targetPath, "target path is required")
		}

		// Load rules from YAML file
		rules, err := replace.LoadRulesFromFile(rulesFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return pkgErrors.NewFileError(rulesFile, "loading rules", pkgErrors.ErrFileNotFound)
			}
			return pkgErrors.NewFileError(rulesFile, "loading rules", err)
		}

		if len(rules) == 0 {
			ui.PrintWarning("No replacement rules found in the file")
			return nil
		}

		// Print rules if in dry-run mode
		if dryRun {
			ui.PrintHeader("Replacement Rules to Apply")
			for i, rule := range rules {
				ui.PrintStep(i+1, len(rules), fmt.Sprintf("Replace '%s' with '%s'", rule.Old, rule.New))
			}
		}

		// Check if target is a file or directory
		info, err := os.Stat(targetPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return pkgErrors.NewFileError(targetPath, "accessing", pkgErrors.ErrFileNotFound)
			}
			if errors.Is(err, os.ErrPermission) {
				return pkgErrors.NewFileError(targetPath, "accessing", pkgErrors.ErrPermissionDenied)
			}
			return pkgErrors.NewFileError(targetPath, "accessing", err)
		}

		// Create backup if requested
		if backup && !dryRun {
			if !quiet {
				ui.PrintInfo("Creating backup of %s...", targetPath)
			}
			if err := createBackup(targetPath, info.IsDir()); err != nil {
				return pkgErrors.NewFileError(targetPath, "creating backup", err)
			}
			if !quiet {
				ui.PrintSuccess("Backup created successfully")
			}
		}

		// Process based on target type
		if info.IsDir() {
			// Process directory
			if dryRun {
				return previewDirectoryReplacements(targetPath, rules, recursive)
			}
			
			var results []replace.ReplaceResult
			var err error
			
			if concurrent {
				// Use concurrent processing for better performance
				opts := replace.DefaultConcurrentOptions()
				if maxWorkers > 0 {
					opts.MaxWorkers = maxWorkers
				}
				opts.ShowProgress = !quiet && !verbose
				
				if verbose {
					ui.PrintInfo("Processing directory with %d workers...", opts.MaxWorkers)
				}
				
				results, err = replace.ReplaceInDirectoryConcurrent(targetPath, rules, recursive, excludeGlob, opts)
			} else {
				results, err = replace.ReplaceInDirectoryWithResultsAndExclude(targetPath, rules, recursive, excludeGlob)
			}
			if err != nil {
				return fmt.Errorf("failed to process directory: %w", err)
			}

			// Print results
			printResults(results)
		} else {
			// Process single file
			ext := strings.ToLower(filepath.Ext(targetPath))
			if ext != ".docx" && ext != ".pptx" {
				return pkgErrors.NewDocumentError(targetPath, ext, "unsupported format (only .docx and .pptx are supported)", pkgErrors.ErrUnsupportedFormat)
			}

			if dryRun {
				ui.PrintInfo("Would process file: %s", targetPath)
				return nil
			}

			if verbose {
				ui.PrintInfo("Processing file: %s", targetPath)
			}
			
			count, err := replace.ReplaceInDocumentWithCount(targetPath, rules)
			if err != nil {
				if errors.Is(err, pkgErrors.ErrDocumentCorrupted) {
					return pkgErrors.NewDocumentError(targetPath, ext, "document appears to be corrupted", err)
				}
				return pkgErrors.NewDocumentError(targetPath, ext, "processing failed", err)
			}
			
			if verbose {
				ui.PrintInfo("Made %d replacements in %s", count, targetPath)
			}

			ui.PrintSuccess("Successfully processed: %s", targetPath)
		}

		return nil
	},
}

// Helper functions

func createBackup(path string, isDir bool) error {
	// Use time-based timestamp for uniqueness
	timestamp := time.Now().Format("20060102_150405")
	
	if isDir {
		// For directories, create a backup directory with timestamp
		backupPath := path + "_backup_" + timestamp
		
		// Copy directory recursively
		return copyDir(path, backupPath)
	}
	
	// For files, create a backup copy with timestamp
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	backupPath := fmt.Sprintf("%s_backup_%s%s", base, timestamp, ext)
	
	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	return os.WriteFile(backupPath, input, 0644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		
		// Create directory or copy file
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		// Copy file
		input, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		return os.WriteFile(dstPath, input, info.Mode())
	})
}

func previewDirectoryReplacements(dirPath string, rules []replace.Rule, recursive bool) error {
	ui.PrintHeader("Files to Process")
	
	count := 0
	// Use the new walk function with exclude support
	err := replace.WalkDocumentFilesWithExclude(dirPath, recursive, excludeGlob, func(path string) error {
		ext := strings.ToLower(filepath.Ext(path))
		ui.PrintFileOperation("Preview", path, ext)
		count++
		return nil
	})
	
	if err != nil {
		return err
	}
	
	ui.PrintInfo("Total files to process: %d", count)
	return nil
}

func printResults(results []replace.ReplaceResult) {
	successCount := 0
	failureCount := 0
	totalReplacements := 0
	
	ui.PrintHeader("Processing Results")
	
	for _, result := range results {
		if result.Success {
			ui.PrintSuccess("%s (%d replacements)", result.FilePath, result.Replacements)
			successCount++
			totalReplacements += result.Replacements
		} else {
			ui.PrintError("%s - %v", result.FilePath, result.Error)
			failureCount++
		}
	}
	
	// Create summary statistics
	stats := map[string]interface{}{
		"Successful":          successCount,
		"Failed":             failureCount,
		"Total Files":        len(results),
		"Total Replacements": totalReplacements,
	}
	
	ui.PrintSummary("Summary", stats)
}

func init() {
	rootCmd.AddCommand(replaceCmd)

	replaceCmd.Flags().StringVarP(&rulesFile, "rules", "r", "", "YAML file containing replacement rules (required)")
	replaceCmd.Flags().StringVarP(&targetPath, "path", "p", "", "Target file or directory (required)")
	replaceCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	replaceCmd.Flags().BoolVar(&backup, "backup", false, "Create backup files before modification")
	replaceCmd.Flags().BoolVar(&recursive, "recursive", true, "Process subdirectories recursively")
	replaceCmd.Flags().StringVar(&excludeGlob, "exclude", "", "Glob pattern for files to exclude")
	replaceCmd.Flags().BoolVar(&concurrent, "concurrent", false, "Process files concurrently for better performance")
	replaceCmd.Flags().IntVar(&maxWorkers, "max-workers", 0, "Maximum number of concurrent workers (default: number of CPUs)")

	replaceCmd.MarkFlagRequired("rules")
	replaceCmd.MarkFlagRequired("path")
}