# OSPFHunter

Esta aplicación analiza un fichero pcap, buscando una ráfaga de paquetes OSPFv2 LSA pdate que cumpla estas condiciones:

- Contener alguna LSA con el campo Age = 3600
- Ser unicast

Si encuentra una ráfaga de mensajes de estas características, vuelca a pantalla los números de las tramas que forman la ráfaga.

Modo de uso:

```bash
ospfhunter.exe [-c tamaño_de_la_rafaga] [-i intervalo_en_segundos] <fichero.pcap>
```

Por defecto, busca ráfagas de al menos 10 paquetes en 60 segundos.

