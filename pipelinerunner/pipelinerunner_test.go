package pipelinerunner_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/pipelinerunner"
	"github.com/robdimsdale/concourse-pipeline-resource/pipelinerunner/pipelinerunnerfakes"
)

var _ = Describe("PipelineRunner", func() {
	const (
		funcSleep = 50 * time.Millisecond
	)

	var (
		fakeLogger *pipelinerunnerfakes.FakeLogger

		aFunc     func(index int, pipeline api.Pipeline) (string, error)
		pipelines []api.Pipeline
	)

	BeforeEach(func() {
		fakeLogger = &pipelinerunnerfakes.FakeLogger{}

		aFunc = func(index int, pipeline api.Pipeline) (string, error) {
			time.Sleep(funcSleep)
			return fmt.Sprintf("completed-%s", pipeline.Name), nil
		}

		pipelines = []api.Pipeline{
			{Name: "p1"},
			{Name: "p2"},
			{Name: "p3"},
			{Name: "p4"},
			{Name: "p5"},
		}
	})

	It("runs in parallel", func() {
		// Cannot use Eventually for this as the timeout does not seem to work

		outChan := make(chan []string, 1)
		errChan := make(chan error, 1)
		go func() {
			out, err := pipelinerunner.RunForAllPipelines(aFunc, pipelines, fakeLogger)
			outChan <- out
			errChan <- err
		}()

		select {
		case output := <-outChan:
			Expect(output).To(HaveLen(len(pipelines)))

			for i, p := range pipelines {
				Expect(output[i]).To(Equal(fmt.Sprintf("completed-%s", p.Name)))
			}
		case err := <-errChan:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(2 * funcSleep):
			Fail("took too long")
		}
	})

	Context("when an error is returned by the function", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("some error")

			aFunc = func(index int, pipeline api.Pipeline) (string, error) {
				time.Sleep(funcSleep)
				return "", expectedErr
			}
		})

		It("forwards the error", func() {
			_, err := pipelinerunner.RunForAllPipelines(aFunc, pipelines, fakeLogger)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})
})
