package openssl

//openssl verify -CAfile RootCert.pem -untrusted Intermediate.pem UserCert.pem
/*
openssl verify -CAfile root.pem.cer   -untrusted inter.pem.cer -verbose A9.pem.cer
A9.pem.cer: OK
*/
