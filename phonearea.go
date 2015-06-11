package phoneareago

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"code.google.com/p/mahonia"
	"github.com/lunny/nodb"
	"github.com/lunny/nodb/config"
)

var (
	gbkDecoder = mahonia.NewDecoder("gbk")
)

//1330000,广西 南宁,中国电信 CDMA,530000,0771
type Area struct {
	Id          int64
	PhonePrefix string
	Province    string
	City        string
	Provider    string
	Model       string
	PostCode    string
	AreaCode    string
}

var (
	dbPath string
	db     *nodb.DB
)

func Init(textPath, sdbPath string) error {
	dbPath = sdbPath
	if IsDBExist() {
		var err error
		db, err = openDB(dbPath)
		return err
	}
	return GenerateDB(textPath, dbPath)
}

func openDB(dbPath string) (*nodb.DB, error) {
	cfg := config.NewConfigDefault()
	cfg.DataDir = dbPath
	ndb, err := nodb.Open(cfg)
	if err != nil {
		return nil, err
	}
	db, err := ndb.Select(0)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GenerateDB(textPath, dbPath string) error {
	f, err := os.Open(textPath)
	if err != nil {
		return err
	}
	defer f.Close()

	db, err = openDB(dbPath)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		err = saveLine(scanner.Text(), db)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	KeyPrefixIncr = "prefix_incr"
	KeyPrefix     = "p:%s"
	KeyProvince   = "pv:%d"
	KeyCity       = "ct:%d"
	KeyProvider   = "pd:%d"
	KeyMode       = "md:%d"
	KeyPostCode   = "pc:%d"
	KeyAreaCode   = "ac:%d"
)

//1,"1330000","广西 南宁市","电信CDMA卡","0771","530000"
func saveLine(line string, db *nodb.DB) error {
	line = strings.Replace(line, `"`, "", -1)
	fmt.Println(line)
	fields := strings.Split(line, ",")
	if len(fields) != 6 {
		return fmt.Errorf("此行数据字段总数不是5: %s", line)
	}

	fields = fields[1:]
	pc := strings.Fields(fields[1])
	if len(pc) == 1 {
		pc = append(pc, pc[0])
	}
	if len(pc) != 2 {
		return fmt.Errorf("此行数据中省份，城市不对: %s", line)
	}

	i, err := db.Incr([]byte(KeyPrefixIncr))
	if err != nil {
		return err
	}

	// 手机前缀
	db.Set([]byte(fmt.Sprintf(KeyPrefix, fields[0])), nodb.PutInt64(i))
	// 省
	db.Set([]byte(fmt.Sprintf(KeyProvince, i)), []byte(pc[0]))
	// 城市
	db.Set([]byte(fmt.Sprintf(KeyCity, i)), []byte(pc[1]))
	// 供应商
	//db.Set([]byte(fmt.Sprintf(KeyProvider, i)), []byte(pd[0]))
	// 制式
	db.Set([]byte(fmt.Sprintf(KeyMode, i)), []byte(fields[2]))
	// 邮编
	db.Set([]byte(fmt.Sprintf(KeyPostCode, i)), []byte(fields[3]))
	// 区号
	db.Set([]byte(fmt.Sprintf(KeyAreaCode, i)), []byte(fields[4]))

	return nil
}

func IsDBExist() bool {
	if len(dbPath) == 0 {
		return false
	}

	f, err := os.Stat(dbPath)
	if err != nil {
		return false
	}
	return f.IsDir()
}

var (
	phoneReg = regexp.MustCompile("^(13[0-9]|14[0-9]|15[0-9]|18[0-9])\\d{8}$")
)

// 判断是否是中国电话号码，并且返回纯号码
func isPhoneNum(phonenum string) (string, bool) {
	phonenum = strings.Replace(phonenum, " ", "", -1)

	if len(phonenum) < 11 {
		return "", false
	}

	if len(phonenum) == 11 {
		if phoneReg.MatchString(phonenum) {
			return phonenum, true
		}
		return "", false
	}

	prefix := phonenum[:len(phonenum)-11]
	if prefix == "86" || prefix == "086" || prefix == "+86" {
		if phoneReg.MatchString(phonenum[len(prefix):]) {
			return phonenum[len(prefix):], true
		}
	}

	return "", false
}

func GetString(db *nodb.DB, key string) (string, error) {
	res, err := db.Get([]byte(key))
	return string(res), err
}

func Convert() {

}

func Query(phoneNum string) (*Area, error) {
	if !IsDBExist() {
		return nil, errors.New("数据库没有初始化")
	}

	newPhoneNum, ok := isPhoneNum(phoneNum)
	if !ok {
		return nil, errors.New("手机号码格式不正确")
	}

	// 手机前缀
	i, err := nodb.Int64(db.Get([]byte(fmt.Sprintf(KeyPrefix, newPhoneNum[:7]))))
	if err != nil {
		return nil, err
	}

	var area Area
	// 省
	area.Province, err = GetString(db, fmt.Sprintf(KeyProvince, i))
	if err != nil {
		return nil, err
	}

	// 城市
	area.City, err = GetString(db, fmt.Sprintf(KeyCity, i))
	if err != nil {
		return nil, err
	}

	// 供应商
	/*area.Provider, err = GetString(db, fmt.Sprintf(KeyProvider, i))
	if err != nil {
		return nil, err
	}*/

	// 制式
	area.Model, err = GetString(db, fmt.Sprintf(KeyMode, i))
	if err != nil {
		return nil, err
	}

	// 邮编
	area.PostCode, err = GetString(db, fmt.Sprintf(KeyPostCode, i))
	if err != nil {
		return nil, err
	}

	// 区号
	area.AreaCode, err = GetString(db, fmt.Sprintf(KeyAreaCode, i))
	if err != nil {
		return nil, err
	}

	return &area, nil
}
