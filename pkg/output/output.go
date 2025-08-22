package output

import "fmt"

// OutputOptions controls the verbosity of output messages
type OutputOptions struct {
	Verbose bool
}

// PrintInfo prints informational messages only in verbose mode
func (o *OutputOptions) PrintInfo(msg string) {
	if o.Verbose {
		fmt.Printf("INFO: %s\n", msg)
	}
}

// PrintStep prints step messages only in verbose mode
func (o *OutputOptions) PrintStep(msg string) {
	if o.Verbose {
		fmt.Printf("STEP: %s\n", msg)
	}
}

// PrintDebug prints debug messages only in verbose mode
func (o *OutputOptions) PrintDebug(msg string) {
	if o.Verbose {
		fmt.Printf("DEBUG: %s\n", msg)
	}
}

// PrintValid prints validation success messages (always visible)
func (o *OutputOptions) PrintValid(file, details string) {
	fmt.Printf("VALID: %s - %s\n", file, details)
}

// PrintWarning prints warning messages (always visible)
func (o *OutputOptions) PrintWarning(file, details string) {
	fmt.Printf("WARNING: %s - %s\n", file, details)
}

// PrintError prints error messages (always visible)
func (o *OutputOptions) PrintError(file, details string) {
	fmt.Printf("ERROR: %s - %s\n", file, details)
}

// PrintProcessed prints file processing messages (always visible)
func (o *OutputOptions) PrintProcessed(source, target, details string) {
	if details != "" {
		fmt.Printf("PROCESSED: %s -> %s (%s)\n", source, target, details)
	} else {
		fmt.Printf("PROCESSED: %s -> %s\n", source, target)
	}
}

// PrintSummary prints summary messages (always visible)
func (o *OutputOptions) PrintSummary(msg string) {
	fmt.Printf("SUMMARY: %s\n", msg)
}

// PrintResult prints intermediate result messages (always visible)
func (o *OutputOptions) PrintResult(msg string) {
	fmt.Printf("RESULT: %s\n", msg)
}

// NewOutputOptions creates a new OutputOptions instance
func NewOutputOptions(verbose bool) *OutputOptions {
	return &OutputOptions{
		Verbose: verbose,
	}
}
