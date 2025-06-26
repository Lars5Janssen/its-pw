package login

func readCreds(config_path string) map[string]string {
	config := readYaml(config_path)
	credentials_store_method, exists := config["credentials_store_method"]
	if !exists {
		log.Fatal("credentials_store_method does not exist")
	}
	if credentials_store_method != "yaml" {
		log.Fatal("not yet implemented")
	}
	credentials_file, exists_file := config["credentials_file"]
	if !exists_file {
		log.Fatal("cred file not found?")
	}
	cred_path := credentials_file
	creds_to_convert := readYaml(cred_path)
	creds := make(map[string]string)
	for k, v := range creds_to_convert {
		key := fmt.Sprintf("%s", k)
		value := fmt.Sprintf("%s", v)
		creds[key] = value
	}
	return creds

}

func hashMe(toHash string) string {
	h := sha256.New()
	h.Write([]byte(toHash))
	return string(h.Sum(nil))
}

func checkLogin(username string, password string, totpCode string) bool {

	creds := readCreds("config.yaml")

	found_password, user_exists := creds[username]
	if !user_exists {
		return false
	}
	if found_password != hashMe(password) {
		return false
	}
	userSecret, exists := totpmap[username]
	if !exists {
		return false
	}
	if totp.Validate(totpCode, userSecret) {
		return true
	}
	return false
}

func AddUser(username string, password string) {
	config := readYaml("config.yaml")
	credentials_file, exists_file := config["credentials_file"]
	if !exists_file {
		log.Fatal("cred file not found?")
	}
	cred_path := credentials_file
	f, os_err := os.Create(cred_path)
	check(os_err)
	creds := readCreds("config.yaml")
	creds[username] = hashMe(password)
	new_creds, cred_err := yaml.Marshal(&creds)
	check(cred_err)
	defer f.Close()
	writer := bufio.NewWriter(f)
	_, w_err := writer.WriteString(string(new_creds))
	check(w_err)
	writer.Flush()

}
func generateTOTP(username string) {
	key, keyErr := totp.Generate(totp.GenerateOpts{
		Issuer:      "localhost",
		AccountName: username,
		SecretSize:  20,
	})

	if keyErr != nil {
		log.Fatal("Something has gone wrong during otp generation")
	}

	totpmap[username] = key.Secret()
	fmt.Println(key.Secret())
}

func checkSessionToken(w http.ResponseWriter, r *http.Request) (bool, int, session) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return false, http.StatusUnauthorized, session{}
		}

		w.WriteHeader(http.StatusBadRequest)
		return false, http.StatusBadRequest, session{}
	}
	sessionToken := c.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, session{}
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, session{}
	}

	return true, http.StatusOK, userSession
}
