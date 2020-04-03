package slurm

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	"model"
)

// check prefix, such as 2001:DB8::/32  or 198.51.100.0/24
func CheckPrefix(prefix string) error {
	if len(prefix) == 0 {
		return nil
	}
	belogs.Debug("CheckPrefix():will check prefix ", prefix)
	pos := strings.Index(prefix, "/")
	lastPos := strings.LastIndex(prefix, "/")
	if pos <= 0 || lastPos <= 0 {
		return errors.New("prefix is not contains '/' , or '/' position is error")
	}
	separatorCount := strings.Count(prefix, "/")
	if separatorCount != 1 {
		return errors.New("prefix can contain only one '/' ")
	}
	ipAndLength := strings.Split(prefix, "/")
	if len(ipAndLength) != 2 {
		return errors.New("prefix is not contains '/' or too much")
	}
	ip := ipAndLength[0]
	belogs.Debug("CheckPrefix():will check by regular expression ", ip)

	//check ipv4
	patternIpv4 := `^(\d+)\.(\d+)\.(\d+)\.(\d+)$`
	matchedIpv4, errIpv4 := regexp.MatchString(patternIpv4, ip)

	//check ipv6
	patternIpv6 := `^\s*((([0-9A-Fa-f]{1,4}:){7}(([0-9A-Fa-f]{1,4})|:))|(([0-9A-Fa-f]{1,4}:){6}(:|((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})|(:[0-9A-Fa-f]{1,4})))|(([0-9A-Fa-f]{1,4}:){5}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){4}(:[0-9A-Fa-f]{1,4}){0,1}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){3}(:[0-9A-Fa-f]{1,4}){0,2}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){2}(:[0-9A-Fa-f]{1,4}){0,3}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:)(:[0-9A-Fa-f]{1,4}){0,4}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(:(:[0-9A-Fa-f]{1,4}){0,5}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})))(%.+)?\s*$`
	matchedIpv6, errIpv6 := regexp.MatchString(patternIpv6, ip)

	//need ipv4 or ipv6 is no err, and one check is ok
	if (errIpv4 != nil || errIpv6 != nil) &&
		matchedIpv4 == false && matchedIpv6 == false {
		return errors.New("prefix is not legal for " + ip)
	}

	//check length
	// will after check,
	return nil
}

func CheckMaxPrefixLength(maxPrefixLength uint64) error {
	if maxPrefixLength == 0 {
		//ignore
	}

	return nil
}

type PrefixAndAsn struct {
	FormatPrefix    string
	PrefixLength    uint64
	MaxPrefixLength uint64
}

func FormatPrefix(ip string) string {
	formatIp := ""

	// format  ipv4
	ipsV4 := strings.Split(ip, ".")
	if len(ipsV4) > 1 {
		for _, ipV4 := range ipsV4 {
			ip, _ := strconv.Atoi(ipV4)
			formatIp += fmt.Sprintf("%02x", ip)
		}
		return formatIp
	}

	// format ipv6
	count := strings.Count(ip, ":")
	if count > 0 {
		count := strings.Count(ip, ":")
		if count < 7 { // total colon is 8
			needCount := 7 - count + 2 //2 is current "::", need add
			colon := strings.Repeat(":", needCount)
			ip = strings.Replace(ip, "::", colon, -1)
			belogs.Debug(ip)
		}
		ipsV6 := strings.Split(ip, ":")
		belogs.Debug(ipsV6)
		for _, ipV6 := range ipsV6 {
			formatIp += fmt.Sprintf("%04s", ipV6)
		}
		return formatIp
	}
	return ""
}

// check prefix and asn
func CheckPrefixAsnAndGetPrefixLength(prefix string, asn model.SlurmAsnModel,
	maxPrefixLength uint64) (PrefixAndAsn, error) {

	// cannot is empty at the same time
	if len(prefix) == 0 && asn.Value == 0 {
		return PrefixAndAsn{}, errors.New("prefix and asn should not is empty at the same time")
	}
	prefixAndAsn := PrefixAndAsn{}
	var (
		errMsg string = ""
	)

	//check prefix

	if err := CheckPrefix(prefix); err != nil {
		errMsg += (prefix + " has error:" + err.Error() + "; ")
	}

	//check maxPrefixLength
	if err := CheckMaxPrefixLength(maxPrefixLength); err != nil {
		errMsg += (string(maxPrefixLength) + " has error:" + err.Error() + "; ")
	}

	// parse prefix to get prefixLength and formatIp, and if there is no maxPrefixLength, it will equal to prefixLength.
	// such as 192.0.2.0/24, the prefixLength is 24. formatip is
	// such as 2001:DB8::/32, the prefixLength is 32. formatip is 20010DB8000000000000000000000000 . filled with 0
	//
	if len(prefix) > 0 {
		//get prefixLength and maxPrefixLength
		ipsAndLength := strings.Split(prefix, "/")
		belogs.Debug("CheckPrefixAsnAndGetPrefixLength():ipsAndLength: ", ipsAndLength)

		ips := ipsAndLength[0]
		belogs.Debug("ips: ", ips)

		prefixAndAsn.FormatPrefix = FormatPrefix(ips)
		belogs.Debug("CheckPrefixAsnAndGetPrefixLength():FormatPrefix: ", prefixAndAsn.FormatPrefix)

		PrefixLength, err := strconv.Atoi(ipsAndLength[1])
		belogs.Debug("CheckPrefixAsnAndGetPrefixLength():PrefixLength: ", PrefixLength)

		if err != nil {
			errMsg += (string(ipsAndLength[1]) + " is not a number, " + err.Error() + "; ")
		} else {
			prefixAndAsn.PrefixLength = uint64(PrefixLength)
		}

		if maxPrefixLength != 0 {
			prefixAndAsn.MaxPrefixLength = maxPrefixLength
		} else {
			prefixAndAsn.MaxPrefixLength = prefixAndAsn.PrefixLength
		}

	}

	// return error
	if len(errMsg) > 0 {
		return PrefixAndAsn{}, errors.New(errMsg)
	}
	belogs.Debug(fmt.Sprintf("CheckPrefixAsnAndGetPrefixLength():will return prefixAndAsn: %+v", prefixAndAsn))
	return prefixAndAsn, nil
}

func AppendRoaLocalIdSql(localIds []int) string {
	roaLocalIdsSql := "("
	for i, idTmp := range localIds {
		if i < len(localIds)-1 {
			roaLocalIdsSql += (strconv.Itoa(idTmp) + ",")
		} else {
			roaLocalIdsSql += (strconv.Itoa(idTmp))
		}
	}
	roaLocalIdsSql += ") "
	return roaLocalIdsSql
}
