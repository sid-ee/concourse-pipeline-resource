package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/validator"
)

var _ = Describe("ValidateCheck", func() {
	var (
		checkRequest concourse.CheckRequest
	)

	BeforeEach(func() {
		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				Target:   "some target",
				Username: "some username",
				Password: "some password",
			},
		}
	})

	Context("when no username is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.Username = ""
		})

		It("returns an error", func() {
			err := validator.ValidateCheck(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided"))
		})
	})

	Context("when no password is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.Password = ""
		})

		It("returns an error", func() {
			err := validator.ValidateCheck(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided"))
		})
	})
})
