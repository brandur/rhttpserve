package common

func GetRemotePath(path string) string {
	return "secret:" + path
}

func Verify(path, expiresAt, signature string) bool {
	return true
}
