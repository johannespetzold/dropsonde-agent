package agent_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"log"
	"io/ioutil"
)

func TestAgent(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Agent Suite")
}
