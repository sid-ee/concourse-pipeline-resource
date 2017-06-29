package api_test

import (
	"fmt"
	"net/http"

	"github.com/concourse/concourse-pipeline-resource/concourse"
	"github.com/concourse/concourse-pipeline-resource/concourse/api"
	"github.com/concourse/concourse-pipeline-resource/logger/loggerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
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
		teams := []concourse.Team{
			{
				Name:     "main",
				Username: "foo",
				Password: "bar",
			},
		}
		client = api.NewClient(target, false, teams)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Pipelines", func() {
		var (
			token              api.AuthToken
			tokenStatusCode    int
			response           []api.Pipeline
			responseStatusCode int
		)

		BeforeEach(func() {
			token = api.AuthToken{
				Type:  "Bearer",
				Value: "foobar",
			}

			response = []api.Pipeline{
				{Name: "p1", URL: "url_p2"},
				{Name: "p2", URL: "url_p1"},
			}

			tokenStatusCode = http.StatusOK
			responseStatusCode = http.StatusOK
		})

		It("returns successfully", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/teams/main/auth/token"),
					ghttp.VerifyBasicAuth("foo", "bar"),
					ghttp.RespondWithJSONEncoded(tokenStatusCode, token),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/teams/main/pipelines"),
					ghttp.VerifyHeaderKV("Authorization", "Bearer foobar"),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
			returnedPipelines, err := client.Pipelines("main")
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedPipelines).To(HaveLen(2))
		})

		Context("when getting pipelines returns unmarshallable body", func() {
			It("returns error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/teams/main/auth/token"),
						ghttp.VerifyBasicAuth("foo", "bar"),
						ghttp.RespondWithJSONEncoded(tokenStatusCode, token),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/teams/main/pipelines",
							apiPrefix,
						)),
						ghttp.VerifyHeaderKV("Authorization", "Bearer foobar"),
						ghttp.RespondWith(
							http.StatusOK,
							`$not%valid-#json`,
						),
					),
				)
				_, err := client.Pipelines("main")
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when getting pipelines on invalid target", func() {
			It("returns error including target url", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/teams/main/auth/token"),
						ghttp.VerifyBasicAuth("foo", "bar"),
						ghttp.RespondWithJSONEncoded(tokenStatusCode, token),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/teams/main/pipelines",
							apiPrefix,
						)),
						ghttp.VerifyHeaderKV("Authorization", "Bearer foobar"),
						ghttp.RespondWith(
							http.StatusNotFound,
							"",
						),
					),
				)
				_, err := client.Pipelines("main")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(target))
			})
		})

		Context("when getting auth token returns unmarshallable body", func() {
			It("returns error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/teams/main/auth/token"),
						ghttp.VerifyBasicAuth("foo", "bar"),
						ghttp.RespondWith(
							http.StatusOK,
							`$not%valid-#json`,
						),
					),
				)
				_, err := client.Pipelines("main")
				Expect(err).To(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
		Context("when getting auth token on invalid target", func() {
			It("returns error including target url", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/teams/main/auth/token"),
						ghttp.VerifyBasicAuth("foo", "bar"),
						ghttp.RespondWith(
							http.StatusNotFound,
							"",
						),
					),
				)
				_, err := client.Pipelines("main")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(target))

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

	})
})
