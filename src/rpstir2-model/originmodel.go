package model

import (
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
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

		ca.nat.moe  arin-rpki-ta
		cb.rg.net    ripe-ncc-ta
		chloe.sobornost.net           ripe-ncc-ta
		krill-eval-ctec.charter.com  arin-rpki-ta
		nostromo.heficed.net     ripe-ncc-ta
		rpki.admin.freerangecloud.com   ripe-ncc-ta
		rpki.apernet.io/repo/APERNET/1/       apnic-rpki-root-iana-origin
		rpki.apernet.io/repo/APERNET/0/       arin-rpki-ta
		rpki.multacom.com        arin-rpki-ta
		rpki.xindi.eu    ripe-ncc-ta
		rpki1.terratransit.de  ripe-ncc-ta

		cc.rg.net  ripe-ncc-ta
		rpki.sailx.co  arin-rpki-ta
		rpki.luys.cloud arin-rpki-ta
		rpki-rsync.mnihyc.com apnic-rpki-root-iana-origin


		rrdp.twnic.tw  ORIGIN_RIR_APNIC
		rpki.blade.sh  ORIGIN_RIR_APNIC
		rpki1.rpki-test.sit.fraunhofer.de  ripe-ncc-ta
		kube-ingress.as207960.net ripe-ncc-ta


		rpki-repo.as207960.net ripe
		rpki.dataplane.org  arin

		rpki.august.tw
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
	} else if strings.Index(filePath, "rpki.sub.apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.sub.apnic.net"
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
	} else if strings.Index(filePath, "ca.nat.moe") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "ca.nat.moe"
	} else if strings.Index(filePath, "cb.rg.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "cb.rg.net"
	} else if strings.Index(filePath, "cc.rg.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "cc.rg.net"
	} else if strings.Index(filePath, "chloe.sobornost.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "chloe.sobornost.net"
	} else if strings.Index(filePath, "krill-eval-ctec.charter.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "krill-eval-ctec.charter.com"
	} else if strings.Index(filePath, "nostromo.heficed.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "nostromo.heficed.net"
	} else if strings.Index(filePath, "rpki.admin.freerangecloud.com") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.admin.freerangecloud.com"
	} else if strings.Index(filePath, "rpki.apernet.io/repo/APERNET/1/") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.apernet.io"
	} else if strings.Index(filePath, "rpki.apernet.io/repo/APERNET/0/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.apernet.io"
	} else if strings.Index(filePath, "rpki.multacom.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.multacom.com"
	} else if strings.Index(filePath, "rpki.xindi.eu") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.xindi.eu"
	} else if strings.Index(filePath, "rpki1.terratransit.de") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki1.terratransit.de"
	} else if strings.Index(filePath, "rpki.sailx.co") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.sailx.co"
	} else if strings.Index(filePath, "rpki.luys.cloud") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.luys.cloud"
	} else if strings.Index(filePath, "rpki-rsync.mnihyc.com") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki-rsync.mnihyc.com"
	} else if strings.Index(filePath, "rrdp.twnic.tw") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rrdp.twnic.tw"
	} else if strings.Index(filePath, "rpki.blade.sh") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.blade.sh"
	} else if strings.Index(filePath, "rpki1.rpki-test.sit.fraunhofer.de") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki1.rpki-test.sit.fraunhofer.de"
	} else if strings.Index(filePath, "kube-ingress.as207960.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "kube-ingress.as207960.net"
	} else if strings.Index(filePath, "rpki-repo.as207960.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki-repo.as207960.net"
	} else if strings.Index(filePath, "rpki.dataplane.org") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.dataplane.org"
	} else if strings.Index(filePath, "magellan.ipxo.com") > 0 {
		// include r.magellan.ipxo.com
		rir = ORIGIN_RIR_ARIN
		repo = "magellan.ipxo.com"
	} else if strings.Index(filePath, "rpki.akrn.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.akrn.net"
	} else if strings.Index(filePath, "0.sb") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "0.sb"
	} else if strings.Index(filePath, "rpki.owl.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.owl.net"
	} else if strings.Index(filePath, "krill.cloud") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.cloud"
	} else if strings.Index(filePath, "rrdp.taaa.eu") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rrdp.taaa.eu"
	} else if strings.Index(filePath, "rpki-rsync.e15f.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki-rsync.e15f.net"
	} else if strings.Index(filePath, "rrdp.e15f.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rrdp.e15f.net"
	} else if strings.Index(filePath, "rpki.e15f.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.e15f.net"
	} else if strings.Index(filePath, "rpki.caramelfox.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.caramelfox.net"
	} else if strings.Index(filePath, "rpki.roa.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.roa.net"
	} else if strings.Index(filePath, "rpki-rps.arin.net") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki-rps.arin.net"
	} else if strings.Index(filePath, "rpki.august.tw") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.august.tw"
	} else if strings.Index(filePath, "rrdp-rps.arin.net") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rrdp-rps.arin.net"
	} else if strings.Index(filePath, "rpki-rrdp.us-east-2.amazonaws.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki-rrdp.us-east-2.amazonaws.com"
	} else if strings.Index(filePath, "rrdp.rp.ki") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rrdp.rp.ki"
	} else if strings.Index(filePath, "rsync.rp.ki") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rsync.rp.ki"
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
		belogs.Info("JudgeOrigin():rir is unknown, filePath:", filePath, "   rir:", rir, "  repo:", repo)
	}
	originModel = OriginModel{Rir: rir, Repo: repo}
	belogs.Debug("JudgeOrigin(): filePath:", filePath, "   originModel:", jsonutil.MarshalJson(originModel))
	return originModel
}
