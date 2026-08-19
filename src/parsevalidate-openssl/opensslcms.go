package openssl

/*
openssl cms -in  AD74E5D0B0DF11E5A886C423C4F9AE02.roa -noout -cmsout -print -inform der
CMS_ContentInfo:
  contentType: pkcs7-signedData (1.2.840.113549.1.7.2)
  d.signedData:
    version: 3
    digestAlgorithms:
        algorithm: sha256 (2.16.840.1.101.3.4.2.1)
        parameter: <ABSENT>
    encapContentInfo:
      eContentType: undefined (1.2.840.113549.1.9.16.1.24)
      eContent:
        0000 - 30 1a 02 03 00 b3 01 30-13 30 11 04 02 00 01   0......0.0.....
        000f - 30 0b 30 09 03 04 00 cb-59 97 02 01 18         0.0.....Y....
    certificates:
      d.certificate:
        cert_info:
          version: 2
          serialNumber: 9673
          signature:
            algorithm: sha256WithRSAEncryption (1.2.840.113549.1.1.11)
            parameter: NULL
          issuer: CN=A91A20A3/serialNumber=D15A5AF9AF8D99DFE5E44D35EE550105FED94511
          validity:
            notBefore: Jun  1 09:50:44 2019 GMT
            notAfter: Aug 31 00:00:00 2020 GMT
          subject: CN=5cf24a73-628f
          key:
            algor:
              algorithm: rsaEncryption (1.2.840.113549.1.1.1)
              parameter: NULL
            public_key:  (0 unused bits)
              0000 - 30 82 01 0a 02 82 01 01-00 d5 f0 f0 22 cc   0...........".
              000e - f3 b9 5f e0 d1 a6 bb 34-5c d1 07 fc 98 e1   .._....4\.....
              001c - f7 ad b8 2b 1f 09 4d c2-ef 04 c4 5a 78 3e   ...+..M....Zx>
              002a - 67 9b 3c ce 59 fa 92 ff-7c 94 7f 98 04 07   g.<.Y...|.....
              0038 - 37 17 00 5f dc 01 3d f0-90 e2 c2 73 97 2e   7.._..=....s..
              0046 - a8 85 33 c6 98 de b2 da-a6 21 09 8b 0b 84   ..3......!....
              0054 - 8d 92 27 0f 44 5a 75 23-dc fc 63 6c 59 10   ..'.DZu#..clY.
              0062 - fe a6 59 1b d9 ad 21 35-f8 37 90 1c 35 ea   ..Y...!5.7..5.
              0070 - 03 d3 70 18 c9 4e 79 57-c2 64 9e 2c d1 ee   ..p..NyW.d.,..
              007e - be 00 f4 f4 33 e4 69 30-15 74 e8 36 cc 5b   ....3.i0.t.6.[
              008c - 78 7d a1 3c b6 91 ef 73-06 e1 5a 67 13 ed   x}.<...s..Zg..
              009a - 4a 4d 44 47 61 25 95 27-cd 95 f5 22 d8 c8   JMDGa%.'..."..
              00a8 - f3 69 8b f0 55 04 64 65-d3 c9 bb 5a f5 bf   .i..U.de...Z..
              00b6 - 31 89 08 d3 82 d9 9e 10-ae bd 44 ac 87 19   1.........D...
              00c4 - 84 76 de 7c 03 0b 09 a6-31 ae ae b1 44 e5   .v.|....1...D.
              00d2 - 20 04 3a 17 8b 8d 52 57-05 b8 30 bf 52 36    .:...RW..0.R6
              00e0 - 65 71 c5 51 6a 3e 7d 80-81 95 8c af a4 84   eq.Qj>}.......
              00ee - bf da da 32 c2 4d 1e 14-56 c2 a5 95 87 54   ...2.M..V....T
              00fc - 26 8e 65 9b 9f 90 e2 f1-50 70 e3 14 6b 02   &.e.....Pp..k.
              010a - 03 01 00 01                                 ....
          issuerUID: <ABSENT>
          subjectUID: <ABSENT>
          extensions:
              object: X509v3 Subject Key Identifier (2.5.29.14)
              critical: BOOL ABSENT
              value:
                0000 - 04 14 b2 2a eb ca cf a3-47 e4 98 52 83   ...*....G..R.
                000d - 19 36 d9 3e d1 f7 6d ff-b4               .6.>..m..

              object: X509v3 Authority Key Identifier (2.5.29.35)
              critical: BOOL ABSENT
              value:
                0000 - 30 16 80 14 d1 5a 5a f9-af 8d 99 df e5   0....ZZ......
                000d - e4 4d 35 ee 55 01 05 fe-d9 45 11         .M5.U....E.

              object: X509v3 Key Usage (2.5.29.15)
              critical: TRUE
              value:
                0000 - 03 02 07 80                              ....

              object: X509v3 CRL Distribution Points (2.5.29.31)
              critical: BOOL ABSENT
              value:
                0000 - 30 7a 30 78 a0 76 a0 74-86 72 72 73 79   0z0x.v.t.rrsy
                000d - 6e 63 3a 2f 2f 72 70 6b-69 2e 61 70 6e   nc://rpki.apn
                001a - 69 63 2e 6e 65 74 2f 6d-65 6d 62 65 72   ic.net/member
                0027 - 5f 72 65 70 6f 73 69 74-6f 72 79 2f 41   _repository/A
                0034 - 39 31 41 32 30 41 33 2f-41 46 34 42 38   91A20A3/AF4B8
                0041 - 41 36 32 31 44 39 45 31-31 45 32 38 45   A621D9E11E28E
                004e - 35 33 43 30 38 45 30 38-42 30 32 43 44   53C08E08B02CD
                005b - 32 2f 30 56 70 61 2d 61-2d 4e 6d 64 5f   2/0Vpa-a-Nmd_
                0068 - 6c 35 45 30 31 37 6c 55-42 42 66 37 5a   l5E017lUBBf7Z
                0075 - 52 52 45 2e 63 72 6c                     RRE.crl

              object: Authority Information Access (1.3.6.1.5.5.7.1.1)
              critical: BOOL ABSENT
              value:
                0000 - 30 70 30 6e 06 08 2b 06-01 05 05 07 30   0p0n..+.....0
                000d - 02 86 62 72 73 79 6e 63-3a 2f 2f 72 70   ..brsync://rp
                001a - 6b 69 2e 61 70 6e 69 63-2e 6e 65 74 2f   ki.apnic.net/
                0027 - 72 65 70 6f 73 69 74 6f-72 79 2f 42 35   repository/B5
                0034 - 32 37 45 46 35 38 31 44-36 36 31 31 45   27EF581D6611E
                0041 - 32 42 42 34 36 38 46 37-43 37 32 46 44   2BB468F7C72FD
                004e - 31 46 46 32 2f 30 56 70-61 2d 61 2d 4e   1FF2/0Vpa-a-N
                005b - 6d 64 5f 6c 35 45 30 31-37 6c 55 42 42   md_l5E017lUBB
                0068 - 66 37 5a 52 52 45 2e 63-65 72            f7ZRRE.cer

              object: X509v3 Certificate Policies (2.5.29.32)
              critical: TRUE
              value:
                0000 - 30 3e 30 3c 06 08 2b 06-01 05 05 07 0e   0>0<..+......
                000d - 02 30 30 30 2e 06 08 2b-06 01 05 05 07   .000...+.....
                001a - 02 01 16 22 68 74 74 70-73 3a 2f 2f 77   ..."https://w
                0027 - 77 77 2e 61 70 6e 69 63-2e 6e 65 74 2f   ww.apnic.net/
                0034 - 52 50 4b 49 2f 43 50 53-2e 70 64 66      RPKI/CPS.pdf

              object: Subject Information Access (1.3.6.1.5.5.7.1.11)
              critical: BOOL ABSENT
              value:
                0000 - 30 81 bb 30 81 83 06 08-2b 06 01 05 05   0..0....+....
                000d - 07 30 0b 86 77 72 73 79-6e 63 3a 2f 2f   .0..wrsync://
                001a - 72 70 6b 69 2e 61 70 6e-69 63 2e 6e 65   rpki.apnic.ne
                0027 - 74 2f 6d 65 6d 62 65 72-5f 72 65 70 6f   t/member_repo
                0034 - 73 69 74 6f 72 79 2f 41-39 31 41 32 30   sitory/A91A20
                0041 - 41 33 2f 41 46 34 42 38-41 36 32 31 44   A3/AF4B8A621D
                004e - 39 45 31 31 45 32 38 45-35 33 43 30 38   9E11E28E53C08
                005b - 45 30 38 42 30 32 43 44-32 2f 41 44 37   E08B02CD2/AD7
                0068 - 34 45 35 44 30 42 30 44-46 31 31 45 35   4E5D0B0DF11E5
                0075 - 41 38 38 36 43 34 32 33-43 34 46 39 41   A886C423C4F9A
                0082 - 45 30 32 2e 72 6f 61 30-33 06 08 2b 06   E02.roa03..+.
                008f - 01 05 05 07 30 0d 86 27-68 74 74 70 73   ....0..'https
                009c - 3a 2f 2f 72 72 64 70 2e-61 70 6e 69 63   ://rrdp.apnic
                00a9 - 2e 6e 65 74 2f 6e 6f 74-69 66 69 63 61   .net/notifica
                00b6 - 74 69 6f 6e 2e 78 6d 6c-                 tion.xml

              object: sbgp-ipAddrBlock (1.3.6.1.5.5.7.1.7)
              critical: TRUE
              value:
                0000 - 30 0e 30 0c 04 02 00 01-30 06 03 04 00   0.0.....0....
                000d - cb 59 97                                 .Y.
        sig_alg:
          algorithm: sha256WithRSAEncryption (1.2.840.113549.1.1.11)
          parameter: NULL
        signature:  (0 unused bits)
          0000 - 48 4d 55 4e 86 2c 7e 18-0f df f5 6b 34 8c 64   HMUN.,~....k4.d
          000f - 05 cb 71 e7 bf 8e fb 0c-63 6f f0 73 19 1d 51   ..q.....co.s..Q
          001e - e3 d3 19 1f 8f 4b 8f 6a-1e 69 bd 02 fd dc 2d   .....K.j.i....-
          002d - a4 71 f8 5e ba 54 75 63-1f d5 7e 05 3e 5d 5e   .q.^.Tuc..~.>]^
          003c - 20 13 37 d7 39 98 07 a5-25 da 11 76 78 5b f7    .7.9...%..vx[.
          004b - a8 d0 c0 01 8d f4 76 ec-11 b6 90 b1 7a 99 2b   ......v.....z.+
          005a - 4c f5 a8 b1 12 d9 00 75-d2 47 6b 12 d2 99 69   L......u.Gk...i
          0069 - 4b 5b ce ef 9e 47 d0 95-59 84 bd 67 f2 e2 b0   K[...G..Y..g...
          0078 - ec 5e ec a3 3c c1 fd 4e-58 b5 bb a8 0a b0 fa   .^..<..NX......
          0087 - 43 3c a6 06 9c da d6 63-b3 dd 2c bb 6e 4d 86   C<.....c..,.nM.
          0096 - 2b ed c3 2e f2 52 52 b1-61 e4 bc b8 54 55 90   +....RR.a...TU.
          00a5 - d6 09 3c a5 be a6 2d 9d-d4 eb ce ed f8 5f ae   ..<...-......_.
          00b4 - 83 e5 51 0e f0 81 a9 8f-60 03 23 65 fd f7 0e   ..Q.....`.#e...
          00c3 - 90 b8 51 6a a4 bd 5a d1-d2 2a f9 ee 61 c3 a0   ..Qj..Z..*..a..
          00d2 - 74 9e 78 fa a1 b0 0f 5e-9b b2 51 ce 06 27 5c   t.x....^..Q..'\
          00e1 - 2f 4b 9e 0d a0 a8 62 ff-96 76 32 87 e1 63 ce   /K....b..v2..c.
          00f0 - e3 dc d6 3f 85 c3 0f 4d-49 df 23 01 94 50 d1   ...?...MI.#..P.
          00ff - 6e                                             n
    crls:
      <ABSENT>
    signerInfos:
        version: 3
        d.subjectKeyIdentifier:
          0000 - b2 2a eb ca cf a3 47 e4-98 52 83 19 36 d9 3e   .*....G..R..6.>
          000f - d1 f7 6d ff b4                                 ..m..
        digestAlgorithm:
          algorithm: sha256 (2.16.840.1.101.3.4.2.1)
          parameter: <ABSENT>
        signedAttrs:
            object: contentType (1.2.840.113549.1.9.3)
            set:
              OBJECT:undefined (1.2.840.113549.1.9.16.1.24)

            object: signingTime (1.2.840.113549.1.9.5)
            set:
              UTCTIME:Jun  1 09:50:44 2019 GMT

            object: messageDigest (1.2.840.113549.1.9.4)
            set:
              OCTET STRING:
                0000 - 29 59 0c eb 66 6a 80 b7-4b af fd 91 dc   )Y..fj..K....
                000d - 37 ad f9 6b f5 7d 82 f6-cb af 22 18 7f   7..k.}...."..
                001a - fb ed 18 f8 98 cf                        ......
        signatureAlgorithm:
          algorithm: rsaEncryption (1.2.840.113549.1.1.1)
          parameter: NULL
        signature:
          0000 - aa 25 e5 66 91 cb 77 73-77 bd 1c b4 d8 6c d6   .%.f..wsw....l.
          000f - 35 57 d4 46 e0 eb a2 b7-c1 70 01 7e 23 63 ee   5W.F.....p.~#c.
          001e - 5d 59 56 5d 1f fe a7 9f-c6 65 93 0f 28 69 cd   ]YV].....e..(i.
          002d - 72 84 55 da f6 9e 1a 2a-19 bc 32 1a 08 c0 28   r.U....*..2...(
          003c - 9d fe 3e 07 47 04 88 10-73 28 1c 4e e6 54 aa   ..>.G...s(.N.T.
          004b - be c6 6d 0d 4c 1e 84 b8-01 49 ab 57 83 fe 98   ..m.L....I.W...
          005a - e1 a1 1a 0b da ec f8 5d-43 3b 59 2a 1b 4e 79   .......]C;Y*.Ny
          0069 - a0 d7 d1 d2 c5 b9 68 11-66 55 d4 e9 19 d8 9b   ......h.fU.....
          0078 - 1c 08 6d 1d 02 be 85 f6-75 3e eb e8 8f 41 60   ..m.....u>...A`
          0087 - f0 da 89 de 00 82 5b b1-1d 01 b8 3c 57 3b 28   ......[....<W;(
          0096 - b7 65 08 f2 0c 85 24 f3-0d 5c 86 66 4e de 29   .e....$..\.fN.)
          00a5 - a7 9e 8b f5 ad 75 46 14-6d 9c 98 3d 68 1e af   .....uF.m..=h..
          00b4 - b7 73 eb f1 c8 51 d6 89-92 d1 f5 de 91 81 43   .s...Q........C
          00c3 - 75 63 20 c6 b6 06 33 cc-e5 87 2e a4 af de 59   uc ...3.......Y
          00d2 - 54 57 a6 f5 7d 5a 5a 63-e0 fd 9b 58 a5 43 03   TW..}ZZc...X.C.
          00e1 - 6e ea d6 a6 78 12 12 48-07 61 2f 5e 78 49 c3   n...x..H.a/^xI.
          00f0 - 0c 60 a6 f1 ea 51 b6 a6-fd 54 d7 6d f9 6b 47   .`...Q...T.m.kG
          00ff - fb                                             .
        unsignedAttrs:
          <ABSENT>

*/
