# Установка на linux
```bash
curl -s https://raw.githubusercontent.com/immisterio/MatriX.Quant/master/install.sh | bash
```

# /opt/matrix/accs.db
```json
{
  "user1": "pass1",
  "user2": "pass2"
}
```

# /opt/matrix/settings.json
```json
{
  "BitTorr": {
    "CacheSize": 96468992,
    "ConnectionsLimit": 30,
    "DisableDHT": false,
    "DisablePEX": false,
    "DisableTCP": false,
    "DisableUPNP": false,
    "DisableUTP": false,
    "DisableUpload": false,
    "DownloadRateLimit": 0,
    "EnableDebug": false,
    "EnableIPv6": false,
    "ForceEncrypt": false,
    "PeersListenPort": 0,
    "PreloadCache": 14,
    "ReaderReadAHead": 86,
    "RemoveCacheOnDrop": false,
    "ResponsiveMode": false,
    "RetrackersMode": 1,
    "SslCert": "",
    "SslKey": "",
    "SslPort": 0,
    "TorrentDisconnectTimeout": 120,
    "TorrentsSavePath": "",
    "UploadRateLimit": 0,
    "UseDisk": false
  }
}
```

# Чем отличается MatriX.Quant от MatriX
* Все пользователи разделены, у каждого своя база
* Пользователи не могут менять настройки BitTorr
* Отключен shutdown, dlna
