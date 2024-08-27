package main

import "gorm.io/gorm"

type User struct {
	ID        int64          `gorm:"primarykey;column:id"`
	Name      string         `gorm:"column:name"`
	Email     string         `gorm:"uniqueIndex;column:email"`
	Password  string         `gorm:"column:password"`
	CreatedAt int64          `gorm:"column:created_at"`
	UpdatedAt int64          `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func (u *User) TableName() string {
	return "t_users"
}
