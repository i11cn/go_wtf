package wtf

import (
	"strconv"
    "os"
)

type (
	Convert string
)

func (s Convert) ToInt() (int, error) {
	ret, err := strconv.ParseInt((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(ret), nil
}

func (s Convert) ToInt64() (int64, error) {
	ret, err := strconv.ParseInt((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (s Convert) ToUInt() (uint, error) {
	ret, err := strconv.ParseUint((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(ret), nil
}

func (s Convert) ToUInt64() (uint64, error) {
	ret, err := strconv.ParseUint((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (s Convert) ToFloat() (float64, error) {
	ret, err := strconv.ParseFloat((string)(s), 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (s Convert) ToBool() (bool, error) {
	ret, err := strconv.ParseBool((string)(s))
	if err != nil {
		return false, err
	}
	return ret, nil
}

func file_exist(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
        return true
    } else if os.IsExist(err) {
		return true
	} else {
		return false
	}
}