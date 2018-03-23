package qingcloud

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestNewConfig(t *testing.T) {
	for _,testcase := range testcases {
		_,warnings,err:=NewConfig(testcase.input)
		for _,item :=range warnings {
			t.Log(item)
		}
		if testcase.expected != nil {
			assert.EqualError(t,err,*testcase.expected)
		} else if err != nil {
			t.Error(err)
		}
	}
}