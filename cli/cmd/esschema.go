package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	parser "github.com/stephane-martin/w3c-extendedlog-parser"
)

var fieldsLine string
var filename string

var shards uint
var replicas uint
var check bool
var refreshInterval int

var esschemaCmd = &cobra.Command{
	Use:   "esschema",
	Short: "Prints an Elasticsearch mapping that can store access logs",
	Run: func(cmd *cobra.Command, args []string) {
		excludes := make(map[string]bool)
		for _, fName := range excludedFields {
			excludes[strings.ToLower(fName)] = true
		}
		excludes["time"] = true
		excludes["date"] = true

		var fieldsNames []string
		fieldsLine = strings.TrimSpace(fieldsLine)
		filename = strings.TrimSpace(filename)
		if len(fieldsLine) == 0 && len(filename) == 0 {
			fmt.Fprintln(os.Stderr, "Please specify fields")
			os.Exit(-1)
		}
		if len(fieldsLine) != 0 && len(filename) != 0 {
			fmt.Fprintln(os.Stderr, "--fields and --filename options are exclusive")
			os.Exit(-1)
		}
		if len(fieldsLine) > 0 {
			fieldsNames = strings.Split(fieldsLine, " ")
		} else {
			f, err := os.Open(filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			defer f.Close()
			p := parser.NewFileParser(f)
			err = p.ParseHeader()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			fieldsNames = p.FieldNames()
		}
		if len(fieldsNames) == 0 {
			fmt.Fprintln(os.Stderr, "field names not found")
			os.Exit(-1)
		}
		opts := newEsOpts(shards, replicas, check, time.Duration(refreshInterval)*time.Second, fieldsNames, excludes)
		b, err := json.MarshalIndent(opts, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(esschemaCmd)
	esschemaCmd.Flags().StringVar(&fieldsLine, "fields", "", "specify the fields that will be present in the access logs")
	esschemaCmd.Flags().StringVar(&filename, "filename", "", "specify the log file from which to extract the fields")
	esschemaCmd.Flags().UintVar(&shards, "shards", 1, "number of shards for the index")
	esschemaCmd.Flags().UintVar(&replicas, "replicas", 0, "number of replicas for the index")
	esschemaCmd.Flags().BoolVar(&check, "check", false, "whether to check the index on startup")
	esschemaCmd.Flags().IntVar(&refreshInterval, "refresh", 1, "refresh interval in seconds")
	esschemaCmd.Flags().StringArrayVar(&excludedFields, "exclude", []string{}, "exclude that field from collection (can be repeated)")
}
