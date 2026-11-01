package bootstrap

import (
	"os"

	vukem "github.com/AlexGalhardo/vukem"
	vconfig "github.com/AlexGalhardo/vukem/config"

	"github.com/AlexGalhardo/hexascore-app/database/seeders"
)

// devSeeders are the load-test seeders; production never registers them. An unreadable .env counts as production.
func devSeeders() []vukem.AppSeeder {
	src, err := vconfig.LoadSource(os.DirFS("."), os.LookupEnv, func(string) ([]byte, error) { return nil, nil })
	if err != nil || src.Env() == vconfig.Production {
		return nil
	}

	return []vukem.AppSeeder{&seeders.CommunityLoad{}}
}
