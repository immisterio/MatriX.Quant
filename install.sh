#!/usr/bin/env bash
DEST="/opt/matrix"

mkdir $DEST -p 
cd $DEST
wget https://github.com/immisterio/MatriX.Quant/releases/latest/download/TorrServer-linux-amd64
chmod +x TorrServer-linux-amd64

cat <<EOF > $DEST/settings.json
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
EOF

cat <<EOF > $DEST/accs.db
{
  "user1": "pass1",
  "user2": "pass2"
}
EOF

echo ""
echo "Install service to /etc/systemd/system/matrix.service ..."
touch /etc/systemd/system/matrix.service && chmod 664 /etc/systemd/system/matrix.service
cat <<EOF > /etc/systemd/system/matrix.service
[Unit]
Description=matrix
Wants=network.target
After=network.target
[Service]
WorkingDirectory=$DEST
ExecStart=$DEST/TorrServer-linux-amd64 -a
#ExecReload=/bin/kill -s HUP $MAINPID
#ExecStop=/bin/kill -s QUIT $MAINPID
Restart=always
LimitNOFILE = 50000
[Install]
WantedBy=multi-user.target
EOF

# Enable service
systemctl daemon-reload
systemctl enable matrix
systemctl start matrix

# iptables drop
cat <<EOF > iptables-drop.sh
#!/bin/sh
echo "Stopping firewall and allowing everyone..."
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X
iptables -t mangle -F
iptables -t mangle -X
iptables -P INPUT ACCEPT
iptables -P FORWARD ACCEPT
iptables -P OUTPUT ACCEPT
EOF

# Note
echo ""
echo "################################################################"
echo ""
echo "Have fun!"
echo ""
echo "Then [re]start it as systemctl [re]start matrix"
echo ""
echo "Clear iptables if port 8090 is not available"
echo "bash $DEST/iptables-drop.sh"
echo ""
