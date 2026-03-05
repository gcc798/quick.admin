package validator

import (
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	macRegex   = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)
	phoneRegex = regexp.MustCompile(`^1\d{10}$`)
	snRegex    = regexp.MustCompile(`^[A-Za-z0-9_-]{6,40}$`)
	once       sync.Once
)

func Register() {
	once.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = v.RegisterValidation("mac", validateMAC)
			_ = v.RegisterValidation("cnphone", validatePhone)
			_ = v.RegisterValidation("sn", validateSN)
		}
	})
}

func validateMAC(fl validator.FieldLevel) bool   { return macRegex.MatchString(fl.Field().String()) }
func validatePhone(fl validator.FieldLevel) bool { return phoneRegex.MatchString(fl.Field().String()) }
func validateSN(fl validator.FieldLevel) bool    { return snRegex.MatchString(fl.Field().String()) }
