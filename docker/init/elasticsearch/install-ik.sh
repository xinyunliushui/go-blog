#!/bin/sh
set -eu

ES_VERSION="${ES_VERSION:?ES_VERSION is required}"
PLUGIN_ZIP="/tmp/elasticsearch-analysis-ik-${ES_VERSION}.zip"
# 8.19.5 实际约 4.6MB，勿设过高阈值导致误判
EXPECT_MIN_BYTES=3500000
download_ok=0

is_zip() {
  [ "$(head -c 2 "${1}" 2>/dev/null | od -An -tx1 | tr -d ' \n')" = "504b" ]
}

for url in \
  "https://get.infini.cloud/elasticsearch/analysis-ik/${ES_VERSION}" \
  "https://release.infinilabs.com/analysis-ik/stable/elasticsearch-analysis-ik-${ES_VERSION}.zip"
do
  echo "Downloading IK plugin from ${url}"
  if curl -fsSL --connect-timeout 30 --max-time 900 --retry 8 --retry-delay 5 \
    -o "${PLUGIN_ZIP}" "${url}"
  then
    size="$(wc -c < "${PLUGIN_ZIP}")"
    if [ "${size}" -ge "${EXPECT_MIN_BYTES}" ] && is_zip "${PLUGIN_ZIP}"; then
      download_ok=1
      echo "Download ok (${size} bytes)"
      break
    fi
    echo "Invalid or incomplete download (${size} bytes), try next mirror..."
  fi
  rm -f "${PLUGIN_ZIP}"
done

if [ "${download_ok}" -ne 1 ]; then
  echo "Failed to download IK plugin from all mirrors"
  exit 1
fi

bin/elasticsearch-plugin install --batch "file://${PLUGIN_ZIP}"
rm -f "${PLUGIN_ZIP}"
echo "IK plugin ${ES_VERSION} installed"
