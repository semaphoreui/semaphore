package tests

import (
	"testing"
	"path/filepath"
	. "gopkg.in/check.v1"
	"github.com/ansible-semaphore/semaphore/api/projects"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}
var _ = Suite(&MySuite{})


func (s *MySuite) TestIsValidInventoryPath(c *C) {
	isValid := projects.IsValidInventoryPath("")
	c.Assert(isValid, Equals, true)
}

func (s *MySuite) TestFilePathMatch(c *C) {
	matched, _ := filepath.Match("test.*", "test.txt")
	c.Assert(matched, Equals, true)
}