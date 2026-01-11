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
	createName             string
	createAmount           float64
	createCurrency         string
	createCloseDate        string
	createStage            string
	createProbability      int
	createCompanyID        string
	createPointOfContactID string
	createData             string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	return builder.NewCreateTypedCommand(builder.CreateTypedConfig[rest.CreateOpportunityInput, types.Opportunity]{
		Use:   "create",
		Short: "Create a new opportunity",
		BuildInput: func(cmd *cobra.Command) (*rest.CreateOpportunityInput, error) {
			var input rest.CreateOpportunityInput
			if createData != "" {
				if err := json.Unmarshal([]byte(createData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			input = rest.CreateOpportunityInput{
				Name:             createName,
				CloseDate:        createCloseDate,
				Stage:            createStage,
				Probability:      createProbability,
				CompanyID:        createCompanyID,
				PointOfContactID: createPointOfContactID,
			}
			if createAmount > 0 {
				// Convert dollars to micros (multiply by 1,000,000)
				amountMicros := fmt.Sprintf("%.0f", createAmount*1000000)
				input.Amount = &types.Currency{
					AmountMicros: amountMicros,
					CurrencyCode: createCurrency,
				}
			}
			return &input, nil
		},
		CreateFunc: func(ctx context.Context, client *rest.Client, input *rest.CreateOpportunityInput) (*types.Opportunity, error) {
			return client.CreateOpportunity(ctx, input)
		},
		OutputText: func(opportunity *types.Opportunity) string {
			return fmt.Sprintf("Created opportunity: %s (%s)", opportunity.ID, opportunity.Name)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&createName, "name", "", "opportunity name")
			cmd.Flags().Float64Var(&createAmount, "amount", 0, "deal amount (e.g., 50000.00)")
			cmd.Flags().StringVar(&createCurrency, "currency", "USD", "currency code (default: USD)")
			cmd.Flags().StringVar(&createCloseDate, "close-date", "", "expected close date (YYYY-MM-DD)")
			cmd.Flags().StringVar(&createStage, "stage", "", "deal stage")
			cmd.Flags().IntVar(&createProbability, "probability", 0, "win probability (0-100)")
			cmd.Flags().StringVar(&createCompanyID, "company-id", "", "associated company ID")
			cmd.Flags().StringVar(&createPointOfContactID, "contact-id", "", "point of contact ID")
			cmd.Flags().StringVarP(&createData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := createCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to create opportunity: %w", err)
	}
	return nil
}
