# gradients5 underway data feed

Data is emitted in lines like this. Par may be unreliable, may contain 0 - N
columns. If there PAR is completely empty (not even a "$PPAR" string) we'll
record it as NA assuming the feed is off. If the feed is there ("$PPAR, ") then
we'll only keep a record if the PAR number is complete (has 3 decimals of
precision).

```
$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5
$SEAFLOW::$GPZDA,213310.00,12,01,2023,00,00*65::$GPGGA,213310.00,4738.983143,N,12218.805821,W,2,17,0.7,15.774,M,-22.2,M,8.0,0402*43::::$PPAR, 157.445, 5
$SEAFLOW::$GPZDA,213311.00,12,01,2023,00,00*64::$GPGGA,213311.00,4738.983147,N,12218.805822,W,2,17,0.7,15.776,M,-22.2,M,5.0,0402*4A::::$PPAR, 157.3
$SEAFLOW::$GPZDA,164825.00,16,01,2023,00,00*6F::$GPGGA,164825.00,3805.763504,N,12319.918325,W,2,14,0.9,12.210,M,-34.0,M,4.0,0402*4A:: 12.3719,  3.64868,  31.2816::$PPAR, 1545.960, 9.42,
```
