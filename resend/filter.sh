#!/bin/bash

if [[ $# -ne 3 ]]; then
    echo "Špatný počet argumentů"
    echo "Použití: filter.sh 'csv soubor' 'od || *' 'do || *'"
    exit 1
fi

if [[ $2 == "*" && $3 == "*" ]]; then
    q -d , "select * from $1"
elif [[ $2 == "*" ]]; then
    q -d , "select * from $1 where c1 < $3"
elif [[ $3 == "*" ]]; then
    q -d , "select * from $1 where c1 > $2"
else
    q -d , "select * from $1 where c1 > $2 AND c1 < $3"
fi
