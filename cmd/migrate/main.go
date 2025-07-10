// This is like the main class in Java - it's our entry point
package main

// Import statements are like Java imports - we bring in code from other packages
import (
	"fmt"           // For printing messages (like System.out.println in Java)
	"log"           // For logging errors (like Logger in Java)
	"os"            // For reading environment variables and command line arguments
	"path/filepath" // For working with file paths safely

	// This is the main migration library - like importing a JAR file
	"github.com/golang-migrate/migrate/v4"
	// The underscore (_) means "import this but don't use it directly"
	// It's like including a JAR that registers itself automatically
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // PostgreSQL driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File reading driver
)

// main() is like public static void main(String[] args) in Java
func main() {
	// os.Args is like String[] args in Java - it holds command line arguments
	// os.Args[0] is the program name, os.Args[1] is the first argument
	// We check if user gave us at least one argument (like "up" or "down")
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate <up|down|force|version>")
	}

	// Get the first argument (the command the user wants to run)
	// This is like args[0] in Java
	command := os.Args[1]

	// os.Getenv is like System.getenv() in Java - reads environment variables
	// We're looking for DATABASE_URL (like a connection string)
	databaseURL := os.Getenv("DATABASE_URL")
	// If no environment variable is set, use a default value
	// This is like: String url = System.getenv("DATABASE_URL") != null ? System.getenv("DATABASE_URL") : "default"
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/enterprise_crud?sslmode=disable"
	}

	// Set up the path to our migration files
	// "file://" is like saying "look in the file system" (not a website)
	migrationsPath := "file://migrations"
	// If user provided a custom path as second argument, use that instead
	if len(os.Args) > 2 {
		migrationsPath = fmt.Sprintf("file://%s", os.Args[2])
	}

	// Get the current working directory (like System.getProperty("user.dir") in Java)
	// In Go, functions can return multiple values! Here we get the directory AND an error
	wd, err := os.Getwd()
	// In Go, we always check for errors explicitly (unlike Java's try/catch)
	if err != nil {
		log.Fatal(err) // This is like throwing a RuntimeException in Java
	}

	// Build the full path to our migrations folder
	// filepath.Join is like Paths.get() in Java - it joins paths safely
	migrationsPath = fmt.Sprintf("file://%s", filepath.Join(wd, "migrations"))

	// Create a new migration instance - like creating a new object in Java
	// Again, Go functions can return multiple values (the migrator AND an error)
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	// defer is special in Go - it means "run this when the function ends"
	// It's like having a finally block that always runs
	defer m.Close()

	// switch is like switch/case in Java, but cleaner
	switch command {
	case "up":
		// Try to run migrations "up" (apply new changes)
		// migrate.ErrNoChange means "nothing to do" - that's OK, not a real error
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		// Try to run migrations "down" (undo changes)
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migrations rolled back successfully")
	case "force":
		// Force the migration to a specific version (dangerous!)
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		version := os.Args[2]
		// Convert string to int using our helper function
		if err := m.Force(parseInt(version)); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Forced migration to version %s\n", version)
	case "version":
		// Check what version we're currently at
		// This function returns THREE values: version, dirty flag, and error
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal(err)
		}
		// %d is for integers, %t is for booleans (true/false)
		fmt.Printf("Version: %d, Dirty: %t\n", version, dirty)
	default:
		// If user typed something we don't understand
		log.Fatal("Unknown command. Use: up, down, force, or version")
	}
}

// This is a helper function - like a private static method in Java
// It converts a string to an integer manually (there's a built-in way too)
func parseInt(s string) int {
	var result int // In Go, this starts at 0 automatically
	// range is like for-each in Java - it loops through each character
	// The underscore _ means "ignore the index, I just want the character"
	for _, r := range s {
		// Check if character is a digit (0-9)
		if r < '0' || r > '9' {
			log.Fatalf("Invalid version number: %s", s)
		}
		// Convert character to number and build the result
		// This is like: result = result * 10 + (character - '0')
		result = result*10 + int(r-'0')
	}
	return result
}
