#!/bin/bash

# executable path
EXE_PATH=$GOPATH/bin/tank

# execute arguments
MysqlHost=127.0.0.1
MysqlPort=3306
MysqlSchema=tank
MysqlUserName=tank
MysqlPassword=Tank_123

AdminUsername=admin
AdminEmail=lish516@126.com
AdminPassword=123456

if [ -f "$EXE_PATH" ]; then
 nohup $EXE_PATH -MysqlHost=$MysqlHost -MysqlPort=$MysqlPort -MysqlSchema=$MysqlSchema -MysqlUserName=$MysqlUserName -MysqlPassword=$MysqlPassword -AdminUsername=$AdminUsername -AdminEmail=$AdminEmail -AdminPassword=$AdminPassword >/dev/null 2>&1 &
else
 echo 'Cannot find $EXE_PATH.'
 exit 1
fi
