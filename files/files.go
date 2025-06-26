package files

import "github.com/Lars5Janssen/its-pw/util"

func writeYAML(path string, content interface{}) {
	f, os_err := os.Create(path)
	defer f.Close()
	check(os_err)
	new_content, cred_err := yaml.Marshal(&content)
	check(cred_err)
	writer := bufio.NewWriter(f)
	_, w_err := writer.WriteString(string(new_content))
	check(w_err)
	writer.Flush()

}

func readYaml(path string) map[string]string {
	file, err := ioutil.ReadFile(path)
	check(err)
	filemap := make(map[interface{}]interface{})
	err2 := yaml.Unmarshal(file, &filemap)
	check(err2)
	finalmap := make(map[string]string)
	for k, v := range filemap {
		key := fmt.Sprintf("%s", k)
		value := fmt.Sprintf("%s", v)
		finalmap[key] = value
	}
	return finalmap
}
