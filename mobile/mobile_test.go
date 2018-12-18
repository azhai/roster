package mobile

import (
	"fmt"
	"testing"
)

var db = LoadDB("./")

func BenchmarkFindPhone(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		var i = 0
		for p.Next() {
			i++
			phone := fmt.Sprintf("%s%d%s", "1897", i&10000, "45")
			_, _, err := Find(db, phone)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func FindPhoneTest(t *testing.T, phone string) {
	area, isp, err := Find(db, phone)
	if err != nil {
		t.Fatal("没有找到数据")
	}
	t.Log(phone, isp)
	t.Log(area)
}

func TestFindPhone1(t *testing.T) {
	FindPhoneTest(t, "15999558910123123213213")
}

func TestFindPhone2(t *testing.T) {
	FindPhoneTest(t, "1300")
}

func TestFindPhone3(t *testing.T) {
	FindPhoneTest(t, "1703576")
}

func TestFindPhone4(t *testing.T) {
	FindPhoneTest(t, "199997922323")
}

func TestFindPhone5(t *testing.T) {
	_, _, err := Find(db, "afsd32323")
	if err == nil {
		t.Fatal("错误的结果")
	}
	t.Log(err)
}
