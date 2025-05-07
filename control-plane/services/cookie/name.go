package cookie

import (
	"fmt"
	"github.com/google/uuid"
)

var NameGenerator = nameGenerator{}

type nameGenerator struct{}

func (nameGenerator) GenerateCookieName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString())
}
