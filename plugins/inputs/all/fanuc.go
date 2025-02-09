//go:build !custom || inputs || inputs.fanuc

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/fanuc" // register plugin