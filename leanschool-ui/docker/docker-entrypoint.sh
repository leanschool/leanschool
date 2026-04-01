#!/bin/sh
set -e
envsubst < /etc/nginx/config.js.template > /usr/share/nginx/html/config.js
exec nginx -g 'daemon off;'
