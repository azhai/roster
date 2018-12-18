package mobile

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/azhai/roster/dataset"
	"github.com/azhai/roster/utils"
	"github.com/azhai/roster/utils/file"
)

type ISP = uint8

const (
	Unknow ISP = uint8(iota)
	ChnTelecom
	ChnMobile
	ChnUnicom
)

var IspArray = []string{"未知", "中国电信", "中国移动", "中国联通"}

func LoadDB(dir string) *dataset.DataSet {
	header := dataset.NewHeader(4, 0)
	header.SetFlags(false, dataset.DROP_U16)
	header.SetVersion("") // 当前日期
	path := filepath.Join(dir, "mobile"+dataset.FILE_EXT)
	fp, size, err := file.OpenFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if size <= 0 {
		var (
			records  []string
			keypairs []*dataset.KeyPair
		)
		records, keypairs, err = ReadMobData(dir)
		if err == nil {
			b := dataset.NewBuilder(header)
			err = b.Build(fp, records, keypairs)
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	return dataset.NewDataSet(fp, header)
}

func Find(db *dataset.DataSet, phone string) (area, isp string, err error) {
	var data, target []byte
	isp = IspArray[int(Unknow)]
	if size := len(phone); size < 7 {
		phone = phone + strings.Repeat("0", 7-size)
	}
	phone = phone[:7] + "9" //最大运营商代码

	target, err = utils.Hex2Bin(phone)
	key, addr := db.SearchIndex(target)
	if addr == nil {
		err = errors.New("Not found")
		return
	}
	data, err = db.GetRecord(addr)
	if err == nil {
		area = string(data)
		x := uint(key[len(key)-1] & 0x0f)
		isp = IspArray[int(x)]
	}
	return
}

func Mob2Bin(phone, ispName string) []byte {
	switch ispName {
	default:
		phone += string(Unknow + '0')
	case IspArray[int(ChnTelecom)]:
		phone += string(ChnTelecom + '0')
	case IspArray[int(ChnMobile)]:
		phone += string(ChnMobile + '0')
	case IspArray[int(ChnUnicom)]:
		phone += string(ChnUnicom + '0')
	}
	data, _ := utils.Hex2Bin(phone)
	return data
}

func ReadMobData(dir string) (rs []string, ks []*dataset.KeyPair, err error) {
	var lines []string
	path := filepath.Join(dir, "data/city.txt")
	lines, err = file.ReadFileLines(path)
	if err != nil {
		return
	}
	for _, line := range lines {
		ps := strings.SplitN(line, "\t", 2)
		rs = append(rs, ps[1])
	}
	path = filepath.Join(dir, "data/mobi.txt")
	lines, err = file.ReadFileLines(path)
	if err != nil {
		return
	}
	for _, line := range lines {
		ps := strings.SplitN(line, "\t", 3)
		n, _ := strconv.Atoi(ps[2])
		key := Mob2Bin(ps[0], ps[1])
		ks = append(ks, &dataset.KeyPair{Key: key, Idx: n})
	}
	return
}
