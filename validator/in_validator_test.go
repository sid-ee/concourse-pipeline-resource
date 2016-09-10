package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/validator"
)

var _ = Describe("ValidateIn", func() {
	var (
		inRequest concourse.InRequest
	)

	BeforeEach(func() {
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target: "some target",
				Teams: []concourse.Team{
					{
						Name:     "some team",
						Username: "some username",
						Password: "some password",
					},
				},
			},
		}
	})

	Context("when no team name is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Teams[0].Name = ""
		})

		It("returns an error", func() {
			err := validator.ValidateIn(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*name.*provided.*team.*0"))
		})
	})

	Context("when no team username is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Teams[0].Username = ""
		})

		It("returns an error", func() {
			err := validator.ValidateIn(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided.*team.*%s", "some team"))
		})
	})

	Context("when no team password is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Teams[0].Password = ""
		})

		It("returns an error", func() {
			err := validator.ValidateIn(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided.*team.*%s", "some team"))
		})
	})
})
