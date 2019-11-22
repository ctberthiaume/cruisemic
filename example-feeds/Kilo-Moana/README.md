# Kilo Moana underway data feed

## Original email describing feed

```
The broadcast on the KM looks like the following:
2017 168 00 30 28 990 bar1   1016.07 mbar
2017 168 00 30 28 998 uthsl 19.968599 0.040550 0.217500 27.397800
$GPDTM,W84,,00.0000,N,00.0000,E,,W84*41
$GPGGA,003029.00,2118.9043,N,15752.6526,W,2,7,0.8,27,M,,M,,*78
2017 168 00 30 29 229 rbgm3 024723 00 978906.152511
2017 168 00 30 29 285 rwd1  10 233   0.0  52.5 208.3  10.0  81.3 
$GPDTM,W84,,00.0000,N,00.0000,E,,W84*41
2017 168 00 30 29 285 rwd2  12 236   0.0  52.5 208.3  12.0  84.3 
2017 168 00 30 29 365 flor 78.000000
$GPGLL,2118.9043,N,15752.6526,W,003029.00,A,D*71
$GPVTG,47.3,T,37.7,M,0.0,N,0.0,K,D*25
$GPZDA,003029.00,17,06,2017,00,00*6A
2017 168 00 30 29 909 met  0.000 28.680  50.900 28.470 24.766  3.758 -0.246  1.097  1.099  0.000 5040.000  1.016 11.9 235.0 11.9   83.3     0.000  0.000
$GPDTM,W84,,00.0000,N,00.0000,E,,W84*41
2017 168 00 30 29 990 bar1   1016.05 mbar
2017 168 00 30 29 998 uthsl 19.968800 0.040540 0.217400 27.398399
$GPGGA,003030.00,2118.9041,N,15752.6526,W,2,7,0.8,27,M,,M,,*72
$GPDTM,W84,,00.0000,N,00.0000,E,,W84*41

The lines of ship data you will be interested in are the uthsl, flor and met
uthsl:
2017 168 00 30 28 998 uthsl 19.968599 0.040550 0.217500 27.397800
Fields 1-6 are a date stamp in the form of YR DAY HR MIN SEC MSEC
Field 7 is an instrument stamp
Fields 8-11 are Temp(C),Cond(S/m),Salinity(PSU),Temp(remote)

flor:
2017 168 00 30 29 365 flor 78.000000
Fields 1-7 are the usual date/instrument stamp
Field 8 is a raw scale count

met: (long string)
The usual date/inst stamp..
PAR (mVolts) can be found in field 19

Let me know if you need anything else.

Trevor
```

## Notes

* Lines in the feed may be CRLF terminated.
* They may also be null terminated.
* The feed normally has a frequency of 1Hz for each kind data,
but it has jumped to 1KHz before.
Any parser should be ready to handle that level of data flow.
* The feed is space separated, sometimes with multiple consecutive spaces.
* Column 24 in the met lines can change from blank to "R-",
which mean any check that counts total columns will probably break.
Just look for the PAR value in column 19 and ignore total column count here.
