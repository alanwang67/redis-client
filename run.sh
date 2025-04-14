#!/bin/bash

create_session() {
    tmux new-session -d -s ${1} -c ${2}
}

# Attach to tmux session
attach_session() {
    tmux attach-session -t $1
}

# Create new tmux window, set starting directory
new_window() {
    tmux new-window -t ${1}:${2} -c ${3}
}

# Create new tmux window split horizontally, set starting directory
new_window_horiz_split() {
    tmux new-window -t ${1}:${2} -c ${3}
    tmux split-window -h -t ${1}:${2}
}

# Name tmux window
name_window() {
    tmux rename-window -t ${1}:${2} ${3}
}

# Run tmux command
run_command() {
    tmux send-keys -t ${1}:${2} "${3}" C-m
}

# Run tmux command in left pane
run_command_left() {
    tmux send-keys -t ${1}:${2}.0 "${3}" C-m
}

# Run tmux command in right pane
run_command_right() {
    tmux send-keys -t ${1}:${2}.1 "${3}" C-m
}

ct=0
tmux kill-session -t experiment 
    
SES="experiment"               
DIR="~/redis-7.4.2/src"   

create_session $SES $DIR       
new_window $SES 1 $DIR
new_window $SES 2 $DIR

sleep 1
name_window $SES 0 server0 
run_command $SES 0 "ssh srg02"

name_window $SES 1 server1
run_command $SES 1 "ssh srg03"

name_window $SES 2 server2
run_command $SES 2 "ssh srg04"

run_command $SES 0 "cd ~/redis-7.4.2/src"
run_command $SES 1 "cd ~/redis-7.4.2/src"
run_command $SES 2 "cd ~/redis-7.4.2/src"

cd ~/redis-client; go build main.go 

sleep 1

declare -a workload=(5 50) 

cd ~/redis-client/generateGraph

for w in "${workload[@]}"
    do
    mkdir workload_$w
    for run in {1..3}
        do
        cd ~/redis-client/generateGraph/workload_$w/
        mkdir run_$run 
        cd ~/redis-client/generateGraph/workload_$w/run_$run
        for i in {1..12}
            do
                run_command $SES 0 "./redis-server ../redis.conf"

                run_command $SES 1 "./redis-server ../redis.conf"

                run_command $SES 2 "./redis-server ../redis.conf"

                sleep 5

                cd ~/redis-client; ./main $(( 2 ** $i )) 10 $w > ./generateGraph/workload_$w/run_$run/$i
                
                tmux send-keys -t server0 C-c
                tmux send-keys -t server1 C-c
                tmux send-keys -t server2 C-c
        done 
    done
done
