package setting

import "gorm.io/gorm"

type Setting struct {
	Key   string `gorm:"primaryKey;size:128" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (Setting) TableName() string {
	return "settings"
}

func Get(db *gorm.DB, key string) (string, error) {
	var s Setting
	if err := db.Where("`key` = ?", key).First(&s).Error; err != nil {
		return "", err
	}
	return s.Value, nil
}

func Set(db *gorm.DB, key string, value string) error {
	return db.Save(&Setting{Key: key, Value: value}).Error
}

func GetMulti(db *gorm.DB, keys []string) (map[string]string, error) {
	var settings []Setting
	if err := db.Where("`key` IN ?", keys).Find(&settings).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func SetMulti(db *gorm.DB, kvMap map[string]string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for k, v := range kvMap {
			if err := tx.Save(&Setting{Key: k, Value: v}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
