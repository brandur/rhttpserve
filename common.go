package rserve

func GetRemotePath(path string) string {
	return "secret:" + path
}
