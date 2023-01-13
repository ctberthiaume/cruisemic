# gradients5 underway data feed

Data is emitted in lines like this. Par may be unreliable, may contain 0 - N
columns. Only keep the first PAR number (microeinstein) and only if it has 3
decimals of precision.

```
$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5
$SEAFLOW::$GPZDA,213310.00,12,01,2023,00,00*65::$GPGGA,213310.00,4738.983143,N,12218.805821,W,2,17,0.7,15.774,M,-22.2,M,8.0,0402*43::::$PPAR, 157.445, 5
$SEAFLOW::$GPZDA,213311.00,12,01,2023,00,00*64::$GPGGA,213311.00,4738.983147,N,12218.805822,W,2,17,0.7,15.776,M,-22.2,M,5.0,0402*4A::::$PPAR, 157.3

```
