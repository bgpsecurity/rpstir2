package model

import ()

/*


https://rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer
rsync://rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer

MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAx9RWSL61YAAYumEiU8z8
qH2ETVIL01ilxZlzIL9JYSORMN5Cmtf8V2JblIealSqgOTGjvSjEsiV73s67zYQI
7C/iSOb96uf3/s86NqbxDiFQGN8qG7RNcdgVuUlAidl8WxvLNI8VhqbAB5uSg/Mr
LeSOvXRja041VptAxIhcGzDMvlAJRwkrYK/Mo8P4E2rSQgwqCgae0ebY1CsJ3Cjf
i67C1nw7oXqJJovvXJ4apGmEv8az23OLC6Ki54Ul/E6xk227BFttqFV3YMtKx42H
cCcDVZZy01n7JjzvO8ccaXmHIgR7utnqhBRNNq5Xc5ZhbkrUsNtiJmrZzVlgU6Ou
0wIDAQAB

TalModel:
TalSyncUrls []talUrl is "https://rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer" and "rsync://rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer"
SubjectPublicKeyInfo string is "MIIBIj*****"

TalSyncUrl[0]:
TalUrl is "https://rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer"

RrdpUrl is RpkiNotify url(https://rrdp.apnic.net/notification.xml), comes from sia in this cer
SupportRrdp is true, when RpkiNotify is exist and get "https" is right

RsyncUrl is same to talUrl
SupportRsync is true, when CaRepository is exist and RsyncUrl start with "rsync:" and rsync is right
*/

type TalModel struct {
	TalSyncUrls []TalSyncUrl `json:"syncUrls"`

	SubjectPublicKeyInfo string `json:"subjectPublicKeyInfo"`
}

type TalSyncUrl struct {

	// url saved in tal file
	TalUrl string `json:"talUrl"`

	// rsync
	// is cer url for rsync
	RsyncUrl     string `json:"syncUrl"`
	SupportRsync bool   `json:"supportRsync"`

	// rrdp
	// is notify.xml for rrdp
	RrdpUrl     string `json:"rrdpUrl"`
	SupportRrdp bool   `json:"supportRrdp"`

	Error error `json:"error"`

	// saved tmp file
	LocalFile string `json:"-"`
}
