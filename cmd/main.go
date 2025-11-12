package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "padel-backend",
	Short: "Padel Backend Server",
	Long:  "Padel Backend API server",
}

// @title Padel Backend API
// @version 1.0
// @description API documentation for the Padel Backend service.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
