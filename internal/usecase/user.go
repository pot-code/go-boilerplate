package usecase

import (
	"context"

	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.elastic.co/apm"
	"golang.org/x/crypto/bcrypt"
)

// UserUseCase ...
type UserUseCase struct {
	UserRepository domain.UserRepository
	UUIDGenerator  infra.UUIDGenerator
}

// NewUserUseCase ...
func NewUserUseCase(
	UserRepository domain.UserRepository,
	UUIDGenerator infra.UUIDGenerator,
) *UserUseCase {
	return &UserUseCase{
		UserRepository: UserRepository,
		UUIDGenerator:  UUIDGenerator,
	}
}

// SignIn sign in
func (uu *UserUseCase) SignIn(ctx context.Context, post *domain.UserModel) (*domain.UserModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "UserUseCase.Login", "service")
	defer apmSpan.End()

	post.Email = post.Username
	user, err := uu.UserRepository.FindByCredential(ctx, post)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrNoSuchUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(post.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, domain.ErrNoSuchUser
		}
		return nil, err
	}
	return user, nil
}

// SignUp create a user
func (uu *UserUseCase) SignUp(ctx context.Context, post *domain.UserModel) (*domain.UserModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "UserUseCase.Register", "service")
	defer apmSpan.End()

	ur := uu.UserRepository
	// search for existence
	if m, err := ur.FindByCredential(ctx, post); err != nil {
		return nil, err
	} else if m != nil {
		return nil, domain.ErrDuplicatedUser
	}

	// generate id
	UUIDGenerator := uu.UUIDGenerator
	if uuid, err := UUIDGenerator.Generate(); err == nil {
		post.ID = uuid
	} else {
		return nil, err
	}

	// hash password
	if password, err := bcrypt.GenerateFromPassword([]byte(post.Password), bcrypt.MinCost); err == nil {
		post.Password = string(password)
	} else {
		return nil, err
	}

	// save user
	if err := ur.SaveUser(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

// Exists find if user exists in database
func (uu *UserUseCase) Exists(ctx context.Context, post *domain.UserModel) (bool, error) {
	apmSpan, _ := apm.StartSpan(ctx, "UserUseCase.Existing", "service")
	defer apmSpan.End()

	user, err := uu.UserRepository.FindByCredential(ctx, post)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return true, nil
}
