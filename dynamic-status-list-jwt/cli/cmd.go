package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (s *Server) Run() {
	// CMD variables
	var (
		out             string
		in              string
		detached        bool
		jti             string
		revoked         bool
		timestamp       int64
		statusListPath  string
		holderProofPath string
	)

	rootCmd := &cobra.Command{
		Use:   "dsl",
		Short: "CLI tool for managing dSL revocation",
		Long: `A command-line tool to issue, print, and revoke 
verifiable credentials using JSON Web Tokens (JWT).`,
	}

	// Issue a mock JWT and store it to a file
	// Default filename: mock-jwt.json
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a mock JWT",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("> Issuing a mock JWT")
			s.IssueJWT(out)
			fmt.Printf("> Mock JWT issued and stored to %s\n", out)
		},
	}
	issueCmd.Flags().StringVarP(&out, "out", "o", "mock-jwt.json", "Path to the output file")

	// Create new status list entry
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new Status List entry",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("> Creating a new status list entry for JWT: %s\n", in)
			err := s.NewDslEntry(in, detached)
			if err != nil {
				fmt.Println("[ERROR]", err)
				return
			}
			fmt.Println("> New status list entry created and stored in dsl.json. JWT jti entries are in dsl-map.json")
		},
	}
	newCmd.Flags().StringVarP(&in, "in", "i", "", "Path to the JWT that will be added to the dSL")
	newCmd.MarkFlagRequired("in")
	newCmd.Flags().BoolVarP(&detached, "detached", "d", false, "Create a detached revocation metadata JWT.")

	// New revocation metadata (proof) command
	proofCmd := &cobra.Command{
		Use:   "wallet",
		Short: "Derive status list identifier (holder/wallet)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("> Deriving status list identifier")
			identifier, err := NewProof(in, revoked, detached, timestamp)
			if err != nil {
				fmt.Println("[ERROR]", err)
				return
			}
			fmt.Println("> Status list identifier:")
			fmt.Println(*identifier)
		},
	}
	proofCmd.Flags().StringVarP(&in, "in", "i", "", "Path to the file to derive the identifier from")
	proofCmd.MarkFlagRequired("in")
	proofCmd.Flags().BoolVarP(&revoked, "revoked", "r", false, "Create a proof for a revoked credential")
	proofCmd.Flags().BoolVarP(&detached, "detached", "d", false, "Create a detached revocation token")
	proofCmd.Flags().Int64VarP(&timestamp, "timestamp", "t", 0, "Unix timestamp when the holder computes the identifier")

	// Recompute DSL command
	recomputeCmd := &cobra.Command{
		Use:   "recompute",
		Short: "Recompute the DSL",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("> Recomputing the DSL")
			var err error
			if timestamp == 0 {
				err = s.RecomputeDslJwt()
			} else {
				err = s.RecomputeDslJwtAt(timestamp)
			}
			if err != nil {
				fmt.Println("[ERROR]", err)
				return
			}
			fmt.Println("> DSL recomputed and stored in dsl.json")
		},
	}
	recomputeCmd.Flags().Int64VarP(&timestamp, "timestamp", "t", 0, "Unix timestamp when the holder computes the identifier")

	// Revoke JWT command
	revokeCmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a JWT",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("> Revoking JWT with jti: %s\n", jti)
			err := s.Revoke(jti)
			if err != nil {
				fmt.Println("[ERROR]", err)
				return
			}
			fmt.Println("> JWT successfully revoked. DSL stored in dsl.json")
		},
	}
	revokeCmd.Flags().StringVarP(&jti, "jti", "j", "", "JTI of the JWT to revoke")

	// Verify proof command
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the holder's proof",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("> Verifying proof: %s\n", holderProofPath)
			revoked, err := Verify(statusListPath, holderProofPath)
			if err != nil {
				fmt.Println("[ERROR]", err)
				return
			}
			fmt.Printf("> Proof successfully verified. Revoked: %t\n", revoked)
		},
	}
	verifyCmd.Flags().StringVarP(&statusListPath, "status-list", "s", "dsl.json", "Path to the status list")
	verifyCmd.Flags().StringVarP(&holderProofPath, "holder-proof", "p", "holder_status-list-identifier.json", "Path to the holder's proof")
	verifyCmd.Flags().StringVarP(&jti, "jti", "j", "", "JTI of the JWT to verify")

	// Print JSON information
	printCmd := &cobra.Command{
		Use:   "print",
		Short: "Print information from a JSON file",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("> Printing information from file: %s\n", in)
			Print(in)
		},
	}
	printCmd.Flags().StringVarP(&in, "in", "i", "", "Path to the JSON file")
	printCmd.MarkFlagRequired("in")

	// Print JWT information
	printJwtCmd := &cobra.Command{
		Use:   "printjwt",
		Short: "Decode and print JWT",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("> Printing JWT from file: %s\n", in)
			PrintJWT(in)
		},
	}
	printJwtCmd.Flags().StringVarP(&in, "in", "i", "", "Path to the JSON file containing JWT")
	printJwtCmd.MarkFlagRequired("in")

	// Add all subcommands to the root
	rootCmd.AddCommand(issueCmd, newCmd, proofCmd, recomputeCmd, revokeCmd, printCmd, printJwtCmd, verifyCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
