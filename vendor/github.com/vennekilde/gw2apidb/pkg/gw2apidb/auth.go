package gw2apidb

import (
	"strconv"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
)

// FindLastUpdatedAPIKey fetches the last updated apikey token from the database
func FindLastUpdatedAPIKey(ignoreOlderThan int) (tokeninfo gw2api.TokenInfo, err error) {
	//For some reason, GORM complains if the interval is provided as a paramter. It will say that 1 parameter was provided, when 0 was required.
	//Anyway, in this case, putting it directly into the query should be fine, as it is an integer
	result := orm.DB().Order("db_updated").Where("last_success >= db_updated - interval '" + strconv.Itoa(ignoreOlderThan) + " seconds' OR last_success IS NULL").First(&tokeninfo)
	return tokeninfo, result.Error
}
