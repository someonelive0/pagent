package utils

import (
	"fmt"
)

// Global version infomation

const (
	APP_VERSION = "3.0.1"

	// date +%FT%T%z  // date +'%Y%m%d'
	BUILD_TIME = "2024-03-31T00:00:00+0800"

	// go version
	GO_VERSION = "1.21.0"

	IDSS_BANNER = `
	╔══╗╔═══╗╔═══╗╔═══╗
	╚╣╠╝╚╗╔╗║║╔═╗║║╔═╗║
	 ║║  ║║║║║╚══╗║╚══╗
	 ║║  ║║║║╚══╗║╚══╗║
	╔╣╠╗╔╝╚╝║║╚═╝║║╚═╝║
	╚══╝╚═══╝╚═══╝╚═══╝	
`
)

func Version(app string) string {
	return fmt.Sprintf("app=%s\nversion=%s\nbuild_time=%s\ngo_version=%s",
		app, APP_VERSION, BUILD_TIME, GO_VERSION)
}

func VersionJson(app string) string {
	return fmt.Sprintf(`{"app": "%s", "version": "%s", "build_time": "%s", "go_version": "%s"}`,
		app, APP_VERSION, BUILD_TIME, GO_VERSION)
}

func ShowBanner() {
	fmt.Printf("%s\n", IDSS_BANNER)
	fmt.Printf("流量转发代理 pagent %s  Copyright (C) 2024 IDSS\n", APP_VERSION)
}

func ShowBannerForApp(app, version, build_time string) {
	fmt.Printf("%s\n", IDSS_BANNER)
	fmt.Printf("流量转发代理 pagent 3.0  Copyright (C) 2024 IDSS\n")
	fmt.Printf("%s version %s, build on %s\n\n", app, version, build_time)
}
