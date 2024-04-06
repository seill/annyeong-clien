package main

import (
	"annyeong-clien/service"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"os"
)

var seleniumService *selenium.Service
var driver selenium.WebDriver

func main() {
	var err error
	var command string
	var id, password *string

	if len(os.Args) < 2 {
		help()
		return
	}

	command = os.Args[1]

	if len(os.Args) > 2 {
		id = &os.Args[2]
	}

	if len(os.Args) > 3 {
		password = &os.Args[3]
	}

	err = initDriver()
	if nil != err {
		println(err.Error())
	}
	defer func(seleniumService *selenium.Service) {
		_ = seleniumService.Stop()
	}(seleniumService)

	crawlerService := service.NewCrawlerService("https://www.clien.net/service", driver)
	switch command {
	case "archive":
		err = crawlerService.Archive(id, password)
	case "delete":
		err = crawlerService.Delete(id, password)
	default:
		help()
	}

	if nil != err {
		println(err.Error())
	}
}

func initDriver() (err error) {
	var caps selenium.Capabilities

	// initialize a Chrome browser instance on port 4444
	seleniumService, err = selenium.NewChromeDriverService("./chromedriver", 4444)
	if nil != err {
		return
	}

	// configure the browser options
	caps = selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"--headless-new", // comment out this line for testing
	}})

	// create a new remote client with the specified options
	driver, err = selenium.NewRemote(caps, "")
	if nil != err {
		return
	}

	// maximize the current window to avoid responsive rendering
	err = driver.MaximizeWindow("")

	return
}

func help() {
	println("Usage: annyeong-clien [command] [id] [pasword]")
	println("command: mandatory")
	println("  archive - archive all your articles")
	println("  delete - delete all your articles")
	println("id: optional")
	println("password: optional")
}
