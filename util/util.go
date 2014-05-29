package util

import (

	// go
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	// local
	"gnvm/util/curl"
	. "gnvm/util/p"
)

const (
	NODE    = "node.exe"
	DIVIDE  = "\\"
	SHASUMS = "SHASUMS.txt"
	UNKNOWN = "unknown"
)

var GlobalNodePath string

func init() {

	// try catch
	defer func() {
		if err := recover(); err != nil {
			Error(ERROR, "initialize gnvm.exe an error has occurred. please check. \nError: ", err)
			os.Exit(0)
		}
	}()

	GlobalNodePath = getGlobalNodePath()
}

func ConverFloat(str string) (float64, error) {
	args := strings.Split(str, ".")
	var newStr string
	for k, v := range args {
		if k == 0 {
			newStr = string(v) + "."
		} else {
			newStr = newStr + string(v)
		}
	}
	version, err := strconv.ParseFloat(newStr, 64)
	return version, err
}

func GetNodeVersion(path string) (string, error) {
	var newout string
	out, err := exec.Command(path+"node", "--version").Output()
	//string(out[:]) bytes to string
	if err == nil {
		// replace \r\n
		newout = strings.Replace(string(string(out[:])[1:]), "\r\n", "", -1)
	}
	return newout, err
}

func GetLatestVersion(url string) string {

	var version string

	// curl
	code, res, _ := curl.Get(url)
	if code != 0 {
		return ""
	}
	// close
	defer res.Body.Close()

	latestVersion := func(content string, line int) bool {
		if content != "" && line == 1 {

			args1 := strings.Split(content, "  ")
			if len(args1) < 2 {
				P(ERROR, "URL %v format error, please change registry. See '%v'.\n", url, "gnvm help config")
				return true
			}

			args2 := strings.Split(args1[1], "-")
			if len(args2) < 2 {
				P(ERROR, "URL %v format error, please change registry. See '%v'.\n", url, "gnvm help config")
				return true
			}

			if len(args2[1]) < 2 {
				P(ERROR, "URL %v format error, please change registry. See '%v'.\n", url, "gnvm help config")
				return true
			}

			version = args2[1][1:]
		}

		return false

	}

	if err := curl.ReadLine(res.Body, latestVersion); err != nil && err != io.EOF {
		P(ERROR, "%v Error: %v\n", "gnvm update latest", err)
	}

	return version

	/*
	for {
		// set line
		line, err := buff.ReadString('\n')

		if line != "" {

			args1 := strings.Split(line, "  ")
			if len(args1) < 2 {
				P(ERROR, "URL %v format error, please change registry. See 'gnvm help config'.\n", url)
				break
			}

			args2 := strings.Split(args1[1], "-")
			if len(args2) < 2 {
				P(ERROR, "URL %v format error, please change registry. See 'gnvm help config'.\n", url)
				break
			}

			if len(args2[1]) < 2 {
				P(ERROR, "URL %v format error, please change registry. See 'gnvm help config'.\n", url)
				break
			}

			// set version
			version = args2[1][1:]
			break
		}

		// when EOF or err break
		if err != nil || err == io.EOF {
			break
		}

	}
	*/

}

func VerifyNodeVersion(version string) bool {
	result := true
	if version == UNKNOWN {
		return true
	}
	arr := strings.Split(version, ".")
	if len(arr) != 3 {
		return false
	}
	for _, v := range arr {
		_, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			result = false
			break
		}
	}
	return result
}

func EqualAbs(key, value string) string {
	if strings.EqualFold(value, key) && value != key {
		P(WARING, "current value is %v, please use %v.\n", value, key)
		value = key
	}
	return value
}

func getGlobalNodePath() string {
	var path string
	file, err := exec.LookPath(NODE)
	if err != nil {
		path = getCurrentPath()
	} else {
		// relpace "\\node.exe"
		path = strings.Replace(file, DIVIDE+NODE, "", -1)
	}

	// gnvm.exe and node.exe the same path
	if path == "." {
		path = getCurrentPath()
	}

	return path
}

func getCurrentPath() string {
	path, err := os.Getwd()
	if err != nil {
		panic("get current path Error: " + err.Error())
	}
	return path
}