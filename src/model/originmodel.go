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
	Rir       string `json:"rir"`
	Repo      string `json:"repo"`
	NotifyUrl string `json:"notifyUrl"`
}

func JudgeOrigin(filePath, notifyUrl string) (originModel *OriginModel) {
	originModel = JudgeOriginByFilePath(filePath)
	originModel.NotifyUrl = notifyUrl
	return originModel
}

func JudgeOriginByFilePath(filePath string) (originModel *OriginModel) {
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

						krill.accuristechnologies.ca  arin-rpki-ta.cer
						cloudie-repo.rpki.app/repo/CLOUDIE-RPKI/2/  arin-rpki-ta
						cloudie-repo.rpki.app/repo/SVENS-RPKI/0/   rpki-rps.arin.net
						cloudie-repo.rpki.app/repo/SVENS-RPKI/1/   rpki-rps.arin.net
						rpki.pedjoeang.group     rpki-rps.arin.net
						rpki-rsync.us-east-2.amazonaws.com   arin-rpki-ta
						repo.kagl.me   arin-rpki-ta
						rpki.zappiehost.com/repo/NORTHLAYER_UID_13864/0/  rpki-rps.arin.net
						rpki.zappiehost.com/repo/NORTHLAYER_UID_13864/1/ rpki-rps.arin.net
						rpki.zappiehost.com/repo/NORTHLAYER_UID_13864/2/
						rpki.zappiehost.com/repo/HAZEL_UID_18860/0/  rpki-rps.arin.net
						rpki.zappiehost.com/repo/TERITUM_UID_18858/0/ rpki-rps.arin.net
						rpki.zappiehost.com/repo/TERITUM_UID_18858/1/
						rpki.cc   arin-rpki-ta


				        rpki.berrybyte.network  ripe-ncc-ta
						rpki-01.pdxnet.uk ripe-ncc-ta
						krill.rayhaan.net  ripe-ncc-ta
						rpki.services.vm.n1.i.bm-x0.w420.net ripe-ncc-ta
						rpki-repository.haruue.net  ripe-ncc-ta
						rsync.roa.tohunet.com  arin
						rpki.folf.systems  ripe ncc


						dev.tw  arin
						rsync.rpki.tianhai.link rrdp.rpki.tianhai.link ripe
						rov-measurements.nlnetlabs.net  ripe
						pub.krill.ausra.cloud  ripe

						rpki.komorebi.network   rpki.komorebi.network/repo/komorebi/1/  "ARIN",
		                                        rpki.komorebi.network/repo/komorebi/0/ /3/ /4/  RIPE
		                rpki.nap.re  /home/rpki/data/rrdprepo/rpki.nap.re/repo/pongery/3/  -->   rpki.komorebi.network/repo/komorebi/1/   -->  ARIN
		                             other RIPE

		                cloudie-repo.rpki.app   /home/rpki/data/rrdprepo/cloudie-repo.rpki.app/repo/CLOUDIE-RPKI/0/  --> RIPE
		                             other ARIN
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
	} else if strings.Index(filePath, "rrdp.ripe.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rrdp.ripe.net"
	} else if strings.Index(filePath, "repository.lacnic.net") > 0 {
		rir = ORIGIN_RIR_LACNIC
		repo = "repository.lacnic.net"
	} else if strings.Index(filePath, "rrdp.lacnic.net") > 0 {
		rir = ORIGIN_RIR_LACNIC
		repo = "rrdp.lacnic.net"
	} else if strings.Index(filePath, "rpki.afrinic.net") > 0 {
		rir = ORIGIN_RIR_AFRINIC
		repo = "rpki.afrinic.net"
	} else if strings.Index(filePath, "rrdp.afrinic.net") > 0 {
		rir = ORIGIN_RIR_AFRINIC
		repo = "rrdp.afrinic.net"
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
	} else if strings.Index(filePath, "rrdp.arin.net") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rrdp.arin.net"
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
	} else if strings.Index(filePath, "rrdp-as0.apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rrdp-as0.apnic.net"
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
	} else if strings.Index(filePath, "rpki-rrdp.mnihyc.com") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki-rrdp.mnihyc.com"
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
	} else if strings.Index(filePath, "rpki.as207960.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.as207960.net"
	} else if strings.Index(filePath, "invalid.rov.koenvanhove.nl") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "invalid.rov.koenvanhove.nl"
	} else if strings.Index(filePath, "child.rov.koenvanhove.nl") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "child.rov.koenvanhove.nl"
	} else if strings.Index(filePath, "rrdp.paas.rpki.ripe.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rrdp.paas.rpki.ripe.net"
	} else if strings.Index(filePath, "parent.rov.koenvanhove.nl/repo/KoenvanHove/0/") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "parent.rov.koenvanhove.nl"
	} else if strings.Index(filePath, "parent.rov.koenvanhove.nl/repo/KoenvanHove/1/") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "parent.rov.koenvanhove.nl"
	} else if strings.Index(filePath, "cloudie-repo.rpki.app") > 0 {
		if strings.Index(filePath, "cloudie-repo.rpki.app/repo/CLOUDIE-RPKI/0/") > 0 {
			rir = ORIGIN_RIR_RIPE_NCC
			repo = "cloudie-repo.rpki.app"
		} else {
			rir = ORIGIN_RIR_ARIN
			repo = "cloudie-repo.rpki.app"
		}
	} else if strings.Index(filePath, "cloudie.rpki.app") > 0 {
		if strings.Index(filePath, "cloudie.rpki.app/repo/CLOUDIE-RPKI/0/") > 0 {
			rir = ORIGIN_RIR_RIPE_NCC
			repo = "cloudie.rpki.app"
		} else {
			rir = ORIGIN_RIR_ARIN
			repo = "cloudie.rpki.app"
		}
	} else if strings.Index(filePath, "rpki.telecentras.lt") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.telecentras.lt"
	} else if strings.Index(filePath, "rpki.zappiehost.com") > 0 {
		if strings.Index(filePath, "rpki.zappiehost.com/repo/ZAPPIE-RPKI/1/") > 0 {
			rir = ORIGIN_RIR_ARIN
			repo = "rpki.zappiehost.com"
		} else if strings.Index(filePath, "rpki.zappiehost.com/repo/ZAPPIE-RPKI/2/") > 0 {
			rir = ORIGIN_RIR_RIPE_NCC
			repo = "rpki.zappiehost.com"
		} else {
			// default is arin
			rir = ORIGIN_RIR_ARIN
			repo = "rpki.zappiehost.com"
		}
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/ZAPPIE-RPKI/3/") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "krill.accuristechnologies.ca") > 0 { //  arin-rpki-ta.cer
		rir = ORIGIN_RIR_ARIN
		repo = "krill.accuristechnologies.ca"
	} else if strings.Index(filePath, "rpki.pedjoeang.group") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.pedjoeang.group"
	} else if strings.Index(filePath, "rpki-rsync.us-east-2.amazonaws.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki-rsync.us-east-2.amazonaws.com"
	} else if strings.Index(filePath, "repo.kagl.me") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "repo.kagl.me"
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/NORTHLAYER_UID_13864/0/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/NORTHLAYER_UID_13864/1/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/NORTHLAYER_UID_13864/2/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/HAZEL_UID_18860/0/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/TERITUM_UID_18858/0/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "rpki.zappiehost.com/repo/TERITUM_UID_18858/1/") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.zappiehost.com"
	} else if strings.Index(filePath, "rpki.cc") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.cc"
	} else if strings.Index(filePath, "rpki.berrybyte.network") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.berrybyte.network"
	} else if strings.Index(filePath, "rpki-01.pdxnet.uk") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki-01.pdxnet.uk"
	} else if strings.Index(filePath, "krill.rayhaan.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.rayhaan.net"
	} else if strings.Index(filePath, "rpki.services.vm.n1.i.bm-x0.w420.net") > 0 { // https://x-0100000000000011.p.u9sv.com/notification.xml
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.services.vm.n1.i.bm-x0.w420.net"
	} else if strings.Index(filePath, "x-0100000000000011.p.u9sv.com") > 0 { // https://x-0100000000000011.p.u9sv.com/notification.xml
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "x-0100000000000011.p.u9sv.com"
	} else if strings.Index(filePath, "rpki-repository.haruue.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki-repository.haruue.net"
	} else if strings.Index(filePath, "rpki-publication.haruue.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki-publication.haruue.net"
	} else if strings.Index(filePath, "rpki.folf.systems") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.folf.systems"
	} else if strings.Index(filePath, "rsync.roa.tohunet.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rsync.roa.tohunet.com"
	} else if strings.Index(filePath, "rrdp.roa.tohunet.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rrdp.roa.tohunet.com"
	} else if strings.Index(filePath, "rrdp.apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rrdp.apnic.net"
	} else if strings.Index(filePath, "dev.tw") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "dev.tw"
	} else if strings.Index(filePath, "rpki.co") > 0 {
		if strings.Index(filePath, "rrdp.rpki.co") > 0 {
			rir = ORIGIN_RIR_ARIN
			repo = "rrdp.rpki.co"
		} else {
			rir = ORIGIN_RIR_ARIN
			repo = "rpki.co"
		}
	} else if strings.Index(filePath, "rpki.tianhai.link") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.tianhai.link"
	} else if strings.Index(filePath, "rov-measurements.nlnetlabs.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rov-measurements.nlnetlabs.net"
	} else if strings.Index(filePath, "pub.krill.ausra.cloud") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "pub.krill.ausra.cloud"
	} else if strings.Index(filePath, "rpki.nap.re") > 0 {
		if strings.Index(filePath, "rpki.nap.re/repo/pongery/3/") > 0 {
			rir = ORIGIN_RIR_ARIN
			repo = "rpki.nap.re"
		} else {
			rir = ORIGIN_RIR_RIPE_NCC
			repo = "rpki.nap.re"
		}
	} else if strings.Index(filePath, "rpki.komorebi.network") > 0 {
		if strings.Index(filePath, "rpki.komorebi.network/repo/komorebi/1/") > 0 {
			rir = ORIGIN_RIR_ARIN
			repo = "rpki.komorebi.network"
		} else {
			rir = ORIGIN_RIR_RIPE_NCC
			repo = "rpki.komorebi.network"
		}
	} else if strings.Index(filePath, "krill.stonham.info") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.stonham.info"
	} else if strings.Index(filePath, "krill.stonham.uk") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.stonham.uk"
	} else if strings.Index(filePath, "rpki.apernet.io") > 0 {
		if strings.Index(filePath, "rpki.apernet.io/repo/APERNET/1/") > 0 {
			rir = ORIGIN_RIR_APNIC
			repo = "rpki.apernet.io"
		} else if strings.Index(filePath, "rpki.apernet.io/repo/APERNET/0/") > 0 {
			rir = ORIGIN_RIR_ARIN
			repo = "rpki.apernet.io"
		} else {
			// default
			rir = ORIGIN_RIR_APNIC
			repo = "rpki.apernet.io"
		}
	} else if strings.Index(filePath, "rpki.netiface.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.netiface.net"
	} else if strings.Index(filePath, "rpki.ssmidge.xyz") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki.ssmidge.xyz"
	} else if strings.Index(filePath, "krill.ca-bc-01.ssmidge.xyz") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "krill.ca-bc-01.ssmidge.xyz"
	} else if strings.Index(filePath, "pub.rpki.win") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "pub.rpki.win"
	} else if strings.Index(filePath, "rpki.0i1.eu") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.0i1.eu"
	} else if strings.Index(filePath, "rpki.ezdomain.ru") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.ezdomain.ru"
	} else if strings.Index(filePath, "andromeda.heficed.net") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "andromeda.heficed.net"
	} else if strings.Index(filePath, "rpki.pudu.be") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.pudu.be"
	} else if strings.Index(filePath, "rpki.sn-p.io") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.sn-p.io"
	} else if strings.Index(filePath, "rpki.services.vm.i.bm-x0.w420.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.services.vm.i.bm-x0.w420.net"
	} else if strings.Index(filePath, "rpki.cernet.edu.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki.cernet.edu.cn"
	} else if strings.Index(filePath, "x-8011.p.u9sv.com") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "x-8011.p.u9sv.com"
	} else if strings.Index(filePath, "repo.rpki.space") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "repo.rpki.space"
	} else if strings.Index(filePath, "rpki.uz") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.uz"
	} else if strings.Index(filePath, "rpki.services.vm.a-x1.w420.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.services.vm.a-x1.w420.net"
	} else if strings.Index(filePath, "repodepot.wildtky.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "repodepot.wildtky.com"
	} else if strings.Index(filePath, "subrepo.wildtky.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "subrepo.wildtky.com"
	} else if strings.Index(filePath, "krill.uta.ng") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.uta.ng"
	} else if strings.Index(filePath, "rpki.nellicus.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.nellicus.net"
	} else if strings.Index(filePath, "rpki-pp.com") > 0 {
		rir = ORIGIN_RIR_ARIN
		repo = "rpki-pp.com"
	} else if strings.Index(filePath, "rpki.athene-center.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.athene-center.net"
	} else if strings.Index(filePath, "rpki-rps.cnnic.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "rpki-rps.cnnic.cn"
	} else if strings.Index(filePath, "krill.47272.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "krill.47272.net"
	} else if strings.Index(filePath, "rpki.axivora.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rpki.axivora.net"
	} else if strings.Index(filePath, "akane.maru.co.jp") > 0 {
		rir = ORIGIN_RIR_APNIC
		repo = "akane.maru.co.jp"
	} else if strings.Index(filePath, "rrdp.as214749.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
		repo = "rrdp.as214749.net"
		/*}

		else if strings.Index(filePath, "rpki.zdns.cn") > 0 {
			rir = ORIGIN_RIR_APNIC
			repo = "rpki.zdns.cn"
		} else if strings.Index(filePath, "cmccrpki.zdns.cn") > 0 {
			rir = ORIGIN_RIR_APNIC
			repo = "cmccrpki.zdns.cn"
		}
		*/
	} else if strings.Index(filePath, ".afrinic.net") > 0 {
		rir = ORIGIN_RIR_AFRINIC
	} else if strings.Index(filePath, ".apnic.net") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".arin.net") > 0 {
		rir = ORIGIN_RIR_ARIN
	} else if strings.Index(filePath, ".lacnic.net") > 0 {
		rir = ORIGIN_RIR_LACNIC
	} else if strings.Index(filePath, ".ripe.net") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
	} else if strings.Index(filePath, ".cnnic.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".edu.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".com.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".net.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".org.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".gov.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".mil.cn") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".eu") > 0 ||
		strings.Index(filePath, ".de") > 0 ||
		strings.Index(filePath, ".fr") > 0 ||
		strings.Index(filePath, ".uk") > 0 ||
		strings.Index(filePath, ".it") > 0 ||
		strings.Index(filePath, ".es") > 0 ||
		strings.Index(filePath, ".nl") > 0 ||
		strings.Index(filePath, ".ch") > 0 ||
		strings.Index(filePath, ".se") > 0 ||
		strings.Index(filePath, ".no") > 0 ||
		strings.Index(filePath, ".dk") > 0 ||
		strings.Index(filePath, ".fi") > 0 ||
		strings.Index(filePath, ".pl") > 0 ||
		strings.Index(filePath, ".at") > 0 ||
		strings.Index(filePath, ".pt") > 0 ||
		strings.Index(filePath, ".ru") > 0 ||
		strings.Index(filePath, ".cz") > 0 ||
		strings.Index(filePath, ".hu") > 0 ||
		strings.Index(filePath, ".ro") > 0 ||
		strings.Index(filePath, ".gr") > 0 {
		rir = ORIGIN_RIR_RIPE_NCC
	} else if strings.Index(filePath, ".jp") > 0 ||
		strings.Index(filePath, ".kr") > 0 ||
		strings.Index(filePath, ".cn") > 0 ||
		strings.Index(filePath, ".hk") > 0 ||
		strings.Index(filePath, ".tw") > 0 ||
		strings.Index(filePath, ".sg") > 0 ||
		strings.Index(filePath, ".au") > 0 ||
		strings.Index(filePath, ".nz") > 0 ||
		strings.Index(filePath, ".in") > 0 ||
		strings.Index(filePath, ".my") > 0 ||
		strings.Index(filePath, ".th") > 0 ||
		strings.Index(filePath, ".vn") > 0 ||
		strings.Index(filePath, ".id") > 0 ||
		strings.Index(filePath, ".ph") > 0 ||
		strings.Index(filePath, ".ae") > 0 ||
		strings.Index(filePath, ".sa") > 0 {
		rir = ORIGIN_RIR_APNIC
	} else if strings.Index(filePath, ".us") > 0 ||
		strings.Index(filePath, ".ca") > 0 {
		rir = ORIGIN_RIR_ARIN
	} else if strings.Index(filePath, ".br") > 0 ||
		strings.Index(filePath, ".ar") > 0 ||
		strings.Index(filePath, ".mx") > 0 ||
		strings.Index(filePath, ".cl") > 0 ||
		strings.Index(filePath, ".pe") > 0 ||
		strings.Index(filePath, ".co") > 0 ||
		strings.Index(filePath, ".uy") > 0 ||
		strings.Index(filePath, ".py") > 0 ||
		strings.Index(filePath, ".bo") > 0 {
		rir = ORIGIN_RIR_LACNIC
	} else if strings.Index(filePath, ".za") > 0 ||
		strings.Index(filePath, ".ng") > 0 ||
		strings.Index(filePath, ".eg") > 0 ||
		strings.Index(filePath, ".ke") > 0 ||
		strings.Index(filePath, ".ma") > 0 {
		rir = ORIGIN_RIR_AFRINIC
	} else {
		rir = "unknown"
	}
	if repo == "" {
		tmp := strings.Replace(filePath, conf.String("rsync::destPath")+osutil.GetPathSeparator(), "", -1)
		tmp = strings.Replace(tmp, conf.String("rrdp::destPath")+osutil.GetPathSeparator(), "", -1)
		split := strings.Split(tmp, osutil.GetPathSeparator())
		if len(split) == 0 {
			repo = filePath
		} else {
			repo = split[0]
		}
	}
	belogs.Info("JudgeOriginByFilePath(): just default rir", "rir", rir, "repo", repo,
		"filePath", filePath)

	originModel = &OriginModel{Rir: rir, Repo: repo}
	belogs.Info("JudgeOriginByFilePath(): ok",
		"filePath", filePath, "originModel", jsonutil.MarshalJson(originModel))
	return originModel
}

func JudgeRir(filePath string) (rir string) {
	originModel := JudgeOriginByFilePath(filePath)
	if originModel.Rir == "unknown" {
		return ""
	}
	return originModel.Rir
}
