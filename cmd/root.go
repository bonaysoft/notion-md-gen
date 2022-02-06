package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"notion-md-gen/pkg/config"
	"notion-md-gen/pkg/generator"

	"github.com/itzg/go-flagsfiller"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "notion-md-gen",
	Short: "A markdown generator for notion",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var config config.BlogConfig
		if err := viper.Unmarshal(&config); err != nil {
			log.Fatal(err)
		}

		if err := generator.Run(config); err != nil {
			log.Println(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is notion-md-gen.yaml)")

	// fill and map struct fields to flags
	var config config.BlogConfig
	filler := flagsfiller.New()
	if err := filler.Fill(flag.CommandLine, &config); err != nil {
		log.Fatal(err)
	}

	var envPrefix string
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		envPrefix = "input_"
	}

	rootCmd.Flags().AddGoFlagSet(flag.CommandLine)
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		key := strings.NewReplacer("-", "").Replace(f.Name)
		envKey := strings.NewReplacer("-", "_").Replace(f.Name)
		_ = viper.BindPFlag(key, f)                               // bind the flag to the config struct
		_ = viper.BindEnv(key, strings.ToUpper(envPrefix+envKey)) // bind the env to the config struct
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("notion-md-gen")
	}

	if err := godotenv.Load(); err == nil {
		fmt.Println("Load .env file")
	}
	// copy to the standard env variable
	// INPUT_DATABASE-ID => INPUT_DATABASE_ID
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		r := strings.NewReplacer("-", "_")
		_ = os.Setenv(r.Replace(key), os.Getenv(key))
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
