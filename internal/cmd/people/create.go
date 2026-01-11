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
	createFirstName string
	createLastName  string
	createEmail     string
	createPhone     string
	createJobTitle  string
	createCity      string
	createData      string
)

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	return builder.NewCreateTypedCommand(builder.CreateTypedConfig[rest.CreatePersonInput, types.Person]{
		Use:   "create",
		Short: "Create a new person",
		BuildInput: func(cmd *cobra.Command) (*rest.CreatePersonInput, error) {
			var input rest.CreatePersonInput
			if createData != "" {
				if err := json.Unmarshal([]byte(createData), &input); err != nil {
					return nil, fmt.Errorf("invalid JSON data: %w", err)
				}
				return &input, nil
			}
			input = rest.CreatePersonInput{
				Name: types.Name{
					FirstName: createFirstName,
					LastName:  createLastName,
				},
				Email: types.Email{
					PrimaryEmail: createEmail,
				},
				Phone: types.Phone{
					PrimaryPhoneNumber: createPhone,
				},
				JobTitle: createJobTitle,
				City:     createCity,
			}
			return &input, nil
		},
		CreateFunc: func(ctx context.Context, client *rest.Client, input *rest.CreatePersonInput) (*types.Person, error) {
			return client.CreatePerson(ctx, input)
		},
		OutputText: func(person *types.Person) string {
			return fmt.Sprintf("Created person: %s (%s %s)", person.ID, person.Name.FirstName, person.Name.LastName)
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&createFirstName, "first-name", "", "first name")
			cmd.Flags().StringVar(&createLastName, "last-name", "", "last name")
			cmd.Flags().StringVar(&createEmail, "email", "", "primary email")
			cmd.Flags().StringVar(&createPhone, "phone", "", "primary phone")
			cmd.Flags().StringVar(&createJobTitle, "job-title", "", "job title")
			cmd.Flags().StringVar(&createCity, "city", "", "city")
			cmd.Flags().StringVarP(&createData, "data", "d", "", "JSON data (overrides other flags)")
		},
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := createCmd.RunE(cmd, args); err != nil {
		if strings.Contains(err.Error(), "invalid JSON data") {
			return err
		}
		return fmt.Errorf("failed to create person: %w", err)
	}
	return nil
}
