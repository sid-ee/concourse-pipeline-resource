package helpers

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/aryann/difflib"
	"github.com/concourse/atc"
	"github.com/mgutz/ansi"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v2"
)

//go:generate counterfeiter . ConfigDiffer

type ConfigDiffer interface {
	Diff(existingConfig atc.Config, newConfig atc.Config) error
}

type configDiffer struct {
	writer io.Writer
}

func NewConfigDiffer(writer io.Writer) ConfigDiffer {
	return &configDiffer{writer: writer}
}

func (p configDiffer) Diff(
	existingConfig atc.Config,
	newConfig atc.Config,
) error {
	indent := gexec.NewPrefixedWriter("  ", p.writer)

	groupDiffs := diffIndices(GroupIndex(existingConfig.Groups), GroupIndex(newConfig.Groups))
	if len(groupDiffs) > 0 {
		fmt.Fprintln(p.writer, "groups:")

		for _, diff := range groupDiffs {
			diff.Render(indent, "group")
		}
	}

	resourceDiffs := diffIndices(ResourceIndex(existingConfig.Resources), ResourceIndex(newConfig.Resources))
	if len(resourceDiffs) > 0 {
		fmt.Fprintln(p.writer, "resources:")

		for _, diff := range resourceDiffs {
			diff.Render(indent, "resource")
		}
	}

	resourceTypeDiffs := diffIndices(ResourceTypeIndex(existingConfig.ResourceTypes), ResourceTypeIndex(newConfig.ResourceTypes))
	if len(resourceTypeDiffs) > 0 {
		fmt.Fprintln(p.writer, "resource types:")

		for _, diff := range resourceTypeDiffs {
			diff.Render(indent, "resource type")
		}
	}

	jobDiffs := diffIndices(JobIndex(existingConfig.Jobs), JobIndex(newConfig.Jobs))
	if len(jobDiffs) > 0 {
		fmt.Fprintln(p.writer, "jobs:")

		for _, diff := range jobDiffs {
			diff.Render(indent, "job")
		}
	}

	return nil
}

type Index interface {
	FindEquivalent(interface{}) (interface{}, bool)
	Slice() []interface{}
}

type Diffs []Diff

type Diff struct {
	Before interface{}
	After  interface{}
}

func name(v interface{}) string {
	return reflect.ValueOf(v).FieldByName("Name").String()
}

func (diff Diff) Render(to io.Writer, label string) {
	indent := gexec.NewPrefixedWriter("  ", to)

	if diff.Before != nil && diff.After != nil {
		fmt.Fprintf(to, ansi.Color("%s %s has changed:", "yellow")+"\n", label, name(diff.Before))

		payloadA, _ := yaml.Marshal(diff.Before)
		payloadB, _ := yaml.Marshal(diff.After)

		renderDiff(indent, string(payloadA), string(payloadB))
	} else if diff.Before != nil {
		fmt.Fprintf(to, ansi.Color("%s %s has been removed:", "yellow")+"\n", label, name(diff.Before))

		payloadA, _ := yaml.Marshal(diff.Before)

		renderDiff(indent, string(payloadA), "")
	} else {
		fmt.Fprintf(to, ansi.Color("%s %s has been added:", "yellow")+"\n", label, name(diff.After))

		payloadB, _ := yaml.Marshal(diff.After)

		renderDiff(indent, "", string(payloadB))
	}
}

type GroupIndex atc.GroupConfigs

func (index GroupIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index GroupIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.GroupConfigs(index).Lookup(name(obj))
}

type JobIndex atc.JobConfigs

func (index JobIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index JobIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.JobConfigs(index).Lookup(name(obj))
}

type ResourceIndex atc.ResourceConfigs

func (index ResourceIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index ResourceIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.ResourceConfigs(index).Lookup(name(obj))
}

type ResourceTypeIndex atc.ResourceTypes

func (index ResourceTypeIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index ResourceTypeIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.ResourceTypes(index).Lookup(name(obj))
}

func diffIndices(oldIndex Index, newIndex Index) Diffs {
	diffs := Diffs{}

	for _, thing := range oldIndex.Slice() {
		newThing, found := newIndex.FindEquivalent(thing)
		if !found {
			diffs = append(diffs, Diff{
				Before: thing,
				After:  nil,
			})
			continue
		}

		if practicallyDifferent(thing, newThing) {
			diffs = append(diffs, Diff{
				Before: thing,
				After:  newThing,
			})
		}
	}

	for _, thing := range newIndex.Slice() {
		_, found := oldIndex.FindEquivalent(thing)
		if !found {
			diffs = append(diffs, Diff{
				Before: nil,
				After:  thing,
			})
			continue
		}
	}

	return diffs
}

func renderDiff(to io.Writer, a, b string) {
	diffs := difflib.Diff(strings.Split(a, "\n"), strings.Split(b, "\n"))

	for _, diff := range diffs {
		text := diff.Payload

		switch diff.Delta {
		case difflib.RightOnly:
			fmt.Fprintf(to, "%s\n", ansi.Color(text, "green"))
		case difflib.LeftOnly:
			fmt.Fprintf(to, "%s\n", ansi.Color(text, "red"))
		case difflib.Common:
			fmt.Fprintf(to, "%s\n", text)
		}
	}
}

func practicallyDifferent(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return false
	}

	// prevent silly things like 300 != 300.0 due to YAML vs. JSON
	// inconsistencies

	marshalledA, _ := yaml.Marshal(a)
	marshalledB, _ := yaml.Marshal(b)

	return !bytes.Equal(marshalledA, marshalledB)
}
