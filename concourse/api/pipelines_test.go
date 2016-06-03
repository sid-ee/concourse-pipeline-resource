package api_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api/gcfakes"

	"github.com/concourse/atc"
	gc "github.com/concourse/go-concourse/concourse"
)

var _ = Describe("Pipeline methods", func() {
	var (
		originalNewGCClientFunc func(target string, httpClient *http.Client) gc.Client
		fakeGCClient            *gcfakes.FakeClient

		client api.Client
		target string
	)

	BeforeEach(func() {
		originalNewGCClientFunc = api.NewGCClientFunc

		fakeGCClient = &gcfakes.FakeClient{}

		api.NewGCClientFunc = func(target string, httpClient *http.Client) gc.Client {
			return fakeGCClient
		}

		target = "some target"

		client = api.NewClient(target, &http.Client{})
	})

	AfterEach(func() {
		api.NewGCClientFunc = originalNewGCClientFunc
	})

	Describe("Pipelines", func() {
		var (
			atcPipelines []atc.Pipeline
			pipelinesErr error
		)

		BeforeEach(func() {
			pipelinesErr = nil

			atcPipelines = []atc.Pipeline{
				{Name: "p1", URL: "url_p2"},
				{Name: "p2", URL: "url_p1"},
			}
		})

		JustBeforeEach(func() {
			fakeGCClient.ListPipelinesReturns(atcPipelines, pipelinesErr)
		})

		It("returns successfully", func() {
			returnedPipelines, err := client.Pipelines()
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedPipelines).To(HaveLen(2))
		})

		Context("when getting pipelines returns an error", func() {
			BeforeEach(func() {
				pipelinesErr = fmt.Errorf("some error")
			})

			It("returns error including target url", func() {
				_, err := client.Pipelines()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring("some error"))
			})
		})
	})

	Describe("PipelineConfig", func() {
		var (
			atcRawConfig      atc.RawConfig
			pipelineExists    bool
			pipelineConfigErr error

			pipelineName string
		)

		BeforeEach(func() {
			pipelineExists = true
			pipelineConfigErr = nil

			atcRawConfig = atc.RawConfig("some raw config")

			pipelineName = "some pipeline"
		})

		JustBeforeEach(func() {
			fakeGCClient.PipelineConfigReturns(atc.Config{}, atcRawConfig, "", pipelineExists, pipelineConfigErr)
		})

		It("returns successfully", func() {
			returnedConfig, err := client.PipelineConfig(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedConfig).To(Equal(atcRawConfig.String()))
		})

		Context("when getting pipelines returns an error", func() {
			BeforeEach(func() {
				pipelineConfigErr = fmt.Errorf("some error")
			})

			It("returns error including target url", func() {
				_, err := client.PipelineConfig(pipelineName)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring("some error"))
			})
		})

		Context("when pipeline does not exist", func() {
			BeforeEach(func() {
				pipelineExists = false
			})

			It("returns error including target url", func() {
				_, err := client.PipelineConfig(pipelineName)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring(pipelineName))
				Expect(err.Error()).Should(ContainSubstring("not found"))
			})
		})
	})
})
