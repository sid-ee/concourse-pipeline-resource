package concourse_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

var _ = Describe("SanitizedSource", func() {
	It("sanitizes the passwords for all teams", func() {
		source := concourse.Source{
			Teams: []concourse.Team{
				{
					Name:     "team-0",
					Username: "username-0",
					Password: "password-0",
				},
				{
					Name:     "team-1",
					Username: "username-1",
					Password: "password-1",
				},
			},
		}
		sanitized := concourse.SanitizedSource(source)

		Expect(sanitized[source.Teams[0].Password]).To(Equal("***REDACTED-PASSWORD-TEAM-0***"))
		Expect(sanitized[source.Teams[1].Password]).To(Equal("***REDACTED-PASSWORD-TEAM-1***"))
	})
})
