package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	NotesDirs []string `mapstructure:"notes_dirs"` // Changed from NotesDir to NotesDirs
	Theme     Theme    `mapstructure:"theme"`
}

// Theme represents the color theme configuration
type Theme struct {
	Primary   string `mapstructure:"primary"`
	Secondary string `mapstructure:"secondary"`
	Success   string `mapstructure:"success"`
	Warning   string `mapstructure:"warning"`
	Error     string `mapstructure:"error"`
	Info      string `mapstructure:"info"`
	Muted     string `mapstructure:"muted"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	notesDir := filepath.Join(homeDir, "notes")

	return &Config{
		NotesDirs: []string{notesDir},
		Theme: Theme{
			Primary:   "#88C0D0", // Nord Blue
			Secondary: "#4C566A", // Nord Gray
			Success:   "#A3BE8C", // Nord Green
			Warning:   "#EBCB8B", // Nord Yellow
			Error:     "#BF616A", // Nord Red
			Info:      "#81A1C1", // Nord Light Blue
			Muted:     "#5E81AC", // Nord Dark Blue
		},
	}
}

// expandTilde expands ~ to the user's home directory
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path // Return original path if we can't get home dir
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// LoadConfig loads configuration from file or creates default
func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	viper.SetConfigFile(configPath) // Use SetConfigFile instead of SetConfigName/AddConfigPath

	// Set defaults
	defaultConfig := DefaultConfig()
	viper.SetDefault("notes_dirs", defaultConfig.NotesDirs)
	viper.SetDefault("theme.primary", defaultConfig.Theme.Primary)
	viper.SetDefault("theme.secondary", defaultConfig.Theme.Secondary)
	viper.SetDefault("theme.success", defaultConfig.Theme.Success)
	viper.SetDefault("theme.warning", defaultConfig.Theme.Warning)
	viper.SetDefault("theme.error", defaultConfig.Theme.Error)
	viper.SetDefault("theme.info", defaultConfig.Theme.Info)
	viper.SetDefault("theme.muted", defaultConfig.Theme.Muted)

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, prompt user for notes directory
			return promptForNotesDirectory(configPath, defaultConfig)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand tilde in notes_dir if present
	for i, dir := range config.NotesDirs {
		config.NotesDirs[i] = expandTilde(dir)
	}

	return &config, nil
}

// promptForNotesDirectory prompts the user to select notes directories
func promptForNotesDirectory(configPath string, defaultConfig *Config) (*Config, error) {
	fmt.Println("Welcome to Burh! This appears to be your first time running the program.")
	fmt.Println("Please enter the path to the directory where you'd like to store your notes.")
	fmt.Printf("Default location: %s\n", defaultConfig.NotesDirs[0])
	fmt.Println()

	// Initialize with empty slice
	var selectedDirs []string

	// Get the first (mandatory) directory
	firstDir, err := getDirectoryFromUser("Enter the path to your primary notes directory", defaultConfig.NotesDirs[0])
	if err != nil {
		return nil, err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(firstDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %w", err)
	}

	selectedDirs = append(selectedDirs, firstDir)
	fmt.Printf("Primary notes directory set to: %s\n", firstDir)

	// Ask if user wants to add more directories
	for {
		fmt.Print("\nWould you like to add another notes directory? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			break
		}

		additionalDir, err := getDirectoryFromUser("Enter the path to an additional notes directory", "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Check if directory already exists in the list
		alreadyExists := false
		for _, dir := range selectedDirs {
			if dir == additionalDir {
				alreadyExists = true
				break
			}
		}

		if alreadyExists {
			fmt.Println("This directory is already in your list.")
			continue
		}

		// Ensure the directory exists
		if err := os.MkdirAll(additionalDir, 0755); err != nil {
			fmt.Printf("Failed to create directory: %v\n", err)
			continue
		}

		selectedDirs = append(selectedDirs, additionalDir)
		fmt.Printf("Additional notes directory added: %s\n", additionalDir)
	}

	// Update config with selected directories
	defaultConfig.NotesDirs = selectedDirs

	fmt.Printf("\nNotes will be loaded from %d directory(ies):\n", len(selectedDirs))
	for i, dir := range selectedDirs {
		fmt.Printf("  %d. %s\n", i+1, dir)
	}
	fmt.Println()

	// Create the config file
	return createDefaultConfig(configPath, defaultConfig)
}

// getDirectoryFromUser prompts the user for a directory path
func getDirectoryFromUser(prompt, defaultPath string) (string, error) {
	fmt.Printf("%s\n", prompt)
	if defaultPath != "" {
		fmt.Printf("Default: %s\n", defaultPath)
	}
	fmt.Println("You can:")
	fmt.Println("1. Press Enter to use the default location")
	fmt.Println("2. Type a custom path")
	fmt.Println("3. Type 'browse' to open file explorer (if supported)")
	fmt.Print("Enter your choice: ")

	var selectedDir string
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// If we can't read input, use default
		if defaultPath != "" {
			fmt.Printf("Using default location: %s\n", defaultPath)
			return defaultPath, nil
		}
		return "", fmt.Errorf("failed to read input")
	}

	input = strings.TrimSpace(input)

	switch {
	case input == "":
		if defaultPath != "" {
			selectedDir = defaultPath
		} else {
			return "", fmt.Errorf("no default path provided and no input given")
		}
	case strings.ToLower(input) == "browse":
		// Try to open file explorer
		selectedDir = openFileExplorer(defaultPath)
		if selectedDir == "" {
			if defaultPath != "" {
				selectedDir = defaultPath
			} else {
				return "", fmt.Errorf("no directory selected")
			}
		}
	default:
		selectedDir = input
	}

	return selectedDir, nil
}

// openFileExplorer attempts to open the system file explorer
func openFileExplorer(defaultPath string) string {
	// This is a simplified approach - in a real implementation,
	// you might want to use platform-specific commands or libraries
	fmt.Println("Opening file explorer...")
	fmt.Println("Please navigate to your desired notes directory and copy the path.")
	fmt.Println("Then paste it here (or press Enter to use default):")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultPath
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultPath
	}

	return input
}

// SaveConfig saves the current configuration to file
func SaveConfig(config *Config) error {
	configPath := getConfigPath()

	// Save the expanded path (without tilde) to avoid confusion
	viper.Set("notes_dirs", config.NotesDirs)
	viper.Set("theme.primary", config.Theme.Primary)
	viper.Set("theme.secondary", config.Theme.Secondary)
	viper.Set("theme.success", config.Theme.Success)
	viper.Set("theme.warning", config.Theme.Warning)
	viper.Set("theme.error", config.Theme.Error)
	viper.Set("theme.info", config.Theme.Info)
	viper.Set("theme.muted", config.Theme.Muted)

	return viper.WriteConfigAs(configPath)
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".burhrc.yaml")
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configPath string, config *Config) (*Config, error) {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save default config
	if err := SaveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save default config: %w", err)
	}

	return config, nil
}

// ValidateAndReloadConfig validates the current configuration and reloads it
func ValidateAndReloadConfig() (*Config, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Validate that all directories exist or can be created
	for i, dir := range config.NotesDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create or access directory %s: %w", dir, err)
		}
		// Update with absolute path
		absPath, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", dir, err)
		}
		config.NotesDirs[i] = absPath
	}

	// Save the validated config back to file
	if err := SaveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save validated config: %w", err)
	}

	return config, nil
}

// AddNotesDirectory adds a new directory to the configuration
func AddNotesDirectory(newDir string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// Expand tilde if present
	newDir = expandTilde(newDir)

	// Check if directory already exists in the list
	for _, dir := range config.NotesDirs {
		if dir == newDir {
			return fmt.Errorf("directory %s is already in the configuration", newDir)
		}
	}

	// Ensure the directory exists
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", newDir, err)
	}

	// Add to configuration
	config.NotesDirs = append(config.NotesDirs, newDir)

	// Save updated configuration
	return SaveConfig(config)
}

// RemoveNotesDirectory removes a directory from the configuration
func RemoveNotesDirectory(dirToRemove string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// Expand tilde if present
	dirToRemove = expandTilde(dirToRemove)

	// Find and remove the directory
	found := false
	var newDirs []string
	for _, dir := range config.NotesDirs {
		if dir == dirToRemove {
			found = true
			continue
		}
		newDirs = append(newDirs, dir)
	}

	if !found {
		return fmt.Errorf("directory %s not found in configuration", dirToRemove)
	}

	// Ensure we don't remove all directories
	if len(newDirs) == 0 {
		return fmt.Errorf("cannot remove all directories - at least one must remain")
	}

	config.NotesDirs = newDirs

	// Save updated configuration
	return SaveConfig(config)
}
