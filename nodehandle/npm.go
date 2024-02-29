package nodehandle

import (

	// lib
	"compress/gzip"

	. "github.com/Kenshin/cprint"
	"github.com/Kenshin/curl"
	"github.com/bitly/go-simplejson"

	// go
	"archive/tar"
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	// local
	"gnvm/config"
	"gnvm/util"
)

const (
	LATNPMURL  = "https://registry.npmjs.org/npm/package.json"
	NPMTAOBAO  = "https://cdn.npmmirror.com/binaries/npm/"
	NPMHUAWEI  = "https://mirrors.huaweicloud.com/npm-software/"
	NPMDEFAULT = "https://registry.npmjs.org/npm/-/"
	ZIP        = ".zip"
	TGZ        = ".tgz"
)

/*
- root:     config.GetConfig(config.NODEROOT)
- zipname:  v3.8.5.zip
- ziproot:  <v3.8.5.zip>/<root_folder>
- zippath:  /<root>/v3.8.5.zip
- tgzname:  npm-6.4.1.tgz
- tgzroot:  <npm-6.4.1.tgz>/<root_folder>
- tgzpath:  /<root>/npm-6.4.1.tgz
- modules:  /<root>/node_modules
- npmpath:  /<root>/node_modules/npm
- npmbin:   /<root>/node_modules/npm/bin
- command1: npm
- command2: npm.cmd
*/
type NPMange struct {
	root     string
	zipname  string
	ziproot  string
	zippath  string
	tgzname  string
	tgzroot  string
	tgzpath  string
	modules  string
	npmpath  string
	npmbin   string
	command1 string
	command2 string
}

var npm = new(NPMange)

/*
Create NPMange
*/
func (this *NPMange) New() *NPMange {
	this.root = config.GetConfig(config.NODEROOT)
	this.modules = this.root + util.DIVIDE + "node_modules"
	this.npmpath = this.modules + util.DIVIDE + util.NPM
	this.npmbin = this.npmpath + util.DIVIDE + "bin"
	this.command1 = "npm"
	this.command2 = "npm.cmd"
	return this
}

/*
Set zip info
*/
func (this *NPMange) SetZip(zip string) {
	this.zipname = zip
	this.zippath = this.root + util.DIVIDE + this.zipname
}

/*
Set tgz info
*/
func (this *NPMange) SetTgz(tgz string) {
	this.tgzname = tgz
	this.tgzpath = this.root + util.DIVIDE + this.tgzname
}

/*
Custom Print
*/
func (this *NPMange) String() string {
	s := fmt.Sprintf("root     = %v\n", this.root)
	s += fmt.Sprintf("zipname  = %v\n", this.zipname)
	s += fmt.Sprintf("ziproot  = %v\n", this.ziproot)
	s += fmt.Sprintf("zippath  = %v\n", this.zippath)
	s += fmt.Sprintf("tgzname  = %v\n", this.tgzname)
	s += fmt.Sprintf("tgzroot  = %v\n", this.tgzroot)
	s += fmt.Sprintf("tgzpath  = %v\n", this.tgzpath)
	s += fmt.Sprintf("modules  = %v\n", this.modules)
	s += fmt.Sprintf("npmpath  = %v\n", this.npmpath)
	s += fmt.Sprintf("npmbin   = %v\n", this.npmbin)
	s += fmt.Sprintf("command1 = %v\n", this.command1)
	s += fmt.Sprintf("command2 = %v", this.command2)
	return s
}

/*
Create node_modules folder
*/
func (this *NPMange) CreateModules() {
	if !util.IsDirExist(this.modules) {
		if err := os.Mkdir(this.modules, 0755); err != nil {
			P(ERROR, "create %v foler error, Error: %v\n", this.modules, err.Error())
		} else {
			P(NOTICE, "%v folder create success.\n", this.modules)
		}
	}
}

