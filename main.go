package main

import (
	"flag"
	"fmt"

	"github.com/azhai/roster/mobile"
)

var (
	datadir string
	phone   string
)

func init() {
	flag.StringVar(&datadir, "d", "./mobile", "数据库文件夹")
	flag.StringVar(&phone, "p", "15999551234", "手机号码")
	flag.Parse()
}

func main() {
	db := mobile.LoadDB(datadir)
	area, isp, err := mobile.Find(db, phone)
	if err == nil {
		fmt.Println(phone, isp)
		fmt.Println(area)
	} else {
		fmt.Println("没有找到数据")
	}
}
