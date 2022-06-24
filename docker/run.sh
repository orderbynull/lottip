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

# Log directory
LOTTIP_LOG_DIRECTORY="${LOTTIP_LOG_DIRECTORY:-./logs}"

# Log filename
LOTTIP_LOGFILE="${LOTTIP_LOGFILE:-logfile.log}"

# Query Log filename
LOTTIP_QUERY_LOGFILE="${LOTTIP_QUERY_LOGFILE:-queries.log}"

# Enable
if [ "${LOTTIP_CONSOLE_LOGGING}" == "true" ];
then
  LOTTIP_CONSOLE_LOGGING="--enable-console-logging"
else
  LOTTIP_CONSOLE_LOGGING=""
fi

# Run lottip
/root/go/bin/lottip \
  --proxy "$LOTTIP_PROXY" \
  --mysql "$LOTTIP_MYSQL" \
  --gui-addr "$LOTTIP_GUI" \
  --mysql-dsn "$LOTTIP_DSN" \
  "$LOTTIP_CONSOLE_LOGGING" \
  --query-log-file "$LOTTIP_QUERY_LOGFILE" \
  --log-file "$LOTTIP_LOGFILE" \
  --log-directory "$LOTTIP_LOG_DIRECTORY"