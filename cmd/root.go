package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fortix/renamefiles/internal/build"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const CONFIG_FILE_NAME = "renamefiles"
const CONFIG_FILE_TYPE = "yaml"
const CONFIG_ENV_PREFIX = "DNS"

var rootCmd = &cobra.Command{
	Use:     "renamefiles",
	Short:   "Rename files based on a CSV.",
	Long:    `This command will take a CSV and rename the files in column 0 to the file name in column 1.`,
	Version: build.Version,
	Args:    cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("csv", cmd.Flags().Lookup("csv"))
		viper.BindEnv("csv", CONFIG_ENV_PREFIX+"_CSV")

		viper.BindPFlag("src-base", cmd.Flags().Lookup("src-base"))
		viper.BindEnv("src-base", CONFIG_ENV_PREFIX+"_SRC_BASE")

		viper.BindPFlag("dst-base", cmd.Flags().Lookup("dst-base"))
		viper.BindEnv("dst-base", CONFIG_ENV_PREFIX+"_DST_BASE")

		viper.BindPFlag("copy-only", cmd.Flags().Lookup("copy-only"))
		viper.BindEnv("copy-only", CONFIG_ENV_PREFIX+"_COPY_ONLY")

	},
	Run: renameCmd,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is "+CONFIG_FILE_NAME+"."+CONFIG_FILE_TYPE+" in the current directory or $HOME/).\nOverrides the "+CONFIG_ENV_PREFIX+"_CONFIG environment variable if set.")
	rootCmd.PersistentFlags().StringP("log-level", "", "info", "Log level (debug, info, warn, error, fatal, panic).\nOverrides the "+CONFIG_ENV_PREFIX+"_LOGLEVEL environment variable if set.")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
		viper.BindEnv("config", CONFIG_ENV_PREFIX+"_CONFIG")
		viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
		viper.BindEnv("log.level", CONFIG_ENV_PREFIX+"_LOGLEVEL")

		// If config file given then use it
		cfgFile := viper.GetString("config")
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
			if err := viper.ReadInConfig(); err != nil {
				log.Fatal().Msgf("missing config file: %s", viper.ConfigFileUsed())
			}
		}

		switch viper.GetString("log.level") {
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		default:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		}
	}

	rootCmd.Flags().StringP("csv", "", "", "The CSV file to use for renaming.\nOverrides the "+CONFIG_ENV_PREFIX+"_CSV environment variable if set.")
	rootCmd.Flags().StringP("src-base", "", "", "Optional base path for source files.\nOverrides the "+CONFIG_ENV_PREFIX+"_SRC_BASE environment variable if set.")
	rootCmd.Flags().StringP("dst-base", "", "", "Optional base path for destination files.\nOverrides the "+CONFIG_ENV_PREFIX+"_DST_BASE environment variable if set.")
	rootCmd.Flags().BoolP("copy-only", "", false, "Optional flag to force copy only instead of moving files. \nOverrides the "+CONFIG_ENV_PREFIX+"_COPY_ONLY environment variable if set.")
}

func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Set search paths for config file
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.SetConfigName(CONFIG_FILE_NAME) // Name of config file without extension
	viper.SetConfigType(CONFIG_FILE_TYPE) // Type of config file
	viper.SetEnvPrefix(CONFIG_ENV_PREFIX)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.ReadInConfig()
}

func renameCmd(cmd *cobra.Command, args []string) {
	// Get the source and destination base paths
	srcBase := strings.TrimRight(viper.GetString("src-base"), "/")
	dstBase := strings.TrimRight(viper.GetString("dst-base"), "/")
	copyOnly := viper.GetBool("copy-only")

	if srcBase != "" {
		srcBase += "/"
	}
	if dstBase != "" {
		dstBase += "/"
	}

	// Get the list of CSV files using a wildcard
	csvFiles, err := filepath.Glob(viper.GetString("csv"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to find CSV files")
	}

	if len(csvFiles) == 0 {
		log.Fatal().Msg("No CSV files found")
	}

	for _, csvFile := range csvFiles {
		// Open the CSV file and read each line outputting it to the log
		log.Info().Msgf("Reading CSV file: %s", csvFile)
		file, err := os.Open(csvFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open CSV file")
		}

		reader := csv.NewReader(file)

		// an array to collect all errors
		var errors []string
		var line int = 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal().Err(err).Msg("Failed to read CSV file")
				errors = append(errors, err.Error())
			}
			line++
			if len(record) < 2 {
				log.Warn().Msgf("Line %d: CSV row does not have at least two columns", line)
				errors = append(errors, "CSV row does not have at least two columns")
			} else {
				// Make the source and destination paths
				from := srcBase + record[0]
				to := dstBase + record[1]

				if copyOnly {
					message := fmt.Sprintf("Line %d: Copying %s to %s", line, from, to)
					log.Info().Msgf(message)
				} else {
					message := fmt.Sprintf("Line %d: Renaming %s to %s", line, from, to)
					log.Info().Msg(message)
				}

				// move or copy the file
				if copyOnly {
					input, err := os.ReadFile(from)
					if err != nil {
						message := fmt.Sprintf("Line %d: Failed to read %s", line, from)
						log.Error().Err(err).Msg(message)
						errors = append(errors, message)
					} else {
						err = os.WriteFile(to, input, 0644)
						if err != nil {
							message := fmt.Sprintf("Line %d: Failed to write %s", line, to)
							log.Error().Err(err).Msg(message)
							errors = append(errors, message)
						}
					}
				} else {
					err = os.Rename(from, to)
					if err != nil {
						message := fmt.Sprintf("Line %d: Failed to rename/copy %s to %s", line, from, to)
						log.Error().Err(err).Msg(message)
						errors = append(errors, message)
					}
				}
			}
		}
		file.Close()

		// If there were no errors then rename the CSV file
		if len(errors) == 0 {
			// Rename the original file to have a .processed extension
			err = os.Rename(csvFile, csvFile+".processed")
			if err != nil {
				log.Error().Err(err).Msgf("Failed to rename %s to %s.processed", csvFile, csvFile)
			}
		} else {
			// dump all errors to a file
			err = os.WriteFile(csvFile+".errors", []byte(strings.Join(errors, "\n")), 0644)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to write %s.errors", csvFile)
			}
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
