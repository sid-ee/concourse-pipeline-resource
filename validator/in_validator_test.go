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
				Target:   "some target",
				Username: "some username",
				Password: "some password",
			},
		}
	})

	Context("when no username is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Username = ""
		})

		It("returns an error", func() {
			err := validator.ValidateIn(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided"))
		})
	})

	Context("when no password is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Password = ""
		})

		It("returns an error", func() {
			err := validator.ValidateIn(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided"))
		})
	})
})
