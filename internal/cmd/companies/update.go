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
	updateName          string
	updateDomain        string
	updateAddress       string
	updateEmployees     int
	updateLinkedinLink  string
	updateXLink         string
	updateAnnualRevenue int
	updateIdealCustomer bool
	updateData          string
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cobra.Command {
	return builder.NewUpdateTypedCommand(builder.UpdateTypedConfig[rest.UpdateCompanyInput, types.Company]{
		Use:   "update",
		Short: "Update a company",
		BuildInput: func(cmd *cobra.Command, id string) (*rest.UpdateCompanyInput, error) {
			var input rest.UpdateCompanyInput
			if updateData != "" {
				if err := json.Unmarshal([]byte(updateData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			if cmd.Flags().Changed("name") {
				input.Name = &updateName
			}
			if cmd.Flags().Changed("domain") {
				input.DomainName = &rest.LinkInput{PrimaryLinkUrl: updateDomain}
			}
			if cmd.Flags().Changed("address") {
				input.Address = &updateAddress
			}
			if cmd.Flags().Changed("employees") {
				input.Employees = &updateEmployees
			}
			if cmd.Flags().Changed("linkedin") {
				input.LinkedinLink = &rest.LinkInput{PrimaryLinkUrl: updateLinkedinLink}
			}
			if cmd.Flags().Changed("x-link") {
				input.XLink = &rest.LinkInput{PrimaryLinkUrl: updateXLink}
			}
			if cmd.Flags().Changed("revenue") {
				input.AnnualRevenue = &updateAnnualRevenue
			}
			if cmd.Flags().Changed("ideal-customer") {
				input.IdealCustomer = &updateIdealCustomer
			}
			return &input, nil
		},
		UpdateFunc: func(ctx context.Context, client *rest.Client, id string, input *rest.UpdateCompanyInput) (*types.Company, error) {
			return client.UpdateCompany(ctx, id, input)
		},
		OutputText: func(company *types.Company) string {
			return fmt.Sprintf("Updated company: %s (%s)", company.ID, company.Name)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&updateName, "name", "", "company name")
			cmd.Flags().StringVar(&updateDomain, "domain", "", "domain name")
			cmd.Flags().StringVar(&updateAddress, "address", "", "address")
			cmd.Flags().IntVar(&updateEmployees, "employees", 0, "number of employees")
			cmd.Flags().StringVar(&updateLinkedinLink, "linkedin", "", "LinkedIn URL")
			cmd.Flags().StringVar(&updateXLink, "x-link", "", "X (Twitter) URL")
			cmd.Flags().IntVar(&updateAnnualRevenue, "revenue", 0, "annual revenue")
			cmd.Flags().BoolVar(&updateIdealCustomer, "ideal-customer", false, "mark as ideal customer")
			cmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := updateCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to update company: %w", err)
	}
	return nil
}
