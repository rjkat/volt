package migrate

import (
	"github.com/pkg/errors"

	"github.com/rjkat/volt/lockjson"
	"github.com/rjkat/volt/transaction"
)

func init() {
	m := &lockjsonMigrater{}
	migrateOps[m.Name()] = m
}

type lockjsonMigrater struct{}

func (*lockjsonMigrater) Name() string {
	return "lockjson"
}

func (m *lockjsonMigrater) Description(brief bool) string {
	if brief {
		return "converts old lock.json format to the latest format"
	}
	return `Usage
  volt migrate [-help] ` + m.Name() + `

Description
  Perform migration of $VOLTPATH/lock.json, which means volt converts old version lock.json structure into the latest version. This is always done automatically when reading lock.json content. For example, 'volt get <repos>' will install plugin, and migrate lock.json structure, and write it to lock.json after all. so the migrated content is written to lock.json automatically.
  But, for example, 'volt list' does not write to lock.json but does read, so every time when running 'volt list' shows warning about lock.json is old.
  To suppress this, running this command simply reads and writes migrated structure to lock.json.`
}

func (*lockjsonMigrater) Migrate() (err error) {
	// Read lock.json
	lockJSON, err := lockjson.ReadNoMigrationMsg()
	if err != nil {
		return errors.Wrap(err, "could not read lock.json")
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		return
	}
	defer func() {
		if e := trx.Done(); e != nil {
			err = e
		}
	}()

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return errors.Wrap(err, "could not write to lock.json")
	}
	return
}
