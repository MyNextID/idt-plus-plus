package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Run() {
	// CMD variables
	var (
		keyPath   string
		crtPath   string
		benchmark bool
		bsl       string
		bslPath   string
		err       error
		compress  bool
	)

	rootCmd := &cobra.Command{
		Use:   "ccrl",
		Short: "CLI tool for managing compact CRL revocations",
		Long:  `A command-line tool to test different CRL profiles`,
	}

	// Benchmarks
	benchmarkCmd := &cobra.Command{
		Use:   "bench",
		Short: "Benchmark CRL with Bit Status List extension",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("> Starting the benchmark")
			// Init CRL manager
			cm := NewCRLManager(&keyPath, &crtPath)
			// Run the Bit String Extension benchmark
			benchmarkCRLBitString(cm)
			fmt.Println("> Benchmark finished")
		},
	}
	benchmarkCmd.Flags().BoolVarP(&benchmark, "bit-string-list", "b", false, "Run the Bit String CRL benchmark")
	benchmarkCmd.Flags().StringVarP(&crtPath, "crt", "c", "certs/rootCA.crt", "Path to the CRL signing certificate")
	benchmarkCmd.Flags().StringVarP(&keyPath, "key", "k", "certs/rootCA.key", "Path to the CRL signing key")

	// CRL Bit String from hex encoded byte array
	// Compressed or uncompressed bit string status list to CRL bit string list
	bslCrlCmd := &cobra.Command{
		Use:   "bsl",
		Short: "Create a CRL with bit-string CRL extension for the provided bit string list",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("> Creating a CRL with bit-string extension")
			// Init CRL manager
			cm := NewCRLManager(&keyPath, &crtPath)
			// Run the Bit String Extension benchmark
			var bitStringBytes []byte
			if bsl != "" {
				bitStringBytes, err = hex.DecodeString(bsl)
				if err != nil {
					fmt.Println("[ERROR]", err)
					return
				}
			} else if bslPath != "" {
				bitStringBytes, err = ReadHexFile(bslPath)
			} else {
				fmt.Println("[ERROR] Missing params bit-string-bytes or bit-string-path", err)
				return
			}
			capacityByte := len(bitStringBytes)
			var compressionRatio float64
			if compress {
				sizeUncompressed := len(bitStringBytes)
				bitStringBytes, err = Compress(bitStringBytes)
				if err != nil {
					fmt.Println("[ERROR]", err)
					return
				}
				// Compress
				sizeCompressed := len(bitStringBytes)
				compressionRatio = float64(sizeUncompressed) / float64(sizeCompressed)
			}
			// Create CRL
			crl, err := cm.NewCRLBitStringExtension(bitStringBytes)
			if err != nil {
				fmt.Println("[ERROR]", err)
				return
			}
			fmt.Println("[r] CRL")
			fmt.Println(string(crl))
			fmt.Printf("[.] PEM CRL size %d bytes\n[r] Bit string capacity: %d\n", len(crl), capacityByte*8)

			if compress {
				fmt.Printf("[c] Compression Ratio: %.2f\n", compressionRatio)
			}
			fmt.Println("> Benchmark finished")
		},
	}
	bslCrlCmd.Flags().StringVarP(&bsl, "bit-string-bytes", "b", "", "Hex-encoded bit string list encoded as bytes")
	bslCrlCmd.Flags().StringVarP(&bslPath, "bit-string-path", "p", "", "Path to a file that contains hex-encoded bit string list encoded as bytes")
	bslCrlCmd.Flags().BoolVarP(&compress, "compress", "z", false, "Compress the bytes using zlib")
	bslCrlCmd.Flags().StringVarP(&crtPath, "crt", "c", "certs/rootCA.crt", "Path to the CRL signing certificate")
	bslCrlCmd.Flags().StringVarP(&keyPath, "key", "k", "certs/rootCA.key", "Path to the CRL signing key")

	// Add all subcommands to the root
	rootCmd.AddCommand(benchmarkCmd, bslCrlCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
