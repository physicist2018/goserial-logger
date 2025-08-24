package domain

import (
	"os"

	"github.com/physicist2018/gomodserial-v1/pkg/utils"
	"github.com/sirupsen/logrus"
)

var DomainLogger = utils.NewLogger(logrus.InfoLevel, os.Stdout)
