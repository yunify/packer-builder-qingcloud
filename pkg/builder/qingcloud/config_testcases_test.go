package qingcloud


type ConfigTestcases struct {
	input map[string]interface{}
	expected *string
}

var testcases = []ConfigTestcases{
	{
		input:map[string]interface{}{
			"zone":"pek3a",
		},
		expected:nil,
	},

}