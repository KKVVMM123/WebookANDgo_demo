package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFind        = gorm.ErrRecordNotFound
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

func (dao *UserDao) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli() //毫秒数
	u.Utime = now
	u.Ctime = now
	err := dao.db.Create(&u).Error //return dao.db.Create(&u).Error db.WithContext(ctx).Create(&u).Error 保持调用
	//解决邮箱冲突  与底层强耦合（mysql就是mysql pgsql就是pgsql）
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			//发生邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	//通过email查找用户
	var u User
	err := dao.db.First(&u, "email = ?", email).Error
	return u, err
}

// User 直接对应数据库表结构
// 不同叫法 entity或model或PO(persistent object)
type User struct {
	Id int64 `gorm:"primary_key,autoincrement"`
	//全部用户唯一 唯一索引
	Email    string `gorm:"unique"`
	Password string
	Ctime    int64 //time.Time //创建时间 两种写法都可以
	Utime    int64 //time.Time //更新时间
}
