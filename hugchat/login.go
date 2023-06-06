package hugchat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"time"
)

type Login struct {
	Email      string
	Passwd     string
	CookiePath string
	Cookies    []*http.Cookie
	Client     *http.Client
}

func NewLogin(email, passwd, cookiePath string) *Login {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: time.Second * 10,
	}
	return &Login{
		Email:      email,
		Passwd:     passwd,
		CookiePath: cookiePath,
		Client:     client,
	}
}

func (l *Login) SaveCookies() error {
	file, err := os.Create(l.CookiePath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(l.Cookies)
}

func (l *Login) LoadCookies() error {
	file, err := os.Open(l.CookiePath)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	return dec.Decode(&l.Cookies)
}

func (l *Login) SigninWithEmail() error {
	Url := "https://huggingface.co/login"
	form := url.Values{
		"username": {l.Email},
		"password": {l.Passwd},
	}

	req, err := http.NewRequest("POST", Url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body := bytes.NewBufferString(form.Encode())
	req.Body = io.NopCloser(body)

	res, err := l.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusBadRequest {
		return errors.New("wrong username or password")
	}

	return nil
}

func (l *Login) GetAuthURL() (string, error) {
	url := "https://huggingface.co/chat/login"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", "https://huggingface.co/chat/login")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.64")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := l.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return "", err
		}

		location, ok := response["location"].(string)
		if !ok {
			return "", errors.New("no authorize URL found, please check your email or password")
		}

		return location, nil
	} else if res.StatusCode == http.StatusSeeOther {
		location := res.Header.Get("Location")
		if location == "" {
			return "", errors.New("no authorize URL found, please check your email or password")
		}

		return location, nil
	} else {
		return "", fmt.Errorf("something went wrong (status code: %d)", res.StatusCode)
	}
}

func (l *Login) GrantAuth(Url string) error {
	res, err := l.Client.Get(Url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("grant auth fatal")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	csrfRe := regexp.MustCompile(`/oauth/authorize.*?name="csrf" value="(.*?)"`)
	csrfMatch := csrfRe.FindStringSubmatch(string(body))
	if len(csrfMatch) == 0 {
		return errors.New("no csrf found")
	}

	csrf := csrfMatch[1]
	form := url.Values{
		"csrf": {csrf},
	}

	req, err := http.NewRequest("POST", Url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://huggingface.co/chat/login")
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.64")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.PostForm = form
	b := bytes.NewBufferString(form.Encode())
	req.Body = io.NopCloser(b)

	res, err = l.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusSeeOther && res.StatusCode != 200 {
		return fmt.Errorf("get hf-chat cookies fatal (status code: %d)", res.StatusCode)
	} else if res.StatusCode == http.StatusOK {
		return nil
	}

	location := res.Header.Get("Location")
	res, err = l.Client.Get(location)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusFound {
		return fmt.Errorf("get hf-chat cookie fatal (status code: %d)", res.StatusCode)
	}

	return nil
}

func (l *Login) Login() error {
	if err := l.SigninWithEmail(); err != nil {
		return err
	}

	location, err := l.GetAuthURL()
	if err != nil {
		return err
	}

	if err := l.GrantAuth(location); err != nil {
		return err
	}

	u, _ := url.Parse("https://huggingface.co/")
	l.Cookies = l.Client.Jar.Cookies(u)
	return nil
}

func main() {
	email := "gormeet711@gmail.com"
	password := "Ganesh@123"

	login := NewLogin(email, password, "usercookies.json")
	if err := login.Login(); err != nil {
		fmt.Println("Login failed:", err)
		return
	}

	if err := login.SaveCookies(); err != nil {
		fmt.Println("Failed to save cookies:", err)
		return
	}

	fmt.Println("Login successful. Cookies saved.")
}
