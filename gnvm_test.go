package main

import (
	"fmt"
	"gnvm/nodehandle"
	"gnvm/util"
	"testing"
)

func TestCurl(t *testing.T) {
	//testSearch()
	//testNodist()
	//testNPManage()
	//testGetNPMVer()
	testIsDirExist()
	//testArch()
	//testVaildPath()
}

func testSearch() {
	nodehandle.Search("x.x.x")
	nodehandle.Search("0.10.x")
	nodehandle.Search("5.x.x")
	nodehandle.Search("5.0.0")
	nodehandle.Search(`/^5(\.([0]|[1-9]\d?)){2}$/`)
	nodehandle.Search("latest")
	nodehandle.Search("1.x.x")
	nodehandle.Search("1.1.x")
	nodehandle.Search("3.x.x")
	nodehandle.Search("3.3.x")
}

func testNodist() {
	if nl, err, code := nodehandle.New("https://cdn.npmmirror.com/binaries/iojs/index.json", nil); err != nil {
		fmt.Println(err)
		fmt.Println(code)
	} else {
		nl.Detail(0)
	}
}

func testNPManage() {
	name := `v3.8.5.zip`
	npm := new(nodehandle.NPMange)
	npm.New().CleanAll()
	npm.SetZip(name)
	npm.Unzip()
	npm.Install()
	fmt.Println(npm)
}

func testGetNPMVer() {
	url := "https://cdn.npmmirror.com/binaries/node/index.json"
	ver := "5.9.0"
	if nd, err := nodehandle.FindNodeDetailByVer(url, ver); err == nil {
		fmt.Println(nd)
	}
}

func testIsDirExist() {
	// empty
	fmt.Println(util.IsDirExist(""))
	// no exist
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools\nodejss`))
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools\nodejs\node_moduless\npm`))
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools\nodejs\5.1.1\node.exe`))
	// exist
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools\nodejs`))
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools\nodejs\node_modules`))
	// not valid path
	fmt.Println(util.IsDirExist("gnvm"))
	// exist
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools\nodejs\5.9.1\node.exe`))
	fmt.Println(util.IsDirExist(`C:\Users\Kenshin\Documents\DevTools`, `nodejs`, `5.1.1`))
}

func testArch() {
	bit32, _ := util.Arch(`C:\Users\Kenshin\Documents\DevTools\nodejs\5.1.1-x86\node.exe`)
	fmt.Println(bit32)
	bit64, _ := util.Arch(`C:\Users\Kenshin\Documents\DevTools\nodejs\5.1.1\node.exe`)
	fmt.Println(bit64)
}

func testVaildPath() {
	path := `C:\Users\Kenshin\Documents\DevTools\nodejs\5.1.1-x86\`
	util.FormatPath(&path)
	fmt.Println(path)
}