/*
Download npm zip

Param:
  - url:  download url
  - name: download file name

Return:
  - error
*/
func (this *NPMange) Download(url, name string) error {
	curl.Options.Header = false
	curl.Options.Footer = false
	if _, errs := curl.New(url, name, name, this.root); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

/*
Unzip file

Return:
  - error
  - code
  - -1: open  zip file error
  - -2: open  file error
  - -3: write file error
  - -4: copy  file error
*/
func (this *NPMange) Unzip() (int, error) {
	path, dest := this.zippath, this.modules
	unzip, err := zip.OpenReader(path)
	if err != nil {
		return -1, err
	}
	defer unzip.Close()

	extractAndWriteFile := func(file *zip.File, idx int) (int, error) {
		rc, err := file.Open()
		if err != nil {
			return -2, err
		}
		defer rc.Close()
		if idx == 0 {
			this.ziproot = strings.Replace(file.Name, "/", "", -1)
		}
		path = filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return -3, err
			}
			defer f.Close()
			if _, err := io.Copy(f, rc); err != nil {
				return -4, err
			}
		}
		return 0, nil
	}

	idx := 0
	for _, file := range unzip.File {
		if code, err := extractAndWriteFile(file, idx); err != nil {
			return code, err
		}
		idx++
	}
	return 0, nil
}

/*
Untgz file

Return:
  - error
  - code
  - -1: open  tgz file error
  - -2: open  file error
  - -3: write file error
  - -4: copy  file error
*/
func (this *NPMange) Untgz() (int, error) {
	path, dest := this.tgzpath, this.modules+"\\"
	srcFile, err := os.Open(path)
	if err != nil {
		return -2, err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return -1, err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	idx := 0
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return -2, err
			}
		}
		//判断tgz文件根目录名称
		if idx == 0 {
			this.tgzroot = strings.Split(hdr.Name, "/")[0]
		}
		//拼接文件名
		filename := dest + hdr.Name
		// filename = strings.Replace(filename, "package", "npm", -1)
		//创建目录
		err = os.MkdirAll(string([]rune(filename)[0:strings.LastIndex(filename, "/")]), hdr.FileInfo().Mode())
		if err != nil {
			return -3, err
		}
		//创建文件
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
		if err != nil {
			return -3, err
		}
		defer file.Close()
		//复制文件
		if _, err := io.Copy(file, tr); err != nil {
			return -4, err
		}
		idx++
	}
	return 0, nil
}

/*
Rename <root>\node_modules\folder to <root>\node_modules\npm
Copy <root>\node_modules\npm\bin\ npm and npm.cmd to <root>\
*/
func (this *NPMange) Install() error {
	//判断是否tgz包
	pkgpath := this.ziproot
	if config.GetConfig(config.REGISTRY) == util.ORIGIN_DEFAULT {
		pkgpath = this.tgzroot
	}
	//重命名文件夹
	if err := os.Rename(filepath.Join(this.modules, util.DIVIDE, pkgpath), this.npmpath); err != nil {
		P(ERROR, "rename fail, Error: %v\n", err.Error())
		return err
	} else {
		files := [2]string{this.command1, this.command2}
		for _, v := range files {
			if err := util.Copy(this.npmbin, this.root, v); err != nil {
				P(ERROR, "copy %v to %v faild, Error: %v \n", this.npmbin, this.root)
				return err
			}
		}
	}
	return nil
}

/*
Remove file

Param:
  - path: olny clude path
  - <root>/node_modules/npm
  - <root>/npm
  - <root>/npm.cmd
  - <root>/<npm.zip>
*/
func (this *NPMange) Clean(path string) error {
	if util.IsDirExist(path) {
		if err := os.RemoveAll(path); err != nil {
			P(ERROR, "remove %v folder Error: %v.\n", path, err.Error())
			return err
		}
	}
	return nil
}

/*
Remove <root>/node_modules/npm, <root>/npm, <root>/npm.cmd

Return:
  - error
*/
func (this *NPMange) CleanAll() error {
	paths := [3]string{this.npmpath, this.root + util.DIVIDE + this.command1, this.root + util.DIVIDE + this.command2}
	for _, v := range paths {
		if err := this.Clean(v); err != nil {
			return err
		}
	}
	return nil
}

/*
Install NPM
*/
func InstallNPM(version string) {
	// try catch
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("'gnvm npm %v' an error has occurred. please check. \nError: ", version)
			Error(ERROR, msg, err)
			os.Exit(0)
		}
	}()

	version = strings.ToLower(version)
	if !util.VerifyNodeVer(version) {
		P(ERROR, "'%v' param only support [%v] [%v] or %v e.g. [%v], please check your input. See '%v'.\n", "gnvm npm", "latest", "global", "valid version", "3.8.1", "gnvm help npm")
		return
	}

	prompt, local, newver := "n", getLocalNPMVer(), version

	if version == util.GLOBAL {
		newver = getNodeNpmVer()
	} else if version == util.LATEST {
		newver = getLatNPMVer()
	}

	cp := CP{Red, false, None, false, newver}
	P(NOTICE, "local    npm version is %v\n", local)
	P(NOTICE, "remote   npm version is %v\n", cp)
	P(NOTICE, "download %v version [Y/n]? ", cp)
	fmt.Scanf("%s\n", &prompt)
	prompt = strings.ToLower(prompt)
	if prompt == "y" {
		downloadNpm(newver)
	} else {
		P(NOTICE, "operation has been cancelled.")
	}
}

