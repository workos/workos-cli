package forms

import (
	"github.com/charmbracelet/huh"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

func OrganizationCreate() (organizations.CreateOrganizationOpts, error) {
	org := organizations.CreateOrganizationOpts{}
	name, err := orgNameForm()

	if err != nil {
		return org, err
	}

	domains, err := orgDomainNameForm()

	if err != nil {
		return org, err
	}

	for _, domain := range domains {

		org.Name = name
		org.DomainData = append(org.DomainData, organizations.OrganizationDomainData{
			Domain: domain,
			State:  organizations.Verified,
		})
	}

	return org, nil
}

func orgNameForm() (string, error) {
	var name string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Organization Name").
				Value(&name),
		),
	)

	err := form.Run()

	return name, err
}

func orgDomainNameForm() ([]string, error) {
	add := true
	domains := []string{}

	for add {
		domain := ""
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Organization Domain (optional)").
					Value(&domain),
			),
		)

		err := form.Run()

		if err != nil || domain == "" {
			return domains, err
		}

		domains = append(domains, domain)
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Would you like to add another Domain (optional)").
					Affirmative("Yes").
					Negative("No").
					Value(&add),
			),
		)

		err = form.Run()

		if err != nil {
			return domains, err
		}
	}

	return domains, nil
}
