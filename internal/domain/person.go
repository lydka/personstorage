package domain

type Person struct {
	ExternalID  string `json:"external_id" gorm:"column:external_id;primaryKey"`
	Name        string `json:"name" gorm:"column:name"`
	Email       string `json:"email" gorm:"column:email;uniqueIndex"`
	DateOfBirth string `json:"date_of_birth" gorm:"column:date_of_birth"`
}

func (Person) TableName() string {
	return "people"
}
