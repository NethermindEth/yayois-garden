#!/bin/bash

function create_binding {
    contract_dir=$1
    contract=$2
    binding_dir=$3

    echo "generating bindings for" $contract

    mkdir -p $binding_dir/${contract}

    contract_json="$contract_dir/out/${contract}.sol/${contract}.json"
    solc_abi=$(cat ${contract_json} | jq -r '.abi')
    solc_bin=$(cat ${contract_json} | jq -r '.bytecode.object')

    abi_file=$(mktemp)
    bin_file=$(mktemp)

    echo ${solc_abi} >${abi_file}
    echo ${solc_bin} >${bin_file}

    rm -f $binding_dir/${contract}/binding.go

    abigen --bin=${bin_file} --abi=${abi_file} --pkg=contract${contract} --out=$binding_dir/${contract}/binding.go

    rm -f ${abi_file} ${bin_file}
}

project_dir=$(realpath ../../..)

create_binding $project_dir/contracts YayoiFactory $project_dir/pkg/bindings
create_binding $project_dir/contracts YayoiCollection $project_dir/pkg/bindings
