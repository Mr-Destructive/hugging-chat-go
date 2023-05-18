package hugchat

import (
	"bufio"
	"encoding/hex"
	"math/rand"
	"os"
	"strings"
)

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	uuidStr := make([]byte, 36)
	hex.Encode(uuidStr[0:8], uuid[0:4])
	uuidStr[8] = '-'
	hex.Encode(uuidStr[9:13], uuid[4:6])
	uuidStr[13] = '-'
	hex.Encode(uuidStr[14:18], uuid[6:8])
	uuidStr[18] = '-'
	hex.Encode(uuidStr[19:23], uuid[8:10])
	uuidStr[23] = '-'
	hex.Encode(uuidStr[24:], uuid[10:])

	return string(uuidStr), nil
}

func LoadEnvFromFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
