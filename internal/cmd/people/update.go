package people

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
	updateFirstName string
	updateLastName  string
	updateEmail     string
	updatePhone     string
	updateJobTitle  string
	updateData      string
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cobra.Command {
	return builder.NewUpdateTypedCommand(builder.UpdateTypedConfig[rest.UpdatePersonInput, types.Person]{
		Use:   "update",
		Short: "Update a person",
		BuildInputWithClient: func(ctx context.Context, client *rest.Client, cmd *cobra.Command, id string) (*rest.UpdatePersonInput, error) {
			var input rest.UpdatePersonInput
			if updateData != "" {
				if err := json.Unmarshal([]byte(updateData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}

			// For name updates, fetch existing person to preserve unchanged fields
			firstNameChanged := cmd.Flags().Changed("first-name")
			lastNameChanged := cmd.Flags().Changed("last-name")

			if firstNameChanged || lastNameChanged {
				existing, err := client.GetPerson(ctx, id, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch existing person: %w", err)
				}

				firstName := existing.Name.FirstName
				lastName := existing.Name.LastName

				if firstNameChanged {
					firstName = updateFirstName
				}
				if lastNameChanged {
					lastName = updateLastName
				}

				input.Name = &types.Name{
					FirstName: firstName,
					LastName:  lastName,
				}
			}
			if cmd.Flags().Changed("email") {
				input.Email = &types.Email{
					PrimaryEmail: updateEmail,
				}
			}
			if cmd.Flags().Changed("phone") {
				input.Phone = &types.Phone{
					PrimaryPhoneNumber: updatePhone,
				}
			}
			if cmd.Flags().Changed("job-title") {
				input.JobTitle = &updateJobTitle
			}

			return &input, nil
		},
		UpdateFunc: func(ctx context.Context, client *rest.Client, id string, input *rest.UpdatePersonInput) (*types.Person, error) {
			return client.UpdatePerson(ctx, id, input)
		},
		OutputText: func(person *types.Person) string {
			return fmt.Sprintf("Updated person: %s (%s %s)", person.ID, person.Name.FirstName, person.Name.LastName)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&updateFirstName, "first-name", "", "first name")
			cmd.Flags().StringVar(&updateLastName, "last-name", "", "last name")
			cmd.Flags().StringVar(&updateEmail, "email", "", "primary email")
			cmd.Flags().StringVar(&updatePhone, "phone", "", "primary phone")
			cmd.Flags().StringVar(&updateJobTitle, "job-title", "", "job title")
			cmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := updateCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		if strings.Contains(err.Error(), "failed to fetch existing person") {
			return err
		}
		return fmt.Errorf("failed to update person: %w", err)
	}
	return nil
}
