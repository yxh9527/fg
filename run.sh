#!/bin/bash
source base.sh
#获取当前程序所在目录
work_pwd=`pwd`
#启动所有程序
function start(){
    if [ "$project_name" != "all" ];then
        game_dir=(${project_name//,/ })
    fi
    for value in ${game_dir[@]};do
        cd $work_pwd"/"$value &&chmod +x $app_header$value 
        echo "nohup ./$app_header$value run >/dev/null 2>&1&" | bash;
        if [ $value == "data-center" ];then
            sleep 5
        else
            sleep 2
        fi
        echo "start $value ok"
    done
}

start
