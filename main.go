package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "bot",
	Short: "run the bot",
	RunE:  run,
}

func init() {
	rootCmd.Flags().StringP("token", "t", "", "Pushbullet API token (TOKEN)")
	viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))
	viper.BindEnv("token")

	rootCmd.Flags().StringP("sku-ids", "", "", "Comma separated list of SKUs to watch (SKU_IDS)")
	viper.BindPFlag("sku_ids", rootCmd.Flags().Lookup("sku-ids"))
	viper.BindEnv("sku_ids")

	rootCmd.Flags().StringP("collection-id", "", "", "ID of Collection to watch (COLLECTION_ID)")
	viper.BindPFlag("collection-id", rootCmd.Flags().Lookup("collection_id"))
	viper.BindEnv("collection_id")

	rootCmd.Flags().StringP("search-query", "", "", "Search result to watch (SEARCH_QUERY)")
	viper.BindPFlag("search_query", rootCmd.Flags().Lookup("search-query"))
	viper.BindEnv("search_query")
}

func run(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("pushbullet api token missing")
	}

	masterSkus := make(map[string]struct{})
	if skuIds := viper.GetString("sku_ids"); skuIds != "" {
		individualSkus := strings.Split(skuIds, ",")
		fmt.Printf("tracking %v SKUs explicitly\n", len(individualSkus))

		for _, s := range individualSkus {
			masterSkus[s] = struct{}{}
		}
	}

	if collectionId := viper.GetString("collection_id"); collectionId != "" {
		collectionSkus, err := getSkusFromCollection(collectionId)
		if err != nil {
			return fmt.Errorf("getting skus from collection: %v", err)
		}

		fmt.Printf("tracking %v SKUs via collection ID\n", len(collectionSkus))
		for _, s := range collectionSkus {
			masterSkus[s] = struct{}{}
		}
	}

	if searchQuery := viper.GetString("search_query"); searchQuery != "" {
		searchSkus, err := getSkusFromSearch(searchQuery)
		if err != nil {
			return fmt.Errorf("getting skus from collection: %v", err)
		}

		fmt.Printf("tracking %v SKUs via search query ID\n", len(searchSkus))
		for _, s := range searchSkus {
			masterSkus[s] = struct{}{}
		}
	}

	skus := make([]string, 0)
	for k := range masterSkus {
		skus = append(skus, k)
	}
	fmt.Printf("tracking %v unique SKUs\n", len(skus))

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			fmt.Println("checking availability")
			available, err := getAvailableSkus(skus)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("resolving product")
			for i, s := range available {
				fmt.Printf("%v/%v\n", i, len(available))
				product, err := getProductFromSku(s)
				if err != nil {
					fmt.Println(err)
					continue
				}

				err = notify(product, token)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
