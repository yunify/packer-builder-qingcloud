package qingcloud

import "io/ioutil"

func loadFileContent(path string) (string,error) {
	content,err:=ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content),nil
}