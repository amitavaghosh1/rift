#!/bin/bash

# Set the target architectures
architectures=("amd64" "arm64" "amd64")

# Set the target operating systems
operating_systems=("darwin" "darwin" "linux")

# Set the output directory
output_dir="bin"

# Set the name of your Go program
program_name="rift"

# Loop through the architectures and operating systems
for ((i=0; i<${#architectures[@]}; i++)); do
    architecture=${architectures[i]}
    os=${operating_systems[i]}
    
    # Set the output file name based on the architecture and operating system
    output_file="${output_dir}/${os}_${architecture}/${program_name}"
    echo "generating ${output_file}"
    
    # Build the executable
    env GOOS=${os} GOARCH=${architecture} go build -o ${output_file}
    
    # Set the executable permissions
    chmod +x ${output_file}
done

