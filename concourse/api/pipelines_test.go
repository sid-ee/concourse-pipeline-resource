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
			atcConfig         atc.Config
			atcRawConfig      atc.RawConfig
			configVersion     string
			pipelineExists    bool
			pipelineConfigErr error

			pipelineName string
		)

		BeforeEach(func() {
			atcConfig = atc.Config{Groups: atc.GroupConfigs{{Name: "some group"}}}
			atcRawConfig = atc.RawConfig("some raw config")
			configVersion = "some config version"
			pipelineExists = true
			pipelineConfigErr = nil

			pipelineName = "some pipeline"
		})

		JustBeforeEach(func() {
			fakeGCClient.PipelineConfigReturns(
				atcConfig,
				atcRawConfig,
				configVersion,
				pipelineExists,
				pipelineConfigErr,
			)
		})

		It("returns successfully", func() {
			returnedATCConfig, returnedConfig, returnedConfigVersion, err := client.PipelineConfig(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedATCConfig).To(Equal(atcConfig))
			Expect(returnedConfig).To(Equal(atcRawConfig.String()))
			Expect(returnedConfigVersion).To(Equal(configVersion))
		})

		Context("when getting pipelines returns an error", func() {
			BeforeEach(func() {
				pipelineConfigErr = fmt.Errorf("some error")
			})

			It("returns error including target url", func() {
				_, _, _, err := client.PipelineConfig(pipelineName)
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
				_, _, _, err := client.PipelineConfig(pipelineName)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring(pipelineName))
				Expect(err.Error()).Should(ContainSubstring("not found"))
			})
		})
	})

	Describe("SetPipelineConfig", func() {
		var (
			pipelineCreated   bool
			pipelineUpdated   bool
			warnings          []gc.ConfigWarning
			pipelineConfigErr error

			pipelineName  string
			configVersion string
			passedConfig  atc.Config
		)

		BeforeEach(func() {
			pipelineCreated = false
			pipelineUpdated = true
			warnings = nil
			pipelineConfigErr = nil

			pipelineName = "some pipeline"
			configVersion = "some version"
			passedConfig = atc.Config{}
		})

		JustBeforeEach(func() {
			fakeGCClient.CreateOrUpdatePipelineConfigReturns(
				pipelineCreated,
				pipelineUpdated,
				warnings,
				pipelineConfigErr,
			)
		})

		It("returns successfully", func() {
			err := client.SetPipelineConfig(
				pipelineName,
				configVersion,
				passedConfig,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeGCClient.CreateOrUpdatePipelineConfigCallCount()).To(Equal(1))
			invokedPipelineName, invokedConfigVersion, invokedPassedConfig :=
				fakeGCClient.CreateOrUpdatePipelineConfigArgsForCall(0)

			Expect(invokedPipelineName).To(Equal(pipelineName))
			Expect(invokedConfigVersion).To(Equal(configVersion))
			Expect(invokedPassedConfig).To(Equal(passedConfig))
		})

		Context("when getting pipelines returns an error", func() {
			BeforeEach(func() {
				pipelineConfigErr = fmt.Errorf("some error")
			})

			It("returns error including target url", func() {
				err := client.SetPipelineConfig(
					pipelineName,
					configVersion,
					passedConfig,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring("some error"))
			})
		})

		Context("when pipeline was not created or updated", func() {
			BeforeEach(func() {
				pipelineCreated = false
				pipelineUpdated = false
			})

			It("returns error including target url", func() {
				err := client.SetPipelineConfig(
					pipelineName,
					configVersion,
					passedConfig,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring(pipelineName))
				Expect(err.Error()).Should(ContainSubstring("not created or updated"))
			})
		})
	})

	Describe("DeletePipeline", func() {
		var (
			pipelineExists    bool
			pipelineConfigErr error

			pipelineName string
		)

		BeforeEach(func() {
			pipelineExists = true
			pipelineConfigErr = nil

			pipelineName = "some pipeline"
		})

		JustBeforeEach(func() {
			fakeGCClient.DeletePipelineReturns(pipelineExists, pipelineConfigErr)
		})

		It("returns successfully", func() {
			err := client.DeletePipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeGCClient.DeletePipelineCallCount()).To(Equal(1))
			Expect(fakeGCClient.DeletePipelineArgsForCall(0)).To(Equal(pipelineName))
		})

		Context("when getting pipelines returns an error", func() {
			BeforeEach(func() {
				pipelineConfigErr = fmt.Errorf("some error")
			})

			It("returns error including target url", func() {
				err := client.DeletePipeline(pipelineName)
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
				err := client.DeletePipeline(pipelineName)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(ContainSubstring(target))
				Expect(err.Error()).Should(ContainSubstring(pipelineName))
				Expect(err.Error()).Should(ContainSubstring("not found"))
			})
		})
	})
})
