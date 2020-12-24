package model

import (
	"strings"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	osutil "github.com/cpusoft/goutil/osutil"
)

const (
	ORIGIN_RIR_AFRINIC  = "AFRINIC"
	ORIGIN_RIR_APNIC    = "APNIC"
	ORIGIN_RIR_ARIN     = "ARIN"
	ORIGIN_RIR_LACNIC   = "LACNIC"
	ORIGIN_RIR_RIPE_NCC = "RIPE NCC"
)

// from rir(tal)->repo
type OriginModel struct {
	Rir  string `json:"rir"`
	Repo string `json:"repo"`
}

func JudgeOrigin(filePath string) (originModel OriginModel) {
	/*
		ca.rg.net
		rpki-repository.nic.ad.jp
		rpki.rand.apnic.net
		krill.heficed.net
		rpki.admin.freerangecloud.com
		rpki.ripe.net
		repository.lacnic.net
		rpki.afrinic.net
		rpki.tools.westconnect.ca
		repository.rpki.rocks
		rpki.apnic.net
		rpki-as0.apnic.net
		rpkica.mckay.com
		rpki.arin.net
		rpkica.twnic.tw
		rpki-ca.idnic.net
		rpki.cnnic.cn
		rsync.rpki.nlnetlabs.nl
		rpki-repo.registro.br
		rpki.qs.nu
		repo-rpki.idnic.net
		sakuya.nat.moe
	*/
	var rir string
	var repo string
	if strings.Index(filePath, "ca.rg.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "ca.rg.net"
	} else if strings.Index(filePath, "rpki-repository.nic.ad.jp") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki-repository.nic.ad.jp"
	} else if strings.Index(filePath, "rpki.rand.apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.rand.apnic.net"
	} else if strings.Index(filePath, "krill.heficed.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.heficed.net"
	} else if strings.Index(filePath, "rpki.admin.freerangecloud.com/repo/FRC-CA/0/") > 0 {
		// /0-->arin
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.admin.freerangecloud.com"
	} else if strings.Index(filePath, "rpki.admin.freerangecloud.com/repo/FRC-CA/1/") > 0 {
		// /1-->ripe ncc
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.admin.freerangecloud.com"
	} else if strings.Index(filePath, "rpki.ripe.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.ripe.net"
	} else if strings.Index(filePath, "repository.lacnic.net") > 0 {
		rir = ORIGIN_RIR_LACNIC
		repo = "repository.lacnic.net"
	} else if strings.Index(filePath, "rpki.afrinic.net") > 0 {
		rir = "AFRINIC"
		repo = "rpki.afrinic.net"
	} else if strings.Index(filePath, "rpki.tools.westconnect.ca") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.tools.westconnect.ca"
	} else if strings.Index(filePath, "repository.rpki.rocks") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "repository.rpki.rocks"
	} else if strings.Index(filePath, "rpki.apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.apnic.net"
	} else if strings.Index(filePath, "rpkica.mckay.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpkica.mckay.com"
	} else if strings.Index(filePath, "rpki.arin.net") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.arin.net"
	} else if strings.Index(filePath, "rpkica.twnic.tw") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpkica.twnic.tw"
	} else if strings.Index(filePath, "rpki-ca.idnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki-ca.idnic.net"
	} else if strings.Index(filePath, "rpki.cnnic.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.cnnic.cn"
	} else if strings.Index(filePath, "rsync.rpki.nlnetlabs.nl") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rsync.rpki.nlnetlabs.nl"
	} else if strings.Index(filePath, "rpki-repo.registro.br") > 0 {
		rir = ORIGIN_RIR_LACNIC
		repo = "rpki-repo.registro.br"
	} else if strings.Index(filePath, "rpki.qs.nu") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.qs.nu"
	} else if strings.Index(filePath, "rpki-as0.apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki-as0.apnic.net"
	} else if strings.Index(filePath, "repo-rpki.idnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "repo-rpki.idnic.net"
	} else if strings.Index(filePath, "sakuya.nat.moe") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "sakuya.nat.moe"
	} else {
		rir = "unknown"
		if strings.Index(filePath, "afrinic.net") > 0 {
			rir = ORIGIN_RIR_AFRINIC
		} else if strings.Index(filePath, "apnic.net") > 0 {
			rir = ORIGIN_RIR_APNIC
		} else if strings.Index(filePath, "arin.net") > 0 {
			rir = ORIGIN_RIR_ARIN
		} else if strings.Index(filePath, "lacnic.net") > 0 {
			rir = ORIGIN_RIR_LACNIC
		} else if strings.Index(filePath, "ripe.net") > 0 {
			rir = ORIGIN_RIR_RIPE_NCC
		}

		tmp := strings.Replace(filePath, conf.VariableString("rsync::destPath")+osutil.GetPathSeparator(), "", -1)
		tmp = strings.Replace(tmp, conf.VariableString("rrdp::destPath")+osutil.GetPathSeparator(), "", -1)
		split := strings.Split(tmp, osutil.GetPathSeparator())
		if len(split) == 0 {
			repo = filePath
		} else {
			repo = split[0]
		}
	}
	originModel = OriginModel{Rir: rir, Repo: repo}
	belogs.Debug("JudgeOrigin(): filePath:", filePath, "   originModel:", originModel)
	return originModel
}
