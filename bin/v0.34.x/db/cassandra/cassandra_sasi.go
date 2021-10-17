package cassandra

import (
	"fmt"
	"github.com/scylladb/gocqlx/v2"
)

const sasiMode = "PREFIX"

func initializeSasiOnKey(session gocqlx.Session, tableName string) error {
	return session.Query(fmt.Sprintf("create custom index if not exists %s_sasi_key on %s (%s) using 'org.apache.cassandra.index.sasi.SASIIndex' with options = { 'mode': '%s' };", tableName, tableName, cKey, sasiMode), []string{}).ExecRelease()
}

func initializeSasiOnHeight(session gocqlx.Session, tableName string) error {
	return session.Query(fmt.Sprintf("create custom index if not exists %s_sasi_height on %s (%s) using 'org.apache.cassandra.index.sasi.SASIIndex' with options = { 'mode': '%s' };", tableName, tableName, cAtHeight, sasiMode), []string{}).ExecRelease()
}
