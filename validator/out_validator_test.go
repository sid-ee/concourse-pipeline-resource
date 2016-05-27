package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/validator"
)

var _ = Describe("ValidateOut", func() {
	var (
		outRequest concourse.OutRequest
	)

	BeforeEach(func() {
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target:   "some target",
				Username: "some username",
				Password: "some password",
			},
			Params: concourse.OutParams{
				Pipelines: []concourse.Pipeline{
					{
						Name:       "p1",
						ConfigFile: "some config",
						VarsFiles: []string{
							"some vars",
						},
					},
				},
			},
		}
	})

	It("returns without error", func() {
		Expect(validator.ValidateOut(outRequest)).Should(Succeed())
	})

	Context("when no username is provided", func() {
		BeforeEach(func() {
			outRequest.Source.Username = ""
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided"))
		})
	})

	Context("when no password is provided", func() {
		BeforeEach(func() {
			outRequest.Source.Password = ""
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided"))
		})
	})

	Context("when pipelines param is nil", func() {
		BeforeEach(func() {
			outRequest.Params.Pipelines = nil
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*pipelines.*provided"))
		})
	})

	Context("when pipelines param is empty", func() {
		BeforeEach(func() {
			outRequest.Params.Pipelines = []concourse.Pipeline{}
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*pipelines.*provided"))
		})
	})

	Context("when pipelines_file param is also provided", func() {
		BeforeEach(func() {
			outRequest.Params.PipelinesFile = "some-file"
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*pipelines.*provided.*one of"))
		})
	})

	Context("when vars files is present but empty", func() {
		BeforeEach(func() {
			outRequest.Params.Pipelines[0].VarsFiles = []string{}
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*vars_files.*provided"))
		})
	})

	Context("when vars files contains an empty string", func() {
		BeforeEach(func() {
			outRequest.Params.Pipelines[0].VarsFiles[0] = ""
		})

		It("returns an error", func() {
			err := validator.ValidateOut(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*vars file.*non-empty"))
		})
	})
})