/*
Uninstall NPM
*/
func UninstallNPM() {
	if getLocalNPMVer() == util.UNKNOWN {
		return
	}
	if err := npm.New().CleanAll(); err == nil {
		P(DEFAULT, "Npm uninstall %v.\n", "success")
	}
}

/*
Get npm version by global( local ) node version

Return:
  - string: npm version
*/
func getNodeNpmVer() string {
	ver, err := util.GetNodeVer(rootPath)
	if err != nil {
		panic(errors.New("not exist global node.exe. please usage 'gnvm install latest -g' frist."))
	}

	url := config.GetConfig(config.REGISTRY)
	if level := util.GetNodeVerLev(util.FormatNodeVer(ver)); level == 3 {
		url = config.GetIOURL(url)
	}
	url += util.NODELIST

	nd, err := FindNodeDetailByVer(url, ver)
	if err != nil {
		panic(err)
	}
	return nd.NPM.Version
}

/*
Get Latest NPM version

Return:
  - string: latest npm version
*/
func getLatNPMVer() string {
	_, res, err := curl.Get(LATNPMURL)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		panic(err)
	}
	ver, _ := json.Get("version").String()
	return ver
}

/*
Get global( local ) NPM version

Return:
  - util.UNKNOWN: current not exist npmCmd
  - version     : current npm version
*/
func getLocalNPMVer() string {
	out, err := exec.Command(rootPath+util.NPM, "-v").Output()
	if err != nil {
		P(WARING, "current path %v not exist npm.\n", rootPath)
		return util.UNKNOWN
	}
	return strings.TrimSpace(string(out[:]))
}

/*
Download / unzip / install npm

Param:
  - ver: npm version
*/
func downloadNpm(ver string) {
	version := "v" + ver + ZIP
	url := NPMTAOBAO + version
	if config.GetConfig(config.REGISTRY) == util.ORIGIN_TAOBAO {
		url = NPMTAOBAO + version
		// create npm
		npm.New().SetZip(version)
	} else if config.GetConfig(config.REGISTRY) == util.ORIGIN_HUAWEI {
		url = NPMHUAWEI + version
		// create npm
		npm.New().SetZip(version)
	} else if config.GetConfig(config.REGISTRY) == util.ORIGIN_DEFAULT {
		version = "npm-" + ver + TGZ
		url = NPMDEFAULT + version
		// create npm
		npm.New().SetTgz(version)
	}

	P(DEFAULT, "Start download new npm version %v\n", version)

	// download
	if err := npm.Download(url, version); err != nil {
		panic(err.Error())
	}

	// create node_modules
	npm.CreateModules()

	// clean all npm files
	npm.CleanAll()

	//判断是否tgz包
	if config.GetConfig(config.REGISTRY) == util.ORIGIN_DEFAULT {
		P(DEFAULT, "Start untgz and install %v tgz file, please wait.\n", version)
		//untgz
		if _, err := npm.Untgz(); err != nil {
			msg := fmt.Sprintf("untgz %v an error has occurred. \nError: ", npm.tgzname, err.Error())
			panic(errors.New(msg))
		}
	} else {
		P(DEFAULT, "Start unzip and install %v zip file, please wait.\n", version)
		// unzip
		if _, err := npm.Unzip(); err != nil {
			msg := fmt.Sprintf("unzip %v an error has occurred. \nError: ", npm.zipname, err.Error())
			panic(errors.New(msg))
		}
	}

	// install
	if err := npm.Install(); err != nil {
		return
	}

	// remove download zip or tgz file
	if config.GetConfig(config.REGISTRY) == util.ORIGIN_DEFAULT {
		npm.Clean(npm.tgzpath)
	} else {
		npm.Clean(npm.zippath)
	}

	P(DEFAULT, "Set success, current npm version is %v.\n", ver)
}
