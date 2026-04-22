#!/bin/bash
source base.sh
#保存参数
#启动所有程序
function kill(){
    if [ -n "$project_name" ] && [ $project_name != "all" ]; then
        game_dir=(${project_name//,/ })
    fi
    for value in ${game_dir[@]};do
        ps -ef|grep $app_header$value|awk '{print $2}'|xargs --no-run-if-empty kill -9
    done
}

kill
