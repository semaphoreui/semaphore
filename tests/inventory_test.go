package tests

import (
	"testing"
	. "gopkg.in/check.v1"
	"github.com/ansible-semaphore/semaphore/api/projects"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}
var _ = Suite(&MySuite{})


func (s *MySuite) TestIsValidInventoryPath(c *C) {
	c.Assert(projects.IsValidInventoryPath("inventories/test"), Equals, true)

	c.Assert(projects.IsValidInventoryPath("inventories/test/../prod"), Equals, true)

	c.Assert(projects.IsValidInventoryPath("/test/../../../inventory"), Equals, false)

	c.Assert(projects.IsValidInventoryPath("/test/inventory"), Equals, false)

	c.Assert(projects.IsValidInventoryPath("c:\\test\\inventory"), Equals, false)
}