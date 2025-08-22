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
	"github.com/spf13/cobra"
)

var (
	rulesFile   string
	targetPath  string
	dryRun      bool
	backup      bool
	recursive   bool
	excludeGlob string
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
  pyhub-docs replace --rules rules.yml --path document.docx

  # Replace text in all documents in a directory
  pyhub-docs replace --rules rules.yml --path ./docs

  # Dry run to preview changes
  pyhub-docs replace --rules rules.yml --path ./docs --dry-run

  # Create backups before modifying
  pyhub-docs replace --rules rules.yml --path ./docs --backup`,
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
			fmt.Println("No replacement rules found in the file")
			return nil
		}

		// Print rules if in dry-run mode
		if dryRun {
			fmt.Println("Replacement rules to be applied:")
			for i, rule := range rules {
				fmt.Printf("  %d. Replace '%s' with '%s'\n", i+1, rule.Old, rule.New)
			}
			fmt.Println()
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
			if verbose {
				fmt.Printf("Creating backup of %s...\n", targetPath)
			}
			if err := createBackup(targetPath, info.IsDir()); err != nil {
				return pkgErrors.NewFileError(targetPath, "creating backup", err)
			}
			if verbose {
				fmt.Println("Backup created successfully")
			}
		}

		// Process based on target type
		if info.IsDir() {
			// Process directory
			if dryRun {
				return previewDirectoryReplacements(targetPath, rules, recursive)
			}
			
			results, err := replace.ReplaceInDirectoryWithResultsAndExclude(targetPath, rules, recursive, excludeGlob)
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
				fmt.Printf("Would process file: %s\n", targetPath)
				return nil
			}

			if verbose {
				fmt.Printf("Processing file: %s\n", targetPath)
			}
			
			count, err := replace.ReplaceInDocumentWithCount(targetPath, rules)
			if err != nil {
				if errors.Is(err, pkgErrors.ErrDocumentCorrupted) {
					return pkgErrors.NewDocumentError(targetPath, ext, "document appears to be corrupted", err)
				}
				return pkgErrors.NewDocumentError(targetPath, ext, "processing failed", err)
			}
			
			if verbose {
				fmt.Printf("Made %d replacements in %s\n", count, targetPath)
			}

			fmt.Printf("Successfully processed: %s\n", targetPath)
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
	fmt.Println("Files that would be processed:")
	
	count := 0
	// Use the new walk function with exclude support
	err := replace.WalkDocumentFilesWithExclude(dirPath, recursive, excludeGlob, func(path string) error {
		fmt.Printf("  - %s\n", path)
		count++
		return nil
	})
	
	if err != nil {
		return err
	}
	
	fmt.Printf("\nTotal files to process: %d\n", count)
	return nil
}

func printResults(results []replace.ReplaceResult) {
	successCount := 0
	failureCount := 0
	totalReplacements := 0
	
	fmt.Println("\nProcessing results:")
	fmt.Println("-------------------")
	
	for _, result := range results {
		if result.Success {
			fmt.Printf("✓ %s - Success (%d replacements)\n", result.FilePath, result.Replacements)
			successCount++
			totalReplacements += result.Replacements
		} else {
			fmt.Printf("✗ %s - Failed: %v\n", result.FilePath, result.Error)
			failureCount++
		}
	}
	
	fmt.Println("\nSummary:")
	fmt.Printf("  Successful: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failureCount)
	fmt.Printf("  Total files: %d\n", len(results))
	fmt.Printf("  Total replacements: %d\n", totalReplacements)
}

func init() {
	rootCmd.AddCommand(replaceCmd)

	replaceCmd.Flags().StringVarP(&rulesFile, "rules", "r", "", "YAML file containing replacement rules (required)")
	replaceCmd.Flags().StringVarP(&targetPath, "path", "p", "", "Target file or directory (required)")
	replaceCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	replaceCmd.Flags().BoolVar(&backup, "backup", false, "Create backup files before modification")
	replaceCmd.Flags().BoolVar(&recursive, "recursive", true, "Process subdirectories recursively")
	replaceCmd.Flags().StringVar(&excludeGlob, "exclude", "", "Glob pattern for files to exclude")

	replaceCmd.MarkFlagRequired("rules")
	replaceCmd.MarkFlagRequired("path")
}