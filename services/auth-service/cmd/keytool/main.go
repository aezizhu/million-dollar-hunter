package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

func main() {
	var (
		keystorePath = flag.String("keystore", "./config/keystore.json", "Path to keystore file")
		generateCmd  = flag.Bool("generate", false, "Generate a new key")
		listCmd      = flag.Bool("list", false, "List all keys")
		activateCmd  = flag.String("activate", "", "Activate key by ID")
		cleanupCmd   = flag.Bool("cleanup", false, "Clean up expired keys")
		bitSize      = flag.Int("bits", 2048, "RSA key size (2048 or 4096)")
		expiresIn    = flag.Duration("expires", 90*24*time.Hour, "Key expiration duration")
		active       = flag.Bool("active", false, "Make new key active immediately")
	)

	flag.Parse()

	ks, err := jwtmgr.NewKeyStore(*keystorePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading keystore: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *generateCmd:
		kid, err := ks.GenerateKey(*bitSize, *active, *expiresIn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated key: %s\n", kid)
		if *active {
			fmt.Println("Key is now active")
		}

	case *listCmd:
		keys := ks.ListKeys()
		if len(keys) == 0 {
			fmt.Println("No keys found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KID\tALGORITHM\tACTIVE\tCREATED\tEXPIRES")
		for _, key := range keys {
			activeStr := ""
			if key.Active {
				activeStr = "✓"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				key.ID,
				key.Algorithm,
				activeStr,
				key.CreatedAt.Format("2006-01-02"),
				key.ExpiresAt.Format("2006-01-02"),
			)
		}
		w.Flush()

	case *activateCmd != "":
		if err := ks.ActivateKey(*activateCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error activating key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Activated key: %s\n", *activateCmd)

	case *cleanupCmd:
		removed, err := ks.CleanupExpiredKeys()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning up keys: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed %d expired keys\n", removed)

	default:
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
