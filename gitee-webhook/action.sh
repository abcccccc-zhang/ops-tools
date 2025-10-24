#!/bin/bash
set -e
echo "123"
rsync -avh /root/workdir/docker-compose/confcenter-docker/data/conf/projects/gitee/ /root/workdir/docker-compose/confcenter-docker/data/conf/projects/gitee-config-test/
