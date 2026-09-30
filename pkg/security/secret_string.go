package security

import "database/sql/driver"

// SecretString 是支持透明加解密的字符串字段类型。
// 通过 GORM 写入数据库时自动加密，读取时自动解密，内存中始终为明文。
type SecretString string

// String 返回明文。
func (s SecretString) String() string { return string(s) }

// Value 实现 driver.Valuer：写入数据库前加密。
func (s SecretString) Value() (driver.Value, error) {
	return Encrypt(string(s)), nil
}

// Scan 实现 sql.Scanner：从数据库读取后解密。
func (s *SecretString) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		*s = ""
	case []byte:
		*s = SecretString(Decrypt(string(v)))
	case string:
		*s = SecretString(Decrypt(v))
	}
	return nil
}
