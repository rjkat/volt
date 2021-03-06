package migrate

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rjkat/volt/lockjson"
	"github.com/rjkat/volt/logger"
	"github.com/rjkat/volt/pathutil"
	"github.com/rjkat/volt/plugconf"
	"github.com/rjkat/volt/subcmd/builder"
	"github.com/rjkat/volt/transaction"
)

func init() {
	m := &plugconfConfigMigrater{}
	migrateOps[m.Name()] = m
}

type plugconfConfigMigrater struct{}

func (*plugconfConfigMigrater) Name() string {
	return "plugconf/config-func"
}

func (m *plugconfConfigMigrater) Description(brief bool) string {
	if brief {
		return "converts s:config() function name to s:on_load_pre() in all plugconf files"
	}
	return `Usage
  volt migrate [-help] ` + m.Name() + `

Description
  Perform migration of the function name of s:config() functions in plugconf files of all plugins. All s:config() functions are renamed to s:on_load_pre().
  "s:config()" is a old function name (see https://github.com/rjkat/volt/issues/196).
  All plugconf files are replaced with new contents.`
}

func (*plugconfConfigMigrater) Migrate() (err error) {
	// Read lock.json
	lockJSON, err := lockjson.ReadNoMigrationMsg()
	if err != nil {
		err = errors.Wrap(err, "could not read lock.json")
		return
	}

	results, parseErr := plugconf.ParseMultiPlugconf(lockJSON.Repos)
	if parseErr.HasErrs() {
		logger.Error("Please fix the following errors before migration:")
		for _, e := range parseErr.Errors().Errors {
			for _, line := range strings.Split(e.Error(), "\n") {
				logger.Errorf("  %s", line)
			}
		}
		err = nil
		return
	}

	type plugInfo struct {
		path    string
		content []byte
	}
	infoList := make([]plugInfo, 0, len(lockJSON.Repos))

	// Collects plugconf infomations and check errors
	results.Each(func(reposPath pathutil.ReposPath, info *plugconf.ParsedInfo) {
		if !info.ConvertConfigToOnLoadPreFunc() {
			return // no s:config() function
		}
		content, err := info.GeneratePlugconf()
		if err != nil {
			logger.Errorf("Could not generate converted plugconf: %s", err)
			return
		}
		infoList = append(infoList, plugInfo{
			path:    reposPath.Plugconf(),
			content: content,
		})
	})

	// After checking errors, write the content to files
	for _, info := range infoList {
		os.MkdirAll(filepath.Dir(info.path), 0755)
		err = ioutil.WriteFile(info.path, info.content, 0644)
		if err != nil {
			return
		}
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

	// Build ~/.vim/pack/volt dir
	err = builder.Build(false)
	if err != nil {
		err = errors.Wrap(err, "could not build "+pathutil.VimVoltDir())
		return
	}

	return
}
