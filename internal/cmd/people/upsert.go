package people

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/graphql"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

var (
	upsertFirstName string
	upsertLastName  string
	upsertEmail     string
	upsertPhone     string
	upsertJobTitle  string
	upsertCompanyID string
	upsertData      string
)

var upsertCmd = &cobra.Command{
	Use:   "upsert",
	Short: "Create or update a person (by email)",
	Long:  "Create a person if they don't exist, or update if they do. Matching is done by email.",
	RunE:  runUpsert,
}

func init() {
	upsertCmd.Flags().StringVar(&upsertFirstName, "first-name", "", "first name")
	upsertCmd.Flags().StringVar(&upsertLastName, "last-name", "", "last name")
	upsertCmd.Flags().StringVar(&upsertEmail, "email", "", "primary email (used for matching)")
	upsertCmd.Flags().StringVar(&upsertPhone, "phone", "", "primary phone")
	upsertCmd.Flags().StringVar(&upsertJobTitle, "job-title", "", "job title")
	upsertCmd.Flags().StringVar(&upsertCompanyID, "company-id", "", "company ID to associate")
	upsertCmd.Flags().StringVarP(&upsertData, "data", "d", "", "JSON data (overrides other flags)")

	_ = upsertCmd.MarkFlagRequired("email")
}

func runUpsert(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	var input graphql.UpsertPersonInput

	if upsertData != "" {
		if err := json.Unmarshal([]byte(upsertData), &input); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		// Ensure email is set from the flag if not in JSON
		if input.Email == "" {
			input.Email = upsertEmail
		}
	} else {
		input = graphql.UpsertPersonInput{
			FirstName: upsertFirstName,
			LastName:  upsertLastName,
			Email:     upsertEmail,
			Phone:     upsertPhone,
			JobTitle:  upsertJobTitle,
			CompanyID: upsertCompanyID,
		}
	}

	client := rt.GraphQLClient()

	person, err := client.UpsertPerson(ctx, &input)
	if err != nil {
		return fmt.Errorf("failed to upsert person: %w", err)
	}

	if rt.Output == "json" {
		return outfmt.WriteJSON(os.Stdout, person, rt.Query)
	}

	fmt.Printf("Upserted person: %s (%s %s)\n", person.ID, person.Name.FirstName, person.Name.LastName)
	return nil
}
