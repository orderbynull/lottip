#/env/bash

# Set defaults

# IP and port to listen for MySQL
LOTTIP_PROXY="${LOTTIP_PROXY:-0.0.0.0:4041}"

# IP and port to connect to MySQL 
LOTTIP_MYSQL="${LOTTIP_MYSQL:-$(/sbin/ip route|awk '/default/ { print $3 }'):3306}"

# IP and port for the GUI to listen to
LOTTIP_GUI="${LOTTIP_GUI:-0.0.0.0:9999}"

# MySQL DSN (credentials)
LOTTIP_DSN="${LOTTIP_DSN:-root:root@/}"


# Run lottip

/root/go/src/github.com/orderbynull/lottip/lottip \
  --proxy "$LOTTIP_PROXY" \
  --mysql "$LOTTIP_MYSQL" \
  --gui "$LOTTIP_GUI" \
  --mysql-dsn "$LOTTIP_DSN"
