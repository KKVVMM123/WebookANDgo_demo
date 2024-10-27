package service

import (
	"context"
	"errors"
	"go_demo/webook/internal/domain"
	"go_demo/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("邮箱或密码不对")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (svc *UserService) SingUp(ctx context.Context, u domain.User) error {
	//1.考虑加密放在哪里
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	//2.将其存起来
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	//调用repo层方法找到对应用户
	u, err := svc.repo.FindByEmail(ctx, email)
	//判断用户对象是否为空
	if err == repository.ErrUserNotFind {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	//比较密码 即判断数据库的密码与前端传过来的密码是否一致
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		//接入日志后要打印日志或debug
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}
