package opportunities

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
	updateName             string
	updateAmount           float64
	updateCurrency         string
	updateCloseDate        string
	updateStage            string
	updateProbability      int
	updateCompanyID        string
	updatePointOfContactID string
	updateData             string
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cobra.Command {
	return builder.NewUpdateTypedCommand(builder.UpdateTypedConfig[rest.UpdateOpportunityInput, types.Opportunity]{
		Use:   "update",
		Short: "Update an opportunity",
		BuildInput: func(cmd *cobra.Command, id string) (*rest.UpdateOpportunityInput, error) {
			var input rest.UpdateOpportunityInput
			if updateData != "" {
				if err := json.Unmarshal([]byte(updateData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			if cmd.Flags().Changed("name") {
				input.Name = &updateName
			}
			if cmd.Flags().Changed("amount") {
				// Convert dollars to micros (multiply by 1,000,000)
				amountMicros := fmt.Sprintf("%.0f", updateAmount*1000000)
				input.Amount = &types.Currency{
					AmountMicros: amountMicros,
					CurrencyCode: updateCurrency,
				}
			}
			if cmd.Flags().Changed("close-date") {
				input.CloseDate = &updateCloseDate
			}
			if cmd.Flags().Changed("stage") {
				input.Stage = &updateStage
			}
			if cmd.Flags().Changed("probability") {
				input.Probability = &updateProbability
			}
			if cmd.Flags().Changed("company-id") {
				input.CompanyID = &updateCompanyID
			}
			if cmd.Flags().Changed("contact-id") {
				input.PointOfContactID = &updatePointOfContactID
			}
			return &input, nil
		},
		UpdateFunc: func(ctx context.Context, client *rest.Client, id string, input *rest.UpdateOpportunityInput) (*types.Opportunity, error) {
			return client.UpdateOpportunity(ctx, id, input)
		},
		OutputText: func(opportunity *types.Opportunity) string {
			return fmt.Sprintf("Updated opportunity: %s (%s)", opportunity.ID, opportunity.Name)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&updateName, "name", "", "opportunity name")
			cmd.Flags().Float64Var(&updateAmount, "amount", 0, "deal amount (e.g., 50000.00)")
			cmd.Flags().StringVar(&updateCurrency, "currency", "USD", "currency code (default: USD)")
			cmd.Flags().StringVar(&updateCloseDate, "close-date", "", "expected close date (YYYY-MM-DD)")
			cmd.Flags().StringVar(&updateStage, "stage", "", "deal stage")
			cmd.Flags().IntVar(&updateProbability, "probability", 0, "win probability (0-100)")
			cmd.Flags().StringVar(&updateCompanyID, "company-id", "", "associated company ID")
			cmd.Flags().StringVar(&updatePointOfContactID, "contact-id", "", "point of contact ID")
			cmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := updateCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to update opportunity: %w", err)
	}
	return nil
}
