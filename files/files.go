package files

import (
	"bufio"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/Lars5Janssen/its-pw/util"
)

func WriteYAML(path string, content interface{}) {
	f, os_err := os.Create(path)
	defer f.Close()
	util.Check(os_err)
	new_content, cred_err := yaml.Marshal(&content)
	util.Check(cred_err)
	writer := bufio.NewWriter(f)
	_, w_err := writer.WriteString(string(new_content))
	util.Check(w_err)
	writer.Flush()

}

func ReadYaml(path string) map[string]string {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error with reading yaml")
		return make(map[string]string)
	}
	filemap := make(map[interface{}]interface{})
	err2 := yaml.Unmarshal(file, &filemap)
	util.Check(err2)
	return util.ItoSmap(filemap)
}
