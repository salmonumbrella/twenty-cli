package companies

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var (
	createName          string
	createDomain        string
	createAddress       string
	createEmployees     int
	createLinkedinLink  string
	createXLink         string
	createAnnualRevenue int
	createIdealCustomer bool
	createData          string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	return builder.NewCreateTypedCommand(builder.CreateTypedConfig[rest.CreateCompanyInput, types.Company]{
		Use:   "create",
		Short: "Create a new company",
		BuildInput: func(cmd *cobra.Command) (*rest.CreateCompanyInput, error) {
			var input rest.CreateCompanyInput
			if createData != "" {
				if err := json.Unmarshal([]byte(createData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			input = rest.CreateCompanyInput{
				Name:          createName,
				Address:       createAddress,
				Employees:     createEmployees,
				AnnualRevenue: createAnnualRevenue,
				IdealCustomer: createIdealCustomer,
			}
			if createDomain != "" {
				input.DomainName = &rest.LinkInput{PrimaryLinkUrl: createDomain}
			}
			if createLinkedinLink != "" {
				input.LinkedinLink = &rest.LinkInput{PrimaryLinkUrl: createLinkedinLink}
			}
			if createXLink != "" {
				input.XLink = &rest.LinkInput{PrimaryLinkUrl: createXLink}
			}
			return &input, nil
		},
		CreateFunc: func(ctx context.Context, client *rest.Client, input *rest.CreateCompanyInput) (*types.Company, error) {
			return client.CreateCompany(ctx, input)
		},
		OutputText: func(company *types.Company) string {
			return fmt.Sprintf("Created company: %s (%s)", company.ID, company.Name)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&createName, "name", "", "company name")
			cmd.Flags().StringVar(&createDomain, "domain", "", "domain name")
			cmd.Flags().StringVar(&createAddress, "address", "", "address")
			cmd.Flags().IntVar(&createEmployees, "employees", 0, "number of employees")
			cmd.Flags().StringVar(&createLinkedinLink, "linkedin", "", "LinkedIn URL")
			cmd.Flags().StringVar(&createXLink, "x-link", "", "X (Twitter) URL")
			cmd.Flags().IntVar(&createAnnualRevenue, "revenue", 0, "annual revenue")
			cmd.Flags().BoolVar(&createIdealCustomer, "ideal-customer", false, "mark as ideal customer")
			cmd.Flags().StringVarP(&createData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := createCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}
