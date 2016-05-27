package api_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/logger/loggerfakes"
)

var _ = Describe("Check", func() {
	var (
		server *ghttp.Server
		client api.Client
		target string

		fakeLogger *loggerfakes.FakeLogger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		target = server.URL()

		fakeLogger = &loggerfakes.FakeLogger{}
		client = api.NewClient(target, "", "")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Pipelines", func() {
		var (
			response           []api.Pipeline
			responseStatusCode int
		)

		BeforeEach(func() {
			response = []api.Pipeline{
				{Name: "p1", URL: "url_p2"},
				{Name: "p2", URL: "url_p1"},
			}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/pipelines"),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns successfully", func() {
			returnedPipelines, err := client.Pipelines()
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedPipelines).To(HaveLen(2))
		})

		Context("when getting pipelines returns unmarshallable body", func() {
			BeforeEach(func() {
				server.Reset()

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/pipelines",
							apiPrefix,
						)),
						ghttp.RespondWith(
							http.StatusOK,
							`$not%valid-#json`,
						),
					),
				)
			})

			It("returns error", func() {
				_, err := client.Pipelines()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
