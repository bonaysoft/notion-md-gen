package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/itzg/go-flagsfiller"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"notion-md-gen/internal"
	notion_blog "notion-md-gen/pkg"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "notion-md-gen",
	Short: "A markdown generator for notion",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var config notion_blog.BlogConfig
		if err := viper.Unmarshal(&config); err != nil {
			log.Fatal(err)
		}

		if err := internal.ParseAndGenerate(config); err != nil {
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
	var config notion_blog.BlogConfig
	filler := flagsfiller.New()
	if err := filler.Fill(flag.CommandLine, &config); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().AddGoFlagSet(flag.CommandLine)
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		// keep same name for the name of config and flag, the flag will overwrite config.
		_ = viper.BindPFlag(strings.Replace(f.Name, "-", "", -1), f)
	})
	// bind the env DATABASE_ID to the databaseId of config struct
	_ = viper.BindEnv("databaseId", "DATABASE_ID")
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
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
