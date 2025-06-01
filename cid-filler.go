package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.design/x/clipboard"
)

const (
	AppVersion = "1.1"
	AppAuthor  = "Gusto F. Chatami (gustof@tuta.io | github.com/9vy)"
)

func appInfo() {
	fmt.Printf("📃 CID Filler version: v%s\n", AppVersion)
	fmt.Printf("🐦‍🔥 Author: %s\n", AppAuthor)
}

// config holds database configuration
type Config struct {
	DBPath       string
	TableName    string
	InputColumn  string
	OutputColumn string
}

func loadConfig() (*Config, error) {
	//load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("⚠️ WARNING: ERROR LOADING .ENV file: %v", err)
	}
	config := &Config{
		DBPath:       os.Getenv("DB_PATH"),
		TableName:    os.Getenv("TABLE_NAME"),
		InputColumn:  os.Getenv("INPUT_COLUMN"),
		OutputColumn: os.Getenv("OUTPUT_COLUMN"),
	}
	// validate all required configuration
	if config.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH is required in .env")
	}
	if config.TableName == "" {
		return nil, fmt.Errorf("TABLE_NAME is required in .env")
	}
	if config.InputColumn == "" {
		return nil, fmt.Errorf("INPUT_COLUMN is required in .env")
	}
	if config.OutputColumn == "" {
		return nil, fmt.Errorf("OUTPUT_COLUMN is required in .env")
	}
	return config, nil
}
func lookupCodes(inputCodes []string, config *Config) ([]string, error) {
	db, err := sql.Open("sqlite3", config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()
	//build dynamic query
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?",
		config.OutputColumn, config.TableName, config.InputColumn)
	var results []string
	for _, code := range inputCodes {
		code = strings.TrimSpace(code)
		var result string
		err := db.QueryRow(query, code).Scan(&result)
		if err != nil {
			if err != sql.ErrNoRows {
				results = append(results, "Not found")
			} else {
				return nil, fmt.Errorf("database query error: %v", err)
			}
		} else {
			results = append(results, result)
		}
	}
	return results, nil
}
func main() {
	appInfo()
	//load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}
	//initialize clipboard
	err = clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	fmt.Printf("♨️ Reading input codes from clipboard...\n")
	fmt.Printf("🗂️ Database: %s\n", config.DBPath)
	fmt.Printf("🔍 Looking up: %s.%s -> %s.%s\n",
		config.TableName, config.InputColumn, config.TableName, config.OutputColumn)
	//read from clipboard
	clipboardContent := string(clipboard.Read(clipboard.FmtText))
	if clipboardContent == "" {
		fmt.Println("❌ No content found in clipboard, recopy and try again")
		return
	}
	// split input codes by newline and filter empty strings
	lines := strings.Split(clipboardContent, "\n")
	var inputCodes []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			inputCodes = append(inputCodes, trimmed)
		}
	}
	if len(inputCodes) == 0 {
		fmt.Println("❌ No input codes found in clipboard. recopy and try again!")
		return
	}
	fmt.Printf("💡 Found %d input string.\n", len(inputCodes))
	// perform lookup
	results, err := lookupCodes(inputCodes, config)
	if err != nil {
		log.Printf("⚠️ CHECK YOUR CLIPBOARD, I THINK IT IS NOT 4 DIGIT SC: %v", err)
		return
	}
	if len(results) > 0 {
		fmt.Println("\nℹ️ Lookup Results:")
		result := strings.Join(results, "\n")
		fmt.Println(result)
		//copy result back to clipboard
		clipboard.Write(clipboard.FmtText, []byte(result))
		fmt.Println("\n✅ Results have been copied to clipboard.")
	} else {
		fmt.Println("⚠️No results found or an error occured")
	}
}
