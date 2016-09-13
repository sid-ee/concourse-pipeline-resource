package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
)

var _ = Describe("HTTP Client", func() {
	Describe("LoginWithBasicAuth", func() {
		var (
			url      string
			teamName string
			username string
			password string
			insecure bool
		)

		Context("when logging in with the client fails", func() {
			BeforeEach(func() {
				url = "invalid-url"
			})

			It("returns an error", func() {
				_, err := api.LoginWithBasicAuth(
					url,
					teamName,
					username,
					password,
					insecure,
				)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
